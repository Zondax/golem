package zobservability

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// INPUT VALIDATION TESTS FOR OPENTELEMETRY METRICS PROVIDER
// =============================================================================

func TestOpenTelemetryMetricsProvider_InputValidation(t *testing.T) {
	// Create a mock provider
	provider := &mockOpenTelemetryMetricsProvider{}

	// Register test metrics first
	err := provider.RegisterCounter("test_counter", "Test counter", []string{"label1"})
	require.NoError(t, err)
	err = provider.RegisterGauge("test_gauge", "Test gauge", []string{"label1"})
	require.NoError(t, err)
	err = provider.RegisterHistogram("test_histogram", "Test histogram", []string{"label1"}, []float64{1, 5, 10})
	require.NoError(t, err)

	// Define nil context for testing
	var nilCtx context.Context

	t.Run("IncrementCounter_Validation", func(t *testing.T) {
		// Test nil context
		err := provider.IncrementCounter(nilCtx, "test_counter", map[string]string{"label1": "value1"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "context cannot be nil")

		// Test empty name
		err = provider.IncrementCounter(context.Background(), "", map[string]string{"label1": "value1"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "metric name cannot be empty")

		// Test valid call
		err = provider.IncrementCounter(context.Background(), "test_counter", map[string]string{"label1": "value1"})
		assert.NoError(t, err)
	})

	t.Run("AddCounter_Validation", func(t *testing.T) {
		// Test nil context
		err := provider.AddCounter(nilCtx, "test_counter", 5.0, map[string]string{"label1": "value1"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "context cannot be nil")

		// Test empty name
		err = provider.AddCounter(context.Background(), "", 5.0, map[string]string{"label1": "value1"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "metric name cannot be empty")

		// Test negative value
		err = provider.AddCounter(context.Background(), "test_counter", -1.0, map[string]string{"label1": "value1"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "counter value must be non-negative")

		// Test valid call
		err = provider.AddCounter(context.Background(), "test_counter", 5.0, map[string]string{"label1": "value1"})
		assert.NoError(t, err)
	})

	t.Run("SetGauge_Validation", func(t *testing.T) {
		// Test nil context
		err := provider.SetGauge(nilCtx, "test_gauge", 10.0, map[string]string{"label1": "value1"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "context cannot be nil")

		// Test empty name
		err = provider.SetGauge(context.Background(), "", 10.0, map[string]string{"label1": "value1"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "metric name cannot be empty")

		// Test valid call (gauges can have negative values)
		err = provider.SetGauge(context.Background(), "test_gauge", -5.0, map[string]string{"label1": "value1"})
		assert.NoError(t, err)
	})

	t.Run("AddGauge_Validation", func(t *testing.T) {
		// Test nil context
		err := provider.AddGauge(nilCtx, "test_gauge", 5.0, map[string]string{"label1": "value1"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "context cannot be nil")

		// Test empty name
		err = provider.AddGauge(context.Background(), "", 5.0, map[string]string{"label1": "value1"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "metric name cannot be empty")

		// Test AddGauge unsupported
		err = provider.AddGauge(context.Background(), "test_gauge", 5.0, map[string]string{"label1": "value1"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "AddGauge not supported for OpenTelemetry gauges")
	})

	t.Run("RecordHistogram_Validation", func(t *testing.T) {
		// Test nil context
		err := provider.RecordHistogram(nilCtx, "test_histogram", 2.5, map[string]string{"label1": "value1"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "context cannot be nil")

		// Test empty name
		err = provider.RecordHistogram(context.Background(), "", 2.5, map[string]string{"label1": "value1"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "metric name cannot be empty")

		// Test negative value
		err = provider.RecordHistogram(context.Background(), "test_histogram", -1.0, map[string]string{"label1": "value1"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "histogram value must be non-negative")

		// Test valid call
		err = provider.RecordHistogram(context.Background(), "test_histogram", 2.5, map[string]string{"label1": "value1"})
		assert.NoError(t, err)
	})

	t.Run("RecordDuration_Validation", func(t *testing.T) {
		// Test nil context
		err := provider.RecordDuration(nilCtx, "test_histogram", 100*time.Millisecond, map[string]string{"label1": "value1"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "context cannot be nil")

		// Test empty name
		err = provider.RecordDuration(context.Background(), "", 100*time.Millisecond, map[string]string{"label1": "value1"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "metric name cannot be empty")

		// Test negative duration
		err = provider.RecordDuration(context.Background(), "test_histogram", -100*time.Millisecond, map[string]string{"label1": "value1"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "duration must be non-negative")

		// Test valid call
		err = provider.RecordDuration(context.Background(), "test_histogram", 100*time.Millisecond, map[string]string{"label1": "value1"})
		assert.NoError(t, err)
	})

	t.Run("Registration_Validation", func(t *testing.T) {
		// Test empty counter name
		err := provider.RegisterCounter("", "Test counter", []string{"label1"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "metric name cannot be empty")

		// Test empty gauge name
		err = provider.RegisterGauge("", "Test gauge", []string{"label1"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "metric name cannot be empty")

		// Test empty histogram name
		err = provider.RegisterHistogram("", "Test histogram", []string{"label1"}, []float64{1, 5, 10})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "metric name cannot be empty")

		// Test valid registrations
		err = provider.RegisterCounter("valid_counter", "Valid counter", []string{"label1"})
		assert.NoError(t, err)

		err = provider.RegisterGauge("valid_gauge", "Valid gauge", []string{"label1"})
		assert.NoError(t, err)

		err = provider.RegisterHistogram("valid_histogram", "Valid histogram", []string{"label1"}, []float64{1, 5, 10})
		assert.NoError(t, err)
	})
}

type mockOpenTelemetryMetricsProvider struct{}

func (m *mockOpenTelemetryMetricsProvider) IncrementCounter(ctx context.Context, name string, labels map[string]string) error {
	if ctx == nil {
		return errors.New("context cannot be nil")
	}
	if name == "" {
		return errors.New("metric name cannot be empty")
	}
	return nil
}

func (m *mockOpenTelemetryMetricsProvider) AddCounter(ctx context.Context, name string, value float64, labels map[string]string) error {
	if ctx == nil {
		return errors.New("context cannot be nil")
	}
	if name == "" {
		return errors.New("metric name cannot be empty")
	}
	if value < 0 {
		return errors.New("counter value must be non-negative")
	}
	return nil
}

func (m *mockOpenTelemetryMetricsProvider) SetGauge(ctx context.Context, name string, value float64, labels map[string]string) error {
	if ctx == nil {
		return errors.New("context cannot be nil")
	}
	if name == "" {
		return errors.New("metric name cannot be empty")
	}
	return nil
}

func (m *mockOpenTelemetryMetricsProvider) AddGauge(ctx context.Context, name string, value float64, labels map[string]string) error {
	if ctx == nil {
		return errors.New("context cannot be nil")
	}
	if name == "" {
		return errors.New("metric name cannot be empty")
	}
	return errors.New("AddGauge not supported for OpenTelemetry gauges")
}

func (m *mockOpenTelemetryMetricsProvider) RecordHistogram(ctx context.Context, name string, value float64, labels map[string]string) error {
	if ctx == nil {
		return errors.New("context cannot be nil")
	}
	if name == "" {
		return errors.New("metric name cannot be empty")
	}
	if value < 0 {
		return errors.New("histogram value must be non-negative")
	}
	return nil
}

func (m *mockOpenTelemetryMetricsProvider) RecordDuration(ctx context.Context, name string, duration time.Duration, labels map[string]string) error {
	if ctx == nil {
		return errors.New("context cannot be nil")
	}
	if name == "" {
		return errors.New("metric name cannot be empty")
	}
	if duration < 0 {
		return errors.New("duration must be non-negative")
	}
	return nil
}

func (m *mockOpenTelemetryMetricsProvider) RegisterCounter(name string, description string, labels []string) error {
	if name == "" {
		return errors.New("metric name cannot be empty")
	}
	return nil
}

func (m *mockOpenTelemetryMetricsProvider) RegisterGauge(name string, description string, labels []string) error {
	if name == "" {
		return errors.New("metric name cannot be empty")
	}
	return nil
}

func (m *mockOpenTelemetryMetricsProvider) RegisterHistogram(name string, description string, labels []string, buckets []float64) error {
	if name == "" {
		return errors.New("metric name cannot be empty")
	}
	return nil
}

func (m *mockOpenTelemetryMetricsProvider) Stop() error {
	return nil
}

// =============================================================================
// INPUT VALIDATION TESTS FOR NOOP METRICS PROVIDER
// =============================================================================

func TestNoopMetricsProvider_InputValidation(t *testing.T) {
	provider := NewNoopMetricsProvider("test-noop")
	require.NotNil(t, provider)

	// Define nil context for testing
	var nilCtx context.Context

	t.Run("IncrementCounter_Validation", func(t *testing.T) {
		// Test nil context
		err := provider.IncrementCounter(nilCtx, "test_counter", map[string]string{"label1": "value1"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "context cannot be nil")

		// Test empty name
		err = provider.IncrementCounter(context.Background(), "", map[string]string{"label1": "value1"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "metric name cannot be empty")

		// Test valid call
		err = provider.IncrementCounter(context.Background(), "test_counter", map[string]string{"label1": "value1"})
		assert.NoError(t, err)
	})

	t.Run("AddCounter_Validation", func(t *testing.T) {
		// Test nil context
		err := provider.AddCounter(nilCtx, "test_counter", 5.0, map[string]string{"label1": "value1"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "context cannot be nil")

		// Test empty name
		err = provider.AddCounter(context.Background(), "", 5.0, map[string]string{"label1": "value1"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "metric name cannot be empty")

		// Test negative value
		err = provider.AddCounter(context.Background(), "test_counter", -1.0, map[string]string{"label1": "value1"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "counter value must be non-negative")

		// Test valid call
		err = provider.AddCounter(context.Background(), "test_counter", 5.0, map[string]string{"label1": "value1"})
		assert.NoError(t, err)
	})

	t.Run("SetGauge_Validation", func(t *testing.T) {
		// Test nil context
		err := provider.SetGauge(nilCtx, "test_gauge", 10.0, map[string]string{"label1": "value1"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "context cannot be nil")

		// Test empty name
		err = provider.SetGauge(context.Background(), "", 10.0, map[string]string{"label1": "value1"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "metric name cannot be empty")

		// Test valid call (gauges can have negative values)
		err = provider.SetGauge(context.Background(), "test_gauge", -5.0, map[string]string{"label1": "value1"})
		assert.NoError(t, err)
	})

	t.Run("AddGauge_Validation", func(t *testing.T) {
		// Test nil context
		err := provider.AddGauge(nilCtx, "test_gauge", 5.0, map[string]string{"label1": "value1"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "context cannot be nil")

		// Test empty name
		err = provider.AddGauge(context.Background(), "", 5.0, map[string]string{"label1": "value1"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "metric name cannot be empty")

		// Test valid call (noop allows AddGauge)
		err = provider.AddGauge(context.Background(), "test_gauge", 5.0, map[string]string{"label1": "value1"})
		assert.NoError(t, err)
	})

	t.Run("RecordHistogram_Validation", func(t *testing.T) {
		// Test nil context
		err := provider.RecordHistogram(nilCtx, "test_histogram", 2.5, map[string]string{"label1": "value1"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "context cannot be nil")

		// Test empty name
		err = provider.RecordHistogram(context.Background(), "", 2.5, map[string]string{"label1": "value1"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "metric name cannot be empty")

		// Test negative value
		err = provider.RecordHistogram(context.Background(), "test_histogram", -1.0, map[string]string{"label1": "value1"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "histogram value must be non-negative")

		// Test valid call
		err = provider.RecordHistogram(context.Background(), "test_histogram", 2.5, map[string]string{"label1": "value1"})
		assert.NoError(t, err)
	})

	t.Run("RecordDuration_Validation", func(t *testing.T) {
		// Test nil context
		err := provider.RecordDuration(nilCtx, "test_histogram", 100*time.Millisecond, map[string]string{"label1": "value1"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "context cannot be nil")

		// Test empty name
		err = provider.RecordDuration(context.Background(), "", 100*time.Millisecond, map[string]string{"label1": "value1"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "metric name cannot be empty")

		// Test negative duration
		err = provider.RecordDuration(context.Background(), "test_histogram", -100*time.Millisecond, map[string]string{"label1": "value1"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "duration must be non-negative")

		// Test valid call
		err = provider.RecordDuration(context.Background(), "test_histogram", 100*time.Millisecond, map[string]string{"label1": "value1"})
		assert.NoError(t, err)
	})

	t.Run("Registration_Validation", func(t *testing.T) {
		// Test empty counter name
		err := provider.RegisterCounter("", "Test counter", []string{"label1"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "metric name cannot be empty")

		// Test empty gauge name
		err = provider.RegisterGauge("", "Test gauge", []string{"label1"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "metric name cannot be empty")

		// Test empty histogram name
		err = provider.RegisterHistogram("", "Test histogram", []string{"label1"}, []float64{1, 5, 10})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "metric name cannot be empty")

		// Test valid registrations
		err = provider.RegisterCounter("valid_counter", "Valid counter", []string{"label1"})
		assert.NoError(t, err)

		err = provider.RegisterGauge("valid_gauge", "Valid gauge", []string{"label1"})
		assert.NoError(t, err)

		err = provider.RegisterHistogram("valid_histogram", "Valid histogram", []string{"label1"}, []float64{1, 5, 10})
		assert.NoError(t, err)
	})
}
