package zobservability

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewNoopMetricsProvider_WhenCalled_ShouldReturnValidProvider(t *testing.T) {
	// Arrange
	name := "test-provider"

	// Act
	provider := NewNoopMetricsProvider(name)

	// Assert
	assert.NotNil(t, provider)
	assert.Implements(t, (*MetricsProvider)(nil), provider)
	assert.Equal(t, name, provider.Name())
}

func TestNewNoopMetricsProvider_WhenCalledWithEmptyName_ShouldReturnProviderWithEmptyName(t *testing.T) {
	// Arrange
	name := ""

	// Act
	provider := NewNoopMetricsProvider(name)

	// Assert
	assert.NotNil(t, provider)
	assert.Equal(t, "", provider.Name())
}

func TestNoopMetricsProvider_WhenIncrementCounter_ShouldNotReturnError(t *testing.T) {
	// Arrange
	provider := NewNoopMetricsProvider("test")
	ctx := context.Background()
	labels := map[string]string{"service": "test", "environment": "dev"}

	// Act
	err := provider.IncrementCounter(ctx, "test_counter", labels)

	// Assert
	assert.NoError(t, err)
}

func TestNoopMetricsProvider_WhenAddToCounter_ShouldNotReturnError(t *testing.T) {
	// Arrange
	provider := NewNoopMetricsProvider("test")
	ctx := context.Background()
	labels := map[string]string{"service": "test"}

	// Act
	err := provider.AddCounter(ctx, "test_counter", 5.5, labels)

	// Assert
	assert.NoError(t, err)
}

func TestNoopMetricsProvider_WhenSetGauge_ShouldNotReturnError(t *testing.T) {
	// Arrange
	provider := NewNoopMetricsProvider("test")
	ctx := context.Background()
	labels := map[string]string{"instance": "test-1"}

	// Act
	err := provider.SetGauge(ctx, "test_gauge", 42.0, labels)

	// Assert
	assert.NoError(t, err)
}

func TestNoopMetricsProvider_WhenAddToGauge_ShouldNotReturnError(t *testing.T) {
	// Arrange
	provider := NewNoopMetricsProvider("test")
	ctx := context.Background()
	labels := map[string]string{"region": "us-west"}

	// Act
	err := provider.AddGauge(ctx, "test_gauge", -1.5, labels)

	// Assert
	assert.NoError(t, err)
}

func TestNoopMetricsProvider_WhenRecordHistogram_ShouldNotReturnError(t *testing.T) {
	// Arrange
	provider := NewNoopMetricsProvider("test")
	ctx := context.Background()
	labels := map[string]string{"operation": "create_user"}

	// Act
	err := provider.RecordHistogram(ctx, "test_histogram", 1.23, labels)

	// Assert
	assert.NoError(t, err)
}

func TestNoopMetricsProvider_WhenRecordDuration_ShouldNotReturnError(t *testing.T) {
	// Arrange
	provider := NewNoopMetricsProvider("test")
	ctx := context.Background()
	labels := map[string]string{"method": "GET"}
	duration := time.Millisecond * 150

	// Act
	err := provider.RecordDuration(ctx, "test_duration", duration, labels)

	// Assert
	assert.NoError(t, err)
}

func TestNoopMetricsProvider_WhenRegisterCounter_ShouldNotReturnError(t *testing.T) {
	// Arrange
	provider := NewNoopMetricsProvider("test")

	// Act
	err := provider.RegisterCounter("test_counter", "A test counter metric", []string{"service", "environment"})

	// Assert
	assert.NoError(t, err)
}

func TestNoopMetricsProvider_WhenRegisterGauge_ShouldNotReturnError(t *testing.T) {
	// Arrange
	provider := NewNoopMetricsProvider("test")

	// Act
	err := provider.RegisterGauge("test_gauge", "A test gauge metric", []string{"instance", "region"})

	// Assert
	assert.NoError(t, err)
}

func TestNoopMetricsProvider_WhenRegisterHistogram_ShouldNotReturnError(t *testing.T) {
	// Arrange
	provider := NewNoopMetricsProvider("test")
	buckets := []float64{0.1, 0.5, 1.0, 2.5, 5.0, 10.0}

	// Act
	err := provider.RegisterHistogram("test_histogram", "A test histogram metric", []string{"operation"}, buckets)

	// Assert
	assert.NoError(t, err)
}

func TestNoopMetricsProvider_WhenStart_ShouldNotReturnError(t *testing.T) {
	// Arrange
	provider := NewNoopMetricsProvider("test")

	// Act
	err := provider.Start()

	// Assert
	assert.NoError(t, err)
}

func TestNoopMetricsProvider_WhenStop_ShouldNotReturnError(t *testing.T) {
	// Arrange
	provider := NewNoopMetricsProvider("test")

	// Act
	err := provider.Stop()

	// Assert
	assert.NoError(t, err)
}

func TestNoopMetricsProvider_WhenName_ShouldReturnCorrectName(t *testing.T) {
	testCases := []struct {
		name     string
		expected string
	}{
		{"simple", "simple"},
		{"complex-name-with-dashes", "complex-name-with-dashes"},
		{"", ""},
		{"noop", "noop"},
		{"opentelemetry", "opentelemetry"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			provider := NewNoopMetricsProvider(tc.name)

			// Act
			result := provider.Name()

			// Assert
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestNoopMetricsProvider_WhenUsedWithNilContext_ShouldNotPanic(t *testing.T) {
	// Arrange
	provider := NewNoopMetricsProvider("test")
	labels := map[string]string{"test": "value"}

	// Act & Assert - should not panic with nil context
	assert.NotPanics(t, func() {
		_ = provider.IncrementCounter(context.TODO(), "test_counter", labels)
		_ = provider.AddCounter(context.TODO(), "test_counter", 1.0, labels)
		_ = provider.SetGauge(context.TODO(), "test_gauge", 42.0, labels)
		_ = provider.AddGauge(context.TODO(), "test_gauge", 1.0, labels)
		_ = provider.RecordHistogram(context.TODO(), "test_histogram", 1.5, labels)
		_ = provider.RecordDuration(context.TODO(), "test_duration", time.Second, labels)
	})
}

func TestNoopMetricsProvider_WhenUsedWithNilLabels_ShouldNotPanic(t *testing.T) {
	// Arrange
	provider := NewNoopMetricsProvider("test")
	ctx := context.Background()

	// Act & Assert - should not panic with nil labels
	assert.NotPanics(t, func() {
		_ = provider.IncrementCounter(ctx, "test_counter", nil)
		_ = provider.AddCounter(ctx, "test_counter", 1.0, nil)
		_ = provider.SetGauge(ctx, "test_gauge", 42.0, nil)
		_ = provider.AddGauge(ctx, "test_gauge", 1.0, nil)
		_ = provider.RecordHistogram(ctx, "test_histogram", 1.5, nil)
		_ = provider.RecordDuration(ctx, "test_duration", time.Second, nil)
	})
}

func TestNoopMetricsProvider_WhenUsedWithEmptyMetricNames_ShouldNotPanic(t *testing.T) {
	// Arrange
	provider := NewNoopMetricsProvider("test")
	ctx := context.Background()
	labels := map[string]string{"test": "value"}

	// Act & Assert - should not panic with empty metric names
	assert.NotPanics(t, func() {
		_ = provider.IncrementCounter(ctx, "", labels)
		_ = provider.AddCounter(ctx, "", 1.0, labels)
		_ = provider.SetGauge(ctx, "", 42.0, labels)
		_ = provider.AddGauge(ctx, "", 1.0, labels)
		_ = provider.RecordHistogram(ctx, "", 1.5, labels)
		_ = provider.RecordDuration(ctx, "", time.Second, labels)
	})
}

func TestNoopMetricsProvider_WhenRegisteringWithEmptyValues_ShouldNotPanic(t *testing.T) {
	// Arrange
	provider := NewNoopMetricsProvider("test")

	// Act & Assert - should not panic with empty values
	assert.NotPanics(t, func() {
		_ = provider.RegisterCounter("", "", nil)
		_ = provider.RegisterGauge("", "", []string{})
		_ = provider.RegisterHistogram("", "", nil, nil)
		_ = provider.RegisterHistogram("test", "help", []string{"label"}, []float64{})
	})
}

func TestNoopMetricsProvider_WhenUsedInComplexScenario_ShouldWorkCorrectly(t *testing.T) {
	// Arrange
	provider := NewNoopMetricsProvider("complex-test")
	ctx := context.Background()

	// Act - simulate a complex metrics scenario
	// Register metrics
	err := provider.RegisterCounter("requests_total", "Total number of requests", []string{"method", "status"})
	assert.NoError(t, err)

	err = provider.RegisterGauge("active_connections", "Number of active connections", []string{"service"})
	assert.NoError(t, err)

	err = provider.RegisterHistogram("request_duration_seconds", "Request duration in seconds",
		[]string{"method", "endpoint"}, []float64{0.1, 0.5, 1.0, 2.5, 5.0, 10.0})
	assert.NoError(t, err)

	// Start provider
	err = provider.Start()
	assert.NoError(t, err)

	// Use metrics
	labels := map[string]string{"method": "GET", "status": "200"}
	err = provider.IncrementCounter(ctx, "requests_total", labels)
	assert.NoError(t, err)

	err = provider.SetGauge(ctx, "active_connections", 42.0, map[string]string{"service": "api"})
	assert.NoError(t, err)

	err = provider.RecordHistogram(ctx, "request_duration_seconds", 0.75,
		map[string]string{"method": "GET", "endpoint": "/users"})
	assert.NoError(t, err)

	// Stop provider
	err = provider.Stop()
	assert.NoError(t, err)

	// Assert
	assert.Equal(t, "complex-test", provider.Name())
}

func TestNoopMetricsProvider_WhenMultipleInstances_ShouldBeIndependent(t *testing.T) {
	// Arrange
	provider1 := NewNoopMetricsProvider("provider-1")
	provider2 := NewNoopMetricsProvider("provider-2")

	// Act & Assert
	assert.NotEqual(t, provider1, provider2)
	assert.Equal(t, "provider-1", provider1.Name())
	assert.Equal(t, "provider-2", provider2.Name())

	// Both should work independently
	ctx := context.Background()
	labels := map[string]string{"test": "value"}

	err1 := provider1.IncrementCounter(ctx, "test", labels)
	err2 := provider2.IncrementCounter(ctx, "test", labels)

	assert.NoError(t, err1)
	assert.NoError(t, err2)
}
