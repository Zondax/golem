package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/zondax/golem/pkg/metrics/collectors"
	"testing"
)

func TestRegisterMetric(t *testing.T) {
	tm := &taskMetrics{metrics: make(map[string]MetricDetail)}
	err := tm.RegisterMetric("test_counter", "help", nil, &collectors.Counter{})

	assert.NoError(t, err)
	assert.NotNil(t, tm.metrics["test_counter"])
	assert.IsType(t, MetricDetail{}, tm.metrics["test_counter"])
	assert.IsType(t, prometheus.NewCounterVec(prometheus.CounterOpts{}, []string{"app_name", "app_version"}), tm.metrics["test_counter"].Collector)
}

func TestFormatMetricName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"router-name", "router_name"},
		{"router name", "router_name"},
		{"router$name", "router_name"},
		{"routerName", "routerName"},
		{"router@name-123", "router_name_123"},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			result := formatMetricName(test.input)
			assert.Equal(t, test.expected, result)
		})
	}
}

func (suite *MetricsTestSuite) TestMetricNameFormatting() {
	tests := []struct {
		metricName    string
		formattedName string
	}{
		{
			metricName:    "test-metric-name",
			formattedName: "test_metric_name",
		},
		{
			metricName:    "another.test@metric",
			formattedName: "another_test_metric",
		},
	}
	for _, tt := range tests {
		suite.Run(tt.metricName, func() {
			// Create a gauge for testing
			realGauge := prometheus.NewGauge(prometheus.GaugeOpts{
				Name: "test_gauge",
				Help: "Test gauge",
			})

			// Create a new task metrics instance
			taskMetric := &taskMetrics{
				metrics: make(map[string]MetricDetail),
			}

			// Store the metric with a formatted name
			taskMetric.metrics[tt.formattedName] = MetricDetail{
				Collector: realGauge,
				Handler:   suite.mockH,
			}

			suite.mockH.On("Update", realGauge, 1.0, mock.Anything).Return(nil).Once()

			// "another.test@metric" should be formatted to "another_test_metric"
			err := taskMetric.UpdateMetric(tt.metricName, 1.0)

			// Verify there are no errors and the mock expectations are met
			suite.Nil(err)
			suite.mockH.AssertExpectations(suite.T())
		})
	}
}
