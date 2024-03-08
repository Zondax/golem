package zcache

import (
	"context"
	"errors"
	"github.com/allegro/bigcache/v3"
	"github.com/stretchr/testify/suite"
	"os"
	"testing"
	"time"
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
		Prefix: prefix,
	}
	suite.cache, err = NewLocalCache(&config)
	suite.Nil(err)
}

func (suite *LocalCacheTestSuite) TestSetAndGet() {
	ctx := context.Background()
	key := "testKey"
	value := "testValue"

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
	value := "testValue"

	suite.NoError(suite.cache.Set(ctx, key, value, 0))

	err := suite.cache.Delete(ctx, key)
	suite.NoError(err)

	err = suite.cache.Get(ctx, key, new(string))
	suite.Error(err)
}

func (suite *LocalCacheTestSuite) TestCacheItemExpiration() {
	item := NewCacheItem([]byte("testValue"), 1*time.Second)
	suite.False(item.IsExpired(), "CacheItem should not be expired right after creation")
	time.Sleep(2 * time.Second)

	suite.True(item.IsExpired(), "CacheItem should be expired after its TTL")
}

func (suite *LocalCacheTestSuite) TestCacheItemNeverExpires() {
	item := NewCacheItem([]byte("testValue"), -1)
	suite.False(item.IsExpired(), "CacheItem with negative TTL should never expire")
	time.Sleep(2 * time.Second)

	suite.False(item.IsExpired(), "CacheItem with negative TTL should never expire, even after some time")
}

func (suite *LocalCacheTestSuite) TestCleanupProcess() {
	cleanupInterval := 1 * time.Second
	ttl := 500 * time.Millisecond

	cache, err := NewLocalCache(&LocalConfig{Prefix: "test", CleanupInterval: cleanupInterval})
	suite.NoError(err)

	ctx := context.Background()
	key := "expireKey"
	value := "testValue"

	err = cache.Set(ctx, key, value, ttl)
	suite.NoError(err)

	time.Sleep(2 * cleanupInterval)

	var result string
	err = cache.Get(ctx, key, &result)

	suite.True(errors.Is(err, bigcache.ErrEntryNotFound), "Expected 'key not found' error, but got a different error")
}

func (suite *LocalCacheTestSuite) TestCleanupProcessItemNeverExpires() {
	cleanupInterval := 1 * time.Second
	cache, err := NewLocalCache(&LocalConfig{Prefix: "test", CleanupInterval: cleanupInterval})
	suite.NoError(err)

	ctx := context.Background()
	key := "expireKey"
	value := "testValue"

	err = cache.Set(ctx, key, value, neverExpires)
	suite.NoError(err)

	time.Sleep(2 * cleanupInterval)

	var result string
	err = cache.Get(ctx, key, &result)

	suite.True(errors.Is(err, bigcache.ErrEntryNotFound), "Expected 'key not found' error, but got a different error")
}
