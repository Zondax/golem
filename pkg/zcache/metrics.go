package zcache

import (
	"fmt"
	"github.com/zondax/golem/pkg/metrics"
	"github.com/zondax/golem/pkg/metrics/collectors"
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

func SetupAndMonitorCacheMetrics(appName string, metricsServer metrics.TaskMetrics, cache ZCache, updateInterval time.Duration) []error {
	if updateInterval <= 0 {
		updateInterval = defaultInterval
	}

	var errs []error
	register := func(name, help string, labels []string, handler metrics.MetricHandler) {
		if err := metricsServer.RegisterMetric(name, help, labels, handler); err != nil {
			errs = append(errs, err)
		}
	}

	register(getMetricName(appName, localCacheHitsMetricName), "Number of successfully found keys", nil, &collectors.Gauge{})
	register(getMetricName(appName, localCacheMissesMetricName), "Number of not found keys", nil, &collectors.Gauge{})
	register(getMetricName(appName, localCacheDelHitsMetricName), "Number of successfully deleted keys", nil, &collectors.Gauge{})
	register(getMetricName(appName, localCacheDelMissesMetricName), "Number of not deleted keys", nil, &collectors.Gauge{})
	register(getMetricName(appName, localCacheCollisionsMetricName), "Number of happened key-collisions", nil, &collectors.Gauge{})

	register(getMetricName(appName, remoteCachePoolHitsMetricName), "Number of times free connection was found in the pool", nil, &collectors.Gauge{})
	register(getMetricName(appName, remoteCachePoolMissesMetricName), "Number of times free connection was NOT found in the pool", nil, &collectors.Gauge{})
	register(getMetricName(appName, remoteCachePoolTimeoutsMetricName), "Number of times a wait timeout occurred", nil, &collectors.Gauge{})
	register(getMetricName(appName, remoteCachePoolTotalConnsMetricName), "Number of total connections in the pool", nil, &collectors.Gauge{})
	register(getMetricName(appName, remoteCachePoolIdleConnsMetricName), "Number of idle connections in the pool", nil, &collectors.Gauge{})
	register(getMetricName(appName, remoteCachePoolStaleConnsMetricName), "Number of stale connections removed from the pool", nil, &collectors.Gauge{})

	if len(errs) > 0 {
		return errs
	}

	go func() {
		ticker := time.NewTicker(updateInterval)
		for range ticker.C {
			stats := cache.GetStats()

			if stats.Bigcache != nil {
				_ = metricsServer.UpdateMetric(getMetricName(appName, localCacheHitsMetricName), float64(stats.Bigcache.Hits))
				_ = metricsServer.UpdateMetric(getMetricName(appName, localCacheMissesMetricName), float64(stats.Bigcache.Misses))
				_ = metricsServer.UpdateMetric(getMetricName(appName, localCacheDelHitsMetricName), float64(stats.Bigcache.DelHits))
				_ = metricsServer.UpdateMetric(getMetricName(appName, localCacheDelMissesMetricName), float64(stats.Bigcache.DelMisses))
				_ = metricsServer.UpdateMetric(getMetricName(appName, localCacheCollisionsMetricName), float64(stats.Bigcache.Collisions))
			}

			if stats.Redis != nil {
				_ = metricsServer.UpdateMetric(getMetricName(appName, remoteCachePoolHitsMetricName), float64(stats.Redis.Pool.Hits))
				_ = metricsServer.UpdateMetric(getMetricName(appName, remoteCachePoolMissesMetricName), float64(stats.Redis.Pool.Misses))
				_ = metricsServer.UpdateMetric(getMetricName(appName, remoteCachePoolTimeoutsMetricName), float64(stats.Redis.Pool.Timeouts))
				_ = metricsServer.UpdateMetric(getMetricName(appName, remoteCachePoolTotalConnsMetricName), float64(stats.Redis.Pool.TotalConns))
				_ = metricsServer.UpdateMetric(getMetricName(appName, remoteCachePoolIdleConnsMetricName), float64(stats.Redis.Pool.IdleConns))
				_ = metricsServer.UpdateMetric(getMetricName(appName, remoteCachePoolStaleConnsMetricName), float64(stats.Redis.Pool.StaleConns))
			}
		}
	}()

	return nil
}

func getMetricName(appName, metricType string) string {
	return fmt.Sprintf("zcache_%s_%s", appName, metricType)
}
