package zcache

import (
	"context"
	"testing"
	"time"

	"github.com/dgraph-io/ristretto"
	"github.com/stretchr/testify/suite"
	"github.com/zondax/golem/pkg/logger"
)

const (
	testValue = "testValue"
)

func TestLocalCacheTestSuite(t *testing.T) {
	suite.Run(t, new(LocalCacheTestSuite))
}

type LocalCacheTestSuite struct {
	suite.Suite
	cache LocalCache
}

func (suite *LocalCacheTestSuite) SetupTest() {
	cache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: 1e7,
		MaxCost:     1 << 20,
		BufferItems: 64,
	})
	if err != nil {
		suite.T().Fatal("Failed to create cache:", err)
	}

	suite.cache = &localCache{
		client: cache,
		prefix: "test_",
		logger: logger.NewLogger(),
	}
}

func (suite *LocalCacheTestSuite) TestDelete() {
	ctx := context.Background()
	key := "testKey"
	value := testValue

	suite.NoError(suite.cache.Set(ctx, key, value, 0))

	err := suite.cache.Delete(ctx, key)
	suite.NoError(err)

	var result string
	err = suite.cache.Get(ctx, key, &result)
	suite.Error(err, "Expected error when getting deleted key")
	suite.Empty(result, "Expected empty result as key is deleted")
}

func TestCacheSetAndGet(t *testing.T) {
	cache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: 1e7,
		MaxCost:     1 << 30,
		BufferItems: 64,
	})
	if err != nil {
		t.Fatalf("Failed to initialize cache: %v", err)
	}

	lc := &localCache{
		client: cache,
		logger: logger.NewLogger(),
	}

	key := "testKey"
	value := testValue

	err = lc.Set(context.Background(), key, value, 60*time.Second)
	if err != nil {
		t.Fatalf("Failed to set cache item: %v", err)
	}

	var result string
	err = lc.Get(context.Background(), key, &result)
	if err != nil {
		t.Fatalf("Failed to get cache item: %v", err)
	}

	if result != value {
		t.Errorf("Expected %s, got %s", value, result)
	}
}

func (suite *LocalCacheTestSuite) TestCacheItemExpiration() {
	ctx := context.Background()
	key := "expireKey"
	value := testValue

	suite.NoError(suite.cache.Set(ctx, key, value, 1*time.Second))

	var result string
	err := suite.cache.Get(ctx, key, &result)
	suite.NoError(err)
	suite.Equal(value, result)

	time.Sleep(2 * time.Second)

	err = suite.cache.Get(ctx, key, &result)
	suite.Error(err, "Expected error for expired key")
}

func (suite *LocalCacheTestSuite) TestCacheItemNeverExpires() {
	ctx := context.Background()
	key := "permanentKey"
	value := testValue

	suite.NoError(suite.cache.Set(ctx, key, value, -1))

	var result string
	err := suite.cache.Get(ctx, key, &result)
	suite.NoError(err)
	suite.Equal(value, result)

	time.Sleep(2 * time.Second)

	err = suite.cache.Get(ctx, key, &result)
	suite.NoError(err)
	suite.Equal(value, result)
}
