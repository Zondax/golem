package httpclient_test

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/zondax/golem/pkg/httpclient"
)

// testSrv is used as a test handler to set custom response body and code and to
// collect statistics on the number of requests and request timings.
type testSrv struct {
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
	tb := []struct {
		name string
		srv  *testSrv

		// retry policy settings
		codes           []int
		ctxDeadline     time.Duration
		backoff         httpclient.Backoff
		maxRetries      int
		perRetryTimeout time.Duration
		initialBackoff  time.Duration

		// expectations
		wantRetry         bool
		wantTotalWait     time.Duration
		wantWaitBetween   time.Duration
		wantCalled        int
		wantCode          int
		wantBody          []byte
		wantErr           error
		wantClientTimeout bool
	}{
		// the request should not be retried if there are no retry codes specified and request did not timeout.
		{
			name: "no retries when codes are not specified",
			srv:  newTestSrv(t, http.StatusInternalServerError, nil, 0),

			backoff:     httpclient.BackoffLinear,
			ctxDeadline: 5 * time.Second,

			wantCalled: 1,
			wantBody:   []byte{},
			wantCode:   http.StatusInternalServerError,
		},
		// if the context deadline exceeds the request execution, the request should be cancelled and
		// retries should be stopped.
		{
			name: "context deadline exceeds retries",
			srv:  newTestSrv(t, http.StatusInternalServerError, nil, 0),

			ctxDeadline:     2 * time.Second,
			backoff:         httpclient.BackoffLinear,
			codes:           []int{http.StatusInternalServerError},
			maxRetries:      3,
			perRetryTimeout: 5 * time.Second,
			initialBackoff:  1 * time.Second,

			wantErr: context.DeadlineExceeded,
		},
		// the request should not be retried if a success response is obtained.
		// the response code and body should not be modified.
		{
			name: "succesful request no retries",
			srv:  newTestSrv(t, http.StatusOK, []byte("OK"), 0),

			backoff:         httpclient.BackoffLinear,
			ctxDeadline:     5 * time.Second,
			codes:           []int{http.StatusInternalServerError},
			maxRetries:      3,
			perRetryTimeout: 5 * time.Second,
			initialBackoff:  1 * time.Second,

			wantCode:   http.StatusOK,
			wantCalled: 1,
			wantBody:   []byte("OK"),
		},
		// the request should timeout after perRetryTimeout ( 400ms )
		// the request should be retried linearly every 100ms a max of 3 times.
		// the total time of the request should be 300ms
		// the time between retries should be 100ms
		{
			name: "linear retry when no codes specified but the request times out",
			srv:  newTestSrv(t, http.StatusOK, []byte("OK"), 450),

			ctxDeadline:     5 * time.Second,
			backoff:         httpclient.BackoffLinear,
			maxRetries:      3,
			perRetryTimeout: 400 * time.Millisecond,
			initialBackoff:  100 * time.Millisecond,

			wantRetry:         true,
			wantClientTimeout: true,
			wantErr:           errors.New("i/o timeout"),
		},

		// the request should be retried linearly every 500ms a max of 3 times.
		// the total time of the request should be 1000ms
		// the time between retries should be 500ms
		{
			name: "linear retry",
			srv:  newTestSrv(t, http.StatusInternalServerError, nil, 0),

			codes:           []int{http.StatusInternalServerError},
			ctxDeadline:     5 * time.Second,
			backoff:         httpclient.BackoffLinear,
			maxRetries:      3,
			perRetryTimeout: 5 * time.Second,
			initialBackoff:  500 * time.Millisecond,

			wantRetry:       true,
			wantCode:        http.StatusInternalServerError,
			wantTotalWait:   1000 * time.Millisecond,
			wantWaitBetween: 500 * time.Millisecond,
			wantCalled:      3,
			wantBody:        []byte{},
		},
		// the request should be retried exponentialy starting from 100ms i.e 100ms * (2 ^ attempt) for a max of 3 times
		// the total time of the request should be:
		// 		total = 0ms  // attempt 1
		//      total += 100ms * (2 ^ 2) = 400ms  // attempt 2
		// 		total += 100ms * (2 ^ 3) = 800ms // attempt 3
		//      total = 1200
		// the time between retries ( we will only check the last retry attempt) = 800ms
		{
			name: "exponential retry",
			srv:  newTestSrv(t, http.StatusInternalServerError, nil, 0),

			codes:           []int{http.StatusInternalServerError},
			ctxDeadline:     5 * time.Second,
			backoff:         httpclient.BackoffExponential,
			maxRetries:      3,
			perRetryTimeout: 5 * time.Second,
			initialBackoff:  100 * time.Millisecond,

			wantRetry:       true,
			wantWaitBetween: 800 * time.Millisecond,
			wantCalled:      3,
			wantCode:        http.StatusInternalServerError,
			wantTotalWait:   1200 * time.Millisecond,
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

			retryPolicy := httpclient.NewRetryPolicy(tt.perRetryTimeout, tt.maxRetries).
				WithCodes(tt.codes...).WithBackoff(tt.backoff, tt.initialBackoff)

			tmp := srv.Client()
			tmp.Timeout = tt.perRetryTimeout
			client := httpclient.NewHTTPClient(httpclient.Config{
				CustomClient: tmp,
				RetryPolicy:  retryPolicy,
			})

			req, err := http.NewRequest(http.MethodGet, srv.URL, nil)
			assert.NoError(t, err)

			// execute the request and measure the time it takes
			start := time.Now().UnixMilli()
			resp, err := client.Do(ctx, req)
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
			defer resp.Body.Close()

			// check that the response is not modified.
			gotBody, err := io.ReadAll(resp.Body)
			assert.NoError(t, err)
			assert.Equal(t, tt.wantBody, gotBody)
			assert.Equal(t, tt.wantCode, resp.StatusCode)

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
