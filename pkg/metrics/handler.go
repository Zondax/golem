package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

type MetricHandler interface {
	Update(collector prometheus.Collector, value float64, labels ...string) error
	Type() string
}

func (t *taskMetrics) UpdateMetric(name string, value float64, labels ...string) {
	t.mux.RLock()
	metricDetail, ok := t.metrics[name]
	t.mux.RUnlock()
	if !ok {
		zap.S().Errorf("Error: Metric not found for %s", name)
		return
	}

	if err := metricDetail.Handler.Update(metricDetail.Collector, value, labels...); err != nil {
		zap.S().Errorf("Error updating metric %s. Err: %s", name, err.Error())
	}
}
