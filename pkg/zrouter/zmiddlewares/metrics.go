package zmiddlewares

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/zondax/golem/pkg/metrics"
	"github.com/zondax/golem/pkg/metrics/collectors"
	"net/http"
	"strconv"
	"sync"
	"time"
)

const (
	activeConnectionsMetricType    = "active_connections"
	durationMillisecondsMetricType = "duration_milliseconds"
	responseSizeMetricType         = "response_size_bytes"
	totalRequestMetricType         = "total_requests"
	pathLabel                      = "path"
	methodLabel                    = "method"
	statusLabel                    = "status"
	TopNRequestsByJTIMetricName    = "topN_requests_by_jti"
)

func RegisterRequestMetrics(appName string, metricsServer metrics.TaskMetrics) []error {
	var errs []error

	register := func(name, help string, labels []string, handler metrics.MetricHandler) {
		if err := metricsServer.RegisterMetric(name, help, labels, handler); err != nil {
			errs = append(errs, err)
		}
	}

	totalRequestsMetricName := getMetricName(appName, "total_requests")
	responseSizeMetricName := getMetricName(appName, "response_size")
	durationMillisecondsMetricName := getMetricName(appName, "request_duration_ms")
	activeConnectionsMetricName := getMetricName(appName, "active_connections")
	register(totalRequestsMetricName, "Total number of HTTP requests made.", []string{"method", "path", "status"}, &collectors.Counter{})
	register(durationMillisecondsMetricName, "Duration of HTTP requests in milliseconds.", []string{"method", "path", "status"}, &collectors.Histogram{})
	register(responseSizeMetricName, "Size of HTTP response in bytes.", []string{"method", "path", "status"}, &collectors.Histogram{})
	register(activeConnectionsMetricName, "Number of active HTTP connections.", nil, &collectors.Gauge{})

	cacheHitsMetricName := getMetricName(appName, cacheHitsMetric)
	cacheMissesMetricName := getMetricName(appName, cacheMissesMetric)
	cacheSetsMetricName := getMetricName(appName, cacheSetsMetric)
	register(cacheHitsMetricName, "Number of cache hits.", []string{pathLabel}, &collectors.Counter{})
	register(cacheMissesMetricName, "Number of cache misses.", []string{pathLabel}, &collectors.Counter{})
	register(cacheSetsMetricName, "Number of responses added to the cache.", []string{pathLabel}, &collectors.Counter{})
	jwtRequestsMetricName := getMetricName(appName, TopNRequestsByJTIMetricName)
	register(jwtRequestsMetricName, "Number of requests made by JWT tokens per path.", []string{"jti", "path"}, &collectors.Gauge{})

	return errs
}

func RequestMetrics(appName string, metricsServer metrics.TaskMetrics) Middleware {
	var activeConnections int64
	var mu sync.Mutex

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			startTime := time.Now()
			activeConnectionsMetricName := getMetricName(appName, activeConnectionsMetricType)

			mu.Lock()
			activeConnections++
			_ = metricsServer.UpdateMetric(activeConnectionsMetricName, float64(activeConnections))
			mu.Unlock()

			mrw := &responseWriter{ResponseWriter: w}
			next.ServeHTTP(mrw, r)

			mu.Lock()
			activeConnections--
			_ = metricsServer.UpdateMetric(activeConnectionsMetricName, float64(activeConnections))
			mu.Unlock()

			duration := float64(time.Since(startTime).Milliseconds())
			path := chi.RouteContext(r.Context()).RoutePattern()

			responseStatus := mrw.status
			bytesWritten := mrw.written

			labels := []string{r.Method, path, strconv.Itoa(responseStatus)}

			durationMillisecondsMetricName := getMetricName(appName, durationMillisecondsMetricType)
			_ = metricsServer.UpdateMetric(durationMillisecondsMetricName, duration, labels...)

			responseSizeMetricName := getMetricName(appName, responseSizeMetricType)
			_ = metricsServer.UpdateMetric(responseSizeMetricName, float64(bytesWritten), labels...)

			totalRequestsMetricName := getMetricName(appName, totalRequestMetricType)
			_ = metricsServer.UpdateMetric(totalRequestsMetricName, 1, labels...)
		})
	}
}

func getMetricName(appName, metricType string) string {
	return fmt.Sprintf("zrouter_request_%s_%s", appName, metricType)
}
