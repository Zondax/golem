package metrics

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"net/http"
	"time"
)

// StartMetricsServer starts a prometheus server.
// Data Url is at localhost:<port>/metrics/<endpoint>
// Normally you would use /metrics as endpoint and 9090 as port

type TaskMetrics struct {
	endpoint string
	port     string
}

func NewTaskMetrics(endpoint string, port string) *TaskMetrics {
	return &TaskMetrics{
		endpoint: endpoint,
		port:     port,
	}
}

func (t *TaskMetrics) Name() string {
	return "metrics"
}

func (t *TaskMetrics) Start() error {
	router := chi.NewRouter()

	zap.S().Infof("Metrics (prometheus) starting: %v", t.port)

	// Prometheus endpoint
	router.Get(t.endpoint, promhttp.Handler().(http.HandlerFunc))

	server := &http.Server{
		Addr:              fmt.Sprintf(":%s", t.port),
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	err := server.ListenAndServe()
	if err != nil {
		zap.S().Errorf("Prometheus server error: %v", err)
	} else {
		zap.S().Infof("Prometheus server serving at port %s", t.port)
	}

	return err
}

func (t *TaskMetrics) Stop() error {
	return nil
}
