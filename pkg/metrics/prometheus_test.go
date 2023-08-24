package metrics

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
	"time"
)

func TestName(t *testing.T) {
	metrics := NewTaskMetrics("/metrics", "9090")
	assert.Equal(t, "metrics", metrics.Name())
}

func TestStartServer(t *testing.T) {
	metrics := NewTaskMetrics("/metrics", "9091")

	go func() {
		err := metrics.StartServer()
		assert.Nil(t, err)
	}()

	time.Sleep(1 * time.Second)

	resp, err := http.Get("http://localhost:9091/metrics")
	assert.Nil(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
