package zcache

import (
	"context"
	"github.com/allegro/bigcache/v3"
	"github.com/go-redis/redis/v8"
	"github.com/zondax/golem/pkg/metrics"
	"go.uber.org/zap"
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
	IsNotFoundError(err error) bool
	SetupAndMonitorMetrics(appName string, metricsServer metrics.TaskMetrics, updateInterval time.Duration) []error
}

func NewLocalCache(config *LocalConfig) (LocalCache, error) {
	bigCacheConfig := config.ToBigCacheConfig()
	client, err := bigcache.New(context.Background(), bigCacheConfig)

	logger := config.Logger
	if logger == nil {
		logger = zap.NewNop()
	}

	return &localCache{client: client, prefix: config.Prefix, logger: logger}, err
}

func NewRemoteCache(config *RemoteConfig) (RemoteCache, error) {
	redisOptions := config.ToRedisConfig()
	client := redis.NewClient(redisOptions)

	logger := config.Logger
	if logger == nil {
		logger = zap.NewNop()
	}

	return &redisCache{client: client, prefix: config.Prefix, logger: logger}, nil
}

func NewCombinedCache(combinedConfig *CombinedConfig) (CombinedCache, error) {
	localCacheConfig := combinedConfig.Local
	remoteCacheConfig := combinedConfig.Remote

	if localCacheConfig == nil {
		localCacheConfig = &LocalConfig{}
	}

	if remoteCacheConfig == nil {
		remoteCacheConfig = &RemoteConfig{}
	}

	// Set global configs on remote cache config
	remoteCacheConfig.Prefix = combinedConfig.GlobalPrefix
	remoteCacheConfig.Logger = combinedConfig.GlobalLogger

	remoteClient, err := NewRemoteCache(remoteCacheConfig)
	if err != nil {
		return nil, err
	}

	// Set global configs on local cache config
	localCacheConfig.EvictionInSeconds = combinedConfig.GlobalTtlSeconds
	localCacheConfig.Prefix = combinedConfig.GlobalPrefix
	localCacheConfig.Logger = combinedConfig.GlobalLogger

	localClient, err := NewLocalCache(localCacheConfig)
	if err != nil {
		return nil, err
	}

	return &combinedCache{
		remoteCache:        remoteClient,
		localCache:         localClient,
		isRemoteBestEffort: combinedConfig.IsRemoteBestEffort,
		ttl:                time.Duration(combinedConfig.GlobalTtlSeconds) * time.Second,
		logger:             combinedConfig.GlobalLogger,
	}, nil
}
