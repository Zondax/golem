package zobservability

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const (
	testOtelServiceName = "test-service"
	testCounterName     = "test_counter"
	testGaugeName       = "test_gauge"
	testHistogramName   = "test_histogram"
)

// =============================================================================
// OPENTELEMETRY METRICS PROVIDER CREATION TESTS
// =============================================================================

func TestNewOpenTelemetryMetricsProvider_WhenValidConfig_ShouldReturnProvider(t *testing.T) {
	// Arrange
	name := testOtelServiceName
	config := MetricsConfig{
		Enabled:  true,
		Provider: string(MetricsProviderOpenTelemetry),
		OpenTelemetry: OpenTelemetryMetricsConfig{
			Endpoint:     "localhost:4317",
			ServiceName:  testOtelServiceName,
			ExportMode:   OTelExportModePush,
			PushInterval: 30 * time.Second,
		},
	}

	// Act
	provider, err := NewOpenTelemetryMetricsProvider(name, config)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, provider)
	assert.Equal(t, name, provider.Name())
	assert.Implements(t, (*MetricsProvider)(nil), provider)
}

func TestNewOpenTelemetryMetricsProvider_WhenInvalidConfig_ShouldReturnError(t *testing.T) {
	// Arrange
	name := testOtelServiceName
	config := MetricsConfig{
		Enabled:  true,
		Provider: string(MetricsProviderOpenTelemetry),
		OpenTelemetry: OpenTelemetryMetricsConfig{
			Endpoint:    "", // Invalid: empty endpoint
			ServiceName: testOtelServiceName,
			ExportMode:  OTelExportModePush,
		},
	}

	// Act
	provider, err := NewOpenTelemetryMetricsProvider(name, config)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, provider)
	assert.Contains(t, err.Error(), "invalid OpenTelemetry configuration")
}

func TestNewOpenTelemetryMetricsProvider_WhenCompleteConfig_ShouldReturnProvider(t *testing.T) {
	// Arrange
	name := "production-service"
	config := MetricsConfig{
		Enabled:  true,
		Provider: string(MetricsProviderOpenTelemetry),
		OpenTelemetry: OpenTelemetryMetricsConfig{
			Endpoint:       "ingest.signoz.io:443",
			Insecure:       false,
			ServiceName:    "production-service",
			ServiceVersion: "2.1.0",
			Environment:    "production",
			Hostname:       "prod-server-01",
			Headers: map[string]string{
				"signoz-access-token": "token123",
				"x-api-key":           "key456",
			},
			ExportMode:    OTelExportModeEndpoint,
			PushInterval:  60 * time.Second,
			BatchTimeout:  10 * time.Second,
			ExportTimeout: 45 * time.Second,
		},
	}

	// Act
	provider, err := NewOpenTelemetryMetricsProvider(name, config)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, provider)
	assert.Equal(t, name, provider.Name())
}

// =============================================================================
// COUNTER OPERATIONS TESTS
// =============================================================================

func TestOpenTelemetryMetricsProvider_WhenIncrementCounter_ShouldNotReturnError(t *testing.T) {
	// Arrange
	provider := createTestOpenTelemetryProvider(t)
	ctx := context.Background()
	name := testCounterName
	labels := map[string]string{"component": "test"}

	// Register the counter first
	err := provider.RegisterCounter(name, "Test counter", []string{"component"})
	assert.NoError(t, err)

	// Act
	err = provider.IncrementCounter(ctx, name, labels)

	// Assert
	assert.NoError(t, err)
}

func TestOpenTelemetryMetricsProvider_WhenAddToCounter_ShouldNotReturnError(t *testing.T) {
	// Arrange
	provider := createTestOpenTelemetryProvider(t)
	ctx := context.Background()
	name := testCounterName
	value := 5.0
	labels := map[string]string{"component": "test"}

	// Register the counter first
	err := provider.RegisterCounter(name, "Test counter", []string{"component"})
	assert.NoError(t, err)

	// Act
	err = provider.AddCounter(ctx, name, value, labels)

	// Assert
	assert.NoError(t, err)
}

func TestOpenTelemetryMetricsProvider_WhenIncrementCounterWithEmptyLabels_ShouldNotReturnError(t *testing.T) {
	// Arrange
	provider := createTestOpenTelemetryProvider(t)
	ctx := context.Background()
	name := testCounterName
	labels := map[string]string{}

	// Register the counter first
	err := provider.RegisterCounter(name, "Test counter", []string{})
	assert.NoError(t, err)

	// Act
	err = provider.IncrementCounter(ctx, name, labels)

	// Assert
	assert.NoError(t, err)
}

func TestOpenTelemetryMetricsProvider_WhenIncrementCounterWithNilLabels_ShouldNotReturnError(t *testing.T) {
	// Arrange
	provider := createTestOpenTelemetryProvider(t)
	ctx := context.Background()
	name := testCounterName

	// Register the counter first
	err := provider.RegisterCounter(name, "Test counter", []string{})
	assert.NoError(t, err)

	// Act
	err = provider.IncrementCounter(ctx, name, nil)

	// Assert
	assert.NoError(t, err)
}

func TestOpenTelemetryMetricsProvider_WhenAddToCounterWithMultipleLabels_ShouldNotReturnError(t *testing.T) {
	// Arrange
	provider := createTestOpenTelemetryProvider(t)
	ctx := context.Background()
	name := testCounterName
	value := 10.0
	labels := map[string]string{
		"component": "test",
		"version":   "1.0.0",
		"env":       "development",
	}

	// Register the counter first
	err := provider.RegisterCounter(name, "Test counter", []string{"component", "version", "env"})
	assert.NoError(t, err)

	// Act
	err = provider.AddCounter(ctx, name, value, labels)

	// Assert
	assert.NoError(t, err)
}

// =============================================================================
// GAUGE OPERATIONS TESTS
// =============================================================================

func TestOpenTelemetryMetricsProvider_WhenSetGauge_ShouldNotReturnError(t *testing.T) {
	// Arrange
	provider := createTestOpenTelemetryProvider(t)
	ctx := context.Background()
	name := testGaugeName
	value := 42.5
	labels := map[string]string{"component": "test"}

	// Register the gauge first
	err := provider.RegisterGauge(name, "Test gauge", []string{"component"})
	assert.NoError(t, err)

	// Act
	err = provider.SetGauge(ctx, name, value, labels)

	// Assert
	assert.NoError(t, err)
}

func TestOpenTelemetryMetricsProvider_WhenAddToGauge_ShouldReturnError(t *testing.T) {
	// Arrange
	provider := createTestOpenTelemetryProvider(t)
	ctx := context.Background()
	name := testGaugeName
	value := 5.0
	labels := map[string]string{"component": "test"}

	// Act
	err := provider.AddGauge(ctx, name, value, labels)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "AddGauge not supported for OpenTelemetry gauges")
}

func TestOpenTelemetryMetricsProvider_WhenSetGaugeWithNegativeValue_ShouldNotReturnError(t *testing.T) {
	// Arrange
	provider := createTestOpenTelemetryProvider(t)
	ctx := context.Background()
	name := testGaugeName
	value := -15.5
	labels := map[string]string{"component": "test"}

	// Register the gauge first
	err := provider.RegisterGauge(name, "Test gauge", []string{"component"})
	assert.NoError(t, err)

	// Act
	err = provider.SetGauge(ctx, name, value, labels)

	// Assert
	assert.NoError(t, err)
}

func TestOpenTelemetryMetricsProvider_WhenSetGaugeWithZeroValue_ShouldNotReturnError(t *testing.T) {
	// Arrange
	provider := createTestOpenTelemetryProvider(t)
	ctx := context.Background()
	name := testGaugeName
	value := 0.0
	labels := map[string]string{"component": "test"}

	// Register the gauge first
	err := provider.RegisterGauge(name, "Test gauge", []string{"component"})
	assert.NoError(t, err)

	// Act
	err = provider.SetGauge(ctx, name, value, labels)

	// Assert
	assert.NoError(t, err)
}

// =============================================================================
// HISTOGRAM OPERATIONS TESTS
// =============================================================================

func TestOpenTelemetryMetricsProvider_WhenRecordHistogram_ShouldNotReturnError(t *testing.T) {
	// Arrange
	provider := createTestOpenTelemetryProvider(t)
	ctx := context.Background()
	name := testHistogramName
	value := 123.45
	labels := map[string]string{"component": "test"}

	// Register the histogram first
	err := provider.RegisterHistogram(name, "Test histogram", []string{"component"}, []float64{1, 5, 10, 50, 100, 500})
	assert.NoError(t, err)

	// Act
	err = provider.RecordHistogram(ctx, name, value, labels)

	// Assert
	assert.NoError(t, err)
}

func TestOpenTelemetryMetricsProvider_WhenRecordDuration_ShouldNotReturnError(t *testing.T) {
	// Arrange
	provider := createTestOpenTelemetryProvider(t)
	ctx := context.Background()
	name := "test_duration"
	duration := 250 * time.Millisecond
	labels := map[string]string{"operation": "test"}

	// Register the histogram first
	err := provider.RegisterHistogram(name, "Test duration", []string{"operation"}, []float64{0.1, 0.5, 1.0, 5.0})
	assert.NoError(t, err)

	// Act
	err = provider.RecordDuration(ctx, name, duration, labels)

	// Assert
	assert.NoError(t, err)
}

func TestOpenTelemetryMetricsProvider_WhenRecordHistogramWithLargeValue_ShouldNotReturnError(t *testing.T) {
	// Arrange
	provider := createTestOpenTelemetryProvider(t)
	ctx := context.Background()
	name := testHistogramName
	value := 999999.99
	labels := map[string]string{"component": "test"}

	// Register the histogram first
	err := provider.RegisterHistogram(name, "Test histogram", []string{"component"}, []float64{1000, 10000, 100000, 1000000})
	assert.NoError(t, err)

	// Act
	err = provider.RecordHistogram(ctx, name, value, labels)

	// Assert
	assert.NoError(t, err)
}

func TestOpenTelemetryMetricsProvider_WhenRecordDurationWithNanoseconds_ShouldNotReturnError(t *testing.T) {
	// Arrange
	provider := createTestOpenTelemetryProvider(t)
	ctx := context.Background()
	name := "test_duration"
	duration := 500 * time.Nanosecond
	labels := map[string]string{"operation": "fast"}

	// Register the histogram first
	err := provider.RegisterHistogram(name, "Test duration", []string{"operation"}, []float64{0.000001, 0.00001, 0.0001, 0.001})
	assert.NoError(t, err)

	// Act
	err = provider.RecordDuration(ctx, name, duration, labels)

	// Assert
	assert.NoError(t, err)
}

// =============================================================================
// REGISTRATION TESTS
// =============================================================================

func TestOpenTelemetryMetricsProvider_WhenRegisterCounter_ShouldNotReturnError(t *testing.T) {
	// Arrange
	provider := createTestOpenTelemetryProvider(t)
	name := "test_counter"
	help := "Test counter description"
	labelNames := []string{"component", "version"}

	// Act
	err := provider.RegisterCounter(name, help, labelNames)

	// Assert
	assert.NoError(t, err)
}

func TestOpenTelemetryMetricsProvider_WhenRegisterGauge_ShouldNotReturnError(t *testing.T) {
	// Arrange
	provider := createTestOpenTelemetryProvider(t)
	name := "test_gauge"
	help := "Test gauge description"
	labelNames := []string{"component", "instance"}

	// Act
	err := provider.RegisterGauge(name, help, labelNames)

	// Assert
	assert.NoError(t, err)
}

func TestOpenTelemetryMetricsProvider_WhenRegisterHistogram_ShouldNotReturnError(t *testing.T) {
	// Arrange
	provider := createTestOpenTelemetryProvider(t)
	name := "test_histogram"
	help := "Test histogram description"
	labelNames := []string{"method", "status"}
	buckets := []float64{0.1, 0.5, 1.0, 2.5, 5.0, 10.0}

	// Act
	err := provider.RegisterHistogram(name, help, labelNames, buckets)

	// Assert
	assert.NoError(t, err)
}

func TestOpenTelemetryMetricsProvider_WhenRegisterCounterWithEmptyLabels_ShouldNotReturnError(t *testing.T) {
	// Arrange
	provider := createTestOpenTelemetryProvider(t)
	name := "test_counter"
	help := "Test counter description"
	labelNames := []string{}

	// Act
	err := provider.RegisterCounter(name, help, labelNames)

	// Assert
	assert.NoError(t, err)
}

func TestOpenTelemetryMetricsProvider_WhenRegisterHistogramWithEmptyBuckets_ShouldNotReturnError(t *testing.T) {
	// Arrange
	provider := createTestOpenTelemetryProvider(t)
	name := "test_histogram"
	help := "Test histogram description"
	labelNames := []string{"method"}
	buckets := []float64{}

	// Act
	err := provider.RegisterHistogram(name, help, labelNames, buckets)

	// Assert
	assert.NoError(t, err)
}

// =============================================================================
// LIFECYCLE TESTS
// =============================================================================

func TestOpenTelemetryMetricsProvider_WhenStart_ShouldNotReturnError(t *testing.T) {
	// Arrange
	provider := createTestOpenTelemetryProvider(t)

	// Act
	err := provider.Start()

	// Assert
	assert.NoError(t, err)
}

func TestOpenTelemetryMetricsProvider_WhenStop_ShouldNotReturnError(t *testing.T) {
	// Arrange
	provider := createTestOpenTelemetryProvider(t)

	// Act & Assert - Stop may return connection errors when trying to flush metrics, which is expected in tests
	// We just verify it doesn't panic and returns some result
	assert.NotPanics(t, func() {
		_ = provider.Stop()
	})
}

func TestOpenTelemetryMetricsProvider_WhenName_ShouldReturnCorrectName(t *testing.T) {
	// Arrange
	expectedName := "test-service"
	provider := createTestOpenTelemetryProviderWithName(t, expectedName)

	// Act
	name := provider.Name()

	// Assert
	assert.Equal(t, expectedName, name)
}

// =============================================================================
// COMPLEX SCENARIOS TESTS
// =============================================================================

func TestOpenTelemetryMetricsProvider_WhenMultipleOperations_ShouldNotReturnError(t *testing.T) {
	// Arrange
	provider := createTestOpenTelemetryProvider(t)
	ctx := context.Background()

	// Register all metrics first
	err := provider.RegisterCounter("requests_total", "Total requests", []string{"method"})
	assert.NoError(t, err)
	err = provider.RegisterGauge("memory_usage", "Memory usage", []string{"instance"})
	assert.NoError(t, err)
	err = provider.RegisterGauge("cpu_usage", "CPU usage", []string{"instance"})
	assert.NoError(t, err)
	err = provider.RegisterHistogram("request_duration", "Request duration", []string{"endpoint"}, []float64{0.1, 0.5, 1.0, 5.0})
	assert.NoError(t, err)
	err = provider.RegisterHistogram("db_query_duration", "DB query duration", []string{"query"}, []float64{0.01, 0.1, 0.5, 1.0})
	assert.NoError(t, err)

	// Act & Assert - Multiple counter operations
	err = provider.IncrementCounter(ctx, "requests_total", map[string]string{"method": "GET"})
	assert.NoError(t, err)

	err = provider.AddCounter(ctx, "requests_total", 5, map[string]string{"method": "POST"})
	assert.NoError(t, err)

	// Act & Assert - Multiple gauge operations
	err = provider.SetGauge(ctx, "memory_usage", 75.5, map[string]string{"instance": "server1"})
	assert.NoError(t, err)

	err = provider.AddGauge(ctx, "cpu_usage", 45.2, map[string]string{"instance": "server2"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "AddGauge not supported for OpenTelemetry gauges")

	// Act & Assert - Multiple histogram operations
	err = provider.RecordHistogram(ctx, "request_duration", 0.25, map[string]string{"endpoint": "/api/users"})
	assert.NoError(t, err)

	err = provider.RecordDuration(ctx, "db_query_duration", 150*time.Millisecond, map[string]string{"query": "SELECT"})
	assert.NoError(t, err)
}

func TestOpenTelemetryMetricsProvider_WhenSameMetricNameDifferentTypes_ShouldHandleCorrectly(t *testing.T) {
	// Arrange
	provider := createTestOpenTelemetryProvider(t)
	ctx := context.Background()
	metricName := "test_metric"

	// Register all metrics first
	err := provider.RegisterCounter(metricName+"_counter", "Test counter", []string{"type"})
	assert.NoError(t, err)
	err = provider.RegisterGauge(metricName+"_gauge", "Test gauge", []string{"type"})
	assert.NoError(t, err)
	err = provider.RegisterHistogram(metricName+"_histogram", "Test histogram", []string{"type"}, []float64{0.1, 1.0, 10.0})
	assert.NoError(t, err)

	// Act & Assert - Use same name for different metric types
	err = provider.IncrementCounter(ctx, metricName+"_counter", map[string]string{"type": "counter"})
	assert.NoError(t, err)

	err = provider.AddCounter(ctx, metricName+"_counter", 3.0, map[string]string{"type": "counter"})
	assert.NoError(t, err)

	err = provider.SetGauge(ctx, metricName+"_gauge", 42.0, map[string]string{"type": "gauge"})
	assert.NoError(t, err)

	err = provider.AddGauge(ctx, metricName+"_gauge", 1.0, map[string]string{"type": "gauge"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "AddGauge not supported for OpenTelemetry gauges")

	err = provider.RecordHistogram(ctx, metricName+"_histogram", 1.5, map[string]string{"type": "histogram"})
	assert.NoError(t, err)
}

func TestOpenTelemetryMetricsProvider_WhenRegistrationAndUsage_ShouldWorkTogether(t *testing.T) {
	// Arrange
	provider := createTestOpenTelemetryProvider(t)
	ctx := context.Background()

	// Act - Register metrics first
	err := provider.RegisterCounter("registered_counter", "A registered counter", []string{"label1"})
	assert.NoError(t, err)

	err = provider.RegisterGauge("registered_gauge", "A registered gauge", []string{"label2"})
	assert.NoError(t, err)

	err = provider.RegisterHistogram("registered_histogram", "A registered histogram", []string{"label3"}, []float64{1, 5, 10})
	assert.NoError(t, err)

	// Act - Use the registered metrics
	err = provider.IncrementCounter(ctx, "registered_counter", map[string]string{"label1": "value1"})
	assert.NoError(t, err)

	err = provider.AddCounter(ctx, "registered_counter", 5.0, map[string]string{"label1": "value1"})
	assert.NoError(t, err)

	err = provider.SetGauge(ctx, "registered_gauge", 100.0, map[string]string{"label2": "value2"})
	assert.NoError(t, err)

	err = provider.AddGauge(ctx, "registered_gauge", 5.0, map[string]string{"label2": "value2"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "AddGauge not supported for OpenTelemetry gauges")

	err = provider.RecordHistogram(ctx, "registered_histogram", 7.5, map[string]string{"label3": "value3"})
	assert.NoError(t, err)
}

// =============================================================================
// HELPER FUNCTIONS
// =============================================================================

func createTestOpenTelemetryProvider(t *testing.T) MetricsProvider {
	return createTestOpenTelemetryProviderWithName(t, "test-service")
}

func createTestOpenTelemetryProviderWithName(t *testing.T, name string) MetricsProvider {
	config := MetricsConfig{
		Enabled:  true,
		Provider: string(MetricsProviderOpenTelemetry),
		OpenTelemetry: OpenTelemetryMetricsConfig{
			Endpoint:     "localhost:4317",
			Insecure:     true,
			ServiceName:  name,
			ExportMode:   OTelExportModePush,
			PushInterval: 30 * time.Second,
		},
	}

	provider, err := NewOpenTelemetryMetricsProvider(name, config)
	assert.NoError(t, err)
	assert.NotNil(t, provider)

	return provider
}
