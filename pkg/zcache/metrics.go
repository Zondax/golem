package zcache

import (
	"github.com/zondax/golem/pkg/logger"
	"github.com/zondax/golem/pkg/metrics"
	"github.com/zondax/golem/pkg/metrics/collectors"
	"go.uber.org/zap"
	"time"
)

const (
	defaultInterval                = time.Minute
	localCacheHitsMetricName       = "local_cache_hits"
	localCacheMissesMetricName     = "local_cache_misses"
	localCacheDelHitsMetricName    = "local_cache_del_hits"
	localCacheDelMissesMetricName  = "local_cache_del_misses"
	localCacheCollisionsMetricName = "local_cache_collisions"

	remoteCachePoolHitsMetricName       = "remote_cache_pool_hits"
	remoteCachePoolMissesMetricName     = "remote_cache_pool_misses"
	remoteCachePoolTimeoutsMetricName   = "remote_cache_pool_timeouts"
	remoteCachePoolTotalConnsMetricName = "remote_cache_pool_total_conns"
	remoteCachePoolIdleConnsMetricName  = "remote_cache_pool_idle_conns"
	remoteCachePoolStaleConnsMetricName = "remote_cache_pool_stale_conns"
)

func setupAndMonitorCacheMetrics(metricsServer metrics.TaskMetrics, cache ZCache, logger *logger.Logger, updateInterval time.Duration) []error {
	if updateInterval <= 0 {
		updateInterval = defaultInterval
	}

	if err := metricsServer.RegisterMetric(localCacheHitsMetricName, "Number of successfully found keys", []string{}, &collectors.Gauge{}); err != nil {
		logger.Errorf("Failed to register cache stats metrics, err: %s", zap.Error(err))
	}
	if err := metricsServer.RegisterMetric(localCacheMissesMetricName, "Number of not found keys", nil, &collectors.Gauge{}); err != nil {
		logger.Errorf("Failed to register cache stats metrics, err: %s", zap.Error(err))
	}
	if err := metricsServer.RegisterMetric(localCacheDelHitsMetricName, "Number of successfully deleted keys", nil, &collectors.Gauge{}); err != nil {
		logger.Errorf("Failed to register cache stats metrics, err: %s", zap.Error(err))
	}
	if err := metricsServer.RegisterMetric(localCacheDelMissesMetricName, "Number of not deleted keys", nil, &collectors.Gauge{}); err != nil {
		logger.Errorf("Failed to register cache stats metrics, err: %s", zap.Error(err))
	}
	if err := metricsServer.RegisterMetric(localCacheCollisionsMetricName, "Number of happened key-collisions", nil, &collectors.Gauge{}); err != nil {
		logger.Errorf("Failed to register cache stats metrics, err: %s", zap.Error(err))
	}

	if err := metricsServer.RegisterMetric(remoteCachePoolHitsMetricName, "Number of times free connection was found in the pool", nil, &collectors.Gauge{}); err != nil {
		logger.Errorf("Failed to register cache stats metrics, err: %s", zap.Error(err))
	}
	if err := metricsServer.RegisterMetric(remoteCachePoolMissesMetricName, "Number of times free connection was NOT found in the pool", nil, &collectors.Gauge{}); err != nil {
		logger.Errorf("Failed to register cache stats metrics, err: %s", zap.Error(err))
	}
	if err := metricsServer.RegisterMetric(remoteCachePoolTimeoutsMetricName, "Number of times a wait timeout occurred", nil, &collectors.Gauge{}); err != nil {
		logger.Errorf("Failed to register cache stats metrics, err: %s", zap.Error(err))
	}
	if err := metricsServer.RegisterMetric(remoteCachePoolTotalConnsMetricName, "Number of total connections in the pool", nil, &collectors.Gauge{}); err != nil {
		logger.Errorf("Failed to register cache stats metrics, err: %s", zap.Error(err))
	}
	if err := metricsServer.RegisterMetric(remoteCachePoolIdleConnsMetricName, "Number of idle connections in the pool", nil, &collectors.Gauge{}); err != nil {
		logger.Errorf("Failed to register cache stats metrics, err: %s", zap.Error(err))
	}
	if err := metricsServer.RegisterMetric(remoteCachePoolStaleConnsMetricName, "Number of stale connections removed from the pool", nil, &collectors.Gauge{}); err != nil {
		logger.Errorf("Failed to register cache stats metrics, err: %s", zap.Error(err))
	}

	go func() {
		ticker := time.NewTicker(updateInterval)
		for range ticker.C {
			stats := cache.GetStats()

			if stats.Local != nil {
				_ = metricsServer.UpdateMetric(localCacheHitsMetricName, float64(stats.Local.Hits))
				_ = metricsServer.UpdateMetric(localCacheMissesMetricName, float64(stats.Local.Misses))
				_ = metricsServer.UpdateMetric(localCacheDelHitsMetricName, float64(stats.Local.DelHits))
				_ = metricsServer.UpdateMetric(localCacheDelMissesMetricName, float64(stats.Local.DelMisses))
				_ = metricsServer.UpdateMetric(localCacheCollisionsMetricName, float64(stats.Local.Collisions))
			}

			if stats.Remote != nil {
				_ = metricsServer.UpdateMetric(remoteCachePoolHitsMetricName, float64(stats.Remote.Pool.Hits))
				_ = metricsServer.UpdateMetric(remoteCachePoolMissesMetricName, float64(stats.Remote.Pool.Misses))
				_ = metricsServer.UpdateMetric(remoteCachePoolTimeoutsMetricName, float64(stats.Remote.Pool.Timeouts))
				_ = metricsServer.UpdateMetric(remoteCachePoolTotalConnsMetricName, float64(stats.Remote.Pool.TotalConns))
				_ = metricsServer.UpdateMetric(remoteCachePoolIdleConnsMetricName, float64(stats.Remote.Pool.IdleConns))
				_ = metricsServer.UpdateMetric(remoteCachePoolStaleConnsMetricName, float64(stats.Remote.Pool.StaleConns))
			}
		}
	}()

	return nil
}
