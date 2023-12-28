package zcache

import (
	"context"
	"github.com/allegro/bigcache/v3"
	"github.com/go-redis/redis/v8"
	"github.com/zondax/golem/pkg/metrics"
	"time"
)

type ZCacheStats struct {
	Local  *bigcache.Stats
	Remote *RedisStats
}

type ZCache interface {
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	Get(ctx context.Context, key string, data interface{}) error
	Delete(ctx context.Context, key string) error
	GetStats() ZCacheStats
	SetupAndMonitorMetrics(appName string, metricsServer metrics.TaskMetrics, updateInterval time.Duration) []error
}

func NewLocalCache(config *LocalConfig) (LocalCache, error) {
	bigCacheConfig := config.ToBigCacheConfig()
	client, err := bigcache.New(context.Background(), bigCacheConfig)
	return &localCache{client: client, prefix: config.Prefix}, err
}

func NewRemoteCache(config *RemoteConfig) (RemoteCache, error) {
	redisOptions := config.ToRedisConfig()
	client := redis.NewClient(redisOptions)
	return &redisCache{client: client, prefix: config.Prefix}, nil
}

func NewCombinedCache(combinedConfig *CombinedConfig) (CombinedCache, error) {
	localCacheConfig := combinedConfig.Local
	remoteCacheConfig := combinedConfig.Remote

	// Set global configs on remote cache config
	remoteCacheConfig.Prefix = combinedConfig.globalPrefix
	remoteClient, err := NewRemoteCache(remoteCacheConfig)
	if err != nil {
		return nil, err
	}

	// Set global configs on local cache config
	localCacheConfig.EvictionInSeconds = combinedConfig.globalTtlSeconds
	localCacheConfig.Prefix = combinedConfig.globalPrefix
	localClient, err := NewLocalCache(localCacheConfig)
	if err != nil {
		return nil, err
	}

	return &combinedCache{
		remoteCache:        remoteClient,
		localCache:         localClient,
		isRemoteBestEffort: combinedConfig.isRemoteBestEffort,
		ttl:                time.Duration(combinedConfig.globalTtlSeconds) * time.Second,
	}, nil
}
