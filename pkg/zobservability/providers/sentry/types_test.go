package sentry

import (
	"context"
	"testing"

	"github.com/getsentry/sentry-go"
	"github.com/stretchr/testify/assert"

	"github.com/zondax/golem/pkg/zobservability"
)

// =============================================================================
// SENTRY TRANSACTION TESTS
// =============================================================================

func TestSentryTransaction_Context_WhenCalled_ShouldReturnContext(t *testing.T) {
	// Arrange
	ctx := context.Background()
	hub := sentry.NewHub(nil, sentry.NewScope())
	tx := sentry.StartTransaction(ctx, "test-transaction")

	sentryTx := &sentryTransaction{
		hub: hub,
		tx:  tx,
		ctx: ctx,
	}

	// Act
	resultCtx := sentryTx.Context()

	// Assert
	assert.Equal(t, ctx, resultCtx)
}

func TestSentryTransaction_SetName_WhenCalled_ShouldSetTransactionName(t *testing.T) {
	// Arrange
	ctx := context.Background()
	hub := sentry.NewHub(nil, sentry.NewScope())
	tx := sentry.StartTransaction(ctx, "original-name")

	sentryTx := &sentryTransaction{
		hub: hub,
		tx:  tx,
		ctx: ctx,
	}

	newName := "updated-transaction-name"

	// Act
	sentryTx.SetName(newName)

	// Assert
	assert.Equal(t, newName, tx.Name)
}

func TestSentryTransaction_SetTag_WhenCalled_ShouldSetTransactionTag(t *testing.T) {
	// Arrange
	ctx := context.Background()
	hub := sentry.NewHub(nil, sentry.NewScope())
	tx := sentry.StartTransaction(ctx, "test-transaction")

	sentryTx := &sentryTransaction{
		hub: hub,
		tx:  tx,
		ctx: ctx,
	}

	key := "service.name"
	value := "user-service"

	// Act & Assert - Should not panic
	assert.NotPanics(t, func() {
		sentryTx.SetTag(key, value)
	})

	// Verify tag was set
	assert.Equal(t, value, tx.Tags[key])
}

func TestSentryTransaction_SetData_WhenCalled_ShouldSetTransactionData(t *testing.T) {
	// Arrange
	ctx := context.Background()
	hub := sentry.NewHub(nil, sentry.NewScope())
	tx := sentry.StartTransaction(ctx, "test-transaction")

	sentryTx := &sentryTransaction{
		hub: hub,
		tx:  tx,
		ctx: ctx,
	}

	key := "user.id"
	value := "12345"

	// Act & Assert - Should not panic
	assert.NotPanics(t, func() {
		sentryTx.SetData(key, value)
	})

	// Verify data was set
	assert.Equal(t, value, tx.Data[key])
}

func TestSentryTransaction_SetData_WhenCalledWithDifferentTypes_ShouldSetTransactionData(t *testing.T) {
	// Arrange
	ctx := context.Background()
	hub := sentry.NewHub(nil, sentry.NewScope())
	tx := sentry.StartTransaction(ctx, "test-transaction")

	sentryTx := &sentryTransaction{
		hub: hub,
		tx:  tx,
		ctx: ctx,
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
				sentryTx.SetData(tc.key, tc.value)
			})

			// Verify data was set
			assert.Equal(t, tc.value, tx.Data[tc.key])
		})
	}
}

func TestSentryTransaction_StartChild_WhenCalled_ShouldReturnSpan(t *testing.T) {
	// Arrange
	ctx := context.Background()
	hub := sentry.NewHub(nil, sentry.NewScope())
	tx := sentry.StartTransaction(ctx, "test-transaction")

	sentryTx := &sentryTransaction{
		hub: hub,
		tx:  tx,
		ctx: ctx,
	}

	operation := "child-operation"

	// Act
	childSpan := sentryTx.StartChild(operation)

	// Assert
	assert.NotNil(t, childSpan)
	assert.Implements(t, (*zobservability.Span)(nil), childSpan)

	// Cleanup
	childSpan.Finish()
}

func TestSentryTransaction_StartChild_WhenCalledWithOptions_ShouldApplyOptions(t *testing.T) {
	// Arrange
	ctx := context.Background()
	hub := sentry.NewHub(nil, sentry.NewScope())
	tx := sentry.StartTransaction(ctx, "test-transaction")

	sentryTx := &sentryTransaction{
		hub: hub,
		tx:  tx,
		ctx: ctx,
	}

	operation := "child-operation-with-options"

	// Act
	childSpan := sentryTx.StartChild(operation,
		zobservability.WithSpanTag("component", "database"),
		zobservability.WithSpanData("query", "SELECT * FROM users"),
	)

	// Assert
	assert.NotNil(t, childSpan)
	assert.Implements(t, (*zobservability.Span)(nil), childSpan)

	// Cleanup
	childSpan.Finish()
}

func TestSentryTransaction_Finish_WhenCalledWithDifferentStatuses_ShouldFinishTransaction(t *testing.T) {
	testCases := []struct {
		name           string
		status         zobservability.TransactionStatus
		expectedStatus sentry.SpanStatus
	}{
		{
			name:           "ok_status",
			status:         zobservability.TransactionOK,
			expectedStatus: sentry.SpanStatusOK,
		},
		{
			name:           "error_status",
			status:         zobservability.TransactionError,
			expectedStatus: sentry.SpanStatusInternalError,
		},
		{
			name:           "cancelled_status",
			status:         zobservability.TransactionCancelled,
			expectedStatus: sentry.SpanStatusCanceled,
		},
		{
			name:           "unknown_status",
			status:         zobservability.TransactionStatus("unknown"),
			expectedStatus: sentry.SpanStatusUnknown,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			ctx := context.Background()
			hub := sentry.NewHub(nil, sentry.NewScope())
			tx := sentry.StartTransaction(ctx, "test-transaction")

			sentryTx := &sentryTransaction{
				hub: hub,
				tx:  tx,
				ctx: ctx,
			}

			// Act & Assert - Should not panic
			assert.NotPanics(t, func() {
				sentryTx.Finish(tc.status)
			})

			// Verify status was set
			assert.Equal(t, tc.expectedStatus, tx.Status)
		})
	}
}

// =============================================================================
// SENTRY SPAN TESTS
// =============================================================================

func TestSentrySpan_SetTag_WhenCalled_ShouldSetSpanTag(t *testing.T) {
	// Arrange
	ctx := context.Background()
	hub := sentry.NewHub(nil, sentry.NewScope())
	tx := sentry.StartTransaction(ctx, "test-transaction")
	span := tx.StartChild("test-span")

	sentrySpan := &sentrySpan{
		hub:       hub,
		operation: "test-operation",
		span:      span,
	}

	key := "db.type"
	value := "postgresql"

	// Act & Assert - Should not panic
	assert.NotPanics(t, func() {
		sentrySpan.SetTag(key, value)
	})

	// Verify tag was set
	assert.Equal(t, value, span.Tags[key])
}

func TestSentrySpan_SetData_WhenCalled_ShouldSetSpanData(t *testing.T) {
	// Arrange
	ctx := context.Background()
	hub := sentry.NewHub(nil, sentry.NewScope())
	tx := sentry.StartTransaction(ctx, "test-transaction")
	span := tx.StartChild("test-span")

	sentrySpan := &sentrySpan{
		hub:       hub,
		operation: "test-operation",
		span:      span,
	}

	key := "query"
	value := "SELECT * FROM users"

	// Act & Assert - Should not panic
	assert.NotPanics(t, func() {
		sentrySpan.SetData(key, value)
	})

	// Verify data was set
	assert.Equal(t, value, span.Data[key])
}

func TestSentrySpan_SetError_WhenCalledWithError_ShouldSetSpanError(t *testing.T) {
	// Arrange
	ctx := context.Background()
	hub := sentry.NewHub(nil, sentry.NewScope())
	tx := sentry.StartTransaction(ctx, "test-transaction")
	span := tx.StartChild("test-span")

	sentrySpan := &sentrySpan{
		hub:       hub,
		operation: "test-operation",
		span:      span,
	}

	testErr := assert.AnError

	// Act & Assert - Should not panic
	assert.NotPanics(t, func() {
		sentrySpan.SetError(testErr)
	})

	// Verify error status was set
	assert.Equal(t, sentry.SpanStatusInternalError, span.Status)
	assert.Equal(t, testErr.Error(), span.Tags["error"])
}

func TestSentrySpan_SetError_WhenCalledWithNilError_ShouldNotSetError(t *testing.T) {
	// Arrange
	ctx := context.Background()
	hub := sentry.NewHub(nil, sentry.NewScope())
	tx := sentry.StartTransaction(ctx, "test-transaction")
	span := tx.StartChild("test-span")

	sentrySpan := &sentrySpan{
		hub:       hub,
		operation: "test-operation",
		span:      span,
	}

	// Act & Assert - Should not panic
	assert.NotPanics(t, func() {
		sentrySpan.SetError(nil)
	})

	// Verify no error status was set
	assert.NotEqual(t, sentry.SpanStatusInternalError, span.Status)
	assert.Empty(t, span.Tags["error"])
}

func TestSentrySpan_Finish_WhenCalled_ShouldFinishSpan(t *testing.T) {
	// Arrange
	ctx := context.Background()
	hub := sentry.NewHub(nil, sentry.NewScope())
	tx := sentry.StartTransaction(ctx, "test-transaction")
	span := tx.StartChild("test-span")

	sentrySpan := &sentrySpan{
		hub:       hub,
		operation: "test-operation",
		span:      span,
	}

	// Act & Assert - Should not panic
	assert.NotPanics(t, func() {
		sentrySpan.Finish()
	})
}

// =============================================================================
// SENTRY EVENT TESTS
// =============================================================================

func TestSentryEvent_SetLevel_WhenCalled_ShouldSetEventLevel(t *testing.T) {
	// Arrange
	hub := sentry.NewHub(nil, sentry.NewScope())
	event := sentry.NewEvent()

	sentryEvent := &sentryEvent{
		hub:   hub,
		event: event,
	}

	level := zobservability.LevelError

	// Act
	sentryEvent.SetLevel(level)

	// Assert
	assert.Equal(t, sentry.LevelError, event.Level)
}

func TestSentryEvent_SetTags_WhenCalled_ShouldSetEventTags(t *testing.T) {
	// Arrange
	hub := sentry.NewHub(nil, sentry.NewScope())
	event := sentry.NewEvent()

	sentryEvent := &sentryEvent{
		hub:   hub,
		event: event,
	}

	tags := map[string]string{
		"component": "database",
		"operation": "query",
		"severity":  "high",
	}

	// Act
	sentryEvent.SetTags(tags)

	// Assert
	assert.NotNil(t, event.Tags)
	for key, value := range tags {
		assert.Equal(t, value, event.Tags[key])
	}
}

func TestSentryEvent_SetTags_WhenCalledMultipleTimes_ShouldMergeTags(t *testing.T) {
	// Arrange
	hub := sentry.NewHub(nil, sentry.NewScope())
	event := sentry.NewEvent()

	sentryEvent := &sentryEvent{
		hub:   hub,
		event: event,
	}

	tags1 := map[string]string{
		"component": "database",
		"operation": "query",
	}

	tags2 := map[string]string{
		"severity": "high",
		"user_id":  "123",
	}

	// Act
	sentryEvent.SetTags(tags1)
	sentryEvent.SetTags(tags2)

	// Assert
	assert.NotNil(t, event.Tags)
	assert.Equal(t, "database", event.Tags["component"])
	assert.Equal(t, "query", event.Tags["operation"])
	assert.Equal(t, "high", event.Tags["severity"])
	assert.Equal(t, "123", event.Tags["user_id"])
}

func TestSentryEvent_SetTag_WhenCalled_ShouldSetSingleTag(t *testing.T) {
	// Arrange
	hub := sentry.NewHub(nil, sentry.NewScope())
	event := sentry.NewEvent()

	sentryEvent := &sentryEvent{
		hub:   hub,
		event: event,
	}

	key := "component"
	value := "api"

	// Act
	sentryEvent.SetTag(key, value)

	// Assert
	assert.NotNil(t, event.Tags)
	assert.Equal(t, value, event.Tags[key])
}

func TestSentryEvent_SetUser_WhenCalled_ShouldSetEventUser(t *testing.T) {
	// Arrange
	hub := sentry.NewHub(nil, sentry.NewScope())
	event := sentry.NewEvent()

	sentryEvent := &sentryEvent{
		hub:   hub,
		event: event,
	}

	id := "user123"
	email := "test@example.com"
	username := "testuser"

	// Act
	sentryEvent.SetUser(id, email, username)

	// Assert
	assert.Equal(t, id, event.User.ID)
	assert.Equal(t, email, event.User.Email)
	assert.Equal(t, username, event.User.Username)
}

func TestSentryEvent_SetFingerprint_WhenCalled_ShouldSetEventFingerprint(t *testing.T) {
	// Arrange
	hub := sentry.NewHub(nil, sentry.NewScope())
	event := sentry.NewEvent()

	sentryEvent := &sentryEvent{
		hub:   hub,
		event: event,
	}

	fingerprint := []string{"database", "connection", "timeout"}

	// Act
	sentryEvent.SetFingerprint(fingerprint)

	// Assert
	assert.Equal(t, fingerprint, event.Fingerprint)
}

func TestSentryEvent_SetError_WhenCalledWithError_ShouldSetEventException(t *testing.T) {
	// Arrange
	hub := sentry.NewHub(nil, sentry.NewScope())
	event := sentry.NewEvent()

	sentryEvent := &sentryEvent{
		hub:   hub,
		event: event,
	}

	testErr := assert.AnError

	// Act
	sentryEvent.SetError(testErr)

	// Assert
	assert.NotEmpty(t, event.Exception)
	assert.Equal(t, testErr.Error(), event.Exception[0].Value)
	assert.Equal(t, "error", event.Exception[0].Type)
	assert.NotNil(t, event.Exception[0].Stacktrace)
}

func TestSentryEvent_SetError_WhenCalledWithNilError_ShouldNotSetException(t *testing.T) {
	// Arrange
	hub := sentry.NewHub(nil, sentry.NewScope())
	event := sentry.NewEvent()

	sentryEvent := &sentryEvent{
		hub:   hub,
		event: event,
	}

	// Act
	sentryEvent.SetError(nil)

	// Assert
	assert.Empty(t, event.Exception)
}

func TestSentryEvent_Capture_WhenCalled_ShouldNotPanic(t *testing.T) {
	// Arrange
	hub := sentry.NewHub(nil, sentry.NewScope())
	event := sentry.NewEvent()

	sentryEvent := &sentryEvent{
		hub:   hub,
		event: event,
	}

	// Act & Assert - Should not panic
	assert.NotPanics(t, func() {
		sentryEvent.Capture()
	})
}

// =============================================================================
// INTEGRATION TESTS
// =============================================================================

func TestSentryTypes_WhenCompleteWorkflow_ShouldWorkCorrectly(t *testing.T) {
	// Arrange
	observer := createTestObserver(t)
	defer func() { _ = observer.Close() }()

	ctx := context.Background()

	// Act - Start transaction
	tx := observer.StartTransaction(ctx, "integration-workflow")
	tx.SetName("updated-workflow-name")
	tx.SetTag("service", "integration-test")
	tx.SetData("workflow_id", "workflow-123")

	// Start child span
	childSpan := tx.StartChild("database-operation")
	childSpan.SetTag("db.type", "postgresql")
	childSpan.SetData("table", "users")

	// Simulate error
	testErr := assert.AnError
	childSpan.SetError(testErr)
	childSpan.Finish()

	// Capture exception
	observer.CaptureException(tx.Context(), testErr,
		zobservability.WithEventTag("error.type", "database_error"),
		zobservability.WithEventUser("user123", "test@example.com", "testuser"),
	)

	// Finish transaction
	tx.Finish(zobservability.TransactionError)

	// Assert - Should complete without errors
	assert.NotNil(t, tx)
}
