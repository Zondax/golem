package metrics

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
)

type MetricHandler interface {
	Update(collector prometheus.Collector, value float64, labels ...string) error
	Type() string
}

type IncrementDecrementHandler interface {
	MetricHandler
	Inc(collector prometheus.Collector, labels ...string) error
	Dec(collector prometheus.Collector, labels ...string) error
}

func (t *taskMetrics) performMetricAction(name string, action func(MetricHandler, prometheus.Collector, ...string) error, labels ...string) error {
	t.mux.RLock()
	metricDetail, ok := t.metrics[name]
	t.mux.RUnlock()

	if !ok {
		return fmt.Errorf("error: Metric not found for %s", name)
	}

	if err := action(metricDetail.Handler, metricDetail.Collector, labels...); err != nil {
		return fmt.Errorf("error performing action on metric %s. Err: %s", name, err.Error())
	}

	return nil
}

func (t *taskMetrics) UpdateMetric(name string, value float64, labels ...string) error {
	return t.performMetricAction(name, func(h MetricHandler, c prometheus.Collector, labels ...string) error {
		return h.Update(c, value, labels...)
	}, labels...)
}

func (t *taskMetrics) IncrementMetric(name string, labels ...string) error {
	return t.performMetricAction(name, func(h MetricHandler, c prometheus.Collector, labels ...string) error {
		if incDecHandler, ok := h.(IncrementDecrementHandler); ok {
			return incDecHandler.Inc(c, labels...)
		}
		return fmt.Errorf("error: Metric %s cannot be incremented", name)
	}, labels...)
}

func (t *taskMetrics) DecrementMetric(name string, labels ...string) error {
	return t.performMetricAction(name, func(h MetricHandler, c prometheus.Collector, labels ...string) error {
		if incDecHandler, ok := h.(IncrementDecrementHandler); ok {
			return incDecHandler.Dec(c, labels...)
		}
		return fmt.Errorf("error: Metric %s cannot be decremented", name)
	}, labels...)
}
