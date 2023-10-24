package zmiddlewares

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/zondax/golem/pkg/metrics"
	"github.com/zondax/golem/pkg/metrics/collectors"
	"net/http"
	"strconv"
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
)

func RegisterRequestMetrics(appName string, metricsServer metrics.TaskMetrics) []error {
	var errs []error

	register := func(name, help string, labels []string, handler metrics.MetricHandler) {
		if err := metricsServer.RegisterMetric(name, help, labels, handler); err != nil {
			errs = append(errs, err)
		}
	}

	totalRequestsMetricName := getMetricName(appName, totalRequestMetricType)
	responseSizeMetricName := getMetricName(appName, responseSizeMetricType)
	durationMillisecondsMetricName := getMetricName(appName, durationMillisecondsMetricType)
	activeConnectionsMetricName := getMetricName(appName, activeConnectionsMetricType)
	register(totalRequestsMetricName, "Total number of HTTP requests made.", []string{methodLabel, pathLabel, statusLabel}, &collectors.Counter{})
	register(durationMillisecondsMetricName, "Duration of HTTP requests.", []string{methodLabel, pathLabel, statusLabel}, &collectors.Histogram{})
	register(responseSizeMetricName, "Size of HTTP response in bytes.", []string{methodLabel, pathLabel, statusLabel}, &collectors.Histogram{})
	register(activeConnectionsMetricName, "Number of active HTTP connections.", nil, &collectors.Gauge{})

	if len(errs) > 0 {
		return errs
	}

	return nil
}

func RequestMetrics(appName string, metricsServer metrics.TaskMetrics) Middleware {
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
