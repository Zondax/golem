package zcache

import (
	"context"
	"github.com/stretchr/testify/suite"
	logger2 "github.com/zondax/golem/pkg/logger"
	"github.com/zondax/golem/pkg/metrics"
	"os"
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
	cacheRemote                    RemoteCache
	ms                             metrics.TaskMetrics
}

func (suite *CombinedCacheTestSuite) SetupSuite() {
	mr, err := miniredis.Run()
	suite.Require().NoError(err)
	suite.mr = mr
	suite.ms = metrics.NewTaskMetrics("", "", "appname")
	logger := logger2.NewLogger()

	prefix := os.Getenv("PREFIX")
	suite.cacheRemoteBrokenBestEffort, err = NewCombinedCache(
		&CombinedConfig{
			Local: &LocalConfig{
				MetricServer: suite.ms,
			},
			Remote: &RemoteConfig{
				Addr: "0.0.0.0",
			},
			IsRemoteBestEffort: true,
			GlobalPrefix:       prefix,
			GlobalLogger:       logger,
		})
	suite.Nil(err)

	suite.cacheOkNotBestEffort, err = NewCombinedCache(&CombinedConfig{
		Local: &LocalConfig{
			MetricServer: suite.ms,
		},
		Remote: &RemoteConfig{
			Addr: mr.Addr(),
		},
		IsRemoteBestEffort: false,
		GlobalPrefix:       prefix,
		GlobalLogger:       logger,
	})
	suite.Nil(err)

	suite.cacheRemoteBrokenNotBestEffort, err = NewCombinedCache(
		&CombinedConfig{
			Local: &LocalConfig{
				MetricServer: suite.ms,
			},
			Remote: &RemoteConfig{
				Addr: "0.0.0.0",
			},
			IsRemoteBestEffort: false,
			GlobalPrefix:       prefix,
			GlobalLogger:       logger,
		})
	suite.Nil(err)

	suite.cacheRemote, err = NewRemoteCache(&RemoteConfig{
		Addr:   mr.Addr(),
		Logger: logger,
		Prefix: prefix,
	})
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

func (suite *CombinedCacheTestSuite) TestGetFromRemoteToLocal() {
	ctx := context.Background()

	// write value remotely directly
	err := suite.cacheRemote.Set(ctx, "onlyOnRemote", "value_on_remote", -1)
	suite.NoError(err)

	// check value on combined cache, it should not find it locally or remotely (it should fail)
	var result1 string
	err = suite.cacheOkNotBestEffort.Get(ctx, "noFound", &result1)
	suite.Error(err)
	suite.Equal(suite.cacheOkNotBestEffort.IsNotFoundError(err), true)

	// check value on combined cache, it should not find it locally, retrieve it remotely and write it back locally and remotely
	err = suite.cacheOkNotBestEffort.Get(ctx, "onlyOnRemote", &result1)
	suite.NoError(err)
	suite.Equal("value_on_remote", result1)

	// check value again on combined cache, it should find it now locally
	var result2 string
	err = suite.cacheOkNotBestEffort.Get(ctx, "onlyOnRemote", &result2)
	suite.NoError(err)
	suite.Equal("value_on_remote", result2)

	// check value directly on remote cache
	var result3 string
	err = suite.cacheRemote.Get(ctx, "onlyOnRemote", &result3)
	suite.NoError(err)
	suite.Equal("value_on_remote", result3)
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
