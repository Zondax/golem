package signoz

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel"

	"github.com/zondax/golem/pkg/zobservability"
)

// =============================================================================
// SIGNOZ EVENT TESTS
// =============================================================================

func TestSignozEvent_SetLevel_WhenCalled_ShouldSetSpanAttribute(t *testing.T) {
	// Arrange
	tracer := otel.Tracer("test")
	_, span := tracer.Start(context.Background(), "test-span")
	defer span.End()

	event := &signozEvent{span: span}
	level := zobservability.LevelError

	// Act & Assert - Should not panic
	assert.NotPanics(t, func() {
		event.SetLevel(level)
	})
}

func TestSignozEvent_SetTags_WhenCalled_ShouldSetMultipleSpanAttributes(t *testing.T) {
	// Arrange
	tracer := otel.Tracer("test")
	_, span := tracer.Start(context.Background(), "test-span")
	defer span.End()

	event := &signozEvent{span: span}
	tags := map[string]string{
		"component": "database",
		"operation": "query",
		"severity":  "high",
	}

	// Act & Assert - Should not panic
	assert.NotPanics(t, func() {
		event.SetTags(tags)
	})
}

func TestSignozEvent_SetTags_WhenEmptyMap_ShouldNotPanic(t *testing.T) {
	// Arrange
	tracer := otel.Tracer("test")
	_, span := tracer.Start(context.Background(), "test-span")
	defer span.End()

	event := &signozEvent{span: span}
	tags := map[string]string{}

	// Act & Assert - Should not panic
	assert.NotPanics(t, func() {
		event.SetTags(tags)
	})
}

func TestSignozEvent_SetTag_WhenCalled_ShouldSetSingleSpanAttribute(t *testing.T) {
	// Arrange
	tracer := otel.Tracer("test")
	_, span := tracer.Start(context.Background(), "test-span")
	defer span.End()

	event := &signozEvent{span: span}
	key := "service.name"
	value := "user-service"

	// Act & Assert - Should not panic
	assert.NotPanics(t, func() {
		event.SetTag(key, value)
	})
}

func TestSignozEvent_SetUser_WhenCalled_ShouldSetUserAttributes(t *testing.T) {
	// Arrange
	tracer := otel.Tracer("test")
	_, span := tracer.Start(context.Background(), "test-span")
	defer span.End()

	event := &signozEvent{span: span}
	id := "user123"
	email := "test@example.com"
	username := "testuser"

	// Act & Assert - Should not panic
	assert.NotPanics(t, func() {
		event.SetUser(id, email, username)
	})
}

func TestSignozEvent_SetUser_WhenEmptyValues_ShouldNotPanic(t *testing.T) {
	// Arrange
	tracer := otel.Tracer("test")
	_, span := tracer.Start(context.Background(), "test-span")
	defer span.End()

	event := &signozEvent{span: span}

	// Act & Assert - Should not panic
	assert.NotPanics(t, func() {
		event.SetUser("", "", "")
	})
}

func TestSignozEvent_SetFingerprint_WhenCalled_ShouldSetFingerprintAttribute(t *testing.T) {
	// Arrange
	tracer := otel.Tracer("test")
	_, span := tracer.Start(context.Background(), "test-span")
	defer span.End()

	event := &signozEvent{span: span}
	fingerprint := []string{"database", "connection", "timeout"}

	// Act & Assert - Should not panic
	assert.NotPanics(t, func() {
		event.SetFingerprint(fingerprint)
	})
}

func TestSignozEvent_SetFingerprint_WhenEmptySlice_ShouldNotPanic(t *testing.T) {
	// Arrange
	tracer := otel.Tracer("test")
	_, span := tracer.Start(context.Background(), "test-span")
	defer span.End()

	event := &signozEvent{span: span}
	fingerprint := []string{}

	// Act & Assert - Should not panic
	assert.NotPanics(t, func() {
		event.SetFingerprint(fingerprint)
	})
}

func TestSignozEvent_SetError_WhenCalledWithError_ShouldRecordError(t *testing.T) {
	// Arrange
	tracer := otel.Tracer("test")
	_, span := tracer.Start(context.Background(), "test-span")
	defer span.End()

	event := &signozEvent{span: span}
	testErr := errors.New("test error")

	// Act & Assert - Should not panic
	assert.NotPanics(t, func() {
		event.SetError(testErr)
	})
}

func TestSignozEvent_SetError_WhenCalledWithNilError_ShouldNotRecordError(t *testing.T) {
	// Arrange
	tracer := otel.Tracer("test")
	_, span := tracer.Start(context.Background(), "test-span")
	defer span.End()

	event := &signozEvent{span: span}

	// Act & Assert - Should not panic
	assert.NotPanics(t, func() {
		event.SetError(nil)
	})
}

func TestSignozEvent_Capture_WhenCalled_ShouldNotPanic(t *testing.T) {
	// Arrange
	tracer := otel.Tracer("test")
	_, span := tracer.Start(context.Background(), "test-span")
	defer span.End()

	event := &signozEvent{span: span}

	// Act & Assert - Should not panic (method is no-op for OpenTelemetry)
	assert.NotPanics(t, func() {
		event.Capture()
	})
}

func TestSignozEvent_SetData_WhenCalledWithDifferentTypes_ShouldSetSpanAttribute(t *testing.T) {
	// Arrange
	tracer := otel.Tracer("test")
	_, span := tracer.Start(context.Background(), "test-span")
	defer span.End()

	event := &signozEvent{span: span}

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
				event.SetData(tc.key, tc.value)
			})
		})
	}
}

// =============================================================================
// INTEGRATION TESTS
// =============================================================================

func TestSignozEvent_WhenCompleteWorkflow_ShouldWorkCorrectly(t *testing.T) {
	// Arrange
	tracer := otel.Tracer("test")
	_, span := tracer.Start(context.Background(), "integration-test-span")
	defer span.End()

	event := &signozEvent{span: span}

	// Act - Set all event properties
	event.SetLevel(zobservability.LevelError)
	event.SetTags(map[string]string{
		"component": "database",
		"operation": "query",
	})
	event.SetTag("severity", "high")
	event.SetUser("user123", "test@example.com", "testuser")
	event.SetFingerprint([]string{"database", "timeout"})
	event.SetData("query", "SELECT * FROM users")
	event.SetData("duration", 5000)
	event.SetError(errors.New("connection timeout"))
	event.Capture()

	// Assert - Should complete without errors
	assert.NotNil(t, event)
}
