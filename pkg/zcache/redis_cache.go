package zcache

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/zondax/golem/pkg/metrics"
	"go.uber.org/zap"
	"time"

	"github.com/go-redis/redis/v8"
)

type RedisStats struct {
	Pool *redis.PoolStats
}

type RemoteCache interface {
	ZCache
	Incr(ctx context.Context, key string) (int64, error)
	Decr(ctx context.Context, key string) (int64, error)
	LPush(ctx context.Context, key string, values ...interface{}) (int64, error)
	RPush(ctx context.Context, key string, values ...interface{}) (int64, error)
	SMembers(ctx context.Context, key string) ([]string, error)
	SAdd(ctx context.Context, key string, members ...interface{}) (int64, error)
	HSet(ctx context.Context, key string, values ...interface{}) (int64, error)
	HGet(ctx context.Context, key, field string) (string, error)
	FlushAll(ctx context.Context) error
	Exists(ctx context.Context, keys ...string) (int64, error)
}

type redisCache struct {
	client        *redis.Client
	prefix        string
	logger        *zap.Logger
	metricsServer *metrics.TaskMetrics
	appName       string
}

func (c *redisCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	realKey := getKeyWithPrefix(c.prefix, key)

	c.logger.Sugar().Debugf("set key on redis cache, fullKey: [%s], value: [%v]", realKey, value)

	val, err := json.Marshal(value)
	if err != nil {
		return err
	}

	err = c.client.Set(ctx, realKey, val, ttl).Err()
	if err != nil {
		c.logger.Sugar().Errorf("error setting new key on redis cache, fullKey: [%s], err: [%s]", realKey, err)
	}

	return err
}

func (c *redisCache) Get(ctx context.Context, key string, data interface{}) error {
	realKey := getKeyWithPrefix(c.prefix, key)

	c.logger.Sugar().Debugf("get key on redis cache, fullKey: [%s]", realKey)

	val, err := c.client.Get(ctx, realKey).Result()
	if err != nil {
		if c.IsNotFoundError(err) {
			c.logger.Sugar().Debugf("key not found on redis cache, fullKey: [%s]", realKey)
		} else {
			c.logger.Sugar().Errorf("error getting key on redis cache, fullKey: [%s], err: [%s]", realKey, err)
		}
		return err
	}
	return json.Unmarshal([]byte(val), &data)
}

func (c *redisCache) Delete(ctx context.Context, key string) error {
	realKey := getKeyWithPrefix(c.prefix, key)
	c.logger.Sugar().Debugf("delete key on redis cache, fullKey: [%s]", realKey)

	return c.client.Del(ctx, realKey).Err()
}

func (c *redisCache) Exists(ctx context.Context, keys ...string) (int64, error) {
	realKeys := getKeysWithPrefix(c.prefix, keys)

	c.logger.Sugar().Debugf("exists keys on redis cache, fullKeys: [%s]", realKeys)

	return c.client.Exists(ctx, realKeys...).Result()
}

func (c *redisCache) Incr(ctx context.Context, key string) (int64, error) {
	realKey := getKeyWithPrefix(c.prefix, key)

	c.logger.Sugar().Debugf("increment on key on redis cache, fullKey: [%s]", realKey)
	return c.client.Incr(ctx, realKey).Result()
}

func (c *redisCache) Decr(ctx context.Context, key string) (int64, error) {
	realKey := getKeyWithPrefix(c.prefix, key)

	c.logger.Sugar().Debugf("decrement on key on redis cache, fullKey: [%s]", realKey)
	return c.client.Decr(ctx, realKey).Result()
}

func (c *redisCache) FlushAll(ctx context.Context) error {
	c.logger.Sugar().Debugf("flush all on redis cache, fullKey")
	return c.client.FlushAll(ctx).Err()
}

func (c *redisCache) LPush(ctx context.Context, key string, values ...interface{}) (int64, error) {
	realKey := getKeyWithPrefix(c.prefix, key)
	c.logger.Sugar().Debugf("lpush on redis cache, fullKey: [%s]", realKey)
	return c.client.LPush(ctx, realKey, values...).Result()
}

func (c *redisCache) RPush(ctx context.Context, key string, values ...interface{}) (int64, error) {
	realKey := getKeyWithPrefix(c.prefix, key)
	c.logger.Sugar().Debugf("rpush on redis cache, fullKey: [%s]", realKey)
	return c.client.RPush(ctx, realKey, values...).Result()
}

func (c *redisCache) SMembers(ctx context.Context, key string) ([]string, error) {
	realKey := getKeyWithPrefix(c.prefix, key)
	c.logger.Sugar().Debugf("smemebers on redis cache, fullKey: [%s]", realKey)
	return c.client.SMembers(ctx, realKey).Result()
}

func (c *redisCache) SAdd(ctx context.Context, key string, members ...interface{}) (int64, error) {
	realKey := getKeyWithPrefix(c.prefix, key)
	c.logger.Sugar().Debugf("sadd on redis cache, fullKey: [%s]", realKey)
	return c.client.SAdd(ctx, realKey, members...).Result()
}

func (c *redisCache) HSet(ctx context.Context, key string, values ...interface{}) (int64, error) {
	realKey := getKeyWithPrefix(c.prefix, key)
	c.logger.Sugar().Debugf("hset on redis cache, fullKey: [%s]", realKey)
	return c.client.HSet(ctx, realKey, values...).Result()
}

func (c *redisCache) HGet(ctx context.Context, key, field string) (string, error) {
	realKey := getKeyWithPrefix(c.prefix, key)
	c.logger.Sugar().Debugf("hget on redis cache, fullKey: [%s]", realKey)
	return c.client.HGet(ctx, realKey, field).Result()
}

func (c *redisCache) GetStats() ZCacheStats {
	poolStats := c.client.PoolStats()
	c.logger.Sugar().Debugf("redis cache pool stats: [%v]", poolStats)

	ctx := context.Background()
	stats, err := c.client.Info(ctx).Result()
	if err != nil {
		c.logger.Sugar().Errorf("error on redis cache stats: [%v]", stats)
	}

	c.logger.Sugar().Debugf("redis cache stats: \n %s", stats)
	// ctx := context.Background()
	// stats, _ := c.client.Info(ctx).Result()

	return ZCacheStats{
		Remote: &RedisStats{
			Pool: poolStats,
		},
	}
}

func (c *redisCache) IsNotFoundError(err error) bool {
	return err.Error() == "redis: nil"
}

func (c *redisCache) SetupAndMonitorMetrics(appName string, metricsServer metrics.TaskMetrics, updateInterval time.Duration) []error {
	c.metricsServer = &metricsServer
	c.appName = appName

	errs := setupAndMonitorCacheMetrics(appName, metricsServer, c, updateInterval)
	errs = append(errs, c.registerInternalCacheMetrics()...)

	return errs
}

func (c *redisCache) registerInternalCacheMetrics() []error {
	if c.metricsServer == nil {
		return []error{fmt.Errorf("metrics server not available")}
	}

	return []error{}
}
