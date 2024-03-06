package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
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
