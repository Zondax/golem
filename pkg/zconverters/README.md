# ZConverters Package

Safe type conversion utilities for Go applications.

## Installation

```bash
go get github.com/zondax/golem/pkg/zconverters
```

## Functions

### IntToUInt64(i int) (uint64, error)
Converts `int` to `uint64`. Returns error if negative.

```go
result, err := zconverters.IntToUInt64(42)  // result = 42, err = nil
result, err := zconverters.IntToUInt64(-1)  // result = 0, err = "cannot convert negative int to uint64"
```

### IntToUInt(i int) (uint, error)
Converts `int` to `uint`. Returns error if negative.

```go
result, err := zconverters.IntToUInt(42)   // result = 42, err = nil
result, err := zconverters.IntToUInt(-1)   // result = 0, err = "cannot convert negative int to uint"
```

### Int64ToUint64(value int64) uint64
Converts `int64` to `uint64`. Returns `0` for negative values (no error).

```go
result := zconverters.Int64ToUint64(42)   // result = 42
result := zconverters.Int64ToUint64(-1)   // result = 0
```

### IntToInt32(value int) int32
Converts `int` to `int32`. Caps at `MaxInt32`/`MinInt32` boundaries.

```go
result := zconverters.IntToInt32(42)              // result = 42
result := zconverters.IntToInt32(math.MaxInt64)   // result = 2147483647 (MaxInt32)
```

## Usage

```go
package main

import (
    "fmt"
    "github.com/zondax/golem/pkg/zconverters"
)

func main() {
    // With error handling
    if result, err := zconverters.IntToUInt64(len(slice)); err == nil {
        fmt.Printf("Length: %d\n", result)
    }
    
    // Safe conversion (no errors)
    safeResult := zconverters.IntToInt32(largeNumber)
}
```

## Performance

All functions have zero allocations and sub-nanosecond performance:
- `0.25 ns/op, 0 allocs/op`

## Testing

```bash
go test -cover ./pkg/zconverters  # 100% coverage
go test -bench=. ./pkg/zconverters
