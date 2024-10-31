# Logger Package

A structured logging package built on top of [uber-go/zap](https://github.com/uber-go/zap), providing both global and context-aware logging capabilities with sensible defaults.

## Features

- Built on top of the high-performance zap logger
- Supports both structured and printf-style logging
- Context-aware logging
- Global and instance-based logging
- Configurable log levels and encoding formats
- Request ID tracking support
- Easy integration with existing applications

## Installation

```go
go get -u go.uber.org/zap
```

## Quick Start

```go
// Initialize with default configuration
logger.InitLogger(logger.Config{
    Level:    "info",
    Encoding: "json",
})

// Basic logging
logger.Info("Server started")
logger.Error("Connection failed")

// Structured logging with fields
log := logger.NewLogger(
    logger.Field{Key: "service", Value: "api"},
    logger.Field{Key: "version", Value: "1.0.0"},
)
log.Info("Service initialized")
```

## Configuration

### Logger Config

```go
type Config struct {
    Level    string `json:"level"`    // Logging level
    Encoding string `json:"encoding"` // Output format
}
```

### Log Levels

Available log levels (in order of increasing severity):
- `debug`: Detailed information for debugging
- `info`: General operational information
- `warn`: Warning messages for potentially harmful situations
- `error`: Error conditions that should be addressed
- `dpanic`: Critical errors in development that cause panic
- `panic`: Critical errors that cause panic in production
- `fatal`: Fatal errors that terminate the program

### Encoding Formats

1. **JSON Format** (Default)
   - Recommended for production
   - Machine-readable structured output
   ```json
   {"level":"INFO","ts":"2024-03-20T10:00:00.000Z","msg":"Server started","service":"api"}
   ```

2. **Console Format**
   - Recommended for development
   - Human-readable output
   ```
   2024-03-20T10:00:00.000Z INFO Server started service=api
   ```

## Advanced Usage

### Context-Aware Logging

```go
// Create a context with logger
ctx := context.Background()
log := logger.NewLogger(logger.Field{
    Key: logger.RequestIDKey,
    Value: "req-123",
})
ctx = logger.ContextWithLogger(ctx, log)

// Get logger from context
contextLogger := logger.GetLoggerFromContext(ctx)
contextLogger.Info("Processing request")
```

### Structured Logging with Fields

```go
log := logger.NewLogger()
log.WithFields(
    zap.String("user_id", "12345"),
    zap.String("action", "login"),
    zap.Int("attempt", 1),
).Info("User login attempt")
```

### Printf-Style Logging

```go
logger.Infof("Processing item %d of %d", current, total)
logger.Errorf("Failed to connect to %s: %v", host, err)
```

## Best Practices

1. **Use Structured Logging**
   ```go
   // Good
   log.WithFields(
       zap.String("user_id", "12345"),
       zap.String("action", "purchase"),
       zap.Float64("amount", 99.99),
   ).Info("Purchase completed")

   // Avoid
   log.Infof("User %s completed purchase of $%.2f", userID, amount)
   ```

2. **Include Request IDs**
   ```go
   log.WithFields(
       zap.String(logger.RequestIDKey, requestID),
   ).Info("Handling request")
   ```

3. **Proper Error Logging**
   ```go
   if err != nil {
       log.WithFields(
           zap.Error(err),
           zap.String("operation", "database_query"),
       ).Error("Query failed")
   }
   ```

4. **Resource Cleanup**
   ```go
   defer logger.Sync()
   ```

## Performance Considerations

- The logger is designed to be zero-allocation in most cases
- JSON encoding is more CPU-intensive but provides structured data
- Log level checks are performed atomically
- Field allocation is optimized for minimal overhead

## Thread Safety

The logger is completely thread-safe and can be used concurrently from multiple goroutines.

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the LICENSE file for details.