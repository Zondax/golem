package zcache

import (
	"context"
	"github.com/stretchr/testify/suite"
	"os"
	"testing"
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
