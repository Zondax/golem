package zobservability

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMetricType_WhenUsed_ShouldHaveCorrectValues(t *testing.T) {
	// Test metric type constants
	assert.Equal(t, MetricType(0), MetricTypeCounter)
	assert.Equal(t, MetricType(1), MetricTypeGauge)
	assert.Equal(t, MetricType(2), MetricTypeHistogram)
}

func TestMetricType_WhenCompared_ShouldBeDistinct(t *testing.T) {
	// Test that metric types are distinct
	types := []MetricType{MetricTypeCounter, MetricTypeGauge, MetricTypeHistogram}

	for i, typeA := range types {
		for j, typeB := range types {
			if i != j {
				assert.NotEqual(t, typeA, typeB, "Metric types should be distinct")
			}
		}
	}
}

func TestMetricDefinition_WhenCreated_ShouldHaveCorrectFields(t *testing.T) {
	// Test MetricDefinition struct
	definition := MetricDefinition{
		Name:       "test_counter",
		Help:       "A test counter metric",
		Type:       MetricTypeCounter,
		LabelNames: []string{"label1", "label2"},
		Buckets:    []float64{0.1, 0.5, 1.0, 5.0},
	}

	assert.Equal(t, "test_counter", definition.Name)
	assert.Equal(t, "A test counter metric", definition.Help)
	assert.Equal(t, MetricTypeCounter, definition.Type)
	assert.Equal(t, []string{"label1", "label2"}, definition.LabelNames)
	assert.Equal(t, []float64{0.1, 0.5, 1.0, 5.0}, definition.Buckets)
}

func TestMetricDefinition_WhenCreatedForDifferentTypes_ShouldWorkCorrectly(t *testing.T) {
	testCases := []struct {
		name              string
		metricType        MetricType
		shouldHaveBuckets bool
	}{
		{"counter", MetricTypeCounter, false},
		{"gauge", MetricTypeGauge, false},
		{"histogram", MetricTypeHistogram, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			definition := MetricDefinition{
				Name:       "test_" + tc.name,
				Help:       "A test " + tc.name + " metric",
				Type:       tc.metricType,
				LabelNames: []string{"service", "environment"},
			}

			if tc.shouldHaveBuckets {
				definition.Buckets = []float64{0.1, 0.5, 1.0, 5.0, 10.0}
			}

			assert.Equal(t, "test_"+tc.name, definition.Name)
			assert.Equal(t, "A test "+tc.name+" metric", definition.Help)
			assert.Equal(t, tc.metricType, definition.Type)
			assert.Equal(t, []string{"service", "environment"}, definition.LabelNames)

			if tc.shouldHaveBuckets {
				assert.NotEmpty(t, definition.Buckets)
			}
		})
	}
}

func TestMetricDefinition_WhenUsedWithEmptyFields_ShouldHandleGracefully(t *testing.T) {
	// Test with minimal definition
	definition := MetricDefinition{
		Name: "minimal_metric",
		Type: MetricTypeCounter,
	}

	assert.Equal(t, "minimal_metric", definition.Name)
	assert.Equal(t, "", definition.Help) // Should be empty string
	assert.Equal(t, MetricTypeCounter, definition.Type)
	assert.Nil(t, definition.LabelNames) // Should be nil
	assert.Nil(t, definition.Buckets)    // Should be nil
}

func TestMetricDefinition_WhenUsedWithComplexLabels_ShouldWorkCorrectly(t *testing.T) {
	// Test with complex label names
	definition := MetricDefinition{
		Name: "complex_metric",
		Help: "A metric with complex labels",
		Type: MetricTypeHistogram,
		LabelNames: []string{
			"service_name",
			"environment",
			"region",
			"instance_id",
			"operation_type",
			"status_code",
		},
		Buckets: []float64{0.001, 0.01, 0.1, 0.5, 1.0, 2.5, 5.0, 10.0},
	}

	assert.Equal(t, "complex_metric", definition.Name)
	assert.Equal(t, "A metric with complex labels", definition.Help)
	assert.Equal(t, MetricTypeHistogram, definition.Type)
	assert.Len(t, definition.LabelNames, 6)
	assert.Contains(t, definition.LabelNames, "service_name")
	assert.Contains(t, definition.LabelNames, "environment")
	assert.Contains(t, definition.LabelNames, "region")
	assert.Contains(t, definition.LabelNames, "instance_id")
	assert.Contains(t, definition.LabelNames, "operation_type")
	assert.Contains(t, definition.LabelNames, "status_code")
	assert.Len(t, definition.Buckets, 8)
}

func TestMetricsProvider_WhenImplemented_ShouldHaveAllRequiredMethods(t *testing.T) {
	// Test that MetricsProvider interface has all required methods
	// This is a compile-time check - if the interface changes, this will fail to compile

	var provider MetricsProvider
	assert.Nil(t, provider) // Provider should be nil when not initialized

	// Test that we can assign a noop implementation
	provider = NewNoopMetricsProvider("test")
	assert.NotNil(t, provider)
	assert.Implements(t, (*MetricsProvider)(nil), provider)
}

func TestMetricsProvider_WhenUsedWithContext_ShouldAcceptContext(t *testing.T) {
	// Test that all methods accept context
	provider := NewNoopMetricsProvider("test")
	ctx := context.Background()
	labels := map[string]string{"test": "value"}

	// Test counter operations
	err := provider.IncrementCounter(ctx, "test_counter", labels)
	assert.NoError(t, err)

	err = provider.AddCounter(ctx, "test_counter", 5.0, labels)
	assert.NoError(t, err)

	// Test gauge operations
	err = provider.SetGauge(ctx, "test_gauge", 42.0, labels)
	assert.NoError(t, err)

	err = provider.AddGauge(ctx, "test_gauge", 1.0, labels)
	assert.NoError(t, err)

	// Test histogram operations
	err = provider.RecordHistogram(ctx, "test_histogram", 1.5, labels)
	assert.NoError(t, err)

	// Test timer operations
	err = provider.RecordDuration(ctx, "test_duration", time.Millisecond*100, labels)
	assert.NoError(t, err)
}

func TestMetricsProvider_WhenRegisteringMetrics_ShouldAcceptDefinitions(t *testing.T) {
	// Test metric registration methods
	provider := NewNoopMetricsProvider("test")

	// Test counter registration
	err := provider.RegisterCounter("test_counter", "A test counter", []string{"label1"})
	assert.NoError(t, err)

	// Test gauge registration
	err = provider.RegisterGauge("test_gauge", "A test gauge", []string{"label1", "label2"})
	assert.NoError(t, err)

	// Test histogram registration
	buckets := []float64{0.1, 0.5, 1.0, 5.0}
	err = provider.RegisterHistogram("test_histogram", "A test histogram", []string{"label1"}, buckets)
	assert.NoError(t, err)
}

func TestMetricsProvider_WhenManagingLifecycle_ShouldProvideStartStop(t *testing.T) {
	// Test lifecycle methods
	provider := NewNoopMetricsProvider("test-provider")

	// Test start
	err := provider.Start()
	assert.NoError(t, err)

	// Test name
	name := provider.Name()
	assert.Equal(t, "test-provider", name)

	// Test stop
	err = provider.Stop()
	assert.NoError(t, err)
}
