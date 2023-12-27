package zcache

import (
	"context"
	"time"
)

type CombinedCache interface {
	ZCache
}

type combinedCache struct {
	localCache         LocalCache
	remoteCache        RemoteCache
	isRemoteBestEffort bool
}

func (c *combinedCache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	if err := c.localCache.Set(ctx, key, value, -1); err != nil {
		return err
	}
	if err := c.remoteCache.Set(ctx, key, value, expiration); err != nil && !c.isRemoteBestEffort {
		return err
	}
	return nil
}

func (c *combinedCache) Get(ctx context.Context, key string, data interface{}) error {
	err := c.localCache.Get(ctx, key, data)
	if err != nil {
		if err := c.remoteCache.Get(ctx, key, data); err != nil {
			return err
		}
		_ = c.localCache.Set(ctx, key, data, -1)
	}

	return nil
}

func (c *combinedCache) Delete(ctx context.Context, key string) error {
	err1 := c.localCache.Delete(ctx, key)
	err2 := c.remoteCache.Delete(ctx, key)

	if err1 != nil {
		return err1
	}
	if err2 != nil && !c.isRemoteBestEffort {
		return err2
	}

	return nil
}
