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
func StartMetricsServer(endpoint string, port string) chan error {
	router := chi.NewRouter()

	zap.S().Infof("Metrics (prometheus) starting: %v", port)

	// Prometheus endpoint
	router.Get(endpoint, promhttp.Handler().(http.HandlerFunc))
	errChan := make(chan error)

	go func() {
		server := &http.Server{
			Addr:              fmt.Sprintf(":%s", port),
			Handler:           router,
			ReadHeaderTimeout: 5 * time.Second,
		}

		err := server.ListenAndServe()
		if err != nil {
			zap.S().Errorf("Prometheus server error: %v", err)
		} else {
			zap.S().Infof("Prometheus server serving at port %s", port)
		}
		errChan <- err
	}()

	return errChan
}
