package httpclient

import (
	"math"
	"time"
)

// Backoff specifies the type of backoff to be applied.
type Backoff int

// BackoffFn is a function that returns a backoff duration.
type BackoffFn func(attempt uint) time.Duration

const (
	// BackoffLinear waits the same duration between retries.
	BackoffLinear Backoff = iota
	// BackoffExponential waits for initialDuration * (2 ^ attempt)
	BackoffExponential
)

type RetryPolicy struct {
	maxAttempts          int
	backoffFn            BackoffFn
	initialBackoff       time.Duration
	perRetryTimeout      time.Duration
	retryableStatusCodes []int
}

// NewRetryPolicy returns a new RetryPolicy.
// The perRetryTimeout is the http client timeout.
func NewRetryPolicy(perRetryTimeout time.Duration, maxAttempts int) *RetryPolicy {
	return &RetryPolicy{
		perRetryTimeout: perRetryTimeout,
		maxAttempts:     maxAttempts,
	}
}

// WithCodes specifies the response status codes which trigger a retry.
func (r *RetryPolicy) WithCodes(codes ...int) *RetryPolicy {
	r.retryableStatusCodes = append(r.retryableStatusCodes, codes...)
	return r
}

// WithBackoff specifies the type of backoff to be applied and the initial duration.
func (r *RetryPolicy) WithBackoff(b Backoff, initial time.Duration) *RetryPolicy {
	r.initialBackoff = initial

	switch b {
	case BackoffLinear:
		r.backoffFn = func(uint) time.Duration {
			return initial
		}
	case BackoffExponential:
		r.backoffFn = func(attempt uint) time.Duration {
			mul := int64(math.Pow(2.0, float64(attempt)))
			return time.Millisecond * time.Duration(initial.Milliseconds()*mul)
		}
	}
	return r
}
