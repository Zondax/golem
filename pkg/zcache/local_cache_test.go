package zcache

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/allegro/bigcache/v3"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
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

func (suite *LocalCacheTestSuite) SetupSuite() {
	prefix := os.Getenv("PREFIX")
	var err error
	config := LocalConfig{
		Prefix:       prefix,
		MetricServer: metrics.NewTaskMetrics("", "", "appname"),
	}
	suite.cache, err = NewLocalCache(&config)
	suite.Nil(err)
}

func (suite *LocalCacheTestSuite) TestSetAndGet() {
	ctx := context.Background()
	key := "testKey"
	value := testValue

	err := suite.cache.Set(ctx, key, value, 0)
	suite.NoError(err)

	var result string
	err = suite.cache.Get(ctx, key, &result)
	suite.NoError(err)
	suite.Equal(value, result)
}

func (suite *LocalCacheTestSuite) TestDelete() {
	ctx := context.Background()
	key := "testKey"
	value := testValue

	suite.NoError(suite.cache.Set(ctx, key, value, 0))

	err := suite.cache.Delete(ctx, key)
	suite.NoError(err)

	err = suite.cache.Get(ctx, key, new(string))
	suite.Error(err)
}

func (suite *LocalCacheTestSuite) TestCacheItemExpiration() {
	item := NewCacheItem([]byte(testValue), 1*time.Second)
	suite.False(item.IsExpired(), "CacheItem should not be expired right after creation")
	time.Sleep(2 * time.Second)

	suite.True(item.IsExpired(), "CacheItem should be expired after its TTL")
}

func (suite *LocalCacheTestSuite) TestCacheItemNeverExpires() {
	item := NewCacheItem([]byte(testValue), -1)
	suite.False(item.IsExpired(), "CacheItem with negative TTL should never expire")
	time.Sleep(2 * time.Second)

	suite.False(item.IsExpired(), "CacheItem with negative TTL should never expire, even after some time")
}

func (suite *LocalCacheTestSuite) TestCleanupProcess() {
	cleanupInterval := 1 * time.Second
	ttl := 10 * time.Millisecond

	cache, err := NewLocalCache(&LocalConfig{
		Prefix:          "test",
		CleanupInterval: cleanupInterval,
		MetricServer:    metrics.NewTaskMetrics("", "", "appname")})
	suite.NoError(err)

	ctx := context.Background()
	key := expireKey
	value := testValue

	err = cache.Set(ctx, key, value, ttl)
	suite.NoError(err)

	time.Sleep(2 * cleanupInterval)

	var result string
	err = cache.Get(ctx, key, &result)

	suite.True(errors.Is(err, bigcache.ErrEntryNotFound), "Expected 'key not found' error, but got a different error")
}

func (suite *LocalCacheTestSuite) TestCleanupProcessBatchLogic() {
	cleanupInterval := 100 * time.Millisecond
	testBatchSize := 5
	itemExpiration := 200 * time.Millisecond

	cache, err := NewLocalCache(&LocalConfig{
		Prefix:          "testBatch",
		CleanupInterval: cleanupInterval,
		MetricServer:    metrics.NewTaskMetrics("", "", "appname"),
		BatchSize:       testBatchSize,
	})
	suite.NoError(err)

	ctx := context.Background()

	for i := 0; i < testBatchSize*2; i++ {
		key := fmt.Sprintf("key%d", i)
		value := fmt.Sprintf("value%d", i)
		err = cache.Set(ctx, key, value, itemExpiration)
		suite.NoError(err)
	}

	time.Sleep(2 * time.Second)

	for i := 0; i < testBatchSize*2; i++ {
		key := fmt.Sprintf("key%d", i)
		var result string
		err = cache.Get(ctx, key, &result)

		suite.NotNil(err, "Expected an error for key: %s, but got nil", key)
		suite.True(errors.Is(err, bigcache.ErrEntryNotFound), "Expected 'ErrEntryNotFound' for key: %s, but got a different error or no error: %s", key, err.Error())
	}
}

func (suite *LocalCacheTestSuite) TestCleanupProcessItemDoesNotExpire() {
	cleanupInterval := 1 * time.Second

	cache, err := NewLocalCache(&LocalConfig{
		Prefix:          "test",
		CleanupInterval: cleanupInterval,
		MetricServer:    metrics.NewTaskMetrics("", "", "appname"),
	})
	suite.NoError(err)

	ctx := context.Background()
	key := "permanentKey"
	value := "thisValueShouldPersist"

	err = cache.Set(ctx, key, value, neverExpires)
	suite.NoError(err)

	time.Sleep(2 * cleanupInterval)

	var result string
	err = cache.Get(ctx, key, &result)

	suite.NoError(err, "Did not expect an error when retrieving a non-expiring item")
	suite.Equal(value, result, "The retrieved value should match the original value")
}

// insert 1 persistent key and 1 key with a ttl.
// after cleanup, there will be 1 key in the cache and 1 deleted expired key.
func (suite *LocalCacheTestSuite) TestCleanupProcessMetrics() {
	cleanupInterval := 1 * time.Second
	ttl := 10 * time.Millisecond

	// label:count
	expected := map[string]int{
		"resident_item_count": 1,
		"deleted_item_count":  1,
	}
	got := map[string]int{}

	tm := &metrics.MockTaskMetrics{}
	tm.On("RegisterMetric", "localCacheCleanupErrors", mock.Anything, []string{"error_type"}, mock.Anything).Once().
		Return(nil)
	tm.On("RegisterMetric", "localCacheCleanupItemCount", mock.Anything, []string{"cleanup_item_count"}, mock.Anything).Once().
		Return(nil)

	tm.On("UpdateMetric", "localCacheCleanupErrors", mock.Anything, mock.Anything).Return(nil)
	tm.On("UpdateMetric", "localCacheCleanupItemCount", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		total := args.Get(1).(float64)
		label := args.Get(2).(string)

		got[label] += int(total)

	}).Return(nil)

	cache, err := NewLocalCache(&LocalConfig{
		Prefix:          "test",
		CleanupInterval: cleanupInterval,
		MetricServer:    tm,
	})
	suite.NoError(err)

	ctx := context.Background()
	key := "permanentKey"
	value := "thisValueShouldPersist"

	err = cache.Set(ctx, key, value, ttl)
	suite.NoError(err)
	err = cache.Set(ctx, key+"2", value, neverExpires)
	suite.NoError(err)

	time.Sleep(2 * cleanupInterval)
	for k, v := range expected {
		suite.Assert().Equal(v, got[k])
	}
}
