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
	startCleanupProcess()
}

type localCache struct {
	client         *ristretto.Cache
	prefix         string
	logger         *logger.Logger
	metricsServer  metrics.TaskMetrics
	cleanupProcess CleanupProcess
	deleteHits     uint64
	deleteMisses   uint64
	keysList       []string
}

func (c *localCache) Set(_ context.Context, key string, value interface{}, ttl time.Duration) error {
	realKey := getKeyWithPrefix(c.prefix, key)

	b, err := json.Marshal(value)
	if err != nil {
		c.logger.Errorf("error marshalling cache item value, key: [%s], err: [%s]", realKey, err)
		return err
	}

	// Create a cache item with the correct TTL
	cacheItem := NewCacheItem(b, ttl)
	itemBytes, err := json.Marshal(cacheItem)
	if err != nil {
		c.logger.Errorf("error marshalling cache item, key: [%s], err: [%s]", realKey, err)
		return err
	}

	c.logger.Debugf("set key on local cache with TTL, key: [%s], value: [%v], ttl: [%v]", realKey, value, ttl)

	// Convert time.Duration to seconds for Ristretto
	var ttlSeconds time.Duration
	if ttl == neverExpires {
		ttlSeconds = 0 // 0 means never expire in Ristretto
	} else {
		ttlSeconds = ttl // Use the provided TTL
	}

	if !c.client.SetWithTTL(realKey, itemBytes, cacheCost, ttlSeconds) {
		c.logger.Errorf("error setting new key on local cache, fullKey: [%s]", realKey)
		return errors.New("failed to set key with TTL")
	}

	// Ensure the item is added to the cache
	c.client.Wait()
	c.keysList = append(c.keysList, realKey)

	return nil
}

// Get retrieves a value from the cache
func (c *localCache) Get(_ context.Context, key string, data interface{}) error {
	realKey := getKeyWithPrefix(c.prefix, key)

	c.logger.Debugf("get key on local cache, fullKey: [%s]", realKey)

	val, found := c.client.Get(realKey)
	if !found {
		c.logger.Debugf("key not found on local cache, fullKey: [%s]", realKey)
		return errors.New("cache miss")
	}

	var cachedItem CacheItem
	if err := json.Unmarshal(val.([]byte), &cachedItem); err != nil {
		c.logger.Errorf("error unmarshalling cache item, key: [%s], err: [%s]", realKey, err)
		return err
	}

	if cachedItem.IsExpired() {
		c.logger.Debugf("key expired on local cache, key: [%s]", realKey)
		c.client.Del(realKey)
		return errors.New("cache item expired")
	}

	c.logger.Debugf("key retrieved from local cache, key: [%s]", realKey)
	return json.Unmarshal(cachedItem.Value, data)
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

func (c *localCache) startCleanupProcess() {
	if c.cleanupProcess.Interval == 0 {
		c.logger.Warn("Cleanup process interval is 0, skipping cleanup.")
		return
	}

	ticker := time.NewTicker(c.cleanupProcess.Interval)
	go func() {
		for range ticker.C {
			if c.client != nil {
				c.cleanupExpiredKeys()
			} else {
				c.logger.Warn("Cache client is nil, cleanup aborted.")
			}
		}
	}()
}

func (c *localCache) cleanupExpiredKeys() {
	var keysToDelete []string
	var totalDeleted int
	var totalResident int

	// Iterate over each key in the keysList
	for _, key := range c.keysList {
		entry, found := c.client.Get(key)
		if !found {
			continue
		}

		// Retrieve the cached item as []byte
		data, ok := entry.([]byte)
		if !ok {
			c.logger.Errorf("[cleanup] - Invalid cache entry type for key: %s", key)
			continue
		}

		var cachedItem CacheItem
		// Unmarshal the data into a CacheItem
		if err := json.Unmarshal(data, &cachedItem); err != nil {
			c.logger.Errorf("[cleanup] - Error unmarshalling cache item: %v", err)
			if err := c.metricsServer.UpdateMetric(cleanupErrorMetricKey, 1, unmarshalErrorLabel); err != nil {
				c.logger.Errorf("[cleanup] - error updating %s metric with label %s: [%s]", cleanupErrorMetricKey, unmarshalErrorLabel, err)
			}
			continue
		}

		totalResident++

		// Skip items with TTL == -1 (neverExpires) during cleanup
		if cachedItem.ExpiresAt == neverExpires {
			c.logger.Debugf("[cleanup] - Skipping permanent item (TTL=-1) with key: %s", key)
			continue
		}

		// If the item is expired, add it to the list of keys to delete
		if cachedItem.IsExpired() {
			keysToDelete = append(keysToDelete, key)
		}

		// If we reach the batch size, delete keys in batch and wait for throttle time
		if len(keysToDelete) >= c.cleanupProcess.BatchSize {
			totalDeleted += c.deleteKeysInBatch(keysToDelete)
			keysToDelete = keysToDelete[:0]           // Reset the list of keys to delete
			time.Sleep(c.cleanupProcess.ThrottleTime) // Throttle the cleanup process
		}
	}

	// Delete any remaining keys after the loop completes
	if len(keysToDelete) > 0 {
		totalDeleted += c.deleteKeysInBatch(keysToDelete)
	}

	// Update metrics for resident and deleted items
	if err := c.metricsServer.UpdateMetric(cleanupItemCountMetricKey, float64(totalResident-totalDeleted), residentItemCountLabel); err != nil {
		c.logger.Errorf("[cleanup] - Failed to update cleanup item count metric, err: %s", err)
	}
	if err := c.metricsServer.UpdateMetric(cleanupDeletedItemCountMetricKey, float64(totalDeleted), deletedItemCountLabel); err != nil {
		c.logger.Errorf("[cleanup] - Failed to update deletion cleanup deleted item count metric, err: %s", err)
	}

	// Update the cleanup last run time metric
	if err := c.metricsServer.UpdateMetric(cleanupLastRunMetricKey, float64(time.Now().Unix())); err != nil {
		c.logger.Errorf("[cleanup] - Failed to update cleanup last run metric, err: %s", err)
	}
}

func (c *localCache) deleteKeysInBatch(keys []string) (deleted int) {
	// Ristretto's Del method simply deletes the key and doesn't return a value.
	for _, key := range keys {
		_, found := c.client.Get(key)
		if found {
			c.client.Del(key)
			c.deleteHits++
		} else {
			c.deleteMisses++
		}

		deleted++
	}

	_ = c.metricsServer.UpdateMetric(localCacheDelHitsMetricName, float64(c.deleteHits))
	_ = c.metricsServer.UpdateMetric(localCacheDelMissesMetricName, float64(c.deleteMisses))

	return
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
