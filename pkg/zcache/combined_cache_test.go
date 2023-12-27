package zcache

import (
	"context"
	"github.com/stretchr/testify/suite"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
)

func TestCombinedCacheTestSuite(t *testing.T) {
	suite.Run(t, new(CombinedCacheTestSuite))
}

type CombinedCacheTestSuite struct {
	suite.Suite
	mr    *miniredis.Miniredis
	cache CombinedCache
}

func (suite *CombinedCacheTestSuite) SetupSuite() {
	mr, err := miniredis.Run()
	suite.Require().NoError(err)
	suite.mr = mr

	remoteConfig := &RemoteConfig{
		Addr: mr.Addr(),
	}

	localConfig := &LocalConfig{
		EvictionInSeconds: 10,
	}

	config := &CombinedConfig{localConfig, remoteConfig, false}

	suite.cache, err = NewCombinedCache(config)
	suite.Nil(err)
}

func (suite *CombinedCacheTestSuite) TearDownSuite() {
	suite.mr.Close()
}

func (suite *CombinedCacheTestSuite) TestSetAndGet() {
	ctx := context.Background()
	err := suite.cache.Set(ctx, "key1", "value1", 10*time.Second)
	suite.NoError(err)

	var result string
	err = suite.cache.Get(ctx, "key1", &result)
	suite.NoError(err)
	suite.Equal("value1", result)
}

func (suite *CombinedCacheTestSuite) TestDelete() {
	ctx := context.Background()

	suite.NoError(suite.cache.Set(ctx, "key2", "value2", 10*time.Second))

	err := suite.cache.Delete(ctx, "key2")
	suite.NoError(err)

	err = suite.cache.Get(ctx, "key2", new(string))
	suite.Error(err)
}
