package collectors

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
)

type Counter struct{}

func (c *Counter) Update(collector prometheus.Collector, value float64, labels ...string) error {
	if len(labels) > 0 {
		if metricVec, ok := collector.(*prometheus.CounterVec); ok {
			metric := metricVec.WithLabelValues(labels...)
			metric.Add(value)
			return nil
		}
		return fmt.Errorf("invalid metric type, expected CounterVec for labels")
	}

	if metric, ok := collector.(prometheus.Counter); ok {
		metric.Add(value)
		return nil
	}

	return fmt.Errorf("invalid metric type, expected Counter")
}
