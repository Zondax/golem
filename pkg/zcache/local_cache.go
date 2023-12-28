package zcache

import (
	"context"
	"encoding/json"
	"github.com/allegro/bigcache/v3"
	"time"
)

type LocalCache interface {
	ZCache
}

type localCache struct {
	client *bigcache.BigCache
}

func (c *localCache) Set(_ context.Context, key string, value interface{}, _ time.Duration) error {
	val, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return c.client.Set(key, val)
}

func (c *localCache) Get(_ context.Context, key string, data interface{}) error {
	val, err := c.client.Get(key)
	if err != nil {
		return err
	}
	return json.Unmarshal(val, &data)
}

func (c *localCache) Delete(_ context.Context, key string) error {
	return c.client.Delete(key)
}

func (c *localCache) GetStats() ZCacheStats {
	stats := c.client.Stats()
	return ZCacheStats{Bigcache: &stats}
}
