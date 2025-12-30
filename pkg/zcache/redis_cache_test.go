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

func (suite *RedisCacheTestSuite) TestIncrByDecrBy() {
	ctx := context.Background()
	key := "incrByKey"

	suite.NoError(suite.cache.Set(ctx, key, 10, 10*time.Second))

	newValue, err := suite.cache.IncrBy(ctx, key, 5)
	suite.NoError(err)
	suite.Equal(int64(15), newValue)

	newValue, err = suite.cache.DecrBy(ctx, key, 3)
	suite.NoError(err)
	suite.Equal(int64(12), newValue)
}

func (suite *RedisCacheTestSuite) TestHIncrBy() {
	ctx := context.Background()
	hashKey := "hincrByHash"

	// Set initial value
	_, err := suite.cache.HSet(ctx, hashKey, "counter", "10")
	suite.NoError(err)

	// Increment
	newValue, err := suite.cache.HIncrBy(ctx, hashKey, "counter", 5)
	suite.NoError(err)
	suite.Equal(int64(15), newValue)

	// Increment again
	newValue, err = suite.cache.HIncrBy(ctx, hashKey, "counter", -3)
	suite.NoError(err)
	suite.Equal(int64(12), newValue)
}

func (suite *RedisCacheTestSuite) TestHSetNX() {
	ctx := context.Background()
	hashKey := "hsetnxHash"

	// First set should succeed
	set, err := suite.cache.HSetNX(ctx, hashKey, "field", "value1")
	suite.NoError(err)
	suite.True(set)

	// Second set should fail (field exists)
	set, err = suite.cache.HSetNX(ctx, hashKey, "field", "value2")
	suite.NoError(err)
	suite.False(set)
}

func (suite *RedisCacheTestSuite) TestHExists() {
	ctx := context.Background()
	hashKey := "hexistsHash"

	// Set a field
	_, err := suite.cache.HSet(ctx, hashKey, "existingField", "value")
	suite.NoError(err)

	// Check existing field
	exists, err := suite.cache.HExists(ctx, hashKey, "existingField")
	suite.NoError(err)
	suite.True(exists)

	// Check non-existing field
	exists, err = suite.cache.HExists(ctx, hashKey, "nonExistingField")
	suite.NoError(err)
	suite.False(exists)
}

func (suite *RedisCacheTestSuite) TestHGetAll() {
	ctx := context.Background()
	hashKey := "hgetallHash"

	// Set multiple fields
	_, err := suite.cache.HSet(ctx, hashKey, "field1", "value1", "field2", "value2")
	suite.NoError(err)

	// Get all
	result, err := suite.cache.HGetAll(ctx, hashKey)
	suite.NoError(err)
	suite.Len(result, 2)
	suite.Equal("value1", result["field1"])
	suite.Equal("value2", result["field2"])
}

func (suite *RedisCacheTestSuite) TestKeys() {
	ctx := context.Background()

	// Set multiple keys with pattern
	suite.NoError(suite.cache.Set(ctx, "pattern:key1", "value1", 10*time.Second))
	suite.NoError(suite.cache.Set(ctx, "pattern:key2", "value2", 10*time.Second))
	suite.NoError(suite.cache.Set(ctx, "other:key3", "value3", 10*time.Second))

	// Find keys matching pattern
	keys, err := suite.cache.Keys(ctx, "pattern:*")
	suite.NoError(err)
	suite.Len(keys, 2)

	// Verify the keys are returned without prefix
	// The returned keys should be the original keys (pattern:key1, pattern:key2)
	suite.Contains(keys, "pattern:key1")
	suite.Contains(keys, "pattern:key2")

	// Verify other keys are not included
	suite.NotContains(keys, "other:key3")
}

func (suite *RedisCacheTestSuite) TestDeleteMulti() {
	ctx := context.Background()

	// Set multiple keys
	suite.NoError(suite.cache.Set(ctx, "delMulti1", "value1", 10*time.Second))
	suite.NoError(suite.cache.Set(ctx, "delMulti2", "value2", 10*time.Second))
	suite.NoError(suite.cache.Set(ctx, "delMulti3", "value3", 10*time.Second))

	// Delete multiple keys
	err := suite.cache.DeleteMulti(ctx, "delMulti1", "delMulti2")
	suite.NoError(err)

	// Verify deletion
	count, err := suite.cache.Exists(ctx, "delMulti1", "delMulti2", "delMulti3")
	suite.NoError(err)
	suite.Equal(int64(1), count) // Only delMulti3 should exist
}

func (suite *RedisCacheTestSuite) TestPipeline() {
	ctx := context.Background()

	// Create pipeline
	pipe := suite.cache.Pipeline()

	// Queue commands
	incrCmd := pipe.IncrBy(ctx, "pipelineCounter", 5)
	hsetCmd := pipe.HSet(ctx, "pipelineHash", "field", "value")

	// Execute
	_, err := pipe.Exec(ctx)
	suite.NoError(err)

	// Check results
	suite.Equal(int64(5), incrCmd.Val())
	suite.Equal(int64(1), hsetCmd.Val())
}

func (suite *RedisCacheTestSuite) TestPipelineHSetNX() {
	ctx := context.Background()

	// Create pipeline
	pipe := suite.cache.Pipeline()

	// Queue HSetNX commands - first should succeed, second should fail
	hsetnxCmd1 := pipe.HSetNX(ctx, "pipelineHSetNXHash", "field1", "value1")
	hsetnxCmd2 := pipe.HSetNX(ctx, "pipelineHSetNXHash", "field1", "value2") // Same field, should not set

	// Execute
	_, err := pipe.Exec(ctx)
	suite.NoError(err)

	// Check results - first set succeeds, second fails
	suite.True(hsetnxCmd1.Val())
	suite.False(hsetnxCmd2.Val())

	// Verify the value is still the first one
	value, err := suite.cache.HGet(ctx, "pipelineHSetNXHash", "field1")
	suite.NoError(err)
	suite.Equal("value1", value)
}

func (suite *RedisCacheTestSuite) TestClient() {
	// Test that we can get the underlying client
	client := suite.cache.Client()
	suite.NotNil(client)

	// Test that the client works
	ctx := context.Background()
	pong, err := client.Ping(ctx).Result()
	suite.NoError(err)
	suite.Equal("PONG", pong)
}
