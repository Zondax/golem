package zcache

import (
	"github.com/allegro/bigcache/v3"
	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
	"math"
	"time"
)

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
	Logger             *zap.Logger
}

type LocalConfig struct {
	EvictionInSeconds int
	Prefix            string
	Logger            *zap.Logger
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
	eviction := time.Duration(c.EvictionInSeconds) * time.Second
	if c.EvictionInSeconds < 0 {
		eviction = time.Duration(math.MaxInt64)
	}

	return bigcache.DefaultConfig(eviction)
}

type CombinedConfig struct {
	Local              *LocalConfig
	Remote             *RemoteConfig
	GlobalLogger       *zap.Logger
	GlobalTtlSeconds   int
	GlobalPrefix       string
	IsRemoteBestEffort bool
}
