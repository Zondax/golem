package metrics

import (
	"fmt"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/zondax/golem/pkg/logger"
)

func TestName(t *testing.T) {
	metrics := NewTaskMetrics("/metrics", "9090", "test")
	assert.Equal(t, "metrics", metrics.Name())
}

func TestStart(t *testing.T) {
	logger.InitLogger(logger.Config{})

	// Get an available port dynamically
	listener, err := net.Listen("tcp", ":0")
	assert.Nil(t, err)
	port := listener.Addr().(*net.TCPAddr).Port
	listener.Close()

	portStr := fmt.Sprintf("%d", port)
	metrics := NewTaskMetrics("/metrics", portStr, "test")

	go func() {
		err := metrics.Start()
		assert.Nil(t, err)
	}()

	time.Sleep(1 * time.Second)

	resp, err := http.Get(fmt.Sprintf("http://localhost:%s/metrics", portStr))
	assert.Nil(t, err)

	defer resp.Body.Close()

	assert.NotNil(t, resp)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
