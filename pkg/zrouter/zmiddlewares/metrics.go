package zmiddlewares

import (
	"fmt"
	"github.com/zondax/golem/pkg/metrics"
	"net/http"
	"strconv"
	"time"
)

const (
	activeConnectionsMetricType    = "active_connections"
	durationMillisecondsMetricType = "duration_milliseconds"
	responseSizeMetricType         = "response_size_bytes"
	totalRequestMetricType         = "total_requests"
)

func routerMetrics(appName string, metricsServer metrics.TaskMetrics) Middleware {
	var activeConnections int64

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			startTime := time.Now()
			activeConnectionsMetricName := getMetricName(appName, activeConnectionsMetricType)
			activeConnections++
			_ = metricsServer.UpdateMetric(activeConnectionsMetricName, float64(activeConnections))

			mrw := &metricsResponseWriter{ResponseWriter: w}
			next.ServeHTTP(mrw, r)

			activeConnections--
			_ = metricsServer.UpdateMetric(activeConnectionsMetricName, float64(activeConnections))

			duration := float64(time.Since(startTime).Milliseconds())
			path := r.URL.Path
			responseStatus := mrw.status
			bytesWritten := mrw.written

			labels := []string{"endpoint", path, "method", r.Method, "status", strconv.Itoa(responseStatus)}

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
