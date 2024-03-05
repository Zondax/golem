package collectors

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGaugeUpdate(t *testing.T) {
	gauge := prometheus.NewGauge(prometheus.GaugeOpts{Name: "test_gauge"})
	g := &Gauge{}

	err := g.Update(gauge, 15)
	assert.NoError(t, err)
	assert.Equal(t, float64(15), testutil.ToFloat64(gauge))
}

func TestGaugeInc(t *testing.T) {
	gauge := prometheus.NewGauge(prometheus.GaugeOpts{Name: "test_gauge"})
	g := &Gauge{}

	err := g.Inc(gauge)
	assert.NoError(t, err)
	assert.Equal(t, float64(1), testutil.ToFloat64(gauge))
}

func TestGaugeDec(t *testing.T) {
	gauge := prometheus.NewGauge(prometheus.GaugeOpts{Name: "test_gauge"})
	g := &Gauge{}

	err := g.Dec(gauge)
	assert.NoError(t, err)
	assert.Equal(t, float64(-1), testutil.ToFloat64(gauge))
}

func TestCounterInc(t *testing.T) {
	counter := prometheus.NewCounter(prometheus.CounterOpts{Name: "test_counter"})
	c := &Counter{}

	err := c.Inc(counter)
	assert.NoError(t, err)
	assert.Equal(t, float64(1), testutil.ToFloat64(counter))
}
