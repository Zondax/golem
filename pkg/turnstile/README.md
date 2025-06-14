# Turnstile Package

A Go package for verifying Cloudflare Turnstile tokens. This package provides a clean, testable interface for integrating Turnstile verification into your Go applications.

## Features

- ✅ Clean interface design with dependency injection
- ✅ Configurable HTTP client with timeout support
- ✅ Comprehensive error handling
- ✅ Full test coverage with standardized testing patterns
- ✅ Context support for request cancellation
- ✅ Zero hardcoded values - everything is configurable via constants
- ✅ Uses standard Go `*http.Client` - no custom interfaces

## Installation

```bash
go get github.com/zondax/golem/pkg/turnstile
```

## Quick Start

### Basic Usage

```go
package main

import (
    "context"
    "log"
    
    "github.com/zondax/golem/pkg/turnstile"
)

func main() {
    // Create a service with basic configuration
    config := turnstile.Config{
        SecretKey: "your-secret-key",
        Endpoint:  "https://challenges.cloudflare.com/turnstile/v0/siteverify",
    }
    
    service := turnstile.NewService(config)
    
    // Verify a token
    err := service.Verify(context.Background(), "user-token")
    if err != nil {
        log.Printf("Verification failed: %v", err)
        return
    }
    
    log.Println("Token verified successfully!")
}
```

### Using Default Configuration

```go
// Start with sensible defaults
config := turnstile.DefaultConfig()
config.SecretKey = "your-secret-key"
config.Endpoint = "https://challenges.cloudflare.com/turnstile/v0/siteverify"

service := turnstile.NewService(config)
```

### Custom HTTP Client

```go
import (
    "net/http"
    "time"
)

// Use a custom HTTP client with specific timeout and transport settings
customClient := &http.Client{
    Timeout: 10 * time.Second,
    Transport: &http.Transport{
        MaxIdleConns:        100,
        MaxIdleConnsPerHost: 10,
    },
}

config := turnstile.Config{
    SecretKey:  "your-secret-key",
    Endpoint:   "https://challenges.cloudflare.com/turnstile/v0/siteverify",
    HTTPClient: customClient,
}

service := turnstile.NewService(config)
```

### With Context and Timeout

```go
import (
    "context"
    "time"
)

// Create a context with timeout for the verification request
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

err := service.Verify(ctx, "user-token")
if err != nil {
    // Handle timeout or other errors
    log.Printf("Verification failed: %v", err)
}
```

## Configuration

### Config Structure

```go
type Config struct {
    SecretKey  string        // Your Turnstile secret key (required)
    Endpoint   string        // Turnstile verification endpoint (required)
    HTTPClient *http.Client  // Custom HTTP client (optional)
    Timeout    time.Duration // Request timeout (optional, default: 30s)
}
```

### Default Values

- **Timeout**: 30 seconds
- **HTTPClient**: `&http.Client{Timeout: 30 * time.Second}`

### Constants

The package defines all field names and headers as constants to avoid hardcoded values:

```go
const (
    // Form field names for Turnstile API
    FieldSecret   = "secret"
    FieldResponse = "response"
    
    // HTTP headers
    HeaderContentType = "Content-Type"
    
    // Default timeout for HTTP requests
    DefaultTimeout = 30 * time.Second
)
```

## Error Handling

The package provides detailed error messages for different failure scenarios:

```go
err := service.Verify(ctx, token)
if err != nil {
    switch {
    case strings.Contains(err.Error(), "failed to make request"):
        // Network error
        log.Println("Network connectivity issue")
    case strings.Contains(err.Error(), "unexpected status code"):
        // HTTP error from Turnstile API
        log.Println("Turnstile API returned an error")
    case strings.Contains(err.Error(), "turnstile verification failed"):
        // Token verification failed
        log.Println("Invalid token provided")
    case strings.Contains(err.Error(), "context canceled"):
        // Request was canceled
        log.Println("Request timed out or was canceled")
    default:
        // Other errors (parsing, etc.)
        log.Printf("Unexpected error: %v", err)
    }
}
```

## Testing

The package follows Go testing best practices and provides comprehensive test coverage.

### Running Tests

```bash
# Run all tests
go test ./pkg/turnstile

# Run tests with coverage
go test -cover ./pkg/turnstile

# Run tests with verbose output
go test -v ./pkg/turnstile
```

### Testing Patterns Used

- **Arrange-Act-Assert**: Clear test structure
- **Table-driven tests**: For testing multiple scenarios
- **Mock servers**: Using `httptest.NewServer` for HTTP testing
- **Helper functions**: Reducing test code duplication
- **Cleanup**: Proper resource cleanup with `t.Cleanup()`
- **Constants testing**: Verification of all defined constants

### Example Test

```go
func TestService_Verify_Success(t *testing.T) {
    // Arrange
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Verify request format
        assert.Equal(t, http.MethodPost, r.Method)
        assert.Contains(t, r.Header.Get(turnstile.HeaderContentType), "multipart/form-data")
        
        // Verify form fields using constants
        err := r.ParseMultipartForm(10 << 20)
        require.NoError(t, err)
        assert.Equal(t, "test-secret", r.FormValue(turnstile.FieldSecret))
        assert.Equal(t, "valid-token", r.FormValue(turnstile.FieldResponse))
        
        // Return success response
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(map[string]interface{}{
            "success": true,
        })
    }))
    defer server.Close()
    
    config := turnstile.Config{
        SecretKey: "test-secret",
        Endpoint:  server.URL,
    }
    service := turnstile.NewService(config)
    
    // Act
    err := service.Verify(context.Background(), "valid-token")
    
    // Assert
    assert.NoError(t, err)
}
```

## Interface Design

The package uses interface-based design for better testability:

```go
// Service interface for easy mocking
type Service interface {
    Verify(ctx context.Context, token string) error
}
```

### Mocking for Tests

```go
// Mock the Service interface
type MockService struct {
    mock.Mock
}

func (m *MockService) Verify(ctx context.Context, token string) error {
    args := m.Called(ctx, token)
    return args.Error(0)
}

// Use in tests
mockService := &MockService{}
mockService.On("Verify", mock.Anything, "valid-token").Return(nil)
```

## Best Practices

1. **Always use context**: Pass context for request cancellation and timeout control
2. **Handle errors appropriately**: Check for different error types and handle accordingly
3. **Configure timeouts**: Set appropriate timeouts for your use case
4. **Use dependency injection**: Inject custom HTTP clients when needed
5. **Test thoroughly**: Use the provided testing patterns for your integration tests
6. **Use constants**: Reference the exported constants for field names and headers

## Architecture Decisions

### Why `*http.Client` instead of custom interface?

- **Simplicity**: Uses the standard Go HTTP client directly
- **Compatibility**: Works with any existing `*http.Client` configuration
- **No abstractions**: Avoids unnecessary interface layers
- **Ecosystem**: Compatible with HTTP middleware and tooling

### Why constants for everything?

- **Maintainability**: Single source of truth for field names and headers
- **Refactoring safety**: Changes are compile-time checked
- **Testing**: Constants can be verified in tests
- **Documentation**: Self-documenting code

## Contributing

When contributing to this package:

1. Follow the existing code style and patterns
2. Add comprehensive tests for new features
3. Update documentation for any API changes
4. Ensure all tests pass before submitting
5. Use the defined constants instead of hardcoded strings

## License

This package is part of the Zondax Golem project.
