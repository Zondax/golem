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
	assert.IsType(t, prometheus.NewCounter(prometheus.CounterOpts{}), tm.metrics["test_counter"].Collector)
}
