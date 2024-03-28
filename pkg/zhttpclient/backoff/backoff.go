package backoff

import (
	"time"

	"github.com/cenkalti/backoff/v4"
)

const (
	exponentialMultiplier = 2
)

type BackOff backoff.BackOff

// Do retries op if it returns an error according to the provided backoff
func Do(op func() error, b backoff.BackOff) error {
	return backoff.Retry(op, b)
}

// LinearBackoff returns a configured constant backoff
func LinearBackoff(maxAttempts int, duration time.Duration) backoff.BackOff {
	return backoff.WithMaxRetries(backoff.NewConstantBackOff(duration), uint64(maxAttempts))
}

// ExponentialBackoff returns a configured exponential backoff
func ExponentialBackoff(maxAttempts int, initial, max time.Duration) backoff.BackOff {
	tmp := backoff.NewExponentialBackOff(backoff.WithInitialInterval(initial), backoff.WithMaxElapsedTime(max), backoff.WithMultiplier(exponentialMultiplier))
	return backoff.WithMaxRetries(tmp, uint64(maxAttempts))
}
