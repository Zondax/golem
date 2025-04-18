# zcache Package

## Overview
The `zcache` package provides an abstraction layer over Redis, allowing easy integration of caching mechanisms into Go applications. It simplifies interacting with Redis by offering a common interface for various caching operations.

## Table of Contents
1. [Features](#features)
2. [Installation](#installation)
3. [Usage](#usage)
4. [Configuration](#configuration)
5. [Mocking Support](#mocking-support)

## Features
- **Unified Caching Interface**: Offers a consistent API for common caching operations, abstracting the complexity of direct Redis interactions.
- **Distributed Mutex Locks**: Supports distributed synchronization using Redis-based mutex locks, crucial for concurrent operations.
- **Extensibility**: Easy to extend with additional methods for more Redis operations.
- **Serialization and Deserialization**: Automatically handles the conversion of Go data structures to and from Redis storage formats.
- **Mocking for Testing**: Includes mock implementations for easy unit testing without a live Redis instance.
- **Connection Pool Management**: Efficiently handles Redis connection pooling.
- **Supported Operations**: Includes a variety of caching operations like Set, Get, Delete, as well as more advanced operations like Incr, Decr, and others.

---

## Installation
```bash
go get github.com/zondax/golem/pkg/zcache
```

---

## Usage Remote cache - Redis

```go
import (
    "github.com/zondax/golem/pkg/zcache"
    "context"
    "time"
)

func main() {
    config := zcache.RemoteConfig{Addr: "localhost:6379"}
    cache := zcache.NewRemoteCache(config)
    ctx := context.Background()

    // Set a value
    cache.Set(ctx, "key1", "value1", 10*time.Minute)

    // Get a value
    if value, err := cache.Get(ctx, "key1"); err == nil {
        fmt.Println("Retrieved value:", value)
    }

    // Delete a value
    cache.Delete(ctx, "key1")
}
```


## Usage Local cache - BigCache

The LocalConfig for zcache not only allows you to specify a CleanupInterval that determines how often the expired keys cleanup process will run but also includes configurations for BatchSize and ThrottleTime to optimize the cleanup process. If CleanupInterval is not set, a default value of 12 hours will be used. Both BatchSize and ThrottleTime also have default values (200 and 1 second respectively) if not explicitly set.
It's important to note that MetricServer is a mandatory configuration field in LocalConfig to facilitate the monitoring of cache operations and errors.

```go
func main() {
    config := zcache.LocalConfig{
        // CleanupInterval is optional; if omitted, a default value of 12 hours will be used
        CleanupProcess: CleanupProcess{
            BatchSize: 100000,
            Interval: 30 * time.Minute,  
            ThrottleTime: time.Second,
        },
        // HardMaxCacheSizeInMB is optional; if omitted, a default value of 512MB will be used
        HardMaxCacheSizeInMB: 1024, // Set maximum cache size to 1GB
        MetricServer: metricServer, 
    }
    
    cache, err := zcache.NewLocalCache(&config)
    if err != nil {
        // Handle error
    }
    
    ctx := context.Background()
    
    cache.Set(ctx, "key1", "value1", 10*time.Minute)
    if value, err := cache.Get(ctx, "key1"); err == nil {
        fmt.Println("Retrieved value:", value)
    }
    cache.Delete(ctx, "key1")
}

```


## Usage Combined cache - Local and Remote

```go
func main() {
    localConfig := zcache.LocalConfig{
        CleanupProcess: zcache.CleanupProcess{
            BatchSize: 100000,           // Size of cleanup batches
            Interval: 5 * time.Minute,   // Cleanup frequency
            ThrottleTime: time.Second,   // Time between batches
        },
        HardMaxCacheSizeInMB: 256,      // Local cache size limit
        MetricServer: metricServer,      // Required for monitoring
    }
    remoteConfig := zcache.RemoteConfig{Addr: "localhost:6379"}
    config := zcache.CombinedConfig{Local: localConfig, Remote: remoteConfig, isRemoteBestEffort: false}
    cache, err := zcache.NewCombinedCache(config)
    if err != nil {
        // Handle error
    }
    
    ctx := context.Background()
    
    cache.Set(ctx, "key1", "value1", 10*time.Minute)
    if value, err := cache.Get(ctx, "key1"); err == nil {
        fmt.Println("Retrieved value:", value)
    }
    cache.Delete(ctx, "key1")
}

```

--- 

## Configuration 

Configure zcache using the Config struct, which includes network settings, server address, timeouts, and other connection parameters. This struct allows you to customize the behavior of your cache and mutex instances to fit your application's needs.

```go
type Config struct {
    Addr               string        // Redis server address
    Password           string        // Redis server password
    DB                 int           // Redis database
    DialTimeout        time.Duration // Timeout for connecting to Redis
    ReadTimeout        time.Duration // Timeout for reading from Redis
    WriteTimeout       time.Duration // Timeout for writing to Redis
    PoolSize           int           // Number of connections in the pool
    MinIdleConns       int           // Minimum number of idle connections
    IdleTimeout        time.Duration // Timeout for idle connections
}
```
---

## Working with mutex

```go
func main() {
    cache := zcache.NewCache(zcache.Config{Addr: "localhost:6379"})
    mutex := cache.NewMutex("mutex_name", 2*time.Minute)

    // Acquire lock
    if err := mutex.Lock(); err != nil {
        log.Fatalf("Failed to acquire mutex: %v", err)
    }

    // Perform operations under lock
    // ...

    // Release lock
    if ok, err := mutex.Unlock(); !ok || err != nil {
        log.Fatalf("Failed to release mutex: %v", err)
    }
}
```
---

## Mocking support

Use MockZCache and MockZMutex for unit testing.

```go
func TestCacheOperation(t *testing.T) {
    mockCache := new(zcache.MockZCache)
    mockCache.On("Get", mock.Anything, "key1").Return("value1", nil)
    // Use mockCache in your tests
}

func TestSomeFunctionWithMutex(t *testing.T) {
    mockMutex := new(zcache.MockZMutex)
    mockMutex.On("Lock").Return(nil)
    mockMutex.On("Unlock").Return(true, nil)
    mockMutex.On("Name").Return("myMutex")
    
    result, err := SomeFunctionThatUsesMutex(mockMutex)
    assert.NoError(t, err)
    assert.Equal(t, expectedResult, result)
    
    mockMutex.AssertExpectations(t)
}
```

## Best Practices - Ristretto Cache

### Memory Management
When using the local cache (Ristretto), memory is managed more efficiently:

1. **Memory Control**:
   - Ristretto uses more precise memory tracking
   - Items are evicted based on both size and access patterns
   - Memory is released more aggressively when items are deleted

2. **Configuration Parameters**:
   - `HardMaxCacheSizeInMB`: Hard limit on cache size (default: 512MB)
   - `BatchSize`: Controls cleanup batch size (default: 200)
   - `ThrottleTime`: Prevents CPU spikes during cleanup (default: 1s)

3. **Cleanup Process**:
   ```go
   CleanupProcess: zcache.CleanupProcess{
       BatchSize: 100000,           // Larger batches for better efficiency
       Interval: 5 * time.Minute,   // More frequent cleanup
       ThrottleTime: time.Second,   // Prevent CPU spikes
   }
   ```

### Memory Monitoring
Monitor cache performance through Ristretto metrics:
- `zcache_local_items_count`: Current number of items
- `zcache_local_memory_usage_bytes`: Actual memory usage
- `zcache_local_cleanup_duration`: Time taken for cleanup
- `zcache_local_cleanup_items`: Items processed in cleanup
- `zcache_local_hit_ratio`: Cache hit rate

### Best Practices
1. **Memory Configuration**:
   - Always set `HardMaxCacheSizeInMB` based on available system memory
   - Use smaller values for `ThrottleTime` in low-latency scenarios
   - Adjust `BatchSize` based on item count and cleanup needs

2. **Cleanup Strategy**:
   - Use shorter cleanup intervals for frequently changing data
   - Increase batch size for large datasets
   - Monitor cleanup duration metrics

3. **Production Recommendations**:
   - Use Combined Cache with Redis for persistence
   - Monitor hit ratios to optimize local cache size
   - Set appropriate TTLs for data freshness

### Notes
- Ristretto provides better memory management than BigCache
- No known memory leak issues
- More predictable memory usage
- Better performance under high load
