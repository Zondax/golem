package domain

import "time"

type CacheConfig struct {
	Paths map[string]time.Duration
}
