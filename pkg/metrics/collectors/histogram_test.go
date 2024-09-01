package collectors

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestHistogramUpdate(t *testing.T) {
	histogram := prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "test_histogram",
		Buckets: prometheus.DefBuckets,
	})
	h := &Histogram{}

	err := h.Update(histogram, 5)
	assert.NoError(t, err)

	metricFamily, err := prometheus.DefaultGatherer.Gather()
	assert.NoError(t, err)

	for _, m := range metricFamily {
		if *m.Name == "test_histogram" {
			assert.Equal(t, 1, *m.Metric[0].Histogram.SampleCount)
			assert.Equal(t, 5.0, *m.Metric[0].Histogram.SampleSum)
		}
	}
}
