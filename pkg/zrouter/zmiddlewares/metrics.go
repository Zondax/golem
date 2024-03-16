package zmiddlewares

import (
	"github.com/zondax/golem/pkg/logger"
	"github.com/zondax/golem/pkg/metrics"
	"github.com/zondax/golem/pkg/metrics/collectors"
	"net/http"
	"strconv"
	"sync"
	"time"
)

const (
	activeConnectionsMetricName    = "active_connections"
	durationMillisecondsMetricName = "request_duration_ms"
	responseSizeMetricName         = "response_size"
	totalRequestsMetricName        = "total_requests"
	pathLabel                      = "path"
	methodLabel                    = "method"
	statusLabel                    = "status"
	TopNRequestsByJTIMetricName    = "topN_requests_by_jti"
)

func RegisterRequestMetrics(metricsServer metrics.TaskMetrics) []error {
	var errs []error

	register := func(name, help string, labels []string, handler metrics.MetricHandler) {
		if err := metricsServer.RegisterMetric(name, help, labels, handler); err != nil {
			errs = append(errs, err)
		}
	}

	register(totalRequestsMetricName, "Total number of HTTP requests made.", []string{"method", "path", "status"}, &collectors.Counter{})
	register(durationMillisecondsMetricName, "Duration of HTTP requests in milliseconds.", []string{"method", "path", "status"}, &collectors.Gauge{})
	register(responseSizeMetricName, "Size of HTTP response in bytes.", []string{"method", "path", "status"}, &collectors.Histogram{})
	register(activeConnectionsMetricName, "Number of active HTTP connections.", []string{"method", "path"}, &collectors.Gauge{})

	cacheHitsMetricName := cacheHitsMetric
	cacheMissesMetricName := cacheMissesMetric
	cacheSetsMetricName := cacheSetsMetric
	register(cacheHitsMetricName, "Number of cache hits.", []string{pathLabel}, &collectors.Counter{})
	register(cacheMissesMetricName, "Number of cache misses.", []string{pathLabel}, &collectors.Counter{})
	register(cacheSetsMetricName, "Number of responses added to the cache.", []string{pathLabel}, &collectors.Counter{})
	register(TopNRequestsByJTIMetricName, "Number of requests made by JWT tokens per path.", []string{"jti", "path"}, &collectors.Gauge{})

	return errs
}

func RequestMetrics(metricsServer metrics.TaskMetrics) Middleware {
	var activeConnections int64
	var mu sync.Mutex

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			path := GetRoutePattern(r)
			startTime := time.Now()

			mu.Lock()
			activeConnections++
			if err := metricsServer.UpdateMetric(activeConnectionsMetricName, float64(activeConnections), r.Method, path); err != nil {
				logger.GetLoggerFromContext(r.Context()).Errorf("error updating active connections metric: %v", err.Error())
			}
			mu.Unlock()

			mrw := &responseWriter{ResponseWriter: w}
			next.ServeHTTP(mrw, r)

			mu.Lock()
			activeConnections--
			if err := metricsServer.UpdateMetric(activeConnectionsMetricName, float64(activeConnections), r.Method, path); err != nil {
				logger.GetLoggerFromContext(r.Context()).Errorf("error updating active connections metric: %v", err.Error())
			}
			mu.Unlock()

			duration := float64(time.Since(startTime).Milliseconds())

			responseStatus := mrw.status
			bytesWritten := mrw.written

			labels := []string{r.Method, path, strconv.Itoa(responseStatus)}

			if err := metricsServer.UpdateMetric(durationMillisecondsMetricName, duration, labels...); err != nil {
				logger.GetLoggerFromContext(r.Context()).Errorf("error updating request duration metric: %v", err.Error())
			}
			if err := metricsServer.UpdateMetric(responseSizeMetricName, float64(bytesWritten), labels...); err != nil {
				logger.GetLoggerFromContext(r.Context()).Errorf("error updating response size metric: %v", err.Error())
			}
			if err := metricsServer.UpdateMetric(totalRequestsMetricName, 1, labels...); err != nil {
				logger.GetLoggerFromContext(r.Context()).Errorf("error updating total requests metric: %v", err.Error())
			}
		})
	}
}
