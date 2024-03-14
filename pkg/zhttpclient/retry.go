package zhttpclient

import (
	"math"
	"net/http"
	"time"
)

// BackoffFn is a function that returns a backoff duration.
type BackoffFn func(attempt uint, r *http.Response, lastErr error) time.Duration

type RetryPolicy struct {
	// MaxAttempts is the maximum number of retries
	MaxAttempts int
	// WaitBeforeRetry is the minimum default wait before retry
	WaitBeforeRetry time.Duration
	// MaxWaitBeforeRetry is the maximum cap for the wait before retry
	MaxWaitBeforeRetry time.Duration
	// backoffFn is a function that returns a custom sleep duration before a retry.
	// It is capped between WaitBeforeRetry and MaxWaitBeforeRetry
	backoffFn        BackoffFn
	retryStatusCodes map[int]struct{}
}

// WithCodes specifies the response status codes which trigger a retry.
func (r *RetryPolicy) WithCodes(codes ...int) *RetryPolicy {
	r.retryStatusCodes = make(map[int]struct{}, len(codes))
	for _, code := range codes {
		r.retryStatusCodes[code] = struct{}{}
	}
	return r
}

// SetBackoff sets a custom backoff function to be used to calculate the sleep duration between retries.
func (r *RetryPolicy) SetBackoff(fn BackoffFn) {
	r.backoffFn = fn
}

// SetLinearBackoff sets a constant sleep duration between retries.
func (r *RetryPolicy) SetLinearBackoff(duration time.Duration) {
	r.backoffFn = func(uint, *http.Response, error) time.Duration {
		return duration
	}
}

// SetExponentialBackoff sets an exponential base 2 delay ( duration * 2 ^ attempt ) for each attempt.
func (r *RetryPolicy) SetExponentialBackoff(duration time.Duration) {
	r.backoffFn = func(attempt uint, _ *http.Response, _ error) time.Duration {
		mul := int64(math.Pow(2.0, float64(attempt)))
		return time.Millisecond * time.Duration(duration.Milliseconds()*mul)
	}
}
