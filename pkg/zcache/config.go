package zcache

import (
	"time"

	"github.com/allegro/bigcache/v3"
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
	Prefix               string
	Logger               *logger.Logger
	MetricServer         metrics.TaskMetrics
	StatsMetrics         StatsMetrics
	CleanupProcess       CleanupProcess
	HardMaxCacheSizeInMB int
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

func (c *LocalConfig) ToBigCacheConfig() bigcache.Config {
	config := bigcache.DefaultConfig(time.Duration(100*365*24) * time.Hour)

	if c.HardMaxCacheSizeInMB <= 0 {
		c.HardMaxCacheSizeInMB = hardMaxCacheSizeDefault
	}
	config.HardMaxCacheSize = c.HardMaxCacheSizeInMB * 1024 * 1024

	return config
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
