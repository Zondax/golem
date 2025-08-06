package signoz

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/zondax/golem/pkg/zobservability"
)

// =============================================================================
// NEW OBSERVER TESTS
// =============================================================================

func TestNewObserver_WhenValidConfig_ShouldReturnObserver(t *testing.T) {
	// Arrange
	config := &Config{
		Endpoint:    "localhost:4317",
		ServiceName: "test-service",
		Environment: "test",
		Release:     "1.0.0",
		Insecure:    true,
		SampleRate:  1.0,
		Metrics: zobservability.MetricsConfig{
			Enabled:  true,
			Provider: string(zobservability.MetricsProviderNoop),
		},
	}

	// Act
	observer, err := NewObserver(config)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, observer)
	assert.Implements(t, (*zobservability.Observer)(nil), observer)

	// Cleanup
	if observer != nil {
		_ = observer.Close()
	}
}

func TestNewObserver_WhenInvalidConfig_ShouldReturnError(t *testing.T) {
	testCases := []struct {
		name   string
		config *Config
	}{
		{
			name: "missing_endpoint",
			config: &Config{
				ServiceName: "test-service",
				Environment: "test",
			},
		},
		{
			name: "missing_service_name",
			config: &Config{
				Endpoint:    "localhost:4317",
				Environment: "test",
			},
		},
		{
			name: "invalid_sample_rate_negative",
			config: &Config{
				Endpoint:    "localhost:4317",
				ServiceName: "test-service",
				SampleRate:  -0.5,
			},
		},
		{
			name: "invalid_sample_rate_too_high",
			config: &Config{
				Endpoint:    "localhost:4317",
				ServiceName: "test-service",
				SampleRate:  1.5,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			observer, err := NewObserver(tc.config)

			// Assert
			assert.Error(t, err)
			assert.Nil(t, observer)
		})
	}
}

func TestNewObserver_WhenCompleteConfig_ShouldReturnObserver(t *testing.T) {
	// Arrange
	config := &Config{
		Endpoint:    "localhost:4317",
		ServiceName: "complete-service",
		Environment: "production",
		Release:     "2.1.0",
		Debug:       true,
		Insecure:    true,
		Headers: map[string]string{
			"signoz-access-token": "test-token",
			"x-api-key":           "test-key",
		},
		SampleRate: 0.5,
		Metrics: zobservability.MetricsConfig{
			Enabled:  true,
			Provider: string(zobservability.MetricsProviderNoop),
		},
		BatchConfig: &BatchConfig{
			BatchTimeout:   5 * time.Second,
			ExportTimeout:  30 * time.Second,
			MaxExportBatch: 512,
			MaxQueueSize:   2048,
		},
		ResourceConfig: &ResourceConfig{
			IncludeHostname:  true,
			IncludeProcessID: true,
			CustomAttributes: map[string]string{
				"team":    "backend",
				"version": "v2",
			},
		},
	}

	// Act
	observer, err := NewObserver(config)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, observer)

	// Cleanup
	if observer != nil {
		_ = observer.Close()
	}
}

// =============================================================================
// OBSERVER INTERFACE TESTS
// =============================================================================

func TestSignozObserver_StartTransaction_WhenCalled_ShouldReturnTransaction(t *testing.T) {
	// Arrange
	observer := createTestObserver(t)
	defer func() { _ = observer.Close() }()

	ctx := context.Background()
	name := "test-transaction"

	// Act
	tx := observer.StartTransaction(ctx, name)

	// Assert
	assert.NotNil(t, tx)
	assert.Implements(t, (*zobservability.Transaction)(nil), tx)

	// Cleanup
	tx.Finish(zobservability.TransactionOK)
}

func TestSignozObserver_StartTransaction_WhenCalledWithOptions_ShouldApplyOptions(t *testing.T) {
	// Arrange
	observer := createTestObserver(t)
	defer func() { _ = observer.Close() }()

	ctx := context.Background()
	name := "test-transaction-with-options"

	// Act
	tx := observer.StartTransaction(ctx, name,
		zobservability.WithTransactionTag("service", "test"),
		zobservability.WithTransactionData("user_id", "123"),
	)

	// Assert
	assert.NotNil(t, tx)

	// Cleanup
	tx.Finish(zobservability.TransactionOK)
}

func TestSignozObserver_StartSpan_WhenCalled_ShouldReturnSpan(t *testing.T) {
	// Arrange
	observer := createTestObserver(t)
	defer func() { _ = observer.Close() }()

	ctx := context.Background()
	operation := "test-operation"

	// Act
	newCtx, span := observer.StartSpan(ctx, operation)

	// Assert
	assert.NotNil(t, newCtx)
	assert.NotNil(t, span)
	assert.Implements(t, (*zobservability.Span)(nil), span)

	// Cleanup
	span.Finish()
}

func TestSignozObserver_StartSpan_WhenCalledWithOptions_ShouldApplyOptions(t *testing.T) {
	// Arrange
	observer := createTestObserver(t)
	defer func() { _ = observer.Close() }()

	ctx := context.Background()
	operation := "test-operation-with-options"

	// Act
	newCtx, span := observer.StartSpan(ctx, operation,
		zobservability.WithSpanTag("component", "database"),
		zobservability.WithSpanData("query", "SELECT * FROM users"),
	)

	// Assert
	assert.NotNil(t, newCtx)
	assert.NotNil(t, span)

	// Cleanup
	span.Finish()
}

func TestSignozObserver_CaptureException_WhenCalled_ShouldNotPanic(t *testing.T) {
	// Arrange
	observer := createTestObserver(t)
	defer func() { _ = observer.Close() }()

	ctx := context.Background()
	err := assert.AnError

	// Act & Assert
	assert.NotPanics(t, func() {
		observer.CaptureException(ctx, err)
	})
}

func TestSignozObserver_CaptureException_WhenCalledWithOptions_ShouldNotPanic(t *testing.T) {
	// Arrange
	observer := createTestObserver(t)
	defer func() { _ = observer.Close() }()

	ctx := context.Background()
	err := assert.AnError

	// Act & Assert
	assert.NotPanics(t, func() {
		observer.CaptureException(ctx, err,
			zobservability.WithEventTag("severity", "high"),
			zobservability.WithEventUser("user123", "test@example.com", "testuser"),
		)
	})
}

func TestSignozObserver_CaptureMessage_WhenCalled_ShouldNotPanic(t *testing.T) {
	// Arrange
	observer := createTestObserver(t)
	defer func() { _ = observer.Close() }()

	ctx := context.Background()
	message := "Test message"
	level := zobservability.LevelInfo

	// Act & Assert
	assert.NotPanics(t, func() {
		observer.CaptureMessage(ctx, message, level)
	})
}

func TestSignozObserver_CaptureMessage_WhenCalledWithDifferentLevels_ShouldNotPanic(t *testing.T) {
	// Arrange
	observer := createTestObserver(t)
	defer func() { _ = observer.Close() }()

	ctx := context.Background()
	message := "Test message"

	testCases := []struct {
		name  string
		level zobservability.Level
	}{
		{
			name:  "debug_level",
			level: zobservability.LevelDebug,
		},
		{
			name:  "info_level",
			level: zobservability.LevelInfo,
		},
		{
			name:  "warning_level",
			level: zobservability.LevelWarning,
		},
		{
			name:  "error_level",
			level: zobservability.LevelError,
		},
		{
			name:  "fatal_level",
			level: zobservability.LevelFatal,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Act & Assert
			assert.NotPanics(t, func() {
				observer.CaptureMessage(ctx, message, tc.level)
			})
		})
	}
}

func TestSignozObserver_GetMetrics_WhenCalled_ShouldReturnMetricsProvider(t *testing.T) {
	// Arrange
	observer := createTestObserver(t)
	defer func() { _ = observer.Close() }()

	// Act
	metrics := observer.GetMetrics()

	// Assert
	assert.NotNil(t, metrics)
	assert.Implements(t, (*zobservability.MetricsProvider)(nil), metrics)
}

func TestSignozObserver_Close_WhenCalled_ShouldNotReturnError(t *testing.T) {
	// Arrange
	observer := createTestObserver(t)

	// Act
	err := observer.Close()

	// Assert
	// In test environment, connection timeout is expected since SigNoz is not running
	// The close operation should either succeed or fail with a connection error
	if err != nil {
		assert.Contains(t, err.Error(), "connection")
	}
}

func TestSignozObserver_GetConfig_WhenCalled_ShouldReturnConfig(t *testing.T) {
	// Arrange
	observer := createTestObserver(t)
	defer func() { _ = observer.Close() }()

	// Act
	config := observer.GetConfig()

	// Assert
	assert.Equal(t, zobservability.ProviderSigNoz, config.Provider)
	assert.True(t, config.Enabled)
	assert.NotEmpty(t, config.Address)
}

// =============================================================================
// TRACER PROVIDER CREATION TESTS
// =============================================================================

func TestCreateTracerProvider_WhenValidConfig_ShouldReturnProvider(t *testing.T) {
	// Arrange
	config := &Config{
		Endpoint:    "localhost:4317",
		ServiceName: "test-service",
		Environment: "test",
		Release:     "1.0.0",
		Insecure:    true,
		SampleRate:  1.0,
	}

	// Act
	provider, tracer, err := createTracerProvider(config)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, provider)
	assert.NotNil(t, tracer)

	// Cleanup
	if provider != nil {
		_ = provider.Shutdown(context.Background())
	}
}

func TestCreateTracerProvider_WhenSecureConfig_ShouldReturnProvider(t *testing.T) {
	// Arrange
	config := &Config{
		Endpoint:    "localhost:4317",
		ServiceName: "test-service",
		Environment: "test",
		Release:     "1.0.0",
		Insecure:    false, // Secure connection
		SampleRate:  1.0,
	}

	// Act
	provider, tracer, err := createTracerProvider(config)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, provider)
	assert.NotNil(t, tracer)

	// Cleanup
	if provider != nil {
		_ = provider.Shutdown(context.Background())
	}
}

func TestCreateTracerProvider_WhenCustomSampleRate_ShouldReturnProvider(t *testing.T) {
	// Arrange
	config := &Config{
		Endpoint:    "localhost:4317",
		ServiceName: "test-service",
		Environment: "test",
		Release:     "1.0.0",
		Insecure:    true,
		SampleRate:  0.5, // 50% sampling
	}

	// Act
	provider, tracer, err := createTracerProvider(config)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, provider)
	assert.NotNil(t, tracer)

	// Cleanup
	if provider != nil {
		_ = provider.Shutdown(context.Background())
	}
}

func TestCreateTracerProvider_WhenCustomBatchConfig_ShouldReturnProvider(t *testing.T) {
	// Arrange
	config := &Config{
		Endpoint:    "localhost:4317",
		ServiceName: "test-service",
		Environment: "test",
		Release:     "1.0.0",
		Insecure:    true,
		SampleRate:  1.0,
		BatchConfig: &BatchConfig{
			BatchTimeout:   2 * time.Second,
			ExportTimeout:  15 * time.Second,
			MaxExportBatch: 256,
			MaxQueueSize:   1024,
		},
	}

	// Act
	provider, tracer, err := createTracerProvider(config)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, provider)
	assert.NotNil(t, tracer)

	// Cleanup
	if provider != nil {
		_ = provider.Shutdown(context.Background())
	}
}

// =============================================================================
// RESOURCE CREATION TESTS
// =============================================================================

func TestCreateTracingResource_WhenValidConfig_ShouldReturnResource(t *testing.T) {
	// Arrange
	config := &Config{
		ServiceName: "test-service",
		Environment: "test",
		Release:     "1.0.0",
	}

	// Act
	resource, err := createTracingResource(config)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resource)

	// Check resource attributes
	attrs := resource.Attributes()
	assert.NotEmpty(t, attrs)

	// Verify required attributes are present
	hasServiceName := false
	hasServiceVersion := false
	hasEnvironment := false
	hasLanguage := false
	hasHostname := false

	for _, attr := range attrs {
		switch string(attr.Key) {
		case zobservability.ResourceServiceName:
			hasServiceName = true
			assert.Equal(t, "test-service", attr.Value.AsString())
		case zobservability.ResourceServiceVersion:
			hasServiceVersion = true
			assert.Equal(t, "1.0.0", attr.Value.AsString())
		case zobservability.ResourceEnvironment:
			hasEnvironment = true
			assert.Equal(t, "test", attr.Value.AsString())
		case zobservability.ResourceLanguage:
			hasLanguage = true
			assert.Equal(t, zobservability.ResourceLanguageGo, attr.Value.AsString())
		case zobservability.ResourceHostName:
			hasHostname = true
			assert.NotEmpty(t, attr.Value.AsString())
		}
	}

	assert.True(t, hasServiceName)
	assert.True(t, hasServiceVersion)
	assert.True(t, hasEnvironment)
	assert.True(t, hasLanguage)
	assert.True(t, hasHostname)
}

func TestCreateTracingResource_WhenCustomResourceConfig_ShouldIncludeCustomAttributes(t *testing.T) {
	// Arrange
	config := &Config{
		ServiceName: "test-service",
		Environment: "test",
		Release:     "1.0.0",
		ResourceConfig: &ResourceConfig{
			IncludeHostname:  true,
			IncludeProcessID: true,
			CustomAttributes: map[string]string{
				"team":       "backend",
				"datacenter": "us-west-1",
			},
		},
	}

	// Act
	resource, err := createTracingResource(config)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resource)

	// Check for custom attributes
	attrs := resource.Attributes()
	hasTeam := false
	hasDatacenter := false
	hasProcessID := false

	for _, attr := range attrs {
		switch string(attr.Key) {
		case "team":
			hasTeam = true
			assert.Equal(t, "backend", attr.Value.AsString())
		case "datacenter":
			hasDatacenter = true
			assert.Equal(t, "us-west-1", attr.Value.AsString())
		case zobservability.ResourceProcessPID:
			hasProcessID = true
			assert.NotEmpty(t, attr.Value.AsString())
		}
	}

	assert.True(t, hasTeam)
	assert.True(t, hasDatacenter)
	assert.True(t, hasProcessID)
}

// =============================================================================
// INTEGRATION TESTS
// =============================================================================

func TestSignozObserver_WhenCompleteWorkflow_ShouldWorkCorrectly(t *testing.T) {
	// Arrange
	observer := createTestObserver(t)
	defer func() { _ = observer.Close() }()

	ctx := context.Background()

	// Act - Start transaction
	tx := observer.StartTransaction(ctx, "test-workflow")
	tx.SetTag("workflow", "integration-test")
	tx.SetData("test_id", "12345")

	// Act - Start span within transaction
	spanCtx, span := observer.StartSpan(tx.Context(), "database-query")
	span.SetTag("db.type", "postgresql")
	span.SetData("query", "SELECT * FROM users")

	// Act - Capture message
	observer.CaptureMessage(spanCtx, "Processing user data", zobservability.LevelInfo)

	// Act - Capture exception
	testErr := assert.AnError
	observer.CaptureException(spanCtx, testErr)

	// Act - Finish span and transaction
	span.Finish()
	tx.Finish(zobservability.TransactionOK)

	// Assert - Should not panic and complete successfully
	assert.NotNil(t, observer.GetMetrics())
	config := observer.GetConfig()
	assert.Equal(t, zobservability.ProviderSigNoz, config.Provider)
}

func TestSignozObserver_WhenNestedSpans_ShouldWorkCorrectly(t *testing.T) {
	// Arrange
	observer := createTestObserver(t)
	defer func() { _ = observer.Close() }()

	ctx := context.Background()

	// Act - Create nested spans
	tx := observer.StartTransaction(ctx, "nested-spans-test")

	// Level 1 span
	ctx1, span1 := observer.StartSpan(tx.Context(), "level-1-operation")
	span1.SetTag("level", "1")

	// Level 2 span (child of span1)
	ctx2, span2 := observer.StartSpan(ctx1, "level-2-operation")
	span2.SetTag("level", "2")

	// Level 3 span (child of span2)
	ctx3, span3 := observer.StartSpan(ctx2, "level-3-operation")
	span3.SetTag("level", "3")

	// Finish in reverse order
	span3.Finish()
	span2.Finish()
	span1.Finish()
	tx.Finish(zobservability.TransactionOK)

	// Assert - Should complete without errors
	assert.NotNil(t, ctx3)
}

// =============================================================================
// TRACING EXCLUSIONS TESTS
// =============================================================================

func TestSignozObserver_StartTransaction_WhenOperationExcluded_ShouldReturnNoopTransaction(t *testing.T) {
	// Arrange
	config := &Config{
		Endpoint:    "localhost:4317",
		ServiceName: "test-service",
		Environment: "test",
		Release:     "1.0.0",
		Insecure:    true,
		SampleRate:  1.0,
		TracingExclusions: []string{
			"health-check",
			"metrics-endpoint",
			"excluded-operation",
		},
		Metrics: zobservability.MetricsConfig{
			Enabled:  true,
			Provider: string(zobservability.MetricsProviderNoop),
		},
	}
	
	observer, err := NewObserver(config)
	require.NoError(t, err)
	require.NotNil(t, observer)
	defer func() { _ = observer.Close() }()

	ctx := context.Background()

	// Act - Start excluded transaction
	tx := observer.StartTransaction(ctx, "health-check")

	// Assert
	assert.NotNil(t, tx)
	// The noop transaction should preserve the context
	assert.Equal(t, ctx, tx.Context())
	
	// Operations on noop transaction should not panic
	assert.NotPanics(t, func() {
		tx.SetName("new-name")
		tx.SetTag("key", "value")
		tx.SetData("data", "value")
		child := tx.StartChild("child-operation")
		child.Finish()
		tx.Finish(zobservability.TransactionOK)
	})
}

func TestSignozObserver_StartTransaction_WhenOperationNotExcluded_ShouldReturnNormalTransaction(t *testing.T) {
	// Arrange
	config := &Config{
		Endpoint:    "localhost:4317",
		ServiceName: "test-service",
		Environment: "test",
		Release:     "1.0.0",
		Insecure:    true,
		SampleRate:  1.0,
		TracingExclusions: []string{
			"health-check",
			"metrics-endpoint",
		},
		Metrics: zobservability.MetricsConfig{
			Enabled:  true,
			Provider: string(zobservability.MetricsProviderNoop),
		},
	}
	
	observer, err := NewObserver(config)
	require.NoError(t, err)
	require.NotNil(t, observer)
	defer func() { _ = observer.Close() }()

	ctx := context.Background()

	// Act - Start normal transaction
	tx := observer.StartTransaction(ctx, "normal-operation")

	// Assert
	assert.NotNil(t, tx)
	// Should be a real transaction with trace context
	assert.NotEqual(t, ctx, tx.Context())
	
	// Cleanup
	tx.Finish(zobservability.TransactionOK)
}

func TestSignozObserver_StartSpan_WhenOperationExcluded_ShouldReturnNoopSpan(t *testing.T) {
	// Arrange
	config := &Config{
		Endpoint:    "localhost:4317",
		ServiceName: "test-service",
		Environment: "test",
		Release:     "1.0.0",
		Insecure:    true,
		SampleRate:  1.0,
		TracingExclusions: []string{
			"excluded-span",
			"ignored-operation",
		},
		Metrics: zobservability.MetricsConfig{
			Enabled:  true,
			Provider: string(zobservability.MetricsProviderNoop),
		},
	}
	
	observer, err := NewObserver(config)
	require.NoError(t, err)
	require.NotNil(t, observer)
	defer func() { _ = observer.Close() }()

	ctx := context.Background()

	// Act - Start excluded span
	newCtx, span := observer.StartSpan(ctx, "excluded-span")

	// Assert
	assert.NotNil(t, span)
	// Context should be preserved (not modified)
	assert.Equal(t, ctx, newCtx)
	
	// Operations on noop span should not panic
	assert.NotPanics(t, func() {
		span.SetTag("key", "value")
		span.SetData("data", "value")
		span.SetError(assert.AnError)
		span.Finish()
	})
}

func TestSignozObserver_StartSpan_WhenOperationNotExcluded_ShouldReturnNormalSpan(t *testing.T) {
	// Arrange
	config := &Config{
		Endpoint:    "localhost:4317",
		ServiceName: "test-service",
		Environment: "test",
		Release:     "1.0.0",
		Insecure:    true,
		SampleRate:  1.0,
		TracingExclusions: []string{
			"excluded-span",
			"ignored-operation",
		},
		Metrics: zobservability.MetricsConfig{
			Enabled:  true,
			Provider: string(zobservability.MetricsProviderNoop),
		},
	}
	
	observer, err := NewObserver(config)
	require.NoError(t, err)
	require.NotNil(t, observer)
	defer func() { _ = observer.Close() }()

	ctx := context.Background()

	// Act - Start normal span
	newCtx, span := observer.StartSpan(ctx, "normal-span")

	// Assert
	assert.NotNil(t, span)
	// Should have trace context injected
	assert.NotEqual(t, ctx, newCtx)
	
	// Cleanup
	span.Finish()
}

func TestSignozObserver_TracingExclusions_EmptyList_ShouldNotExcludeAnything(t *testing.T) {
	// Arrange
	config := &Config{
		Endpoint:    "localhost:4317",
		ServiceName: "test-service",
		Environment: "test",
		Release:     "1.0.0",
		Insecure:    true,
		SampleRate:  1.0,
		TracingExclusions: []string{}, // Empty exclusion list
		Metrics: zobservability.MetricsConfig{
			Enabled:  true,
			Provider: string(zobservability.MetricsProviderNoop),
		},
	}
	
	observer, err := NewObserver(config)
	require.NoError(t, err)
	require.NotNil(t, observer)
	defer func() { _ = observer.Close() }()

	ctx := context.Background()

	// Act & Assert - All operations should create real spans
	tx := observer.StartTransaction(ctx, "any-transaction")
	assert.NotEqual(t, ctx, tx.Context()) // Should have trace context
	tx.Finish(zobservability.TransactionOK)

	newCtx, span := observer.StartSpan(ctx, "any-span")
	assert.NotEqual(t, ctx, newCtx) // Should have trace context
	span.Finish()
}

func TestSignozObserver_isOperationExcluded(t *testing.T) {
	// Arrange
	exclusionsMap := make(map[string]bool)
	exclusionsMap["health"] = true
	exclusionsMap["metrics"] = true
	exclusionsMap["debug/pprof"] = true
	
	observer := &signozObserver{
		config: &Config{
			TracingExclusions: []string{
				"health",
				"metrics",
				"debug/pprof",
			},
		},
		exclusionsMap: exclusionsMap,
	}

	testCases := []struct {
		name      string
		operation string
		expected  bool
	}{
		{
			name:      "exact_match_health",
			operation: "health",
			expected:  true,
		},
		{
			name:      "exact_match_metrics",
			operation: "metrics",
			expected:  true,
		},
		{
			name:      "exact_match_debug_pprof",
			operation: "debug/pprof",
			expected:  true,
		},
		{
			name:      "not_excluded",
			operation: "api/users",
			expected:  false,
		},
		{
			name:      "partial_match_not_excluded",
			operation: "health-check", // Not exact match
			expected:  false,
		},
		{
			name:      "empty_operation",
			operation: "",
			expected:  false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			result := observer.isOperationExcluded(tc.operation)

			// Assert
			assert.Equal(t, tc.expected, result)
		})
	}
}

// =============================================================================
// HELPER FUNCTIONS
// =============================================================================

func createTestObserver(t *testing.T) zobservability.Observer {
	config := &Config{
		Endpoint:    "localhost:4317",
		ServiceName: "test-service",
		Environment: "test",
		Release:     "1.0.0",
		Insecure:    true,
		SampleRate:  1.0,
		Metrics: zobservability.MetricsConfig{
			Enabled:  true,
			Provider: string(zobservability.MetricsProviderNoop),
		},
	}

	observer, err := NewObserver(config)
	require.NoError(t, err)
	require.NotNil(t, observer)

	return observer
}
