package zcache

import (
	"context"
	"fmt"
	"github.com/allegro/bigcache/v3"
	"github.com/go-redis/redis/v8"
	"time"
)

type CacheType int

const (
	LocalCacheType CacheType = iota
	RemoteCacheType
)

type ZCache interface {
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Get(ctx context.Context, key string, data interface{}) error
	Delete(ctx context.Context, key string) error
}

func NewCache(config *Config, cacheType CacheType) (ZCache, error) {
	if cacheType == LocalCacheType {
		return newLocalCacheClient(config)
	}

	if cacheType == RemoteCacheType {
		return newRedisClient(config), nil

	}

	return nil, fmt.Errorf("cache type is invalid")
}

func newRedisClient(config *Config) ZCache {
	redisOptions := config.ToRedisConfig()
	client := redis.NewClient(redisOptions)
	return &redisCache{client: client}
}

func newLocalCacheClient(config *Config) (ZCache, error) {
	bigCacheConfig := config.ToBigCacheConfig()
	client, err := bigcache.New(context.Background(), bigCacheConfig)
	return &localCache{client: client}, err
}
