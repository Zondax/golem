package zcache

import (
	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/suite"
	"testing"
	"time"
)

func TestZMutexTestSuite(t *testing.T) {
	suite.Run(t, new(ZMutexTestSuite))
}

type ZMutexTestSuite struct {
	suite.Suite
	mr    *miniredis.Miniredis
	cache *redisCache
}

func (suite *ZMutexTestSuite) SetupSuite() {
	mr, err := miniredis.Run()
	suite.Require().NoError(err)
	suite.mr = mr

	config := &Config{
		Addr: mr.Addr(),
	}

	suite.cache = NewCache(config).(*redisCache)
}

func (suite *ZMutexTestSuite) TearDownSuite() {
	suite.mr.Close()
}

func (suite *ZMutexTestSuite) TestNewMutex() {
	mutexName := "testMutex"
	expiry := 10 * time.Second

	mutex := suite.cache.NewMutex(mutexName, expiry)
	suite.NotNil(mutex)
	suite.Equal(mutexName, mutex.Name())
}

func (suite *ZMutexTestSuite) TestMutexLockUnlock() {
	mutex := suite.cache.NewMutex("testMutex", 10*time.Second)
	err := mutex.Lock()
	suite.NoError(err)

	unlocked, err := mutex.Unlock()
	suite.NoError(err)
	suite.True(unlocked)
}
