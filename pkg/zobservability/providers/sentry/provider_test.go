package sentry

import (
	"context"
	"testing"

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
		DSN:           "https://test@sentry.io/123456",
		Environment:   "test",
		Release:       "1.0.0",
		Debug:         false,
		ServiceName:   "test-service",
		SampleRate:    1.0,
		CaptureErrors: true,
	}

	// Act
	observer, err := NewObserver(config)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, observer)
	assert.Implements(t, (*zobservability.Observer)(nil), observer)

	// Cleanup
	defer func() { _ = observer.Close() }()
}

func TestNewObserver_WhenMinimalConfig_ShouldReturnObserver(t *testing.T) {
	// Arrange
	config := &Config{
		DSN: "https://test@sentry.io/123456",
	}

	// Act
	observer, err := NewObserver(config)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, observer)

	// Cleanup
	defer func() { _ = observer.Close() }()
}

func TestNewObserver_WhenCompleteConfig_ShouldReturnObserver(t *testing.T) {
	// Arrange
	config := &Config{
		DSN:           "https://test@sentry.io/123456",
		Environment:   "production",
		Release:       "v2.1.1",
		Debug:         true,
		ServiceName:   "complete-service",
		SampleRate:    0.5,
		CaptureErrors: true,
	}

	// Act
	observer, err := NewObserver(config)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, observer)

	// Cleanup
	defer func() { _ = observer.Close() }()
}

func TestNewObserver_WhenInvalidDSN_ShouldReturnError(t *testing.T) {
	// Arrange
	config := &Config{
		DSN: "invalid-dsn",
	}

	// Act
	observer, err := NewObserver(config)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, observer)
}

// =============================================================================
// OBSERVER INTERFACE TESTS
// =============================================================================

func TestSentryObserver_StartTransaction_WhenCalled_ShouldReturnTransaction(t *testing.T) {
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

func TestSentryObserver_StartTransaction_WhenCalledWithOptions_ShouldApplyOptions(t *testing.T) {
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

func TestSentryObserver_StartSpan_WhenCalled_ShouldReturnSpan(t *testing.T) {
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

func TestSentryObserver_StartSpan_WhenCalledWithOptions_ShouldApplyOptions(t *testing.T) {
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

func TestSentryObserver_StartSpan_WhenCalledWithParentSpan_ShouldCreateChildSpan(t *testing.T) {
	// Arrange
	observer := createTestObserver(t)
	defer func() { _ = observer.Close() }()

	ctx := context.Background()

	// Create parent transaction first
	tx := observer.StartTransaction(ctx, "parent-transaction")
	parentCtx := tx.Context()

	// Act - Create child span
	childCtx, childSpan := observer.StartSpan(parentCtx, "child-operation")

	// Assert
	assert.NotNil(t, childCtx)
	assert.NotNil(t, childSpan)

	// Cleanup
	childSpan.Finish()
	tx.Finish(zobservability.TransactionOK)
}

func TestSentryObserver_CaptureException_WhenCalled_ShouldNotPanic(t *testing.T) {
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

func TestSentryObserver_CaptureException_WhenCalledWithOptions_ShouldNotPanic(t *testing.T) {
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

func TestSentryObserver_CaptureMessage_WhenCalled_ShouldNotPanic(t *testing.T) {
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

func TestSentryObserver_CaptureMessage_WhenCalledWithDifferentLevels_ShouldNotPanic(t *testing.T) {
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

func TestSentryObserver_GetMetrics_WhenCalled_ShouldReturnNoopProvider(t *testing.T) {
	// Arrange
	observer := createTestObserver(t)
	defer func() { _ = observer.Close() }()

	// Act
	metrics := observer.GetMetrics()

	// Assert
	assert.NotNil(t, metrics)
	assert.Implements(t, (*zobservability.MetricsProvider)(nil), metrics)
	// Sentry returns a noop metrics provider since it doesn't support metrics
}

func TestSentryObserver_Close_WhenCalled_ShouldNotReturnError(t *testing.T) {
	// Arrange
	observer := createTestObserver(t)

	// Act
	err := observer.Close()

	// Assert
	assert.NoError(t, err)
}

func TestSentryObserver_GetConfig_WhenCalled_ShouldReturnConfig(t *testing.T) {
	// Arrange
	observer := createTestObserver(t)
	defer func() { _ = observer.Close() }()

	// Act
	config := observer.GetConfig()

	// Assert
	assert.Equal(t, zobservability.ProviderSentry, config.Provider)
	assert.True(t, config.Enabled)
	assert.NotEmpty(t, config.Address)
}

// =============================================================================
// LEVEL CONVERSION TESTS
// =============================================================================

func TestConvertLevel_WhenCalledWithDifferentLevels_ShouldReturnCorrectSentryLevel(t *testing.T) {
	testCases := []struct {
		name           string
		observLevel    zobservability.Level
		expectedSentry string // We'll check the string representation
	}{
		{
			name:           "debug_level",
			observLevel:    zobservability.LevelDebug,
			expectedSentry: "debug",
		},
		{
			name:           "info_level",
			observLevel:    zobservability.LevelInfo,
			expectedSentry: "info",
		},
		{
			name:           "warning_level",
			observLevel:    zobservability.LevelWarning,
			expectedSentry: "warning",
		},
		{
			name:           "error_level",
			observLevel:    zobservability.LevelError,
			expectedSentry: "error",
		},
		{
			name:           "fatal_level",
			observLevel:    zobservability.LevelFatal,
			expectedSentry: "fatal",
		},
		{
			name:           "unknown_level",
			observLevel:    zobservability.Level(999), // Unknown level value
			expectedSentry: "error",                   // Should default to error
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			sentryLevel := convertLevel(tc.observLevel)

			// Assert
			assert.Equal(t, tc.expectedSentry, string(sentryLevel))
		})
	}
}

// =============================================================================
// INTEGRATION TESTS
// =============================================================================

func TestSentryObserver_WhenCompleteWorkflow_ShouldWorkCorrectly(t *testing.T) {
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
	assert.Equal(t, zobservability.ProviderSentry, config.Provider)
}

func TestSentryObserver_WhenNestedSpans_ShouldWorkCorrectly(t *testing.T) {
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

func TestSentryObserver_WhenErrorScenario_ShouldCaptureErrors(t *testing.T) {
	// Arrange
	observer := createTestObserver(t)
	defer func() { _ = observer.Close() }()

	ctx := context.Background()

	// Act - Simulate error scenario
	tx := observer.StartTransaction(ctx, "error-workflow")
	tx.SetTag("scenario", "error-test")

	// Simulate some work that fails
	errorSpan := tx.StartChild("failing-operation")
	errorSpan.SetTag("operation.type", "database")
	errorSpan.SetData("error.message", "connection timeout")

	// Simulate error in span
	testErr := assert.AnError
	errorSpan.SetError(testErr)
	errorSpan.Finish()

	// Capture exception at transaction level
	observer.CaptureException(tx.Context(), testErr,
		zobservability.WithEventTag("error.type", "database_error"),
		zobservability.WithEventUser("user123", "test@example.com", "testuser"),
	)

	// Capture error message
	observer.CaptureMessage(tx.Context(), "Database operation failed", zobservability.LevelError,
		zobservability.WithEventTag("component", "database"),
	)

	// Finish transaction with error status
	tx.Finish(zobservability.TransactionError)

	// Assert - Should complete without panicking
	assert.NotNil(t, tx)
}

func TestSentryObserver_WhenMultipleTransactions_ShouldHandleCorrectly(t *testing.T) {
	// Arrange
	observer := createTestObserver(t)
	defer func() { _ = observer.Close() }()

	ctx := context.Background()

	// Act - Create multiple concurrent transactions
	tx1 := observer.StartTransaction(ctx, "transaction-1")
	tx1.SetTag("transaction.id", "1")

	tx2 := observer.StartTransaction(ctx, "transaction-2")
	tx2.SetTag("transaction.id", "2")

	// Each transaction has its own spans
	span1 := tx1.StartChild("operation-1")
	span1.SetData("data", "from-tx1")

	span2 := tx2.StartChild("operation-2")
	span2.SetData("data", "from-tx2")

	// Finish everything
	span1.Finish()
	span2.Finish()
	tx1.Finish(zobservability.TransactionOK)
	tx2.Finish(zobservability.TransactionOK)

	// Assert - Should complete without errors
	assert.NotNil(t, tx1)
	assert.NotNil(t, tx2)
}

// =============================================================================
// HELPER FUNCTIONS
// =============================================================================

func createTestObserver(t *testing.T) zobservability.Observer {
	config := &Config{
		DSN:         "https://test@sentry.io/123456",
		Environment: "test",
		Release:     "1.0.0",
		Debug:       false,
		ServiceName: "test-service",
		SampleRate:  1.0,
	}

	observer, err := NewObserver(config)
	require.NoError(t, err)
	require.NotNil(t, observer)

	return observer
}
