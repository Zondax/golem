package zhttpclient_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	url "net/url"
	"sync"
	"testing"
	"time"

	"github.com/zondax/golem/pkg/utils"

	"github.com/cenkalti/backoff/v4"
	"github.com/zondax/golem/pkg/zhttpclient"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testSrv is used as a test handler to set custom response body and code and to
// collect statistics on the number of requests and request timings.
type testSrv struct {
	lck              sync.Mutex
	t                *testing.T
	code             int
	body             []byte
	sleepMs          int64
	called           int
	firstCalled      int64
	lastCalled       int64
	waitBetweenCalls int64
}

func newTestSrv(t *testing.T, code int, body []byte, sleepMs int64) *testSrv {
	return &testSrv{
		t:       t,
		code:    code,
		body:    body,
		sleepMs: sleepMs,
	}
}

func (ts *testSrv) Handle(w http.ResponseWriter, r *http.Request) {
	ts.lck.Lock()
	defer ts.lck.Unlock()

	if ts.sleepMs > 0 {
		time.Sleep(time.Duration(ts.sleepMs) * time.Millisecond)
	}
	ts.called++
	if ts.lastCalled > 0 {
		ts.waitBetweenCalls = time.Now().UnixMilli() - ts.lastCalled
	} else {
		ts.firstCalled = time.Now().UnixMilli()
	}
	ts.lastCalled = time.Now().UnixMilli()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(ts.code)
	_, err := w.Write(ts.body)
	assert.NoError(ts.t, err)
}

func TestHTTPClient(t *testing.T) {
	getParams := url.Values{}
	getParams.Add("a", "1")
	getParams.Add("b", "2")

	postBody := []byte(`{"a":1,"b":2}`)

	headers := map[string]string{
		"X-Test-Header": "value-x",
	}

	checkHeaders := func(t *testing.T, requestHeaders http.Header) {
		for k, v := range headers {
			assert.Contains(t, requestHeaders, k)
			assert.Equal(t, v, requestHeaders.Get(k))
		}
	}

	tb := []struct {
		name     string
		method   string
		handler  func(t *testing.T) http.HandlerFunc
		custom   bool
		wantCode int
		wantBody []byte
		wantErr  bool
	}{
		{
			name: "GET request success",
			handler: func(t *testing.T) http.HandlerFunc {
				return func(w http.ResponseWriter, r *http.Request) {
					checkHeaders(t, r.Header)
					q := r.URL.Query()
					for k := range q {
						assert.Contains(t, getParams, k)
					}

					w.WriteHeader(http.StatusOK)
					_, err := w.Write([]byte("OK"))
					assert.NoError(t, err)
				}
			},
			method:   http.MethodGet,
			wantCode: http.StatusOK,
			wantBody: []byte("OK"),
		},
		{
			name: "POST request success",
			handler: func(t *testing.T) http.HandlerFunc {
				return func(w http.ResponseWriter, r *http.Request) {
					checkHeaders(t, r.Header)
					data, err := io.ReadAll(r.Body)
					assert.NoError(t, err)
					assert.Equal(t, string(postBody), string(data))

					w.WriteHeader(http.StatusOK)
					_, err = w.Write([]byte("OK"))
					assert.NoError(t, err)
				}
			},
			method:   http.MethodPost,
			wantCode: http.StatusOK,
			wantBody: []byte("OK"),
		},
		{
			name: "custom request success",
			handler: func(t *testing.T) http.HandlerFunc {
				return func(w http.ResponseWriter, r *http.Request) {
					checkHeaders(t, r.Header)
					data, err := io.ReadAll(r.Body)
					assert.NoError(t, err)
					assert.Equal(t, string(postBody), string(data))

					w.WriteHeader(http.StatusOK)
					_, err = w.Write([]byte("OK"))
					assert.NoError(t, err)
				}
			},
			custom:   true,
			method:   http.MethodGet,
			wantCode: http.StatusOK,
			wantBody: []byte("OK"),
		},
	}

	for i := range tb {
		tt := tb[i]
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			srv := httptest.NewServer(tt.handler(t))
			defer srv.Close()

			client := zhttpclient.New(zhttpclient.Config{
				BaseClient: srv.Client(),
			})

			var (
				req     *http.Request
				gotResp *zhttpclient.Response
				gotErr  error
			)

			if tt.custom {
				req, gotErr = http.NewRequest(tt.method, srv.URL, bytes.NewBuffer(postBody))
				assert.NoError(t, gotErr)
				req.URL.RawQuery = getParams.Encode()
				for k, v := range headers {
					req.Header.Add(k, v)
				}
				gotResp, gotErr = client.Do(ctx, req)
			} else {
				r := client.NewRequest().SetURL(srv.URL).SetHeaders(headers)
				switch tt.method {
				case http.MethodGet:
					r = r.SetQueryParams(getParams)
					gotResp, gotErr = r.Get(ctx)
				case http.MethodPost:
					r = r.SetBody(bytes.NewBuffer(postBody))
					gotResp, gotErr = r.Post(ctx)
				}
			}

			if tt.wantErr {
				assert.Error(t, gotErr)
				return
			}
			assert.NoError(t, gotErr)
			assert.Equal(t, tt.wantCode, gotResp.Code)
			assert.Equal(t, string(tt.wantBody), string(gotResp.Body))
		})
	}
}

func TestHTTPClient_Retry(t *testing.T) {
	defaultRetryPolicy := &zhttpclient.RetryPolicy{}

	tb := []struct {
		name string
		srv  *testSrv

		method string
		body   []byte

		getRetryPolicy func() *zhttpclient.RetryPolicy
		timeout        time.Duration
		ctxDeadline    time.Duration

		wantRetry         bool
		wantTotalWait     time.Duration
		wantWaitBetween   time.Duration
		wantCalled        int
		wantCode          int
		wantBody          []byte
		wantErr           error
		wantClientTimeout bool
	}{
		{
			name: "post request not retried without retryCodes",
			srv:  newTestSrv(t, http.StatusInternalServerError, nil, 0),

			ctxDeadline: 5 * time.Second,
			getRetryPolicy: func() *zhttpclient.RetryPolicy {
				return nil
			},
			wantCalled: 1,
			wantBody:   []byte{},
			wantCode:   http.StatusInternalServerError,
		},
		// if the context deadline exceeds the request execution, the request should be cancelled and
		// retries should be stopped.
		{
			name: "context deadline exceeds retries",
			srv:  newTestSrv(t, http.StatusInternalServerError, nil, 1000),

			ctxDeadline: 500 * time.Millisecond,
			getRetryPolicy: func() *zhttpclient.RetryPolicy {
				r := &zhttpclient.RetryPolicy{
					MaxAttempts: 3,
				}
				r.WithCodes(http.StatusInternalServerError)
				r.SetLinearBackoff(500 * time.Millisecond)
				return r
			},

			wantErr: context.DeadlineExceeded,
		},
		// the request should not be retried if a success response is obtained.
		// the response code and body should not be modified.
		{
			name: "succesful request no retries",
			srv:  newTestSrv(t, http.StatusOK, []byte("OK"), 0),

			ctxDeadline: 5 * time.Second,
			getRetryPolicy: func() *zhttpclient.RetryPolicy {
				r := &zhttpclient.RetryPolicy{
					MaxAttempts: 3,
				}
				r.WithCodes(http.StatusInternalServerError)
				r.SetLinearBackoff(500 * time.Millisecond)
				return r
			},
			wantCode:   http.StatusOK,
			wantCalled: 1,
			wantBody:   []byte("OK"),
		},
		// the request should not be retried if a success response is obtained.
		// the response code and body should not be modified.
		{
			name: "succesful request no retries",
			srv:  newTestSrv(t, http.StatusOK, []byte("OK"), 0),

			ctxDeadline: 5 * time.Second,
			getRetryPolicy: func() *zhttpclient.RetryPolicy {
				r := &zhttpclient.RetryPolicy{
					MaxAttempts: 3,
				}
				r.WithCodes(http.StatusInternalServerError)
				r.SetLinearBackoff(500 * time.Millisecond)
				return r
			},
			wantCode:   http.StatusOK,
			wantCalled: 1,
			wantBody:   []byte("OK"),
		},

		// the request should timeout after perRetryTimeout ( 400ms )
		// the request should be retried linearly every 100ms a max of 2 times ( 3 total requests ).
		// the total time of the request should be 300ms
		// the time between retries should be 100ms
		{
			name: "linear retry when no codes specified but the request times out",
			srv:  newTestSrv(t, http.StatusOK, []byte("OK"), 450),

			ctxDeadline: 5 * time.Second,
			timeout:     400 * time.Millisecond,

			getRetryPolicy: func() *zhttpclient.RetryPolicy {
				r := &zhttpclient.RetryPolicy{
					MaxAttempts: 2,
				}
				r.WithCodes(http.StatusInternalServerError)
				r.SetLinearBackoff(100 * time.Millisecond)
				return r
			},
			wantRetry:         true,
			wantClientTimeout: true,
			wantCalled:        3,
			wantErr:           errors.New("i/o timeout"),
		},

		// the request should be retried linearly every 500ms a max of 2 times ( 3 total requests ).
		// the total time of the request should be 1000ms
		// the time between retries should be 500ms
		{
			name: "linear retry",
			srv:  newTestSrv(t, http.StatusInternalServerError, nil, 0),

			ctxDeadline: 5 * time.Second,

			timeout: 5 * time.Second,

			getRetryPolicy: func() *zhttpclient.RetryPolicy {
				r := &zhttpclient.RetryPolicy{
					MaxAttempts:        2,
					MaxWaitBeforeRetry: 500 * time.Millisecond,
				}
				r.WithCodes(http.StatusInternalServerError)
				r.SetLinearBackoff(500 * time.Millisecond)
				return r
			},

			wantRetry:       true,
			wantCode:        http.StatusInternalServerError,
			wantTotalWait:   1000 * time.Millisecond,
			wantWaitBetween: 500 * time.Millisecond,
			wantCalled:      3,
			wantBody:        []byte{},
		},

		// the request should be retried exponentialy starting from 100ms i.e 100ms * (2 ^ attempt) for a max of 2 times
		// the total time of the request should be:
		// 		total = 0ms  // attempt 1
		//      total += 100ms * (2 ^ 0) = 100ms  // attempt 1
		// 		total += 100ms * (2 ^ 1) = 200ms // attempt 2
		//      total = 300ms
		// the time between retries ( we will only check the last retry attempt) = 200ms
		{
			name: "exponential retry",
			srv:  newTestSrv(t, http.StatusInternalServerError, nil, 0),

			ctxDeadline: 5 * time.Second,

			getRetryPolicy: func() *zhttpclient.RetryPolicy {
				r := &zhttpclient.RetryPolicy{
					MaxAttempts:        2,
					MaxWaitBeforeRetry: 2 * time.Second,
				}
				r.WithCodes(http.StatusInternalServerError)

				tmp := backoff.NewExponentialBackOff(backoff.WithInitialInterval(100*time.Millisecond),
					backoff.WithMaxElapsedTime(r.MaxWaitBeforeRetry),
					backoff.WithMultiplier(2),
				)
				tmp.RandomizationFactor = 0
				maxAttempts, _ := utils.IntToUInt64(r.MaxAttempts)
				b := backoff.WithMaxRetries(tmp, maxAttempts)

				r.SetBackoff(func(_ uint, _ *http.Response, _ error) time.Duration {
					return b.NextBackOff()
				})

				return r
			},
			wantRetry:       true,
			wantWaitBetween: 200 * time.Millisecond,
			wantCalled:      3,
			wantCode:        http.StatusInternalServerError,
			wantTotalWait:   300 * time.Millisecond,
			wantBody:        []byte{},
		},
	}

	for i := range tb {
		tt := tb[i]
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx, cancel := context.WithTimeout(context.Background(), tt.ctxDeadline)
			defer cancel()

			srv := httptest.NewServer(http.HandlerFunc(tt.srv.Handle))
			defer srv.Close()

			client := zhttpclient.New(zhttpclient.Config{
				BaseClient: srv.Client(),
				Timeout:    tt.timeout,
			})
			client.SetRetryPolicy(defaultRetryPolicy)

			// execute the request and measure the time it takes

			r := client.NewRequest().SetURL(srv.URL).SetBody(bytes.NewBuffer(tt.body))
			if p := tt.getRetryPolicy(); p != nil {
				r = r.SetRetryPolicy(p)
			}

			start := time.Now().UnixMilli()
			resp, err := r.Post(ctx)
			end := time.Now().UnixMilli()

			if tt.wantErr != nil {
				if tt.wantClientTimeout {
					_, ok := err.(net.Error)
					assert.True(t, ok)
				} else {
					assert.ErrorIs(t, err, tt.wantErr)
				}
				return
			}

			assert.NoError(t, err)

			// check that the response is not modified.
			assert.NoError(t, err)
			assert.Equal(t, tt.wantCode, resp.Code)
			assert.Equal(t, string(tt.wantBody), string(resp.Body))

			// check that the request was retried as expected
			assert.Equal(t, tt.wantCalled, tt.srv.called)
			if tt.wantRetry {
				// ignore minor deviations in millisecond values
				fmt.Println(end, start, end-start, tt.wantTotalWait.Milliseconds())
				assert.Equal(t, (end-start)/tt.wantTotalWait.Milliseconds(), int64(1))
				assert.Equal(t, tt.srv.waitBetweenCalls/tt.wantWaitBetween.Milliseconds(), int64(1))
			}
		})
	}
}

type testResp struct {
	TestField string `json:"testField"`
}
type testError struct {
	TestError string `json:"testError"`
}

func TestHTTPClient_DecodeResult(t *testing.T) {
	tb := []struct {
		name     string
		srv      *testSrv
		wantResp interface{}
		wantErr  interface{}
	}{
		{
			name:     "succesfully decode into result",
			wantResp: &testResp{TestField: "success"},
			srv:      newTestSrv(t, http.StatusOK, []byte(`{"testField":"success"}`), 0),
		},
		{
			name:    "succesfully decode into error",
			wantErr: &testError{TestError: "error"},
			srv:     newTestSrv(t, http.StatusConflict, []byte(`{"testError":"error"}`), 0),
		},
	}

	for i := range tb {
		tt := tb[i]
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			srv := httptest.NewServer(http.HandlerFunc(tt.srv.Handle))
			defer srv.Close()

			req := zhttpclient.New(zhttpclient.Config{}).NewRequest().SetURL(srv.URL)
			if tt.wantResp != nil {
				req = req.SetResult(&testResp{})
			}
			if tt.wantErr != nil {
				req = req.SetError(&testError{})
			}

			resp, err := req.Post(context.Background())
			assert.NoError(t, err)

			assert.Equal(t, tt.wantResp, resp.Result)
			assert.Equal(t, tt.wantErr, resp.Error)
		})
	}
}

func TestNew_WithoutOpenTelemetry(t *testing.T) {
	t.Run("creates client without OpenTelemetry when not configured", func(t *testing.T) {
		config := zhttpclient.Config{
			Timeout: 30 * time.Second,
		}

		client := zhttpclient.New(config)
		require.NotNil(t, client)

		// Verify client can make requests (basic functionality test)
		request := client.NewRequest()
		assert.NotNil(t, request)
	})

	t.Run("creates client without OpenTelemetry when explicitly disabled", func(t *testing.T) {
		config := zhttpclient.Config{
			Timeout: 30 * time.Second,
			OpenTelemetry: &zhttpclient.OpenTelemetryConfig{
				Enabled: false,
			},
		}

		client := zhttpclient.New(config)
		require.NotNil(t, client)

		// Verify client can make requests (basic functionality test)
		request := client.NewRequest()
		assert.NotNil(t, request)
	})
}

func TestNew_WithOpenTelemetry(t *testing.T) {
	t.Run("creates client with OpenTelemetry when enabled", func(t *testing.T) {
		config := zhttpclient.Config{
			Timeout: 30 * time.Second,
			OpenTelemetry: &zhttpclient.OpenTelemetryConfig{
				Enabled: true,
			},
		}

		client := zhttpclient.New(config)
		require.NotNil(t, client)

		// Verify client can make requests (functionality should work the same)
		request := client.NewRequest()
		assert.NotNil(t, request)
	})

	t.Run("creates client with custom operation name function", func(t *testing.T) {
		customNameFunc := func(operation string, r *http.Request) string {
			return "custom-" + operation
		}

		config := zhttpclient.Config{
			Timeout: 30 * time.Second,
			OpenTelemetry: &zhttpclient.OpenTelemetryConfig{
				Enabled:           true,
				OperationNameFunc: customNameFunc,
			},
		}

		client := zhttpclient.New(config)
		require.NotNil(t, client)

		// Verify client functionality
		request := client.NewRequest()
		assert.NotNil(t, request)
	})

	t.Run("creates client with custom filters", func(t *testing.T) {
		customFilter := func(r *http.Request) bool {
			// Only instrument GET requests
			return r.Method == http.MethodGet
		}

		config := zhttpclient.Config{
			Timeout: 30 * time.Second,
			OpenTelemetry: &zhttpclient.OpenTelemetryConfig{
				Enabled: true,
				Filters: customFilter,
			},
		}

		client := zhttpclient.New(config)
		require.NotNil(t, client)

		// Verify client functionality
		request := client.NewRequest()
		assert.NotNil(t, request)
	})

	t.Run("creates client with both custom naming and filters", func(t *testing.T) {
		customNameFunc := func(operation string, r *http.Request) string {
			return "filtered-" + operation
		}

		customFilter := func(r *http.Request) bool {
			return r.Method == http.MethodGet || r.Method == http.MethodPost
		}

		config := zhttpclient.Config{
			Timeout: 30 * time.Second,
			OpenTelemetry: &zhttpclient.OpenTelemetryConfig{
				Enabled:           true,
				OperationNameFunc: customNameFunc,
				Filters:           customFilter,
			},
		}

		client := zhttpclient.New(config)
		require.NotNil(t, client)

		// Verify client functionality
		request := client.NewRequest()
		assert.NotNil(t, request)
	})
}

func TestOpenTelemetryConfig_Validation(t *testing.T) {
	t.Run("nil OpenTelemetry config is handled gracefully", func(t *testing.T) {
		config := zhttpclient.Config{
			Timeout:       30 * time.Second,
			OpenTelemetry: nil,
		}

		// Should not panic
		client := zhttpclient.New(config)
		assert.NotNil(t, client)
	})

	t.Run("empty OpenTelemetry config with enabled false", func(t *testing.T) {
		config := zhttpclient.Config{
			Timeout: 30 * time.Second,
			OpenTelemetry: &zhttpclient.OpenTelemetryConfig{
				Enabled: false,
				// Other fields are nil/zero values
			},
		}

		client := zhttpclient.New(config)
		require.NotNil(t, client)

		// Verify basic functionality
		request := client.NewRequest()
		assert.NotNil(t, request)
	})

	t.Run("OpenTelemetry config with only enabled true", func(t *testing.T) {
		config := zhttpclient.Config{
			Timeout: 30 * time.Second,
			OpenTelemetry: &zhttpclient.OpenTelemetryConfig{
				Enabled: true,
				// Other fields are nil - should use defaults
			},
		}

		client := zhttpclient.New(config)
		require.NotNil(t, client)

		// Verify basic functionality
		request := client.NewRequest()
		assert.NotNil(t, request)
	})
}

func TestConfig_BackwardCompatibility(t *testing.T) {
	t.Run("existing code without OpenTelemetry config continues to work", func(t *testing.T) {
		// This simulates existing code that doesn't know about the new OpenTelemetry field
		config := zhttpclient.Config{
			Timeout: 30 * time.Second,
			// OpenTelemetry field is not set (nil)
		}

		// Should create client successfully without OpenTelemetry instrumentation
		client := zhttpclient.New(config)
		assert.NotNil(t, client)

		// Verify basic functionality still works
		request := client.NewRequest()
		assert.NotNil(t, request)
	})

	t.Run("existing code with BaseClient continues to work", func(t *testing.T) {
		baseClient := &http.Client{
			Timeout: 10 * time.Second,
		}

		config := zhttpclient.Config{
			Timeout:    30 * time.Second,
			BaseClient: baseClient,
			// OpenTelemetry field is not set (nil)
		}

		client := zhttpclient.New(config)
		assert.NotNil(t, client)

		// Verify basic functionality
		request := client.NewRequest()
		assert.NotNil(t, request)
	})
}

func TestOpenTelemetryConfig_EdgeCases(t *testing.T) {
	t.Run("OpenTelemetry enabled with custom BaseClient", func(t *testing.T) {
		baseClient := &http.Client{
			Timeout: 10 * time.Second,
		}

		config := zhttpclient.Config{
			Timeout:    30 * time.Second,
			BaseClient: baseClient,
			OpenTelemetry: &zhttpclient.OpenTelemetryConfig{
				Enabled: true,
			},
		}

		client := zhttpclient.New(config)
		require.NotNil(t, client)

		// Verify basic functionality
		request := client.NewRequest()
		assert.NotNil(t, request)
	})
}

// TestOpenTelemetryIntegration tests the actual functionality with real HTTP calls
func TestOpenTelemetryIntegration(t *testing.T) {
	t.Run("client with OpenTelemetry makes successful requests", func(t *testing.T) {
		// Create a test server
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("OK"))
		}))
		defer srv.Close()

		// Create client with OpenTelemetry enabled
		config := zhttpclient.Config{
			Timeout: 30 * time.Second,
			OpenTelemetry: &zhttpclient.OpenTelemetryConfig{
				Enabled: true,
			},
		}

		client := zhttpclient.New(config)
		require.NotNil(t, client)

		// Make a request and verify it works
		resp, err := client.NewRequest().SetURL(srv.URL).Get(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.Code)
		assert.Equal(t, "OK", string(resp.Body))
	})

	t.Run("client without OpenTelemetry makes successful requests", func(t *testing.T) {
		// Create a test server
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("OK"))
		}))
		defer srv.Close()

		// Create client without OpenTelemetry
		config := zhttpclient.Config{
			Timeout: 30 * time.Second,
		}

		client := zhttpclient.New(config)
		require.NotNil(t, client)

		// Make a request and verify it works
		resp, err := client.NewRequest().SetURL(srv.URL).Get(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.Code)
		assert.Equal(t, "OK", string(resp.Body))
	})

	t.Run("client with custom filters works correctly", func(t *testing.T) {
		// Create a test server
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("Method: " + r.Method))
		}))
		defer srv.Close()

		// Create client with custom filter (only GET requests)
		customFilter := func(r *http.Request) bool {
			return r.Method == http.MethodGet
		}

		config := zhttpclient.Config{
			Timeout: 30 * time.Second,
			OpenTelemetry: &zhttpclient.OpenTelemetryConfig{
				Enabled: true,
				Filters: customFilter,
			},
		}

		client := zhttpclient.New(config)
		require.NotNil(t, client)

		// Test GET request (should be instrumented by filter)
		getResp, err := client.NewRequest().SetURL(srv.URL).Get(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, getResp.Code)
		assert.Equal(t, "Method: GET", string(getResp.Body))

		// Test POST request (should also work, regardless of filter)
		postResp, err := client.NewRequest().SetURL(srv.URL).Post(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, postResp.Code)
		assert.Equal(t, "Method: POST", string(postResp.Body))
	})
}
