package zcache

import (
	"context"
	"fmt"
	"github.com/allegro/bigcache/v3"
	"github.com/go-redis/redis/v8"
	"github.com/zondax/golem/pkg/metrics"
	"time"
)

type CombinedCache interface {
	ZCache
}

type combinedCache struct {
	localCache         LocalCache
	remoteCache        RemoteCache
	ttl                time.Duration
	isRemoteBestEffort bool
	metricsServer      *metrics.TaskMetrics
	appName            string
}

func (c *combinedCache) Set(ctx context.Context, key string, value interface{}, _ time.Duration) error {
	if err := c.remoteCache.Set(ctx, key, value, c.ttl); err != nil && !c.isRemoteBestEffort {
		return err
	}

	// ttl is controlled by cache instantiation, so it does not matter here
	if err := c.localCache.Set(ctx, key, value, c.ttl); err != nil {
		return err
	}
	return nil
}

func (c *combinedCache) Get(ctx context.Context, key string, data interface{}) error {
	err := c.localCache.Get(ctx, key, data)
	if err != nil {
		if err := c.remoteCache.Get(ctx, key, data); err != nil {
			return err
		}

		// Refresh data TTL on both caches
		_ = c.localCache.Set(ctx, key, data, c.ttl)
		_ = c.remoteCache.Set(ctx, key, data, c.ttl)
	}

	return nil
}

func (c *combinedCache) Delete(ctx context.Context, key string) error {
	err2 := c.remoteCache.Delete(ctx, key)
	if err2 != nil && !c.isRemoteBestEffort {
		return err2
	}

	if err1 := c.localCache.Delete(ctx, key); err1 != nil {
		return err1
	}

	return nil
}

func (c *combinedCache) GetStats() ZCacheStats {
	localStats := c.localCache.(interface{}).(*bigcache.BigCache).Stats()
	remotePoolStats := c.remoteCache.(interface{}).(*redis.Client).PoolStats()
	return ZCacheStats{
		Local: &localStats,
		Remote: &RedisStats{
			Pool: remotePoolStats,
		},
	}
}

func (c *combinedCache) SetupAndMonitorCacheMetrics(appName string, metricsServer metrics.TaskMetrics, updateInterval time.Duration) []error {
	c.metricsServer = &metricsServer
	c.appName = appName

	errs := setupAndMonitorCacheMetrics(appName, metricsServer, c, updateInterval)
	errs = append(errs, c.registerInternalCacheMetrics()...)

	return errs
}

func (c *combinedCache) registerInternalCacheMetrics() []error {
	if c.metricsServer == nil {
		return []error{fmt.Errorf("metrics server not available")}
	}

	return []error{}
}
