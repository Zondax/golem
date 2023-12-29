package zcache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/allegro/bigcache/v3"
	"github.com/zondax/golem/pkg/metrics"
	"go.uber.org/zap"
	"time"
)

type LocalCache interface {
	ZCache
}

type localCache struct {
	client        *bigcache.BigCache
	prefix        string
	logger        *zap.Logger
	metricsServer *metrics.TaskMetrics
	appName       string
}

func (c *localCache) Set(_ context.Context, key string, value interface{}, _ time.Duration) error {
	realKey := getKeyWithPrefix(c.prefix, key)

	c.logger.Sugar().Debugf("set key on local cache, fullKey: [%s], value: [%v]", realKey, value)
	val, err := json.Marshal(value)
	if err != nil {
		return err
	}

	err = c.client.Set(realKey, val)
	if err != nil {
		c.logger.Sugar().Errorf("error setting new key on local cache, fullKey: [%s], err: [%s]", realKey, err)
	}

	return err
}

func (c *localCache) Get(_ context.Context, key string, data interface{}) error {
	realKey := getKeyWithPrefix(c.prefix, key)

	c.logger.Sugar().Debugf("get key on local cache, fullKey: [%s]", realKey)

	val, err := c.client.Get(realKey)
	if err != nil {
		if c.IsNotFoundError(err) {
			c.logger.Sugar().Debugf("key not found on local cache, fullKey: [%s]", realKey)
		} else {
			c.logger.Sugar().Errorf("error getting key on local cache, fullKey: [%s], err: [%s]", realKey, err)
		}

		return err
	}
	return json.Unmarshal(val, &data)
}

func (c *localCache) Delete(_ context.Context, key string) error {
	realKey := getKeyWithPrefix(c.prefix, key)
	c.logger.Sugar().Debugf("delete key on local cache, fullKey: [%s]", realKey)

	return c.client.Delete(realKey)
}

func (c *localCache) GetStats() ZCacheStats {
	stats := c.client.Stats()
	c.logger.Sugar().Debugf("local cache stats: [%v]", stats)

	return ZCacheStats{Local: &stats}
}

func (c *localCache) IsNotFoundError(err error) bool {
	return errors.Is(err, bigcache.ErrEntryNotFound)
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
