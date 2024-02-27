package zcache

import (
	"context"
	"fmt"
	"github.com/allegro/bigcache/v3"
	"github.com/go-redis/redis/v8"
	"github.com/zondax/golem/pkg/metrics"
	"go.uber.org/zap"
	"time"
)

type CombinedCache interface {
	ZCache
}

type combinedCache struct {
	localCache         LocalCache
	remoteCache        RemoteCache
	logger             *zap.Logger
	isRemoteBestEffort bool
	metricsServer      *metrics.TaskMetrics
	appName            string
}

func (c *combinedCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	c.logger.Sugar().Debugf("set key on combined cache, key: [%s]", key)

	if err := c.remoteCache.Set(ctx, key, value, ttl); err != nil {
		c.logger.Sugar().Errorf("error setting key on combined/remote cache, key: [%s], err: %s", key, err)
		if !c.isRemoteBestEffort {
			c.logger.Sugar().Debugf("emitting error as remote best effort is false, key: [%s]", key)
			return err
		}
	}

	if err := c.localCache.Set(ctx, key, value, ttl); err != nil {
		c.logger.Sugar().Errorf("error setting key on combined/local cache, key: [%s], err: %s", key, err)
		return err
	}
	return nil
}

func (c *combinedCache) Get(ctx context.Context, key string, data interface{}) error {
	c.logger.Sugar().Debugf("get key on combined cache, key: [%s]", key)

	err := c.localCache.Get(ctx, key, data)
	if err != nil {
		if c.localCache.IsNotFoundError(err) {
			c.logger.Sugar().Debugf("key not found on combined/local cache, key: [%s]", key)
		} else {
			c.logger.Sugar().Debugf("error getting key on combined/local cache, key: [%s], err: %s", key, err)
		}

		if err := c.remoteCache.Get(ctx, key, data); err != nil {
			if c.remoteCache.IsNotFoundError(err) {
				c.logger.Sugar().Debugf("key not found on combined/remote cache, key: [%s]", key)
			} else {
				c.logger.Sugar().Debugf("error getting key on combined/remote cache, key: [%s], err: %s", key, err)
			}

			return err
		}

		c.logger.Sugar().Debugf("set value found on remote cache in the local cache, key: [%s]", key)
		ttl, ttlErr := c.remoteCache.TTL(ctx, key)
		if ttlErr != nil {
			c.logger.Sugar().Errorf("error getting TTL for key [%s] from remote cache, err: %s", key, ttlErr)
		}

		// Refresh data TTL on both caches
		if ttl != 0 {
			_ = c.localCache.Set(ctx, key, data, ttl)
		}
	}

	return nil
}

func (c *combinedCache) Delete(ctx context.Context, key string) error {
	c.logger.Sugar().Debugf("delete key on combined cache, key: [%s]", key)
	err2 := c.remoteCache.Delete(ctx, key)
	if err2 != nil {
		c.logger.Sugar().Errorf("error deleting key on combined/remote cache, key: [%s], err: %s", key, err2)
		if !c.isRemoteBestEffort {
			c.logger.Sugar().Debugf("emitting error as remote best effort is false, key: [%s]")
			return err2
		}
	}

	if err1 := c.localCache.Delete(ctx, key); err1 != nil {
		c.logger.Sugar().Errorf("error deleting key on combined/local cache, key: [%s], err: %s", key, err1)
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
