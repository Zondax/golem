package zcache

import (
	"time"

	"github.com/dgraph-io/ristretto"
	"github.com/go-redis/redis/v8"
	"github.com/zondax/golem/pkg/logger"
	"github.com/zondax/golem/pkg/metrics"
)

const hardMaxCacheSizeDefault = 512

type StatsMetrics struct {
	Enable         bool
	UpdateInterval time.Duration
}

type CleanupProcess struct {
	Interval     time.Duration
	BatchSize    int
	ThrottleTime time.Duration
}

type RemoteConfig struct {
	Network            string
	Addr               string
	Password           string
	DB                 int
	DialTimeout        time.Duration
	ReadTimeout        time.Duration
	WriteTimeout       time.Duration
	PoolSize           int
	MinIdleConns       int
	MaxConnAge         time.Duration
	PoolTimeout        time.Duration
	IdleTimeout        time.Duration
	IdleCheckFrequency time.Duration
	Prefix             string
	Logger             *logger.Logger
	MetricServer       metrics.TaskMetrics
	StatsMetrics       StatsMetrics
}

type LocalConfig struct {
	Prefix           string
	Logger           *logger.Logger
	MetricServer     metrics.TaskMetrics
	StatsMetrics     StatsMetrics
	CleanupProcess   CleanupProcess
	CacheSizeInBytes int64
}

func (c *RemoteConfig) ToRedisConfig() *redis.Options {
	return &redis.Options{
		Network:            c.Network,
		Addr:               c.Addr,
		Password:           c.Password,
		DB:                 c.DB,
		DialTimeout:        c.DialTimeout,
		ReadTimeout:        c.ReadTimeout,
		WriteTimeout:       c.WriteTimeout,
		PoolSize:           c.PoolSize,
		MinIdleConns:       c.MinIdleConns,
		MaxConnAge:         c.MaxConnAge,
		PoolTimeout:        c.PoolTimeout,
		IdleTimeout:        c.IdleTimeout,
		IdleCheckFrequency: c.IdleCheckFrequency,
	}
}

func (c *LocalConfig) ToRistrettoConfig() *ristretto.Config {
	if c.CacheSizeInBytes <= 0 {
		c.CacheSizeInBytes = hardMaxCacheSizeDefault
	}

	return &ristretto.Config{
		NumCounters: c.CacheSizeInBytes / 64, // Number of counters (approx 10x the capacity)
		MaxCost:     c.CacheSizeInBytes,      // Max cost in bytes
		BufferItems: 64,                      // Number of keys per Get buffer
	}
}

type CombinedConfig struct {
	Local              *LocalConfig
	Remote             *RemoteConfig
	GlobalLogger       *logger.Logger
	GlobalPrefix       string
	GlobalMetricServer metrics.TaskMetrics
	GlobalStatsMetrics StatsMetrics
	IsRemoteBestEffort bool
}
