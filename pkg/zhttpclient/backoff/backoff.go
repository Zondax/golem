package backoff

import (
	"time"

	"github.com/cenkalti/backoff/v4"
)

const (
	exponentialMultiplier = 2
)

type BackOff struct {
	maxAttempts     int
	maxDuration     time.Duration
	initialDuration time.Duration
}

func New() *BackOff {
	return &BackOff{}
}

func (b *BackOff) WithMaxAttempts(maxAttempts int) *BackOff {
	b.maxAttempts = maxAttempts
	return b
}
func (b *BackOff) WithMaxDuration(max time.Duration) *BackOff {
	b.maxDuration = max
	return b
}
func (b *BackOff) WithInitialDuration(initial time.Duration) *BackOff {
	b.initialDuration = initial
	return b
}

func (b *BackOff) Exponential() backoff.BackOff {
	tmp := backoff.NewExponentialBackOff(backoff.WithInitialInterval(b.initialDuration), backoff.WithMaxElapsedTime(b.maxDuration), backoff.WithMultiplier(exponentialMultiplier))
	return backoff.WithMaxRetries(tmp, uint64(b.maxAttempts))
}

func (b *BackOff) Linear() backoff.BackOff {
	return backoff.WithMaxRetries(backoff.NewConstantBackOff(b.initialDuration), uint64(b.maxAttempts))
}

// Do retries op if it returns an error according to the provided backoff
func Do(op func() error, b backoff.BackOff) error {
	return backoff.Retry(op, b)
}
