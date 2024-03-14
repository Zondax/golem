package zhttpclient_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	url "net/url"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	httpclient "github.com/zondax/golem/pkg/zhttpclient"
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

			client := httpclient.New(httpclient.Config{
				BaseClient: srv.Client(),
			})

			var (
				req       *http.Request
				gotStatus int
				gotBody   []byte
				gotErr    error
			)

			if tt.custom {
				req, gotErr = http.NewRequest(tt.method, srv.URL, bytes.NewBuffer(postBody))
				assert.NoError(t, gotErr)
				req.URL.RawQuery = getParams.Encode()
				for k, v := range headers {
					req.Header.Add(k, v)
				}
				gotStatus, gotBody, gotErr = client.Do(ctx, req)
			} else {
				r := client.NewRequest().SetURL(srv.URL).SetHeaders(headers)
				switch tt.method {
				case http.MethodGet:
					r = r.SetQueryParams(getParams)
					gotStatus, gotBody, gotErr = r.Get(ctx)
				case http.MethodPost:
					r = r.SetBody(bytes.NewBuffer(postBody))
					gotStatus, gotBody, gotErr = r.Post(ctx)
				}
			}

			if tt.wantErr {
				assert.Error(t, gotErr)
				return
			}
			assert.NoError(t, gotErr)
			assert.Equal(t, tt.wantCode, gotStatus)
			assert.Equal(t, string(tt.wantBody), string(gotBody))
		})
	}
}

func TestHTTPClient_Retry(t *testing.T) {
	defaultRetryPolicy := &httpclient.RetryPolicy{}

	tb := []struct {
		name string
		srv  *testSrv

		method string
		body   []byte

		getRetryPolicy func() *httpclient.RetryPolicy
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
			getRetryPolicy: func() *httpclient.RetryPolicy {
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
			getRetryPolicy: func() *httpclient.RetryPolicy {
				r := &httpclient.RetryPolicy{
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
			getRetryPolicy: func() *httpclient.RetryPolicy {
				r := &httpclient.RetryPolicy{
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
			getRetryPolicy: func() *httpclient.RetryPolicy {
				r := &httpclient.RetryPolicy{
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

			getRetryPolicy: func() *httpclient.RetryPolicy {
				r := &httpclient.RetryPolicy{
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

			getRetryPolicy: func() *httpclient.RetryPolicy {
				r := &httpclient.RetryPolicy{
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
		//      total += 100ms * (2 ^ 1) = 200ms  // attempt 1
		// 		total += 100ms * (2 ^ 2) = 400ms // attempt 2
		//      total = 600ms
		// the time between retries ( we will only check the last retry attempt) = 400ms
		{
			name: "exponential retry",
			srv:  newTestSrv(t, http.StatusInternalServerError, nil, 0),

			ctxDeadline: 5 * time.Second,

			getRetryPolicy: func() *httpclient.RetryPolicy {
				r := &httpclient.RetryPolicy{
					MaxAttempts:        2,
					MaxWaitBeforeRetry: 2 * time.Second,
				}
				r.WithCodes(http.StatusInternalServerError)
				r.SetExponentialBackoff(100 * time.Millisecond)
				return r
			},
			wantRetry:       true,
			wantWaitBetween: 400 * time.Millisecond,
			wantCalled:      3,
			wantCode:        http.StatusInternalServerError,
			wantTotalWait:   600 * time.Millisecond,
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

			client := httpclient.New(httpclient.Config{
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
			code, resp, err := r.Post(ctx)
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
			assert.Equal(t, tt.wantCode, code)
			assert.Equal(t, string(tt.wantBody), string(resp))

			// check that the request was retried as expected
			assert.Equal(t, tt.wantCalled, tt.srv.called)
			if tt.wantRetry {
				// ignore minor deviations in millisecond values
				assert.Equal(t, (end-start)/tt.wantTotalWait.Milliseconds(), int64(1))
				assert.Equal(t, tt.srv.waitBetweenCalls/tt.wantWaitBetween.Milliseconds(), int64(1))
			}
		})
	}
}
