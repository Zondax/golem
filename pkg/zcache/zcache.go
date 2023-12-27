package zcache

import (
	"context"
	"github.com/allegro/bigcache/v3"
	"github.com/go-redis/redis/v8"
	"time"
)

type ZCache interface {
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	Get(ctx context.Context, key string, data interface{}) error
	Delete(ctx context.Context, key string) error
}

func NewLocalCache(config *LocalConfig) (LocalCache, error) {
	bigCacheConfig := config.ToBigCacheConfig()
	client, err := bigcache.New(context.Background(), bigCacheConfig)
	return &localCache{client: client}, err
}

func NewRemoteCache(config *RemoteConfig) (RemoteCache, error) {
	redisOptions := config.ToRedisConfig()
	client := redis.NewClient(redisOptions)
	return &redisCache{client: client}, nil
}

func NewCombinedCache(combinedConfig *CombinedConfig) (CombinedCache, error) {
	remoteClient, err := NewRemoteCache(combinedConfig.Remote)
	if err != nil && !combinedConfig.isRemoteBestEffort {
		return nil, err
	}

	localClient, err := NewLocalCache(combinedConfig.Local)
	if err != nil {
		return nil, err
	}

	return &combinedCache{remoteCache: remoteClient, localCache: localClient, isRemoteBestEffort: combinedConfig.isRemoteBestEffort}, nil
}
