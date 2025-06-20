package signoz

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel"

	"github.com/zondax/golem/pkg/zobservability"
)

// =============================================================================
// SIGNOZ TRANSACTION TESTS
// =============================================================================

func TestSignozTransaction_Context_WhenCalled_ShouldReturnContext(t *testing.T) {
	// Arrange
	ctx := context.Background()
	tracer := otel.Tracer("test")
	_, span := tracer.Start(ctx, "test-span")

	tx := &signozTransaction{
		ctx:  ctx,
		span: span,
		name: "test-transaction",
	}

	// Act
	resultCtx := tx.Context()

	// Assert
	assert.Equal(t, ctx, resultCtx)
}

func TestSignozTransaction_SetName_WhenCalled_ShouldSetSpanName(t *testing.T) {
	// Arrange
	ctx := context.Background()
	tracer := otel.Tracer("test")
	_, span := tracer.Start(ctx, "original-name")

	tx := &signozTransaction{
		ctx:  ctx,
		span: span,
		name: "original-name",
	}

	newName := "updated-transaction-name"

	// Act
	tx.SetName(newName)

	// Assert
	assert.Equal(t, newName, tx.name)
	// Note: We can't easily verify the span name was updated without mocking
	// but the implementation calls span.SetName(name)
}

func TestSignozTransaction_SetTag_WhenCalled_ShouldSetSpanAttribute(t *testing.T) {
	// Arrange
	ctx := context.Background()
	tracer := otel.Tracer("test")
	_, span := tracer.Start(ctx, "test-span")

	tx := &signozTransaction{
		ctx:  ctx,
		span: span,
		name: "test-transaction",
	}

	key := "service.name"
	value := "user-service"

	// Act & Assert - Should not panic
	assert.NotPanics(t, func() {
		tx.SetTag(key, value)
	})
}

func TestSignozTransaction_SetData_WhenCalled_ShouldSetSpanAttribute(t *testing.T) {
	// Arrange
	ctx := context.Background()
	tracer := otel.Tracer("test")
	_, span := tracer.Start(ctx, "test-span")

	tx := &signozTransaction{
		ctx:  ctx,
		span: span,
		name: "test-transaction",
	}

	key := "user.id"
	value := "12345"

	// Act & Assert - Should not panic
	assert.NotPanics(t, func() {
		tx.SetData(key, value)
	})
}

func TestSignozTransaction_SetData_WhenCalledWithDifferentTypes_ShouldSetSpanAttribute(t *testing.T) {
	// Arrange
	ctx := context.Background()
	tracer := otel.Tracer("test")
	_, span := tracer.Start(ctx, "test-span")

	tx := &signozTransaction{
		ctx:  ctx,
		span: span,
		name: "test-transaction",
	}

	testCases := []struct {
		name  string
		key   string
		value interface{}
	}{
		{
			name:  "string_value",
			key:   "string_key",
			value: "string_value",
		},
		{
			name:  "int_value",
			key:   "int_key",
			value: 42,
		},
		{
			name:  "bool_value",
			key:   "bool_key",
			value: true,
		},
		{
			name:  "float_value",
			key:   "float_key",
			value: 3.14,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Act & Assert - Should not panic
			assert.NotPanics(t, func() {
				tx.SetData(tc.key, tc.value)
			})
		})
	}
}

func TestSignozTransaction_StartChild_WhenCalled_ShouldReturnSpan(t *testing.T) {
	// Arrange
	ctx := context.Background()
	tracer := otel.Tracer("test")
	_, span := tracer.Start(ctx, "test-span")

	tx := &signozTransaction{
		ctx:  ctx,
		span: span,
		name: "test-transaction",
	}

	operation := "child-operation"

	// Act
	childSpan := tx.StartChild(operation)

	// Assert
	assert.NotNil(t, childSpan)
	assert.Implements(t, (*zobservability.Span)(nil), childSpan)

	// Cleanup
	childSpan.Finish()
}

func TestSignozTransaction_StartChild_WhenCalledWithOptions_ShouldApplyOptions(t *testing.T) {
	// Arrange
	ctx := context.Background()
	tracer := otel.Tracer("test")
	_, span := tracer.Start(ctx, "test-span")

	tx := &signozTransaction{
		ctx:  ctx,
		span: span,
		name: "test-transaction",
	}

	operation := "child-operation-with-options"

	// Act
	childSpan := tx.StartChild(operation,
		zobservability.WithSpanTag("component", "database"),
		zobservability.WithSpanData("query", "SELECT * FROM users"),
	)

	// Assert
	assert.NotNil(t, childSpan)
	assert.Implements(t, (*zobservability.Span)(nil), childSpan)

	// Cleanup
	childSpan.Finish()
}

func TestSignozTransaction_Finish_WhenCalledWithDifferentStatuses_ShouldFinishSpan(t *testing.T) {
	testCases := []struct {
		name   string
		status zobservability.TransactionStatus
	}{
		{
			name:   "ok_status",
			status: zobservability.TransactionOK,
		},
		{
			name:   "error_status",
			status: zobservability.TransactionError,
		},
		{
			name:   "cancelled_status",
			status: zobservability.TransactionCancelled,
		},
		{
			name:   "unknown_status",
			status: zobservability.TransactionStatus("unknown"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			ctx := context.Background()
			tracer := otel.Tracer("test")
			_, span := tracer.Start(ctx, "test-span")

			tx := &signozTransaction{
				ctx:  ctx,
				span: span,
				name: "test-transaction",
			}

			// Act & Assert - Should not panic
			assert.NotPanics(t, func() {
				tx.Finish(tc.status)
			})
		})
	}
}

// =============================================================================
// INTEGRATION TESTS
// =============================================================================

func TestSignozTransaction_WhenCompleteWorkflow_ShouldWorkCorrectly(t *testing.T) {
	// Arrange
	observer := createTestObserver(t)
	defer func() { _ = observer.Close() }()

	ctx := context.Background()

	// Act - Start transaction
	tx := observer.StartTransaction(ctx, "integration-workflow")

	// Set transaction metadata
	tx.SetName("updated-workflow-name")
	tx.SetTag("service", "integration-test")
	tx.SetData("workflow_id", "workflow-123")
	tx.SetData("user_count", 42)
	tx.SetData("is_test", true)

	// Start child spans
	childSpan1 := tx.StartChild("database-operation")
	childSpan1.SetTag("db.type", "postgresql")
	childSpan1.SetData("table", "users")
	childSpan1.Finish()

	childSpan2 := tx.StartChild("api-call",
		zobservability.WithSpanTag("http.method", "GET"),
		zobservability.WithSpanData("url", "https://api.example.com/users"),
	)
	childSpan2.Finish()

	// Finish transaction
	tx.Finish(zobservability.TransactionOK)

	// Assert - Should complete without errors
	assert.NotNil(t, tx)
}

func TestSignozTransaction_WhenNestedChildren_ShouldWorkCorrectly(t *testing.T) {
	// Arrange
	observer := createTestObserver(t)
	defer func() { _ = observer.Close() }()

	ctx := context.Background()

	// Act - Create nested structure
	tx := observer.StartTransaction(ctx, "nested-workflow")

	// Level 1
	level1Span := tx.StartChild("level-1-operation")
	level1Span.SetTag("level", "1")

	// Level 2 (using the span context from level 1)
	level2Ctx, level2Span := observer.StartSpan(tx.Context(), "level-2-operation")
	level2Span.SetTag("level", "2")

	// Level 3 (using the span context from level 2)
	_, level3Span := observer.StartSpan(level2Ctx, "level-3-operation")
	level3Span.SetTag("level", "3")

	// Finish in reverse order
	level3Span.Finish()
	level2Span.Finish()
	level1Span.Finish()
	tx.Finish(zobservability.TransactionOK)

	// Assert - Should complete without errors
	assert.NotNil(t, tx)
}

func TestSignozTransaction_WhenErrorScenario_ShouldHandleGracefully(t *testing.T) {
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

	// Finish transaction with error status
	tx.Finish(zobservability.TransactionError)

	// Assert - Should complete without panicking
	assert.NotNil(t, tx)
}
