package zcache

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/zondax/golem/pkg/logger"

	"github.com/allegro/bigcache/v3"
	"github.com/zondax/golem/pkg/metrics"
	"github.com/zondax/golem/pkg/metrics/collectors"
)

const (
	neverExpires = -1

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
	client         *bigcache.BigCache
	prefix         string
	logger         *logger.Logger
	metricsServer  metrics.TaskMetrics
	cleanupProcess CleanupProcess
}

func (c *localCache) Set(_ context.Context, key string, value interface{}, ttl time.Duration) error {
	realKey := getKeyWithPrefix(c.prefix, key)

	b, err := json.Marshal(value)
	if err != nil {
		c.logger.Errorf("error marshalling cache item value, key: [%s], err: [%s]", realKey, err)
		return err
	}

	cacheItem := NewCacheItem(b, ttl)
	itemBytes, err := json.Marshal(cacheItem)
	if err != nil {
		c.logger.Errorf("error marshalling cache item, key: [%s], err: [%s]", realKey, err)
		return err
	}

	c.logger.Debugf("set key on local cache with TTL, key: [%s], value: [%v], ttl: [%v]", realKey, value, ttl)
	if err = c.client.Set(realKey, itemBytes); err != nil {
		c.logger.Errorf("error setting new key on local cache, fullKey: [%s], err: [%s]", realKey, err)
	}

	return err
}

func (c *localCache) Get(_ context.Context, key string, data interface{}) error {
	realKey := getKeyWithPrefix(c.prefix, key)

	c.logger.Debugf("get key on local cache, fullKey: [%s]", realKey)

	val, err := c.client.Get(realKey)
	if err != nil {
		if c.IsNotFoundError(err) {
			c.logger.Debugf("key not found on local cache, fullKey: [%s]", realKey)
		} else {
			c.logger.Errorf("error getting key on local cache, fullKey: [%s], err: [%s]", realKey, err)
		}

		return err
	}

	var cachedItem CacheItem
	if err := json.Unmarshal(val, &cachedItem); err != nil {
		c.logger.Errorf("error unmarshalling cache item, key: [%s], err: [%s]", realKey, err)
		return err
	}

	if cachedItem.IsExpired() {
		c.logger.Debugf("key expired on local cache, key: [%s]", realKey)
		_ = c.client.Delete(realKey)
		return errors.New("cache item expired")
	}

	c.logger.Debugf("key retrieved from local cache, key: [%s]", realKey)
	return json.Unmarshal(cachedItem.Value, data)
}

func (c *localCache) Delete(_ context.Context, key string) error {
	realKey := getKeyWithPrefix(c.prefix, key)
	c.logger.Debugf("delete key on local cache, fullKey: [%s]", realKey)

	return c.client.Delete(realKey)
}

func (c *localCache) GetStats() ZCacheStats {
	stats := c.client.Stats()
	c.logger.Debugf("local cache stats: [%v]", stats)

	return ZCacheStats{Local: &stats}
}

func (c *localCache) IsNotFoundError(err error) bool {
	return errors.Is(err, bigcache.ErrEntryNotFound)
}

func (c *localCache) setupAndMonitorMetrics(updateInterval time.Duration) {
	setupAndMonitorCacheMetrics(c.metricsServer, c, c.logger, updateInterval)
}

func (c *localCache) startCleanupProcess() {
	ticker := time.NewTicker(c.cleanupProcess.Interval)
	go func() {
		for range ticker.C {
			c.cleanupExpiredKeys()
		}
	}()
}

func (c *localCache) cleanupExpiredKeys() {
	iterator := c.client.Iterator()
	var keysToDelete []string
	var totalDeleted int
	var totalResident int

	for iterator.SetNext() {
		entry, err := iterator.Value()
		if err != nil {
			c.logger.Errorf("Error iterating over cache entries: %v", err)
			if err = c.metricsServer.UpdateMetric(cleanupErrorMetricKey, 1, iterationErrorLabel); err != nil {
				c.logger.Errorf("error updating %s metric with label %s: [%s]", cleanupErrorMetricKey, iterationErrorLabel, err)
			}
			continue
		}

		var cachedItem CacheItem
		if err = json.Unmarshal(entry.Value(), &cachedItem); err != nil {
			c.logger.Errorf("Error unmarshalling cache item: %v", err)
			if err = c.metricsServer.UpdateMetric(cleanupErrorMetricKey, 1, unmarshalErrorLabel); err != nil {
				c.logger.Errorf("error updating %s metric with label %s: [%s]", cleanupErrorMetricKey, unmarshalErrorLabel, err)
			}
			continue
		}

		totalResident++

		if cachedItem.IsExpired() {
			keysToDelete = append(keysToDelete, entry.Key())
		}

		if len(keysToDelete) >= c.cleanupProcess.BatchSize {
			totalDeleted += c.deleteKeysInBatch(keysToDelete)
			keysToDelete = keysToDelete[:0]
			time.Sleep(c.cleanupProcess.ThrottleTime)
		}
	}

	if len(keysToDelete) > 0 {
		totalDeleted += c.deleteKeysInBatch(keysToDelete)
	}

	// update metrics
	if err := c.metricsServer.UpdateMetric(cleanupItemCountMetricKey, float64(totalResident-totalDeleted), residentItemCountLabel); err != nil {
		c.logger.Errorf("Failed to update cleanup item count metric, err: %s", err)
	}
	if err := c.metricsServer.UpdateMetric(cleanupDeletedItemCountMetricKey, float64(totalDeleted), deletedItemCountLabel); err != nil {
		c.logger.Errorf("Failed to update deletion cleanup deleted item count metric, err: %s", err)
	}

	if err := c.metricsServer.UpdateMetric(cleanupLastRunMetricKey, float64(time.Now().Unix())); err != nil {
		c.logger.Errorf("Failed to update cleanup last run metric, err: %s", err)
	}
}

func (c *localCache) deleteKeysInBatch(keys []string) (deleted int) {
	for _, key := range keys {
		if err := c.client.Delete(key); err != nil {
			c.logger.Errorf("Error deleting key %s: %v", key, err)
			if err = c.metricsServer.UpdateMetric(cleanupErrorMetricKey, 1, deletionErrorLabel); err != nil {
				c.logger.Errorf("error updating %s metric with label %s: [%s]", cleanupErrorMetricKey, deletionErrorLabel, err)
			}
			continue
		}
		deleted++
	}
	return
}

func (c *localCache) registerCleanupMetrics() {
	if err := c.metricsServer.RegisterMetric(cleanupErrorMetricKey, "Counts different types of errors occurred during cache cleanup process", []string{errorTypeLabel}, &collectors.Counter{}); err != nil {
		c.logger.Errorf("Failed to register cleanup metrics, err: %s", err)
	}

	if err := c.metricsServer.RegisterMetric(cleanupItemCountMetricKey, "Counts the valid items in the cache during cache cleanup process", []string{itemCountLabel}, &collectors.Gauge{}); err != nil {
		c.logger.Errorf("Failed to register cleanup metrics, err: %s", err)
	}

	if err := c.metricsServer.RegisterMetric(cleanupDeletedItemCountMetricKey, "Counts the expired (deleted) items in the cache during cache cleanup process", []string{itemCountLabel}, &collectors.Gauge{}); err != nil {
		c.logger.Errorf("Failed to register cleanup metrics, err: %s", err)
	}

	if err := c.metricsServer.RegisterMetric(cleanupLastRunMetricKey, "Timestamp of the last cleanup process execution", []string{}, &collectors.Gauge{}); err != nil {
		c.logger.Errorf("Failed to register cleanup last run metric, err: %s", err)
	}
}
