package zcache

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/dgraph-io/ristretto"
	"github.com/stretchr/testify/suite"
	"github.com/zondax/golem/pkg/logger"
	"github.com/zondax/golem/pkg/metrics"
)

const (
	testValue = "testValue"
	expireKey = "expireKey"
)

func TestLocalCacheTestSuite(t *testing.T) {
	suite.Run(t, new(LocalCacheTestSuite))
}

type LocalCacheTestSuite struct {
	suite.Suite
	cache LocalCache
}

func (suite *LocalCacheTestSuite) SetupTest() {
	// Initialize the cache before each test
	cache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: 1e7,     // Number of keys to track frequency of access
		MaxCost:     1 << 20, // Maximum cost of cache
		BufferItems: 64,      // Number of keys to buffer for eviction
	})
	if err != nil {
		suite.T().Fatal("Failed to create cache:", err)
	}

	// Initialize the localCache instance
	suite.cache = &localCache{
		client: cache,
		prefix: "test_",            // Add any prefix or other necessary fields
		logger: logger.NewLogger(), // Ensure you provide a logger
	}
}

func (suite *LocalCacheTestSuite) TestDelete() {
	ctx := context.Background()
	key := "testKey"
	value := "testValue"

	// Ensure cache is initialized before testing
	suite.NoError(suite.cache.Set(ctx, key, value, 0))

	// Delete the cache item
	err := suite.cache.Delete(ctx, key)
	suite.NoError(err)

	// Attempt to get the value after deletion
	var result string
	err = suite.cache.Get(ctx, key, &result)

	// Check for cache miss error
	suite.Error(err, "Expected error when getting deleted key")
	suite.Empty(result, "Expected empty result as key is deleted")
}

func TestCacheSetAndGet(t *testing.T) {
	t.Log("Initializing cache")

	// Initialize ristretto cache
	cache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: 1e7,
		MaxCost:     1 << 30,
		BufferItems: 64,
	})
	if err != nil {
		t.Fatalf("Failed to initialize cache: %v", err)
	}

	t.Log("Cache initialized successfully")

	// Initialize logger (ensure the logger is initialized before use)
	logr := logger.NewLogger() // Ensure this creates a valid, non-nil logger

	// Initialize localCache
	lc := &localCache{
		client: cache,
		logger: logr,
	}
	key := "testKey"
	value := "testValue"

	t.Log("Setting key-value pair in cache")
	err = lc.Set(context.Background(), key, value, 60*time.Second)
	if err != nil {
		t.Fatalf("Failed to set cache item: %v", err)
	}

	t.Log("Retrieving value from cache")
	var result string
	err = lc.Get(context.Background(), key, &result)
	if err != nil {
		t.Fatalf("Failed to get cache item: %v", err)
	}

	t.Logf("Expected value: %s, Retrieved value: %s", value, result)

	if result != value {
		t.Errorf("Expected %s, got %s", value, result)
	}

	t.Log("Test completed successfully")
}

// func (suite *LocalCacheTestSuite) TestDelete() {
// 	ctx := context.Background()
// 	key := "testKey"
// 	value := testValue

// 	// Ensure cache is initialized before testing
// 	if suite.cache == nil {
// 		suite.T().Fatal("Cache is not initialized")
// 	}

// 	// Set a key-value pair in the cache
// 	suite.NoError(suite.cache.Set(ctx, key, value, 0))

// 	// Delete the cache item
// 	err := suite.cache.Delete(ctx, key)
// 	suite.NoError(err)

// 	// Attempt to get the value after deletion
// 	var result string
// 	err = suite.cache.Get(ctx, key, &result)

// 	// Check for cache miss error
// 	suite.Error(err, "Expected error when getting deleted key")
// 	suite.Empty(result, "Expected empty result as key is deleted")
// }

func (suite *LocalCacheTestSuite) TestCacheItemExpiration() {
	cache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: 1e7,
		MaxCost:     100,
		BufferItems: 64,
	})
	suite.NoError(err)

	cache.SetWithTTL("key", testValue, 1, 1*time.Second)
	cache.Wait()
	value, found := cache.Get("key")
	suite.True(found, "CacheItem should be available immediately after creation")
	suite.Equal(testValue, value, "CacheItem value should match")

	time.Sleep(2 * time.Second)

	_, found = cache.Get("key")
	suite.False(found, "CacheItem should be expired after its TTL")
}

func (suite *LocalCacheTestSuite) TestCacheItemNeverExpires() {
	cache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: 1e7,
		MaxCost:     100,
		BufferItems: 64,
	})
	suite.NoError(err)

	cache.SetWithTTL("key", testValue, 1, 0)

	cache.Wait()

	value, found := cache.Get("key")
	suite.True(found, "CacheItem should be available immediately after creation")
	suite.Equal(testValue, value, "CacheItem value should match")

	time.Sleep(2 * time.Second)

	value, found = cache.Get("key")
	suite.True(found, "CacheItem should not expire with TTL of 0")
	suite.Equal(testValue, value, "CacheItem value should still match")
}
func (suite *LocalCacheTestSuite) TestCleanupProcess() {
	cleanupInterval := 100 * time.Millisecond
	ttl := 300 * time.Millisecond // Use a reasonable TTL

	cache, err := NewLocalCache(&LocalConfig{
		Prefix: "test",
		CleanupProcess: CleanupProcess{
			Interval: cleanupInterval,
		},
		MetricServer: metrics.NewTaskMetrics("", "", "appname"),
	})
	suite.NoError(err)

	ctx := context.Background()
	key := "expireKey"
	value := "testValue"

	// Set the cache key with TTL
	err = cache.Set(ctx, key, value, ttl)
	suite.NoError(err)

	// First verify the key exists
	var result string
	err = cache.Get(ctx, key, &result)
	suite.NoError(err)
	suite.Equal(value, result)

	// Wait for key to expire
	time.Sleep(ttl + 500*time.Millisecond) // Give extra time for cleanup

	// Clear the result to avoid confusion
	result = ""

	// Now check for expiration
	err = cache.Get(ctx, key, &result)

	// We should get a cache miss or item expired error
	suite.Error(err, "Expected an error but got nil")
	if err != nil {
		suite.Contains([]string{"cache miss", "cache item expired"}, err.Error(),
			"Expected either 'cache miss' or 'cache item expired'")
	}
}
func (suite *LocalCacheTestSuite) TestCleanupProcessBatchLogic() {
	cleanupInterval := 100 * time.Millisecond
	testBatchSize := 5
	itemExpiration := 200 * time.Millisecond

	cache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: 1e7,
		MaxCost:     100,
		BufferItems: 64,
	})
	suite.NoError(err)

	// Set items in the cache
	for i := 0; i < testBatchSize*2; i++ {
		key := fmt.Sprintf("key%d", i)
		value := fmt.Sprintf("value%d", i)
		ok := cache.Set(key, value, itemExpiration.Milliseconds())
		suite.True(ok, "Failed to set key %s in cache", key)
	}

	// Wait for expiration + cleanup interval + additional time
	time.Sleep(itemExpiration + cleanupInterval + 2*time.Second)

	// Check that the keys have expired
	for i := 0; i < testBatchSize*2; i++ {
		key := fmt.Sprintf("key%d", i)
		_, found := cache.Get(key)
		suite.False(found, "Expected key %s to be expired, but it was found", key)
	}
}

func (suite *LocalCacheTestSuite) TestCleanupProcessItemDoesNotExpire() {
	cleanupInterval := 1 * time.Second // Cleanup interval for cache management

	cache, err := NewLocalCache(&LocalConfig{
		Prefix: "test",
		CleanupProcess: CleanupProcess{
			Interval: cleanupInterval, // Set cleanup interval
		},
		MetricServer: metrics.NewTaskMetrics("", "", "appname"),
	})
	suite.NoError(err)

	// Start the cleanup process (if it isn't already started automatically)
	cache.startCleanupProcess()

	ctx := context.Background()
	key := "permanentKey"
	value := "thisValueShouldPersist"

	// Set item with "neverExpires" TTL, meaning it should not expire
	err = cache.Set(ctx, key, value, neverExpires)
	suite.NoError(err)

	// Wait for cleanup interval (though the item should not expire)
	time.Sleep(2 * cleanupInterval)

	// Try to retrieve the non-expiring item
	var result string
	err = cache.Get(ctx, key, &result)

	// Verify the item was not deleted by cleanup and still exists
	suite.NoError(err, "Did not expect an error when retrieving a non-expiring item")
	suite.Equal(value, result, "The retrieved value should match the original value")
}
