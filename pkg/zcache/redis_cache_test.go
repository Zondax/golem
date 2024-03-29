package zcache

import (
	"context"
	"github.com/stretchr/testify/suite"
	"os"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
)

func TestRedisTestSuite(t *testing.T) {
	suite.Run(t, new(RedisCacheTestSuite))
}

type RedisCacheTestSuite struct {
	suite.Suite
	mr    *miniredis.Miniredis
	cache RemoteCache
}

func (suite *RedisCacheTestSuite) SetupSuite() {
	mr, err := miniredis.Run()
	suite.Require().NoError(err)
	suite.mr = mr

	prefix := os.Getenv("PREFIX")
	config := &RemoteConfig{
		Addr:   mr.Addr(),
		Prefix: prefix,
	}

	suite.cache, err = NewRemoteCache(config)
	suite.Nil(err)
}

func (suite *RedisCacheTestSuite) TearDownSuite() {
	suite.mr.Close()
}

func (suite *RedisCacheTestSuite) TestSetAndGet() {
	ctx := context.Background()
	err := suite.cache.Set(ctx, "key1", "value1", 10*time.Second)
	suite.NoError(err)

	var result string
	err = suite.cache.Get(ctx, "key1", &result)
	suite.NoError(err)
	suite.Equal("value1", result)
}

func (suite *RedisCacheTestSuite) TestDelete() {
	ctx := context.Background()

	suite.NoError(suite.cache.Set(ctx, "key2", "value2", 10*time.Second))

	err := suite.cache.Delete(ctx, "key2")
	suite.NoError(err)

	err = suite.cache.Get(ctx, "key2", new(string))
	suite.Error(err)
}

func (suite *RedisCacheTestSuite) TestExists() {
	ctx := context.Background()

	suite.NoError(suite.cache.Set(ctx, "key3", "value3", 10*time.Second))
	suite.NoError(suite.cache.Set(ctx, "key4", "value4", 10*time.Second))

	count, err := suite.cache.Exists(ctx, "key3", "key4", "nonExistingKey")
	suite.NoError(err)
	suite.Equal(int64(2), count)
}

func (suite *RedisCacheTestSuite) TestIncrDecr() {
	ctx := context.Background()
	key := "counterKey"

	suite.NoError(suite.cache.Set(ctx, key, 0, 10*time.Second))

	newValue, err := suite.cache.Incr(ctx, key)
	suite.NoError(err)
	suite.Equal(int64(1), newValue)

	newValue, err = suite.cache.Decr(ctx, key)
	suite.NoError(err)
	suite.Equal(int64(0), newValue)
}

func (suite *RedisCacheTestSuite) TestFlushAll() {
	ctx := context.Background()
	suite.NoError(suite.cache.Set(ctx, "key5", "value5", 10*time.Second))

	err := suite.cache.FlushAll(ctx)
	suite.NoError(err)

	count, err := suite.cache.Exists(ctx, "key5")
	suite.NoError(err)
	suite.Equal(int64(0), count)
}

func (suite *RedisCacheTestSuite) TestLPushAndRPush() {
	ctx := context.Background()
	listKey := "listKey"

	lLen, err := suite.cache.LPush(ctx, listKey, "value6")
	suite.NoError(err)
	suite.Equal(int64(1), lLen)

	rLen, err := suite.cache.RPush(ctx, listKey, "value7")
	suite.NoError(err)
	suite.Equal(int64(2), rLen)
}

func (suite *RedisCacheTestSuite) TestSMembersAndSAdd() {
	ctx := context.Background()
	setKey := "setKey"

	addCount, err := suite.cache.SAdd(ctx, setKey, "member1", "member2")
	suite.NoError(err)
	suite.Equal(int64(2), addCount)

	members, err := suite.cache.SMembers(ctx, setKey)
	suite.NoError(err)
	suite.Contains(members, "member1")
	suite.Contains(members, "member2")
}

func (suite *RedisCacheTestSuite) TestHSetAndHGet() {
	ctx := context.Background()
	hashKey := "hashKey"

	hSetCount, err := suite.cache.HSet(ctx, hashKey, "field1", "value8")
	suite.NoError(err)
	suite.Equal(int64(1), hSetCount)

	value, err := suite.cache.HGet(ctx, hashKey, "field1")
	suite.NoError(err)
	suite.Equal("value8", value)
}
