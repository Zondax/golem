# Clock Package

A simple and testable time abstraction package for Go applications.

## Overview

The `clock` package provides a clean abstraction over Go's `time` package, making it easier to write testable code that depends on time operations. By using an interface-based approach, you can easily mock time operations in your tests while using the real system time in production.

## Features

- ✅ **Simple Interface**: Clean abstraction with minimal API surface
- ✅ **Testable**: Easy to mock for unit testing
- ✅ **Production Ready**: Uses actual system time in production
- ✅ **Zero Dependencies**: Only depends on Go's standard library
- ✅ **Mock Generation**: Includes generated mocks using `testify/mock`

## Installation

```bash
go get github.com/zondax/golem/pkg/clock
```

## Usage

### Basic Usage

```go
package main

import (
    "fmt"
    "github.com/zondax/golem/pkg/clock"
)

func main() {
    // Create a new clock instance
    clk := clock.New()
    
    // Get current time
    now := clk.Now()
    fmt.Printf("Current time: %v\n", now)
}
```

### Dependency Injection

```go
type Service struct {
    clock clock.Clock
}

func NewService(clk clock.Clock) *Service {
    return &Service{
        clock: clk,
    }
}

func (s *Service) ProcessWithTimestamp() {
    timestamp := s.clock.Now()
    // Process with timestamp...
}

// In production
func main() {
    clk := clock.New()
    service := NewService(clk)
    service.ProcessWithTimestamp()
}
```

### Testing with Mocks

```go
package main

import (
    "testing"
    "time"
    
    "github.com/stretchr/testify/assert"
    "github.com/zondax/golem/pkg/clock"
)

func TestServiceWithMockClock(t *testing.T) {
    // Create mock clock
    mockClock := clock.NewMockClock(t)
    
    // Set up expectations
    fixedTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
    mockClock.EXPECT().Now().Return(fixedTime).Once()
    
    // Create service with mock
    service := NewService(mockClock)
    
    // Test with predictable time
    result := service.ProcessWithTimestamp()
    
    // Verify behavior with known time
    assert.Equal(t, fixedTime, result.Timestamp)
}
```

## API Reference

### Interface

```go
type Clock interface {
    // Now returns the current time
    Now() time.Time
}
```

### Functions

#### `New() Clock`

Creates a new clock instance that uses the actual system time.

**Returns:**
- `Clock`: A clock implementation using `time.Now()`

**Example:**
```go
clk := clock.New()
currentTime := clk.Now()
```

### Mock

The package includes a generated mock (`MockClock`) that implements the `Clock` interface using `testify/mock`.

#### `NewMockClock(t *testing.T) *MockClock`

Creates a new mock clock for testing.

**Parameters:**
- `t *testing.T`: The testing context

**Returns:**
- `*MockClock`: A mock implementation of the Clock interface

**Example:**
```go
func TestWithMockClock(t *testing.T) {
    mockClock := clock.NewMockClock(t)
    
    // Set up expectations
    expectedTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
    mockClock.EXPECT().Now().Return(expectedTime).Times(2)
    
    // Use mock in tests
    assert.Equal(t, expectedTime, mockClock.Now())
    assert.Equal(t, expectedTime, mockClock.Now())
}
```

## Testing Patterns

### Fixed Time Testing

```go
func TestWithFixedTime(t *testing.T) {
    mockClock := clock.NewMockClock(t)
    fixedTime := time.Date(2024, 6, 14, 12, 0, 0, 0, time.UTC)
    
    mockClock.EXPECT().Now().Return(fixedTime).Once()
    
    service := NewService(mockClock)
    result := service.GetTimestamp()
    
    assert.Equal(t, fixedTime, result)
}
```

### Time Progression Testing

```go
func TestWithTimeProgression(t *testing.T) {
    mockClock := clock.NewMockClock(t)
    baseTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
    
    // Set up time progression
    mockClock.EXPECT().Now().Return(baseTime).Once()
    mockClock.EXPECT().Now().Return(baseTime.Add(time.Hour)).Once()
    mockClock.EXPECT().Now().Return(baseTime.Add(2 * time.Hour)).Once()
    
    service := NewService(mockClock)
    
    // Test time progression
    t1 := service.GetTimestamp()
    t2 := service.GetTimestamp()
    t3 := service.GetTimestamp()
    
    assert.Equal(t, baseTime, t1)
    assert.Equal(t, baseTime.Add(time.Hour), t2)
    assert.Equal(t, baseTime.Add(2*time.Hour), t3)
}
```

## Best Practices

### 1. Use Dependency Injection

Always inject the clock as a dependency rather than creating it directly in your business logic:

```go
// ✅ Good - Testable
type Service struct {
    clock clock.Clock
}

func NewService(clk clock.Clock) *Service {
    return &Service{clock: clk}
}

// ❌ Bad - Hard to test
type Service struct{}

func (s *Service) Process() {
    now := time.Now() // Hard to mock
}
```

### 2. Use Interface Types

Always use the `clock.Clock` interface type in your function signatures:

```go
// ✅ Good - Flexible
func ProcessData(clk clock.Clock, data []byte) error {
    timestamp := clk.Now()
    // ...
}

// ❌ Bad - Tightly coupled
func ProcessData(clk *clock.clock, data []byte) error {
    // ...
}
```

### 3. Set Clear Mock Expectations

Always set up your mock expectations clearly and verify they're met:

```go
func TestProcess(t *testing.T) {
    mockClock := clock.NewMockClock(t)
    expectedTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
    
    // Clear expectation
    mockClock.EXPECT().Now().Return(expectedTime).Once()
    
    result := ProcessData(mockClock, []byte("test"))
    
    assert.NoError(t, result)
    // Mock expectations are automatically verified by testify
}
```

## File Structure

```
pkg/clock/
├── README.md           # This documentation
├── clock.go           # Main implementation
├── clock_mock.go      # Generated mock (DO NOT EDIT)
└── clock_test.go      # Unit tests
```

## Mock Generation

The mock is generated using [mockery](https://github.com/vektra/mockery). To regenerate the mock:

```bash
# Install mockery if not already installed
go install github.com/vektra/mockery/v2@latest

# Generate mocks (run from project root)
mockery --name=Clock --dir=pkg/clock --output=pkg/clock --filename=clock_mock.go
```

## Contributing

When contributing to this package:

1. **Keep it simple**: The clock package should remain minimal and focused
2. **Maintain backward compatibility**: Any changes should not break existing code
3. **Update tests**: Ensure all tests pass and add new tests for new functionality
4. **Update documentation**: Keep this README up to date with any changes
5. **Don't edit generated files**: Never manually edit `clock_mock.go`

## License

This package is part of the Zondax Golem project and follows the same licensing terms.
