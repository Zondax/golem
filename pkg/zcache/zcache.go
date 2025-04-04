package zcache

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/dgraph-io/ristretto"
	"github.com/go-redis/redis/v8"
	"github.com/zondax/golem/pkg/logger"
)

const (
	defaultCleanupInterval = 12 * time.Hour
	defaultBatchSize       = 200
	defaultThrottleTime    = time.Second
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
	// Ensure MetricServer is provided
	if config.MetricServer == nil {
		return nil, fmt.Errorf("metric server is mandatory")
	}

	// Use ToRistrettoConfig to create the Ristretto config
	ristrettoConfig := config.ToRistrettoConfig()

	// Set up the Ristretto cache using the config from ToRistrettoConfig
	client, err := ristretto.NewCache(ristrettoConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Ristretto cache: %w", err)
	}

	// Use the provided logger or fallback to a default one
	loggerInst := config.Logger
	if loggerInst == nil {
		loggerInst = logger.NewLogger()
	}

	// Set default cleanup process parameters if not provided
	if config.CleanupProcess.Interval <= 0 {
		config.CleanupProcess.Interval = defaultCleanupInterval
	}
	if config.CleanupProcess.BatchSize <= 0 {
		config.CleanupProcess.BatchSize = defaultBatchSize
	}
	if config.CleanupProcess.ThrottleTime <= 0 {
		config.CleanupProcess.ThrottleTime = defaultThrottleTime
	}

	// Create the local cache instance
	lc := &localCache{
		client:         client,
		prefix:         config.Prefix,
		logger:         loggerInst,
		cleanupProcess: config.CleanupProcess,
		metricsServer:  config.MetricServer,
		keyListMap:     sync.Map{},
	}

	// Register cleanup metrics and start the cleanup process
	lc.registerCleanupMetrics()
	lc.startCleanupProcess()

	// Setup and monitor cache metrics if enabled
	if config.StatsMetrics.Enable {
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
