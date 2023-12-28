package zcache

import (
	"context"
	"encoding/json"
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
	client *redis.Client
}

func (c *redisCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	val, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, key, val, ttl).Err()
}

func (c *redisCache) Get(ctx context.Context, key string, data interface{}) error {
	val, err := c.client.Get(ctx, key).Result()
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(val), &data)
}

func (c *redisCache) Delete(ctx context.Context, key string) error {
	return c.client.Del(ctx, key).Err()
}

func (c *redisCache) Exists(ctx context.Context, keys ...string) (int64, error) {
	return c.client.Exists(ctx, keys...).Result()
}

func (c *redisCache) Incr(ctx context.Context, key string) (int64, error) {
	return c.client.Incr(ctx, key).Result()
}

func (c *redisCache) Decr(ctx context.Context, key string) (int64, error) {
	return c.client.Decr(ctx, key).Result()
}

func (c *redisCache) FlushAll(ctx context.Context) error {
	return c.client.FlushAll(ctx).Err()
}

func (c *redisCache) LPush(ctx context.Context, key string, values ...interface{}) (int64, error) {
	return c.client.LPush(ctx, key, values...).Result()
}

func (c *redisCache) RPush(ctx context.Context, key string, values ...interface{}) (int64, error) {
	return c.client.RPush(ctx, key, values...).Result()
}

func (c *redisCache) SMembers(ctx context.Context, key string) ([]string, error) {
	return c.client.SMembers(ctx, key).Result()
}

func (c *redisCache) SAdd(ctx context.Context, key string, members ...interface{}) (int64, error) {
	return c.client.SAdd(ctx, key, members...).Result()
}

func (c *redisCache) HSet(ctx context.Context, key string, values ...interface{}) (int64, error) {
	return c.client.HSet(ctx, key, values...).Result()
}

func (c *redisCache) HGet(ctx context.Context, key, field string) (string, error) {
	return c.client.HGet(ctx, key, field).Result()
}

func (c *redisCache) GetStats() ZCacheStats {
	poolStats := c.client.PoolStats()
	// ctx := context.Background()
	// stats, _ := c.client.Info(ctx).Result()

	return ZCacheStats{
		Redis: &RedisStats{
			Pool: poolStats,
		},
	}
}
