package backoff

import (
	"errors"
	"testing"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	t.Run("ShouldCreateNewBackOffWithZeroValues", func(t *testing.T) {
		backOff := New()

		assert.NotNil(t, backOff)
		assert.Equal(t, 0, backOff.maxAttempts)
		assert.Equal(t, time.Duration(0), backOff.maxDuration)
		assert.Equal(t, time.Duration(0), backOff.initialDuration)
	})
}

func TestBackOff_WithMaxAttempts(t *testing.T) {
	t.Run("ShouldSetMaxAttempts", func(t *testing.T) {
		backOff := New().WithMaxAttempts(5)

		assert.Equal(t, 5, backOff.maxAttempts)
	})

	t.Run("ShouldReturnSameInstance", func(t *testing.T) {
		original := New()
		result := original.WithMaxAttempts(3)

		assert.Same(t, original, result)
	})

	t.Run("ShouldAllowZeroAttempts", func(t *testing.T) {
		backOff := New().WithMaxAttempts(0)

		assert.Equal(t, 0, backOff.maxAttempts)
	})

	t.Run("ShouldAllowNegativeAttempts", func(t *testing.T) {
		backOff := New().WithMaxAttempts(-1)

		assert.Equal(t, -1, backOff.maxAttempts)
	})

	t.Run("ShouldAllowLargeAttempts", func(t *testing.T) {
		backOff := New().WithMaxAttempts(1000000)

		assert.Equal(t, 1000000, backOff.maxAttempts)
	})
}

func TestBackOff_WithMaxDuration(t *testing.T) {
	t.Run("ShouldSetMaxDuration", func(t *testing.T) {
		duration := 30 * time.Second
		backOff := New().WithMaxDuration(duration)

		assert.Equal(t, duration, backOff.maxDuration)
	})

	t.Run("ShouldReturnSameInstance", func(t *testing.T) {
		original := New()
		result := original.WithMaxDuration(time.Minute)

		assert.Same(t, original, result)
	})

	t.Run("ShouldAllowZeroDuration", func(t *testing.T) {
		backOff := New().WithMaxDuration(0)

		assert.Equal(t, time.Duration(0), backOff.maxDuration)
	})

	t.Run("ShouldAllowNegativeDuration", func(t *testing.T) {
		duration := -5 * time.Second
		backOff := New().WithMaxDuration(duration)

		assert.Equal(t, duration, backOff.maxDuration)
	})

	t.Run("ShouldAllowLargeDuration", func(t *testing.T) {
		duration := 24 * time.Hour
		backOff := New().WithMaxDuration(duration)

		assert.Equal(t, duration, backOff.maxDuration)
	})
}

func TestBackOff_WithInitialDuration(t *testing.T) {
	t.Run("ShouldSetInitialDuration", func(t *testing.T) {
		duration := 100 * time.Millisecond
		backOff := New().WithInitialDuration(duration)

		assert.Equal(t, duration, backOff.initialDuration)
	})

	t.Run("ShouldReturnSameInstance", func(t *testing.T) {
		original := New()
		result := original.WithInitialDuration(time.Second)

		assert.Same(t, original, result)
	})

	t.Run("ShouldAllowZeroDuration", func(t *testing.T) {
		backOff := New().WithInitialDuration(0)

		assert.Equal(t, time.Duration(0), backOff.initialDuration)
	})

	t.Run("ShouldAllowNegativeDuration", func(t *testing.T) {
		duration := -100 * time.Millisecond
		backOff := New().WithInitialDuration(duration)

		assert.Equal(t, duration, backOff.initialDuration)
	})

	t.Run("ShouldAllowMicrosecondPrecision", func(t *testing.T) {
		duration := 500 * time.Microsecond
		backOff := New().WithInitialDuration(duration)

		assert.Equal(t, duration, backOff.initialDuration)
	})
}

func TestBackOff_FluentInterface(t *testing.T) {
	t.Run("ShouldAllowMethodChaining", func(t *testing.T) {
		backOff := New().
			WithMaxAttempts(5).
			WithMaxDuration(30 * time.Second).
			WithInitialDuration(100 * time.Millisecond)

		assert.Equal(t, 5, backOff.maxAttempts)
		assert.Equal(t, 30*time.Second, backOff.maxDuration)
		assert.Equal(t, 100*time.Millisecond, backOff.initialDuration)
	})

	t.Run("ShouldAllowPartialConfiguration", func(t *testing.T) {
		backOff := New().WithMaxAttempts(3)

		assert.Equal(t, 3, backOff.maxAttempts)
		assert.Equal(t, time.Duration(0), backOff.maxDuration)
		assert.Equal(t, time.Duration(0), backOff.initialDuration)
	})
}

func TestBackOff_Exponential(t *testing.T) {
	t.Run("ShouldReturnExponentialBackOff", func(t *testing.T) {
		backOff := New().
			WithMaxAttempts(3).
			WithMaxDuration(10 * time.Second).
			WithInitialDuration(100 * time.Millisecond).
			Exponential()

		assert.NotNil(t, backOff)

		// Test that it behaves like an exponential backoff
		interval1 := backOff.NextBackOff()
		interval2 := backOff.NextBackOff()

		// Should return valid intervals (not backoff.Stop initially)
		assert.NotEqual(t, backoff.Stop, interval1)
		assert.NotEqual(t, backoff.Stop, interval2)

		// Second interval should be larger (exponential growth)
		assert.Greater(t, interval2, interval1)
	})

	t.Run("ShouldRespectMaxAttempts", func(t *testing.T) {
		backOff := New().
			WithMaxAttempts(2).
			WithInitialDuration(10 * time.Millisecond).
			Exponential()

		// First attempt
		interval1 := backOff.NextBackOff()
		assert.NotEqual(t, backoff.Stop, interval1)

		// Second attempt
		interval2 := backOff.NextBackOff()
		assert.NotEqual(t, backoff.Stop, interval2)

		// Third attempt should stop
		interval3 := backOff.NextBackOff()
		assert.Equal(t, backoff.Stop, interval3)
	})

	t.Run("ShouldHandleZeroAttempts", func(t *testing.T) {
		backOff := New().
			WithMaxAttempts(0).
			WithInitialDuration(10 * time.Millisecond).
			Exponential()

		// Should immediately stop with 0 attempts
		interval := backOff.NextBackOff()
		assert.Equal(t, backoff.Stop, interval)
	})

	t.Run("ShouldHandleZeroInitialDuration", func(t *testing.T) {
		backOff := New().
			WithMaxAttempts(3).
			WithInitialDuration(0).
			Exponential()

		assert.NotNil(t, backOff)

		// Should still work with zero initial duration
		interval := backOff.NextBackOff()
		assert.NotEqual(t, backoff.Stop, interval)
	})

	t.Run("ShouldRespectMaxDuration", func(t *testing.T) {
		backOff := New().
			WithMaxAttempts(100).                   // High attempts
			WithMaxDuration(50 * time.Millisecond). // Low max duration
			WithInitialDuration(10 * time.Millisecond).
			Exponential()

		start := time.Now()

		// Keep getting intervals until stop
		for {
			interval := backOff.NextBackOff()
			if interval == backoff.Stop {
				break
			}
			time.Sleep(1 * time.Millisecond) // Small sleep to advance time
		}

		elapsed := time.Since(start)

		// Should respect max duration (allowing some tolerance for test execution)
		assert.Less(t, elapsed, 200*time.Millisecond)
	})
}

func TestBackOff_Linear(t *testing.T) {
	t.Run("ShouldReturnLinearBackOff", func(t *testing.T) {
		backOff := New().
			WithMaxAttempts(3).
			WithInitialDuration(100 * time.Millisecond).
			Linear()

		assert.NotNil(t, backOff)

		// Test that it behaves like a constant/linear backoff
		interval1 := backOff.NextBackOff()
		interval2 := backOff.NextBackOff()

		// Should return valid intervals
		assert.NotEqual(t, backoff.Stop, interval1)
		assert.NotEqual(t, backoff.Stop, interval2)

		// Intervals should be equal (constant backoff)
		assert.Equal(t, interval1, interval2)
		assert.Equal(t, 100*time.Millisecond, interval1)
	})

	t.Run("ShouldRespectMaxAttempts", func(t *testing.T) {
		backOff := New().
			WithMaxAttempts(2).
			WithInitialDuration(10 * time.Millisecond).
			Linear()

		// First attempt
		interval1 := backOff.NextBackOff()
		assert.NotEqual(t, backoff.Stop, interval1)

		// Second attempt
		interval2 := backOff.NextBackOff()
		assert.NotEqual(t, backoff.Stop, interval2)

		// Third attempt should stop
		interval3 := backOff.NextBackOff()
		assert.Equal(t, backoff.Stop, interval3)
	})

	t.Run("ShouldHandleZeroAttempts", func(t *testing.T) {
		backOff := New().
			WithMaxAttempts(0).
			WithInitialDuration(10 * time.Millisecond).
			Linear()

		// Should immediately stop with 0 attempts
		interval := backOff.NextBackOff()
		assert.Equal(t, backoff.Stop, interval)
	})

	t.Run("ShouldHandleZeroInitialDuration", func(t *testing.T) {
		backOff := New().
			WithMaxAttempts(3).
			WithInitialDuration(0).
			Linear()

		assert.NotNil(t, backOff)

		interval := backOff.NextBackOff()
		assert.Equal(t, time.Duration(0), interval)
	})

	t.Run("ShouldMaintainConstantInterval", func(t *testing.T) {
		duration := 50 * time.Millisecond
		backOff := New().
			WithMaxAttempts(5).
			WithInitialDuration(duration).
			Linear()

		// All intervals should be the same
		for i := 0; i < 5; i++ {
			interval := backOff.NextBackOff()
			assert.Equal(t, duration, interval)
		}

		// Should stop after max attempts
		interval := backOff.NextBackOff()
		assert.Equal(t, backoff.Stop, interval)
	})
}

func TestDo(t *testing.T) {
	t.Run("ShouldReturnNilWhenOperationSucceeds", func(t *testing.T) {
		callCount := 0
		operation := func() error {
			callCount++
			return nil
		}

		backOff := New().WithMaxAttempts(3).WithInitialDuration(1 * time.Millisecond).Exponential()

		err := Do(operation, backOff)
		assert.NoError(t, err)
		assert.Equal(t, 1, callCount)
	})

	t.Run("ShouldRetryOnFailureAndEventuallySucceed", func(t *testing.T) {
		callCount := 0
		operation := func() error {
			callCount++
			if callCount < 3 {
				return errors.New("temporary error")
			}
			return nil
		}

		backOff := New().WithMaxAttempts(5).WithInitialDuration(1 * time.Millisecond).Exponential()

		err := Do(operation, backOff)
		assert.NoError(t, err)
		assert.Equal(t, 3, callCount)
	})

	t.Run("ShouldReturnErrorWhenMaxAttemptsReached", func(t *testing.T) {
		callCount := 0
		expectedError := errors.New("persistent error")
		operation := func() error {
			callCount++
			return expectedError
		}

		backOff := New().WithMaxAttempts(3).WithInitialDuration(1 * time.Millisecond).Exponential()

		err := Do(operation, backOff)
		assert.Error(t, err)
		assert.Equal(t, expectedError, err)
		assert.GreaterOrEqual(t, callCount, 3)
		assert.LessOrEqual(t, callCount, 4)
	})

	t.Run("ShouldWorkWithLinearBackOff", func(t *testing.T) {
		callCount := 0
		operation := func() error {
			callCount++
			if callCount < 2 {
				return errors.New("temporary error")
			}
			return nil
		}

		backOff := New().WithMaxAttempts(3).WithInitialDuration(1 * time.Millisecond).Linear()

		err := Do(operation, backOff)
		assert.NoError(t, err)
		assert.Equal(t, 2, callCount)
	})

	t.Run("ShouldRespectBackOffTiming", func(t *testing.T) {
		callCount := 0
		var callTimes []time.Time

		operation := func() error {
			callCount++
			callTimes = append(callTimes, time.Now())
			if callCount < 3 {
				return errors.New("temporary error")
			}
			return nil
		}

		backOff := New().WithMaxAttempts(5).WithInitialDuration(10 * time.Millisecond).Linear()

		start := time.Now()
		err := Do(operation, backOff)
		elapsed := time.Since(start)
		assert.NoError(t, err)

		assert.Equal(t, 3, callCount)
		assert.Len(t, callTimes, 3)

		// Verify timing between calls (allowing for some tolerance)
		if len(callTimes) >= 2 {
			timeBetweenCalls := callTimes[1].Sub(callTimes[0])
			assert.GreaterOrEqual(t, timeBetweenCalls, 8*time.Millisecond) // Allow some tolerance
			assert.LessOrEqual(t, timeBetweenCalls, 50*time.Millisecond)   // Upper bound for test stability
		}

		assert.GreaterOrEqual(t, elapsed, 15*time.Millisecond) // At least 2 * 10ms intervals
	})

	t.Run("ShouldHandleNilOperation", func(t *testing.T) {
		backOff := New().WithMaxAttempts(1).WithInitialDuration(1 * time.Millisecond).Exponential()

		// This should panic, but we test that the function can handle it
		assert.Panics(t, func() {
			_ = Do(nil, backOff)
		})
	})

	t.Run("ShouldHandleZeroAttempts", func(t *testing.T) {
		callCount := 0
		operation := func() error {
			callCount++
			return errors.New("should not be called")
		}

		backOff := New().WithMaxAttempts(0).WithInitialDuration(1 * time.Millisecond).Exponential()

		err := Do(operation, backOff)
		assert.Error(t, err)
		// With 0 max attempts, the operation is still called once (initial attempt)
		assert.Equal(t, 1, callCount)
	})
}

// Integration tests
func TestBackOff_Integration(t *testing.T) {
	t.Run("ShouldWorkEndToEndWithExponentialBackOff", func(t *testing.T) {
		callCount := 0
		operation := func() error {
			callCount++
			if callCount < 4 {
				return errors.New("temporary failure")
			}
			return nil
		}

		backOff := New().
			WithMaxAttempts(5).
			WithMaxDuration(1 * time.Second).
			WithInitialDuration(5 * time.Millisecond).
			Exponential()

		start := time.Now()
		err := Do(operation, backOff)
		assert.NoError(t, err)
		elapsed := time.Since(start)

		assert.Equal(t, 4, callCount)
		assert.Greater(t, elapsed, 5*time.Millisecond) // Should take some time due to backoff
		assert.Less(t, elapsed, 500*time.Millisecond)  // But not too long for test
	})

	t.Run("ShouldWorkEndToEndWithLinearBackOff", func(t *testing.T) {
		callCount := 0
		operation := func() error {
			callCount++
			if callCount < 3 {
				return errors.New("temporary failure")
			}
			return nil
		}

		backOff := New().
			WithMaxAttempts(5).
			WithInitialDuration(5 * time.Millisecond).
			Linear()

		start := time.Now()
		err := Do(operation, backOff)
		assert.NoError(t, err)
		elapsed := time.Since(start)

		assert.Equal(t, 3, callCount)
		assert.Greater(t, elapsed, 10*time.Millisecond) // Should take at least 2 * 5ms
		assert.Less(t, elapsed, 100*time.Millisecond)   // But not too long for test
	})
}

// Benchmark tests
func BenchmarkNew(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		New()
	}
}

func BenchmarkBackOff_FluentInterface(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		New().
			WithMaxAttempts(5).
			WithMaxDuration(30 * time.Second).
			WithInitialDuration(100 * time.Millisecond)
	}
}

func BenchmarkBackOff_Exponential(b *testing.B) {
	backOff := New().
		WithMaxAttempts(3).
		WithMaxDuration(1 * time.Second).
		WithInitialDuration(10 * time.Millisecond)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		backOff.Exponential()
	}
}

func BenchmarkBackOff_Linear(b *testing.B) {
	backOff := New().
		WithMaxAttempts(3).
		WithInitialDuration(10 * time.Millisecond)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		backOff.Linear()
	}
}

func BenchmarkDo_Success(b *testing.B) {
	operation := func() error {
		return nil
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		backOff := New().WithMaxAttempts(1).WithInitialDuration(1 * time.Microsecond).Exponential()
		_ = Do(operation, backOff)
	}
}

func BenchmarkDo_WithRetries(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		callCount := 0
		operation := func() error {
			callCount++
			if callCount < 3 {
				return errors.New("temporary error")
			}
			return nil
		}

		backOff := New().WithMaxAttempts(5).WithInitialDuration(1 * time.Microsecond).Exponential()
		_ = Do(operation, backOff)
	}
}
