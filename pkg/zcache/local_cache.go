package zcache

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/allegro/bigcache/v3"
	"github.com/zondax/golem/pkg/metrics"
	"time"
)

type LocalCache interface {
	ZCache
}

type localCache struct {
	client        *bigcache.BigCache
	metricsServer *metrics.TaskMetrics
	appName       string
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
	return ZCacheStats{Local: &stats}
}

func (c *localCache) SetupAndMonitorMetrics(appName string, metricsServer metrics.TaskMetrics, updateInterval time.Duration) []error {
	c.metricsServer = &metricsServer
	c.appName = appName

	errs := setupAndMonitorCacheMetrics(appName, metricsServer, c, updateInterval)
	errs = append(errs, c.registerInternalCacheMetrics()...)

	return errs
}

func (c *localCache) registerInternalCacheMetrics() []error {
	if c.metricsServer == nil {
		return []error{fmt.Errorf("metrics server not available")}
	}

	return []error{}
}
