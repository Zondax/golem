package httpclient

import (
	"context"
	"net"
	"net/http"
	"time"
)

// Config is used to modify the HTTPClient behaviour when executing requests.
type Config struct {
	RetryPolicy  *RetryPolicy
	CustomClient *http.Client
}

// HTTPClient abstracts over the std http.Client and provides a retry mechanism.
type HTTPClient struct {
	client *http.Client
	config Config
}

// NewHTTPClient returns a new HTTPClient with the provided config.
// If a CustomClient is provided in the config, the client Timeout will override the RetryPolicy.perRetryTimeout.
func NewHTTPClient(config Config) *HTTPClient {
	c := &http.Client{}
	if config.RetryPolicy != nil {
		c.Timeout = config.RetryPolicy.perRetryTimeout
	}
	if config.CustomClient != nil {
		c = config.CustomClient
	}

	return &HTTPClient{
		client: c,
		config: config,
	}
}

// Do executes the request and applies the RetryPolicy specified when creating the client if any.
func (c *HTTPClient) Do(ctx context.Context, req *http.Request) (resp *http.Response, err error) {
	var attempt uint
	var wait time.Duration
	var deadlineExceeded bool

	done := make(chan bool)

	for !deadlineExceeded && shouldRetry(attempt, c.config.RetryPolicy, resp, err) {
		attempt++

		// do not wait before executing the request the first time.
		if attempt > 1 {
			wait = c.config.RetryPolicy.backoffFn(attempt)
		}

		t := time.AfterFunc(wait, func() {
			resp, err = c.client.Do(req.WithContext(ctx))
			done <- true
		})

		for {
			select {
			case <-done:
			case <-ctx.Done():
				t.Stop()
				err = ctx.Err()
				deadlineExceeded = true
			}
			break
		}
	}

	return
}

func shouldRetry(attempt uint, r *RetryPolicy, resp *http.Response, err error) bool {
	// proceed request if no attempt has been made
	if attempt == 0 {
		return true
	}

	// if no retry policy, do not retry.
	if r == nil {
		return false
	}

	// do not retry if we have reached max attempts.
	if attempt >= uint(r.maxAttempts) {
		return false
	}

	if err != nil {
		// only retry if client error is a timeout.
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			return true
		} else {
			return false
		}
	}

	for _, code := range r.retryableStatusCodes {
		if code == resp.StatusCode {
			return true
		}
	}
	return false
}
