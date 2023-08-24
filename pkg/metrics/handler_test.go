package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/mock"
	"testing"
)

type mockHandler struct {
	mock.Mock
}

func (m *mockHandler) Update(collector prometheus.Collector, value float64, labels ...string) error {
	args := m.Called(collector, value, labels)
	return args.Error(0)
}

func (m *mockHandler) Type() string {
	args := m.Called()
	return args.String(0)
}

func TestUpdateMetric(t *testing.T) {
	mockH := new(mockHandler)

	realCounter := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "test_counter",
		Help: "Test counter",
	})

	metricDetail := MetricDetail{
		Collector: realCounter,
		Handler:   mockH,
	}

	taskMetric := &taskMetrics{
		metrics: make(map[string]MetricDetail),
	}
	taskMetric.metrics["test_metric"] = metricDetail

	mockH.On("Update", realCounter, 1.0, mock.Anything).Return(nil).Once()

	taskMetric.UpdateMetric("test_metric", 1.0)
	mockH.AssertExpectations(t)
}
