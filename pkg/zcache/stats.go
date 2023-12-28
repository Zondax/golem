package zcache

import (
	"github.com/allegro/bigcache/v3"
)

type ZCacheStats struct {
	Bigcache *bigcache.Stats
	Redis    *RedisStats
}
