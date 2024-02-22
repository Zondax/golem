package zcache

import (
	"context"
	"fmt"
	"github.com/allegro/bigcache/v3"
	"github.com/go-redis/redis/v8"
	"github.com/zondax/golem/pkg/logger"
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

func (c *combinedCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	logger.Log(ctx).Debugf("set key on combined cache, key: [%s]", key)

	if err := c.remoteCache.Set(ctx, key, value, c.ttl); err != nil {
		logger.Log(ctx).Errorf("error setting key on combined/remote cache, key: [%s], err: %s", key, err)
		if !c.isRemoteBestEffort {
			logger.Log(ctx).Debugf("emitting error as remote best effort is false, key: [%s]", key)
			return err
		}
	}

	if err := c.localCache.Set(ctx, key, value, ttl); err != nil {
		logger.Log(ctx).Errorf("error setting key on combined/local cache, key: [%s], err: %s", key, err)
		return err
	}
	return nil
}

func (c *combinedCache) Get(ctx context.Context, key string, data interface{}) error {
	logger.Log(ctx).Debugf("get key on combined cache, key: [%s]", key)

	err := c.localCache.Get(ctx, key, data)
	if err != nil {
		if c.localCache.IsNotFoundError(err) {
			logger.Log(ctx).Debugf("key not found on combined/local cache, key: [%s]", key)
		} else {
			logger.Log(ctx).Debugf("error getting key on combined/local cache, key: [%s], err: %s", key, err)
		}

		if err := c.remoteCache.Get(ctx, key, data); err != nil {
			if c.remoteCache.IsNotFoundError(err) {
				logger.Log(ctx).Debugf("key not found on combined/remote cache, key: [%s]", key)
			} else {
				logger.Log(ctx).Debugf("error getting key on combined/remote cache, key: [%s], err: %s", key, err)
			}

			return err
		}

		logger.Log(ctx).Debugf("set value found on remote cache in the local cache, key: [%s]", key)

		// Refresh data TTL on both caches
		_ = c.localCache.Set(ctx, key, data, c.ttl)
		_ = c.remoteCache.Set(ctx, key, data, c.ttl)
	}

	return nil
}

func (c *combinedCache) Delete(ctx context.Context, key string) error {
	logger.Log(ctx).Debugf("delete key on combined cache, key: [%s]", key)
	err2 := c.remoteCache.Delete(ctx, key)
	if err2 != nil {
		logger.Log(ctx).Errorf("error deleting key on combined/remote cache, key: [%s], err: %s", key, err2)
		if !c.isRemoteBestEffort {
			logger.Log(ctx).Debugf("emitting error as remote best effort is false, key: [%s]")
			return err2
		}
	}

	if err1 := c.localCache.Delete(ctx, key); err1 != nil {
		logger.Log(ctx).Errorf("error deleting key on combined/local cache, key: [%s], err: %s", key, err1)
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

func (c *combinedCache) IsNotFoundError(err error) bool {
	return c.remoteCache.IsNotFoundError(err) || c.localCache.IsNotFoundError(err)
}

func (c *combinedCache) SetupAndMonitorMetrics(appName string, metricsServer metrics.TaskMetrics, updateInterval time.Duration) []error {
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
