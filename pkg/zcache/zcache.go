package zcache

import (
	"context"
	"fmt"
	"time"

	"github.com/dgraph-io/ristretto"
	"github.com/go-redis/redis/v8"
	"github.com/zondax/golem/pkg/logger"
)

type ZCacheStats struct {
	Local  *ristretto.Metrics
	Remote *RedisStats
}

type ZCache interface {
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	Get(ctx context.Context, key string, data interface{}) error
	Delete(ctx context.Context, key string) error
	GetStats() ZCacheStats
	IsNotFoundError(err error) bool
}

func NewLocalCache(config *LocalConfig) (LocalCache, error) {
	if config.MetricServer == nil {
		return nil, fmt.Errorf("metric server is mandatory")
	}

	ristrettoConfig := config.ToRistrettoConfig()

	client, err := ristretto.NewCache(ristrettoConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Ristretto cache: %w", err)
	}

	loggerInst := config.Logger
	if loggerInst == nil {
		loggerInst = logger.NewLogger()
	}

	lc := &localCache{
		client:        client,
		prefix:        config.Prefix,
		logger:        loggerInst,
		metricsServer: config.MetricServer,
	}

	if config.StatsMetrics.Enable {
		lc.setupAndMonitorMetrics(config.StatsMetrics.UpdateInterval)
	}

	return lc, nil
}

func NewRemoteCache(config *RemoteConfig) (RemoteCache, error) {
	redisOptions, err := config.ToRedisConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to create redis config: %w", err)
	}

	client := redis.NewClient(redisOptions)

	// Validate connection with Ping
	dialTimeout := config.DialTimeout
	if dialTimeout == 0 {
		dialTimeout = 5 * time.Second // Default timeout
	}
	ctx, cancel := context.WithTimeout(context.Background(), dialTimeout)
	defer cancel()
	if _, err := client.Ping(ctx).Result(); err != nil {
		return nil, fmt.Errorf("redis connection failed: %w", err)
	}

	loggerInst := config.Logger
	if loggerInst == nil {
		loggerInst = logger.NewLogger()
	}

	rc := &redisCache{
		client:        client,
		prefix:        config.Prefix,
		logger:        loggerInst,
		metricsServer: config.MetricServer,
	}

	if config.StatsMetrics.Enable {
		rc.setupAndMonitorMetrics(config.StatsMetrics.UpdateInterval)
	}

	return rc, nil
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

	// Disable stats metrics registration on inner caches to avoid possible collisions
	localCacheConfig.StatsMetrics = StatsMetrics{}
	remoteCacheConfig.StatsMetrics = StatsMetrics{}

	// Remote cache
	// Set global configs on remote cache config
	remoteCacheConfig.Prefix = combinedConfig.GlobalPrefix
	remoteCacheConfig.Logger = combinedConfig.GlobalLogger
	remoteCacheConfig.MetricServer = combinedConfig.GlobalMetricServer

	remoteClient, err := NewRemoteCache(remoteCacheConfig)
	if err != nil {
		return nil, err
	}

	// Local cache
	// Set global configs on local cache config
	localCacheConfig.Prefix = combinedConfig.GlobalPrefix
	localCacheConfig.Logger = combinedConfig.GlobalLogger
	localCacheConfig.MetricServer = combinedConfig.GlobalMetricServer

	localClient, err := NewLocalCache(localCacheConfig)
	if err != nil {
		return nil, err
	}

	// Combined cache
	cc := &combinedCache{
		remoteCache:        remoteClient,
		localCache:         localClient,
		isRemoteBestEffort: combinedConfig.IsRemoteBestEffort,
		metricsServer:      combinedConfig.GlobalMetricServer,
		logger:             combinedConfig.GlobalLogger,
	}

	if combinedConfig.GlobalStatsMetrics.Enable {
		cc.setupAndMonitorMetrics(combinedConfig.GlobalStatsMetrics.UpdateInterval)
	}

	return cc, nil
}
