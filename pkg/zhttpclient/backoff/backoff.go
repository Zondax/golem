package backoff

import (
	"github.com/zondax/golem/pkg/utils"
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
	maxAttempts, _ := utils.IntToUInt64(b.maxAttempts)
	return backoff.WithMaxRetries(tmp, maxAttempts)
}

func (b *BackOff) Linear() backoff.BackOff {
	maxAttempts, _ := utils.IntToUInt64(b.maxAttempts)
	return backoff.WithMaxRetries(backoff.NewConstantBackOff(b.initialDuration), maxAttempts)
}
