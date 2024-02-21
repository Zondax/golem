package zdb

import (
	"context"
	"fmt"
	"github.com/zondax/golem/pkg/logger"
	"github.com/zondax/golem/pkg/metrics"
	"github.com/zondax/golem/pkg/metrics/collectors"
	"time"
)

var waitDurationBuckets = []float64{50, 100, 250, 500, 1000, 2500, 5000, 10000, 20000, 50000, 100000}

const (
	defaultInterval                = time.Minute
	dbOpenConnectionsMetricName    = "db_open_connections"
	dbIdleConnectionsMetricName    = "db_idle_connections"
	dbMaxOpenConnectionsMetricName = "db_max_open_connections"
	dbWaitDurationMetricName       = "db_wait_duration"
	dbInUseConnectionsMetricName   = "db_in_use_connections"
)

func SetupAndMonitorDBMetrics(appName string, metricsServer metrics.TaskMetrics, db ZDatabase, updateInterval time.Duration) []error {
	if updateInterval <= 0 {
		updateInterval = defaultInterval
	}

	var errs []error
	register := func(name, help string, labels []string, handler metrics.MetricHandler) {
		if err := metricsServer.RegisterMetric(name, help, labels, handler); err != nil {
			errs = append(errs, err)
		}
	}

	register(getMetricName(appName, dbOpenConnectionsMetricName), "Number of open database connections.", nil, &collectors.Gauge{})
	register(getMetricName(appName, dbIdleConnectionsMetricName), "Number of idle database connections in the pool.", nil, &collectors.Gauge{})
	register(getMetricName(appName, dbMaxOpenConnectionsMetricName), "Maximum number of open database connections.", nil, &collectors.Gauge{})
	register(getMetricName(appName, dbWaitDurationMetricName), "Total time waited for new database connections.", nil, &collectors.Histogram{Buckets: waitDurationBuckets})
	register(getMetricName(appName, dbInUseConnectionsMetricName), "Number of database connections currently in use.", nil, &collectors.Gauge{})

	if len(errs) > 0 {
		return errs
	}

	go func() {
		ticker := time.NewTicker(updateInterval)
		for range ticker.C {
			stats, err := db.GetDBStats()
			if err != nil {
				logger.Log().Errorf(context.Background(), "Error while getting db stats: %v", err)
				continue
			}

			_ = metricsServer.UpdateMetric(getMetricName(appName, dbOpenConnectionsMetricName), float64(stats.OpenConnections))
			_ = metricsServer.UpdateMetric(getMetricName(appName, dbIdleConnectionsMetricName), float64(stats.Idle))
			_ = metricsServer.UpdateMetric(getMetricName(appName, dbMaxOpenConnectionsMetricName), float64(stats.MaxOpenConnections))
			_ = metricsServer.UpdateMetric(getMetricName(appName, dbWaitDurationMetricName), float64(stats.WaitDuration.Milliseconds()))
			_ = metricsServer.UpdateMetric(getMetricName(appName, dbInUseConnectionsMetricName), float64(stats.InUse))
		}
	}()

	return nil
}

func getMetricName(appName, metricType string) string {
	return fmt.Sprintf("zdatabase_%s_%s", appName, metricType)
}
