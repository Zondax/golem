package zobservability

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

// =============================================================================
// NOOP OBSERVER TESTS
// =============================================================================

func TestNewNoopObserver(t *testing.T) {
	// Act
	observer := NewNoopObserver()

	// Assert
	assert.NotNil(t, observer)
	assert.Implements(t, (*Observer)(nil), observer)

	// Verify metrics provider is set
	metrics := observer.GetMetrics()
	assert.NotNil(t, metrics)
	assert.Equal(t, "noop", metrics.Name())
}

func TestNoopObserver_StartTransaction_WhenCalled_ShouldReturnNoopTransaction(t *testing.T) {
	// Arrange
	observer := NewNoopObserver()
	ctx := context.Background()
	name := "test-transaction"

	// Act
	tx := observer.StartTransaction(ctx, name)

	// Assert
	assert.NotNil(t, tx)
	assert.Implements(t, (*Transaction)(nil), tx)
}

func TestNoopObserver_StartSpan_WhenCalled_ShouldReturnContextAndNoopSpan(t *testing.T) {
	// Arrange
	observer := NewNoopObserver()
	ctx := context.Background()
	operation := "test-operation"

	// Act
	newCtx, span := observer.StartSpan(ctx, operation)

	// Assert
	assert.NotNil(t, newCtx)
	assert.NotNil(t, span)
	assert.Implements(t, (*Span)(nil), span)
	assert.Equal(t, ctx, newCtx) // Should return the same context
}

func TestNoopObserver_CaptureException(t *testing.T) {
	// Arrange
	observer := NewNoopObserver()
	ctx := context.Background()
	testError := errors.New("test error")

	// Act & Assert - Should not panic
	assert.NotPanics(t, func() {
		observer.CaptureException(ctx, testError)
	})
}

func TestNoopObserver_CaptureMessage(t *testing.T) {
	// Arrange
	observer := NewNoopObserver()
	ctx := context.Background()
	message := "test message"
	level := LevelInfo

	// Act & Assert - Should not panic
	assert.NotPanics(t, func() {
		observer.CaptureMessage(ctx, message, level)
	})
}

func TestNoopObserver_GetMetrics_WhenCalled_ShouldReturnMetricsProvider(t *testing.T) {
	// Arrange
	observer := NewNoopObserver()

	// Act
	metrics := observer.GetMetrics()

	// Assert
	assert.NotNil(t, metrics)
	assert.Implements(t, (*MetricsProvider)(nil), metrics)
}

func TestNoopObserver_Close_WhenCalled_ShouldReturnNil(t *testing.T) {
	// Arrange
	observer := NewNoopObserver()

	// Act
	err := observer.Close()

	// Assert
	assert.NoError(t, err)
}

func TestNoopObserver_GetConfig(t *testing.T) {
	// Arrange
	observer := NewNoopObserver()

	// Act
	config := observer.GetConfig()

	// Assert
	assert.Equal(t, "noop", config.Provider)
	assert.False(t, config.Enabled)
	assert.Equal(t, "development", config.Environment)
	assert.False(t, config.Debug)
	assert.Equal(t, float64(0), config.SampleRate)
	assert.False(t, config.Middleware.CaptureErrors)
	assert.False(t, config.Metrics.Enabled)
	assert.Equal(t, "noop", config.Metrics.Provider)
	assert.Equal(t, "/metrics", config.Metrics.Path)
	assert.Equal(t, 9090, config.Metrics.Port)
}

// =============================================================================
// NOOP TRANSACTION TESTS
// =============================================================================

func TestNoopTransaction_Context_WhenCalled_ShouldReturnBackgroundContext(t *testing.T) {
	// Arrange
	observer := NewNoopObserver()
	tx := observer.StartTransaction(context.Background(), "test")

	// Act
	ctx := tx.Context()

	// Assert
	assert.NotNil(t, ctx)
	assert.Equal(t, context.Background(), ctx)
}

func TestNoopTransaction_SetName(t *testing.T) {
	// Arrange
	transaction := &noopTransaction{}

	// Act & Assert - Should not panic
	assert.NotPanics(t, func() {
		transaction.SetName("test transaction")
	})
}

func TestNoopTransaction_SetTag(t *testing.T) {
	// Arrange
	transaction := &noopTransaction{}

	// Act & Assert - Should not panic
	assert.NotPanics(t, func() {
		transaction.SetTag("key", "value")
	})
}

func TestNoopTransaction_SetData(t *testing.T) {
	// Arrange
	transaction := &noopTransaction{}

	// Act & Assert - Should not panic
	assert.NotPanics(t, func() {
		transaction.SetData("key", map[string]interface{}{"test": "data"})
	})
}

func TestNoopTransaction_StartChild(t *testing.T) {
	// Arrange
	transaction := &noopTransaction{}

	// Act
	span := transaction.StartChild("test operation")

	// Assert
	assert.NotNil(t, span)
	assert.Implements(t, (*Span)(nil), span)
}

func TestNoopTransaction_Finish(t *testing.T) {
	// Arrange
	transaction := &noopTransaction{}

	// Act & Assert - Should not panic
	assert.NotPanics(t, func() {
		transaction.Finish(TransactionOK)
	})
}

// =============================================================================
// NOOP SPAN TESTS
// =============================================================================

func TestNoopSpan_SetTag(t *testing.T) {
	// Arrange
	span := &noopSpan{}

	// Act & Assert - Should not panic
	assert.NotPanics(t, func() {
		span.SetTag("key", "value")
	})
}

func TestNoopSpan_SetData(t *testing.T) {
	// Arrange
	span := &noopSpan{}

	// Act & Assert - Should not panic
	assert.NotPanics(t, func() {
		span.SetData("key", "value")
	})
}

func TestNoopSpan_SetError(t *testing.T) {
	// Arrange
	span := &noopSpan{}
	testError := errors.New("test error")

	// Act & Assert - Should not panic
	assert.NotPanics(t, func() {
		span.SetError(testError)
	})
}

func TestNoopSpan_Finish(t *testing.T) {
	// Arrange
	span := &noopSpan{}

	// Act & Assert - Should not panic
	assert.NotPanics(t, func() {
		span.Finish()
	})
}

// =============================================================================
// INTEGRATION TESTS
// =============================================================================

func TestNoopObserver_Integration(t *testing.T) {
	// Arrange
	observer := NewNoopObserver()
	ctx := context.Background()

	// Act - Test complete workflow
	transaction := observer.StartTransaction(ctx, "test transaction")
	transaction.SetName("updated name")
	transaction.SetTag("env", "test")
	transaction.SetData("metadata", map[string]string{"key": "value"})

	span := transaction.StartChild("child operation")
	span.SetTag("operation", "test")
	span.SetData("result", "success")
	span.SetError(errors.New("test error"))
	span.Finish()

	transaction.Finish(TransactionOK)

	observer.CaptureException(ctx, errors.New("test exception"))
	observer.CaptureMessage(ctx, "test message", LevelError)

	// Assert - Should complete without panics
	assert.NotNil(t, observer)
	assert.NotNil(t, transaction)
	assert.NotNil(t, span)
}
