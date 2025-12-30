package zcache

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

// RedisPipeline provides pipelined execution of Redis commands
type RedisPipeline interface {
	IncrBy(ctx context.Context, key string, value int64) *redis.IntCmd
	HIncrBy(ctx context.Context, key, field string, incr int64) *redis.IntCmd
	HSet(ctx context.Context, key string, values ...interface{}) *redis.IntCmd
	HSetNX(ctx context.Context, key, field string, value interface{}) *redis.BoolCmd
	HExists(ctx context.Context, key, field string) *redis.BoolCmd
	HGetAll(ctx context.Context, key string) *redis.StringStringMapCmd
	Expire(ctx context.Context, key string, expiration time.Duration) *redis.BoolCmd
	Del(ctx context.Context, keys ...string) *redis.IntCmd
	Exec(ctx context.Context) ([]redis.Cmder, error)
}

// redisPipeline wraps redis.Pipeliner to implement RedisPipeline
type redisPipeline struct {
	pipeliner redis.Pipeliner
	prefix    string
}

// Pipeline returns a new pipeline for batching commands
func (c *redisCache) Pipeline() RedisPipeline {
	return &redisPipeline{
		pipeliner: c.client.Pipeline(),
		prefix:    c.prefix,
	}
}

// TxPipeline returns a new transactional pipeline
func (c *redisCache) TxPipeline() RedisPipeline {
	return &redisPipeline{
		pipeliner: c.client.TxPipeline(),
		prefix:    c.prefix,
	}
}

func (p *redisPipeline) IncrBy(ctx context.Context, key string, value int64) *redis.IntCmd {
	realKey := getKeyWithPrefix(p.prefix, key)
	return p.pipeliner.IncrBy(ctx, realKey, value)
}

func (p *redisPipeline) HIncrBy(ctx context.Context, key, field string, incr int64) *redis.IntCmd {
	realKey := getKeyWithPrefix(p.prefix, key)
	return p.pipeliner.HIncrBy(ctx, realKey, field, incr)
}

func (p *redisPipeline) HSet(ctx context.Context, key string, values ...interface{}) *redis.IntCmd {
	realKey := getKeyWithPrefix(p.prefix, key)
	return p.pipeliner.HSet(ctx, realKey, values...)
}

func (p *redisPipeline) HSetNX(ctx context.Context, key, field string, value interface{}) *redis.BoolCmd {
	realKey := getKeyWithPrefix(p.prefix, key)
	return p.pipeliner.HSetNX(ctx, realKey, field, value)
}

func (p *redisPipeline) HExists(ctx context.Context, key, field string) *redis.BoolCmd {
	realKey := getKeyWithPrefix(p.prefix, key)
	return p.pipeliner.HExists(ctx, realKey, field)
}

func (p *redisPipeline) HGetAll(ctx context.Context, key string) *redis.StringStringMapCmd {
	realKey := getKeyWithPrefix(p.prefix, key)
	return p.pipeliner.HGetAll(ctx, realKey)
}

func (p *redisPipeline) Expire(ctx context.Context, key string, expiration time.Duration) *redis.BoolCmd {
	realKey := getKeyWithPrefix(p.prefix, key)
	return p.pipeliner.Expire(ctx, realKey, expiration)
}

func (p *redisPipeline) Del(ctx context.Context, keys ...string) *redis.IntCmd {
	realKeys := getKeysWithPrefix(p.prefix, keys)
	return p.pipeliner.Del(ctx, realKeys...)
}

func (p *redisPipeline) Exec(ctx context.Context) ([]redis.Cmder, error) {
	return p.pipeliner.Exec(ctx)
}
