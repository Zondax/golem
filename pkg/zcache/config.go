package zcache

import (
	"github.com/allegro/bigcache/v3"
	"github.com/go-redis/redis/v8"
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
}

type LocalConfig struct {
	EvictionInSeconds int
	Prefix            string
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
	return bigcache.DefaultConfig(time.Second * time.Duration(c.EvictionInSeconds))
}

type CombinedConfig struct {
	Local              *LocalConfig
	Remote             *RemoteConfig
	GlobalTtlSeconds   int
	GlobalPrefix       string
	IsRemoteBestEffort bool
}
