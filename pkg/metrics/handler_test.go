package metrics

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"testing"
)

type MetricsTestSuite struct {
	suite.Suite
	mockH *mockHandler
}

func (suite *MetricsTestSuite) SetupTest() {
	suite.mockH = new(mockHandler)
}

func TestMetricsTestSuite(t *testing.T) {
	suite.Run(t, new(MetricsTestSuite))
}

type mockHandler struct {
	mock.Mock
}

func (m *mockHandler) Update(collector prometheus.Collector, value float64, labels ...string) error {
	args := m.Called(collector, value, labels)
	return args.Error(0)
}

func (m *mockHandler) Inc(collector prometheus.Collector, labels ...string) error {
	args := m.Called(collector, labels)
	return args.Error(0)
}

func (m *mockHandler) Dec(collector prometheus.Collector, labels ...string) error {
	args := m.Called(collector, labels)
	return args.Error(0)
}

func (m *mockHandler) Type() string {
	args := m.Called()
	return args.String(0)
}

func (suite *MetricsTestSuite) TestUpdateMetric() {
	realCounter := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "test_counter",
		Help: "Test counter",
	})
	taskMetric := &taskMetrics{
		metrics: make(map[string]MetricDetail),
	}
	taskMetric.metrics["test_metric"] = MetricDetail{
		Collector: realCounter,
		Handler:   suite.mockH,
	}
	suite.mockH.On("Update", realCounter, 1.0, mock.Anything).Return(nil).Once()
	err := taskMetric.UpdateMetric("test_metric", 1.0)
	suite.Nil(err)
	suite.mockH.AssertExpectations(suite.T())
}

func (suite *MetricsTestSuite) TestIncrementMetric() {
	realGauge := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "test_gauge",
		Help: "Test gauge",
	})
	taskMetric := &taskMetrics{
		metrics: make(map[string]MetricDetail),
	}
	taskMetric.metrics["test_metric"] = MetricDetail{
		Collector: realGauge,
		Handler:   suite.mockH,
	}
	suite.mockH.On("Inc", realGauge, mock.Anything).Return(nil).Once()
	err := taskMetric.IncrementMetric("test_metric")
	suite.Nil(err)
	suite.mockH.AssertExpectations(suite.T())
}

func (suite *MetricsTestSuite) TestDecrementMetric() {
	realGauge := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "test_gauge",
		Help: "Test gauge",
	})
	taskMetric := &taskMetrics{
		metrics: make(map[string]MetricDetail),
	}
	taskMetric.metrics["test_metric"] = MetricDetail{
		Collector: realGauge,
		Handler:   suite.mockH,
	}
	suite.mockH.On("Dec", realGauge, mock.Anything).Return(nil).Once()
	err := taskMetric.DecrementMetric("test_metric")
	suite.Nil(err)
	suite.mockH.AssertExpectations(suite.T())
}

func (suite *MetricsTestSuite) TestIncrementNonGaugeMetric() {
	realCounter := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "test_counter",
		Help: "Test counter",
	})
	taskMetric := &taskMetrics{
		metrics: make(map[string]MetricDetail),
	}
	taskMetric.metrics["test_metric"] = MetricDetail{
		Collector: realCounter,
		Handler:   suite.mockH,
	}
	suite.mockH.On("Inc", realCounter, mock.Anything).Return(fmt.Errorf("Error: Metric test_metric cannot be incremented")).Once()
	err := taskMetric.IncrementMetric("test_metric")
	suite.Error(err)
	suite.mockH.AssertExpectations(suite.T())
}
