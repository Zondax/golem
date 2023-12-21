package zcache

import (
	"context"
	"github.com/allegro/bigcache/v3"
	"github.com/stretchr/testify/suite"
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
	bigCacheConfig := bigcache.DefaultConfig(10 * time.Minute)
	client, err := bigcache.New(context.Background(), bigCacheConfig)
	suite.Require().NoError(err)

	suite.cache = &localCache{client: client}
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
