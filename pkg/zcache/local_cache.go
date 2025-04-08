package zcache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/dgraph-io/ristretto"
	"github.com/zondax/golem/pkg/logger"
	"github.com/zondax/golem/pkg/metrics"
	"github.com/zondax/golem/pkg/metrics/collectors"
)

const (
	neverExpires = -1
	cacheCost    = 1

	errorTypeLabel         = "error_type"
	itemCountLabel         = "item_count"
	residentItemCountLabel = "resident_item_count"
	deletedItemCountLabel  = "deleted_item_count"
	iterationErrorLabel    = "iteration_error"
	unmarshalErrorLabel    = "unmarshal_error"
	deletionErrorLabel     = "deletion_error"

	cleanupItemCountMetricKey        = "local_cache_cleanup_item_count"
	cleanupDeletedItemCountMetricKey = "local_cache_cleanup_deleted_item_count"
	cleanupErrorMetricKey            = "local_cache_cleanup_errors"
	cleanupLastRunMetricKey          = "local_cache_cleanup_last_run"
)

type CacheItem struct {
	Value     []byte `json:"value"`
	ExpiresAt int64  `json:"expires_at"`
}

func NewCacheItem(value []byte, ttl time.Duration) CacheItem {
	expiresAt := time.Now().Add(ttl).Unix()
	if ttl < 0 {
		expiresAt = neverExpires
	}

	return CacheItem{
		Value:     value,
		ExpiresAt: expiresAt,
	}
}

func (item CacheItem) IsExpired() bool {
	if item.ExpiresAt < 0 {
		return false
	}
	return time.Now().Unix() > item.ExpiresAt
}

type LocalCache interface {
	ZCache
}

type localCache struct {
	client        *ristretto.Cache
	prefix        string
	logger        *logger.Logger
	metricsServer metrics.TaskMetrics
	deleteHits    uint64
	deleteMisses  uint64
}

func (c *localCache) Set(_ context.Context, key string, value interface{}, ttl time.Duration) error {
	realKey := getKeyWithPrefix(c.prefix, key)

	b, err := json.Marshal(value)
	if err != nil {
		c.logger.Errorf("error marshalling value, key: [%s], err: [%s]", realKey, err)
		return err
	}

	ttlSeconds := ttl
	if ttl < 0 {
		ttlSeconds = 0 // 0 means never expire in Ristretto
	}

	if !c.client.SetWithTTL(realKey, b, cacheCost, ttlSeconds) {
		c.logger.Errorf("error setting new key on local cache, fullKey: [%s]", realKey)
		return errors.New("failed to set key with TTL")
	}

	c.client.Wait()
	return nil
}

func (c *localCache) Get(_ context.Context, key string, data interface{}) error {
	realKey := getKeyWithPrefix(c.prefix, key)

	val, found := c.client.Get(realKey)
	if !found {
		c.logger.Debugf("key not found on local cache, fullKey: [%s]", realKey)
		return errors.New("cache miss")
	}

	return json.Unmarshal(val.([]byte), data)
}

// Delete removes a value from the cache
func (c *localCache) Delete(ctx context.Context, key string) error {
	if c.client == nil {
		return fmt.Errorf("cache client is not initialized")
	}
	realKey := getKeyWithPrefix(c.prefix, key)
	c.logger.Debugf("delete key on local cache, fullKey: [%s]", realKey)
	c.client.Del(realKey) // Del might be accessing a nil client
	return nil
}

func (c *localCache) GetStats() ZCacheStats {
	stats := c.client.Metrics
	c.logger.Debugf("local cache stats: [%v]", stats)

	return ZCacheStats{Local: stats}
}

func (c *localCache) IsNotFoundError(err error) bool {
	return err != nil && err.Error() == "cache miss"
}

func (c *localCache) setupAndMonitorMetrics(updateInterval time.Duration) {
	setupAndMonitorCacheMetrics(c.metricsServer, c, c.logger, updateInterval)
}

func (c *localCache) registerCleanupMetrics() {
	if err := c.metricsServer.RegisterMetric(cleanupErrorMetricKey, "Counts different types of errors occurred during cache cleanup process", []string{errorTypeLabel}, &collectors.Counter{}); err != nil {
		c.logger.Errorf("[cleanup] - Failed to register cleanup metrics, err: %s", err)
	}

	if err := c.metricsServer.RegisterMetric(cleanupItemCountMetricKey, "Counts the valid items in the cache during cache cleanup process", []string{itemCountLabel}, &collectors.Gauge{}); err != nil {
		c.logger.Errorf("[cleanup] - Failed to register cleanup metrics, err: %s", err)
	}

	if err := c.metricsServer.RegisterMetric(cleanupDeletedItemCountMetricKey, "Counts the expired (deleted) items in the cache during cache cleanup process", []string{itemCountLabel}, &collectors.Gauge{}); err != nil {
		c.logger.Errorf("[cleanup] - Failed to register cleanup metrics, err: %s", err)
	}

	if err := c.metricsServer.RegisterMetric(cleanupLastRunMetricKey, "Timestamp of the last cleanup process execution", []string{}, &collectors.Gauge{}); err != nil {
		c.logger.Errorf("[cleanup] - Failed to register cleanup last run metric, err: %s", err)
	}
}
