package metrics

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
)

func (t *taskMetrics) ResetMetric(name string) error {
	t.mux.Lock()
	metricDetail, ok := t.metrics[name]
	t.mux.Unlock()

	if !ok {
		return fmt.Errorf("metric %s not found", name)
	}

	if r := prometheus.Unregister(metricDetail.Collector); !r {
		return fmt.Errorf("failed to unregister metric %s", name)
	}

	t.mux.Lock()
	t.metrics[name] = MetricDetail{
		Collector: metricDetail.Collector,
		Handler:   metricDetail.Handler,
		Help:      metricDetail.Help,
		Labels:    metricDetail.Labels,
	}
	delete(t.metrics, name)
	t.mux.Unlock()

	var labels []string
	if len(metricDetail.Labels) > 0 {
		labels = metricDetail.Labels[:len(metricDetail.Labels)-1] // Delete "app_name" label to avoid duplicate error
	}
	if err := t.RegisterMetric(name, metricDetail.Help, labels, metricDetail.Handler); err != nil {
		return err
	}
	return nil
}
