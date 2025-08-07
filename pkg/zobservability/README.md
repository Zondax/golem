# Observability Package

A comprehensive observability system for Go applications providing unified tracing, metrics, and error tracking with support for multiple providers including OpenTelemetry, SigNoz, and Sentry.

## Table of Contents

- [Architecture](#architecture)
- [Features](#features)
- [Quick Start](#quick-start)
- [Configuration](#configuration)
- [Providers](#providers)
- [Metrics System](#metrics-system)
- [Tracing & Error Tracking](#tracing--error-tracking)
- [Usage Examples](#usage-examples)
- [Best Practices](#best-practices)
- [Troubleshooting](#troubleshooting)

## Architecture

### Core Components

```
zobservability/
├── observer.go              # Main Observer interface
├── config.go               # Configuration structures
├── constants.go            # Provider and attribute constants
├── metrics.go              # Metrics interfaces
├── metrics_config.go       # Metrics configuration
├── metrics_factory.go      # Metrics provider factory
├── metrics_opentelemetry.go # OpenTelemetry metrics implementation
├── metrics_noop.go         # No-op metrics for testing
├── helpers.go              # Utility functions
├── transaction.go          # Transaction interface
├── span.go                 # Span interface
├── event.go               # Event interface
├── noop.go                # No-op observer implementation
├── factory/               # Observer factory
│   └── factory.go
├── providers/             # Provider implementations
│   ├── signoz/
│   └── sentry/
└── examples/              # Usage examples
    └── metrics_example.go
```

### Key Interfaces

1. **Observer** - Main interface for all observability operations
2. **MetricsProvider** - Interface for metrics operations (counters, gauges, histograms)
3. **Transaction** - Interface for distributed tracing transactions
4. **Span** - Interface for tracing spans
5. **Event** - Interface for error and message capture

## Features

### Unified Observability
- Single interface for tracing, metrics, and error tracking
- Provider-agnostic design with pluggable backends
- Consistent API across all providers

### Multiple Providers
- **OpenTelemetry** - Industry standard for metrics and tracing
- **SigNoz** - Open-source observability platform
- **Sentry** - Error tracking and performance monitoring
- **No-op** - For testing and development

### Flexible Metrics
- **Push Mode** - Automatic periodic export to backends
- **Endpoint Mode** - On-demand metrics via HTTP endpoints
- Support for Counters, Gauges, and Histograms
- Custom labels and dimensions

### Advanced Configuration
- Environment-specific settings
- Custom headers and authentication
- Batch processing configuration
- Resource attribute customization

### Tracing Exclusions
- Selective tracing with method exclusion support
- Performance optimization for high-frequency endpoints
- Reduce noise from health checks and metrics endpoints

### External API Monitoring
- Automatic detection and monitoring of external service calls

## Quick Start

### 1. Basic Setup

```go
import (
    "github.com/zondax/golem/pkg/zobservability"
    "github.com/zondax/golem/pkg/zobservability/factory"
)

// Initialize observer
config := &zobservability.Config{
    Provider:    zobservability.ProviderSigNoz,
    Enabled:     true,
    Environment: "production",
    Release:     "v1.0.0",
    Address:     "http://signoz:4317",
}

observer, err := factory.NewObserver(config, "my-service")
if err != nil {
    log.Fatal(err)
}
defer observer.Close()
```

### 2. Using Metrics

```go
// Get metrics provider
metrics := observer.GetMetrics()

// Register metrics (safe to call multiple times)
metrics.RegisterCounter("api_requests_total", "Total API requests", []string{"method", "status"})
metrics.RegisterHistogram("request_duration_seconds", "Request duration", []string{"endpoint"}, nil)

// Record metrics
metrics.IncrementCounter("api_requests_total", map[string]string{
    "method": "GET",
    "status": "200",
})

start := time.Now()
// ... your business logic ...
metrics.RecordDuration("request_duration_seconds", time.Since(start), map[string]string{
    "endpoint": "/api/users",
})
```

### 3. Using Tracing

```go
// Start a transaction
tx := observer.StartTransaction(ctx, "user-registration")
defer tx.Finish(zobservability.TransactionOK)

// Add context
tx.SetTag("user_id", "12345")
tx.SetTag("plan", "premium")

// Start a span
ctx, span := observer.StartSpan(tx.Context(), "validate-email")
defer span.Finish()

// Capture events
observer.CaptureMessage(ctx, "Email validation started", zobservability.LevelInfo)

// Handle errors
if err != nil {
    observer.CaptureException(ctx, err)
    tx.Finish(zobservability.TransactionError)
    return err
}
```

## Configuration

### YAML Configuration

```yaml
zobservability:
  provider: "opentelemetry"
  enabled: true
  environment: "production"
  release: "1.0.0"
  debug: false
  address: "localhost:4317"
  sample_rate: 1.0
  middleware:
    capture_errors: true
  tracing_exclusions:  # Optional: List of operations/endpoints to exclude from tracing
    # Direct operation names
    - "HealthCheck"
    - "Ping"
    - "GetMetrics"
    # gRPC methods (full path)
    - "/grpc.health.v1.Health/Check"
    - "/grpc.health.v1.Health/Watch"
    - "/grpc.reflection.v1alpha.ServerReflection/ServerReflectionInfo"
    # HTTP routes
    - "/health"
    - "/metrics"
    - "/api/v1/health"
  metrics:
    enabled: true
    provider: "opentelemetry"
    opentelemetry:
      endpoint: "localhost:4317"
      insecure: false
      service_name: "my-service"
      service_version: "1.0.0"
      environment: "production"
      export_mode: "push"
      push_interval: "30s"
      headers:
        "authorization": "Bearer token"
  custom_config:
    insecure: "false"
    batch_profile: "production"
```

### Environment-Specific Configurations

#### Development
```yaml
zobservability:
  provider: "opentelemetry"
  enabled: true
  environment: "development"
  address: "localhost:4317"
  debug: true
  sample_rate: 1.0
  metrics:
    opentelemetry:
      export_mode: "endpoint"  # For debugging
      insecure: true
  custom_config:
    insecure: "true"
    batch_profile: "development"
```

#### Production
```yaml
zobservability:
  provider: "signoz"
  enabled: true
  environment: "production"
  address: "ingest.eu.signoz.cloud:443"
  debug: false
  sample_rate: 0.1
  metrics:
    opentelemetry:
      export_mode: "push"      # Automatic export
      push_interval: "30s"
      insecure: false
  custom_config:
    header_signoz-access-token: "your-token"
    batch_profile: "production"
```

## Providers

### OpenTelemetry

The default provider supporting industry-standard observability.

**Features:**
- OTLP gRPC and HTTP export
- Automatic resource detection
- Configurable batch processing
- Custom headers support

**Configuration:**
```yaml
zobservability:
  provider: "opentelemetry"
  address: "localhost:4317"
  custom_config:
    insecure: "true"
    batch_profile: "development"
```

### SigNoz

Open-source observability platform with advanced features.

**Features:**
- Native OpenTelemetry support
- Custom batch profiles
- Advanced resource configuration
- Authentication via headers

**Configuration:**
```yaml
zobservability:
  provider: "signoz"
  address: "ingest.eu.signoz.cloud:443"
  custom_config:
    header_signoz-access-token: "your-token"
    batch_profile: "production"
    insecure: "false"
```

### Sentry

Error tracking and performance monitoring.

**Features:**
- Error capture and grouping
- Performance monitoring
- Release tracking
- User context

**Configuration:**
```yaml
zobservability:
  provider: "sentry"
  address: "https://your-dsn@sentry.io/project"
  release: "1.0.0"
  sample_rate: 0.1
```

## Metrics System

### Export Modes

#### Push Mode (`"push"`)

Automatically exports metrics at regular intervals.

**When to use:**
- Production environments
- Continuous monitoring
- Low-latency applications

**Configuration:**
```yaml
metrics:
  opentelemetry:
    export_mode: "push"
    push_interval: "30s"
```

**Advantages:**
- Automatic and consistent
- Lower application latency
- No manual intervention required

#### Endpoint Mode (`"endpoint"`)

Exports metrics on-demand via HTTP endpoints.

**When to use:**
- Development and debugging
- Prometheus scraping
- Custom export timing

**Configuration:**
```yaml
metrics:
  opentelemetry:
    export_mode: "endpoint"
```

**Advantages:**
- Full control over export timing
- Compatible with Prometheus
- Ideal for debugging

### Metric Types

#### Counters
Monotonically increasing values (requests, errors, etc.)

```go
// Register
metrics.RegisterCounter("api_requests_total", "Total API requests", []string{"method", "status"})

// Use
metrics.IncrementCounter("api_requests_total", map[string]string{
    "method": "GET",
    "status": "200",
})
```

#### Gauges
Values that can increase or decrease (memory usage, active connections)

```go
// Register
metrics.RegisterGauge("active_connections", "Active database connections", []string{"database"})

// Use
metrics.SetGauge("active_connections", 42.0, map[string]string{
    "database": "postgres",
})
```

#### Histograms
Distribution of values (request duration, response sizes)

```go
// Register
metrics.RegisterHistogram("request_duration_seconds", "Request duration", []string{"endpoint"}, nil)

// Use
metrics.RecordDuration("request_duration_seconds", time.Since(start), map[string]string{
    "endpoint": "/api/events",
})
```

## Tracing & Error Tracking

### Transactions

Represent top-level operations in your application.

```go
// Start transaction
transaction := observer.StartTransaction(ctx, "user_service.create_user")
defer transaction.Finish()

// Set transaction context
transaction.SetTag("user.email", email)
transaction.SetUser(zobservability.User{
    ID:    userID,
    Email: email,
})

// Handle errors
if err != nil {
    transaction.SetStatus(zobservability.SpanStatusError)
    observer.CaptureException(ctx, err)
}
```

### Spans

Represent individual operations within transactions.

```go
// Start span
ctx, span := observer.StartSpan(ctx, "database.create_user")
defer span.Finish()

// Add context
span.SetTag("db.statement", "INSERT INTO users...")
span.SetTag("db.table", "users")

// Handle errors
if err != nil {
    span.SetStatus(zobservability.SpanStatusError)
    span.SetTag("error", true)
}
```

### Error Capture

```go
// Capture exceptions
observer.CaptureException(ctx, err, zobservability.WithLevel(zobservability.LevelError))

// Capture messages
observer.CaptureMessage(ctx, "User login failed", zobservability.LevelWarning)
```

## Tracing Exclusions

The tracing exclusions feature allows you to selectively exclude specific methods or operations from being traced. This is particularly useful for:

- **High-frequency endpoints**: Health checks, metrics endpoints that would generate excessive trace data
- **Performance optimization**: Reduce overhead on operations that don't need monitoring
- **Noise reduction**: Keep your traces focused on business-critical operations

### Configuration

Add the `tracing_exclusions` list to your configuration:

```yaml
zobservability:
  provider: "signoz"
  enabled: true
  tracing_exclusions:
    - "HealthCheck"           # Exclude health check endpoints
    - "Ping"                  # Exclude ping/heartbeat endpoints
    - "GetMetrics"            # Exclude metrics collection
    - "db.query.select_health" # Exclude specific database queries
    - "cache.get"             # Exclude cache operations
```

### How It Works

When a method name matches an entry in the `tracing_exclusions` list:
- The method returns a no-op transaction or span that performs no operations
- No trace data is sent to the observability backend
- Child spans of excluded transactions are also excluded
- The method continues to execute normally, just without tracing

### Usage Example

```go
// Configuration with exclusions
config := &zobservability.Config{
    Provider: zobservability.ProviderSigNoz,
    Enabled: true,
    TracingExclusions: []string{
        "HealthCheck",
        "GetMetrics",
        "cache.get",
    },
}

observer, _ := factory.NewObserver(config, "my-service")

// This transaction will be excluded (no-op)
tx1 := observer.StartTransaction(ctx, "HealthCheck")
tx1.SetTag("key", "value") // No-op, does nothing
tx1.Finish(zobservability.TransactionOK) // No-op, does nothing

// This transaction will be traced normally
tx2 := observer.StartTransaction(ctx, "ProcessOrder")
tx2.SetTag("order_id", "12345") // Actually sets the tag
tx2.Finish(zobservability.TransactionOK) // Sends trace to backend

// Spans work the same way
_, span1 := observer.StartSpan(ctx, "cache.get") // Excluded (no-op)
span1.Finish() // No-op

_, span2 := observer.StartSpan(ctx, "db.query.insert") // Traced normally
span2.Finish() // Sends span to backend
```

### Best Practices

1. **Be Specific**: Use exact method names to avoid accidentally excluding important operations
2. **Case Sensitive**: Method names are case-sensitive - "HealthCheck" is different from "healthcheck"
3. **No Wildcards**: The current implementation requires exact matches (no regex or wildcards)
4. **Document Exclusions**: Keep a list of excluded methods in your documentation for debugging purposes
5. **Monitor Impact**: Periodically review excluded methods to ensure you're not missing important data

## External API Monitoring

The External API Monitoring feature automatically detects and categorizes external service calls in SigNoz. This provides visibility into all external dependencies and their performance characteristics.

### How It Works

External API Monitoring leverages OpenTelemetry semantic conventions to automatically detect external API calls. When properly configured, SigNoz will automatically categorize calls based on these attributes:

- `net.peer.name`: Domain or host of the external service (e.g., "api.stripe.com")
- `http.url`: Complete URL of the request (e.g., "https://api.stripe.com/v1/charges")
- `http.target`: Path portion of the URL (e.g., "/v1/charges")
- `rpc.system`: RPC system identifier (e.g., "grpc")

### Automatic Instrumentation

#### HTTP Calls

HTTP calls are automatically instrumented when using the `zhttpclient` package:

```go
// HTTP client is automatically instrumented with OpenTelemetry
httpClient := zhttpclient.New(zhttpclient.Config{
    Timeout: 30 * time.Second,
})

// This call will automatically appear in SigNoz External API Monitoring
resp, err := httpClient.NewRequest().
    SetURL("https://api.stripe.com/v1/charges").
    SetHeaders(map[string]string{
        "Authorization": "Bearer sk_test_...",
    }).
    Get(ctx)
```

#### gRPC Calls

gRPC calls are automatically instrumented when using the configured interceptors:

```go
// gRPC client with OpenTelemetry interceptor
dialOpts := []grpc.DialOption{
    grpc.WithStatsHandler(interceptors.NewOTelClientHandler(otelConfig)),
}

conn, err := grpc.NewClient(target, dialOpts...)
client := pb.NewServiceClient(conn)

// This call will automatically appear in SigNoz External API Monitoring
response, err := client.GetData(ctx, request)
```

## Usage Examples

### Complete Service Example

```go
type UserService struct {
    observer zobservability.Observer
    repo     UserRepository
}

func (s *UserService) CreateUser(ctx context.Context, req *CreateUserRequest) (*User, error) {
    // Start transaction
    transaction := s.observer.StartTransaction(ctx, "user_service.create_user")
    defer transaction.Finish()

    // Get metrics
    metrics := s.observer.GetMetrics()
    
    // Register metrics (safe to call multiple times)
    metrics.RegisterCounter("user_operations_total", "Total user operations", []string{"operation", "status"})
    metrics.RegisterHistogram("user_operation_duration_seconds", "User operation duration", []string{"operation"}, nil)
    
    start := time.Now()
    
    // Record operation start
    metrics.IncrementCounter("user_operations_total", map[string]string{
        "operation": "create",
        "status":    "started",
    })
    
    defer func() {
        status := "success"
        if err != nil {
            status = "error"
            transaction.SetStatus(zobservability.SpanStatusError)
            s.observer.CaptureException(ctx, err)
        }
        
        // Record final metrics
        metrics.RecordDuration("user_operation_duration_seconds", time.Since(start), map[string]string{
            "operation": "create",
        })
        metrics.IncrementCounter("user_operations_total", map[string]string{
            "operation": "create",
            "status":    status,
        })
    }()
    
    // Validate input
    if err := req.Validate(); err != nil {
        return nil, fmt.Errorf("invalid request: %w", err)
    }
    
    // Database operation with span
    ctx, span := s.observer.StartSpan(ctx, "database.create_user")
    span.SetTag("db.table", "users")
    defer span.Finish()
    
    user, err := s.repo.Create(ctx, req)
    if err != nil {
        span.SetStatus(zobservability.SpanStatusError)
        return nil, err
    }
    
    // Set transaction context
    transaction.SetTag("user.id", user.ID)
    transaction.SetUser(zobservability.User{
        ID:    user.ID,
        Email: user.Email,
    })
    
    return user, nil
}
```

### Middleware Integration

```go
func ObservabilityMiddleware(observer zobservability.Observer) gin.HandlerFunc {
    return func(c *gin.Context) {
        // Start transaction
        transaction := observer.StartTransaction(c.Request.Context(), 
            fmt.Sprintf("%s %s", c.Request.Method, c.FullPath()))
        defer transaction.Finish()
        
        // Set request context
        transaction.SetTag("http.method", c.Request.Method)
        transaction.SetTag("http.url", c.Request.URL.String())
        
        // Process request
        c.Next()
        
        // Set response context
        transaction.SetTag("http.status_code", c.Writer.Status())
        
        if c.Writer.Status() >= 400 {
            transaction.SetStatus(zobservability.SpanStatusError)
        }
    }
}
```

## Best Practices

### 1. Metric Naming
```go
// Good - descriptive with units
"api_requests_total"
"request_duration_seconds"
"memory_usage_bytes"

// Bad - unclear or missing units
"requests"
"duration"
"memory"
```

### 2. Label Usage
```go
// Good - consistent labels
metrics.IncrementCounter("api_requests_total", map[string]string{
    "method":   "GET",
    "endpoint": "/api/users",
    "status":   "200",
})

// Bad - inconsistent or high cardinality
metrics.IncrementCounter("api_requests_total", map[string]string{
    "user_id": "12345", // High cardinality!
    "Method":  "GET",   // Inconsistent casing
})
```

### 3. Error Handling
```go
// Good - structured error capture
if err != nil {
    observer.CaptureException(ctx, err, 
        zobservability.WithLevel(zobservability.LevelError),
        zobservability.WithTag("component", "user_service"),
        zobservability.WithTag("operation", "create_user"),
    )
    return nil, err
}
```

### 4. Resource Management
```go
// Good - proper cleanup
observer, err := factory.NewObserver(config, serviceName)
if err != nil {
    return err
}
defer observer.Close() // Always close!

// Good - span cleanup
ctx, span := observer.StartSpan(ctx, "operation")
defer span.Finish() // Always finish spans!
```

### 5. Configuration Management
```go
// Good - environment-specific config
func getObservabilityConfig(env string) *zobservability.Config {
    config := &zobservability.Config{
        Provider:    "opentelemetry",
        Enabled:     true,
        Environment: env,
    }
    
    switch env {
    case "production":
        config.Address = "ingest.signoz.cloud:443"
        config.SampleRate = 0.1
        config.Debug = false
    case "development":
        config.Address = "localhost:4317"
        config.SampleRate = 1.0
        config.Debug = true
    }
    
    return config
}
```

## Troubleshooting

### Common Issues

#### 1. Metrics Not Appearing
```bash
# Check configuration
Verify endpoint is correct
Confirm export_mode is "push" for automatic export
Check authentication headers
Verify network connectivity

# Debug steps
- Enable debug mode: debug: true
- Check logs for export errors
- Test with endpoint mode for manual export
```

#### 2. High Memory Usage
```bash
# Optimize batch configuration
Reduce batch_timeout
Lower max_export_batch_size
Increase export frequency

# Configuration example
custom_config:
  batch_profile: "low_memory"
```

#### 3. Connection Errors
```bash
# Check network and security
Verify endpoint URL and port
Check TLS/SSL settings (insecure flag)
Validate authentication tokens
Test network connectivity

# Debug configuration
custom_config:
  insecure: "true"  # For development only
  debug: "true"
```

#### 4. Performance Issues
```bash
# Optimize for performance
Reduce sample_rate in production
Use appropriate batch_profile
Minimize high-cardinality labels
Use async export modes

# Production configuration
sample_rate: 0.1
custom_config:
  batch_profile: "production"
```

### Debug Configuration

```yaml
# Maximum debugging
zobservability:
  provider: "opentelemetry"
  debug: true
  sample_rate: 1.0
  metrics:
    opentelemetry:
      export_mode: "endpoint"
      insecure: true
  custom_config:
    insecure: "true"
    batch_profile: "development"
```

### Monitoring Health

```go
// Check observer health
if err := observer.GetMetrics().Start(); err != nil {
    log.Errorf("Metrics provider failed to start: %v", err)
}

// Monitor export success
metrics.RegisterCounter("zobservability_exports_total", "Total exports", []string{"status"})
```

---

For more information, see the [examples](examples/) directory and provider-specific documentation in [providers/](providers/).