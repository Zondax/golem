package zcache

import (
	"context"
	"github.com/allegro/bigcache/v3"
	"github.com/go-redis/redis/v8"
	"github.com/zondax/golem/pkg/logger"
	"time"
)

const (
	defaultCleanupInterval = 12 * time.Hour
	defaultBatchSize       = 200
	defaultThrottleTime    = time.Second
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
}

func NewLocalCache(config *LocalConfig) (LocalCache, error) {
	if config.MetricServer == nil {
		panic("metric server is mandatory")
	}

	bigCacheConfig := config.ToBigCacheConfig()
	client, err := bigcache.New(context.Background(), bigCacheConfig)
	if err != nil {
		return nil, err
	}

	loggerInst := config.Logger
	if loggerInst == nil {
		loggerInst = logger.NewLogger()
	}

	if config.CleanupProcess.Interval <= 0 {
		config.CleanupProcess.Interval = defaultCleanupInterval
	}

	if config.CleanupProcess.BatchSize <= 0 {
		config.CleanupProcess.BatchSize = defaultBatchSize
	}

	if config.CleanupProcess.ThrottleTime <= 0 {
		config.CleanupProcess.ThrottleTime = defaultThrottleTime
	}

	lc := &localCache{
		client:         client,
		prefix:         config.Prefix,
		logger:         loggerInst,
		cleanupProcess: config.CleanupProcess,
		metricsServer:  config.MetricServer,
	}

	lc.registerCleanupMetrics()
	lc.startCleanupProcess()

	if config.StatsMetrics.Enable {
		if config.MetricServer == nil {
			panic("metric server is mandatory")
		}

		lc.setupAndMonitorMetrics(config.StatsMetrics.UpdateInterval)
	}

	return lc, nil
}

func NewRemoteCache(config *RemoteConfig) (RemoteCache, error) {
	redisOptions := config.ToRedisConfig()
	client := redis.NewClient(redisOptions)

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
