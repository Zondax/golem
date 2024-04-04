package zcache

import (
	"github.com/allegro/bigcache/v3"
	"github.com/go-redis/redis/v8"
	"github.com/zondax/golem/pkg/logger"
	"github.com/zondax/golem/pkg/metrics"
	"time"
)

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
	Prefix         string
	Logger         *logger.Logger
	MetricServer   metrics.TaskMetrics
	StatsMetrics   StatsMetrics
	CleanupProcess CleanupProcess
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
	evictionTime := time.Duration(100*365*24) * time.Hour // 100 years
	return bigcache.DefaultConfig(evictionTime)
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
