package zcache

import (
	"context"
	"os"
	"testing"
	"time"

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

func (suite *LocalCacheTestSuite) TestCleanupProcess() {
	cleanupInterval := 1 * time.Second
	ttl := 10 * time.Millisecond

	cache, err := NewLocalCache(&LocalConfig{
		Prefix: "test",
		CleanupProcess: CleanupProcess{
			Interval: cleanupInterval,
		},
		MetricServer: metrics.NewTaskMetrics("", "", "appname"),
	})
	suite.NoError(err)

	ctx := context.Background()
	key := expireKey
	value := testValue

	err = cache.Set(ctx, key, value, ttl)
	suite.NoError(err)

	// Use polling to check for key expiration
	expired := false
	maxWaitTime := 10 * cleanupInterval       // Maximum wait time
	pollingInterval := 100 * time.Millisecond // Polling interval
	timeout := time.After(maxWaitTime)
	tick := time.Tick(pollingInterval)

	for !expired {
		select {
		case <-timeout:
			suite.FailNow("Timeout reached, key did not expire as expected")
			return
		case <-tick:
			var result string
			err = cache.Get(ctx, key, &result)
			if err != nil {
				suite.FailNow("Unexpected error during cache get: %v", err)
				return
			}
			// If result is an empty string, this indicates a cache miss
			if result == "" {
				expired = true
			}
		}
	}

	// Ensure the key has expired as expected
	suite.True(expired, "Key should have expired")
}
