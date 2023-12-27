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
	mr                             *miniredis.Miniredis
	cacheRemoteBrokenBestEffort    CombinedCache
	cacheRemoteBrokenNotBestEffort CombinedCache
	cacheOkNotBestEffort           CombinedCache
}

func (suite *CombinedCacheTestSuite) SetupSuite() {
	mr, err := miniredis.Run()
	suite.Require().NoError(err)
	suite.mr = mr

	suite.cacheRemoteBrokenBestEffort, err = NewCombinedCache(&CombinedConfig{Local: &LocalConfig{
		EvictionInSeconds: 10,
	}, Remote: &RemoteConfig{
		Addr: "0.0.0.0",
	}, isRemoteBestEffort: true})

	suite.cacheOkNotBestEffort, err = NewCombinedCache(&CombinedConfig{Local: &LocalConfig{
		EvictionInSeconds: 10,
	}, Remote: &RemoteConfig{
		Addr: mr.Addr(),
	}, isRemoteBestEffort: false})

	suite.cacheRemoteBrokenNotBestEffort, err = NewCombinedCache(&CombinedConfig{Local: &LocalConfig{
		EvictionInSeconds: 10,
	}, Remote: &RemoteConfig{
		Addr: "0.0.0.0",
	}, isRemoteBestEffort: false})
	suite.Nil(err)
}

func (suite *CombinedCacheTestSuite) TearDownSuite() {
	suite.mr.Close()
}

func (suite *CombinedCacheTestSuite) TestSetAndGet() {
	ctx := context.Background()

	err := suite.cacheRemoteBrokenBestEffort.Set(ctx, "key1", "value1", 10*time.Second)
	suite.NoError(err)

	var result1 string
	err = suite.cacheRemoteBrokenBestEffort.Get(ctx, "key1", &result1)
	suite.NoError(err)
	suite.Equal("value1", result1)

	err = suite.cacheOkNotBestEffort.Set(ctx, "key1", "value1", 10*time.Second)
	suite.NoError(err)

	var result2 string
	err = suite.cacheOkNotBestEffort.Get(ctx, "key1", &result2)
	suite.NoError(err)
	suite.Equal("value1", result2)

	err = suite.cacheRemoteBrokenNotBestEffort.Set(ctx, "key1", "value1", 10*time.Second)
	suite.Error(err)

	var result3 string
	err = suite.cacheRemoteBrokenNotBestEffort.Get(ctx, "key1", &result3)
	suite.Error(err)
	suite.Equal("", result3)
}

func (suite *CombinedCacheTestSuite) TestDelete() {
	ctx := context.Background()

	suite.NoError(suite.cacheRemoteBrokenBestEffort.Set(ctx, "key2", "value2", 10*time.Second))

	err := suite.cacheRemoteBrokenBestEffort.Delete(ctx, "key2")
	suite.NoError(err)

	err = suite.cacheRemoteBrokenBestEffort.Get(ctx, "key2", new(string))
	suite.Error(err)

	suite.NoError(suite.cacheOkNotBestEffort.Set(ctx, "key2", "value2", 10*time.Second))

	err = suite.cacheOkNotBestEffort.Delete(ctx, "key2")
	suite.NoError(err)

	err = suite.cacheOkNotBestEffort.Get(ctx, "key2", new(string))
	suite.Error(err)

	err = suite.cacheRemoteBrokenNotBestEffort.Delete(ctx, "key2")
	suite.Error(err)
}
