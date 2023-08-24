package collectors

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCounterUpdate(t *testing.T) {
	counter := prometheus.NewCounter(prometheus.CounterOpts{Name: "test_counter"})
	c := &Counter{}

	err := c.Update(counter, 10)
	assert.NoError(t, err)
	assert.Equal(t, float64(10), testutil.ToFloat64(counter))
}
