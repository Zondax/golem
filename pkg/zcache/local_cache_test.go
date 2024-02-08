package zcache

import (
	"context"
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
