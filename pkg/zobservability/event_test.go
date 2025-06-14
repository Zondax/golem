package zobservability

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWithEventLevel_WhenAppliedToEvent_ShouldNotPanic(t *testing.T) {
	// Arrange
	event := NewMockEvent(t)
	option := WithEventLevel(LevelError)

	event.On("SetLevel", LevelError).Return()

	// Act & Assert - should not panic
	assert.NotPanics(t, func() {
		option.ApplyEvent(event)
	})

	event.AssertExpectations(t)
}

func TestWithEventLevel_WhenAppliedWithDifferentLevels_ShouldNotPanic(t *testing.T) {
	// Arrange
	levels := []Level{LevelDebug, LevelInfo, LevelWarning, LevelError, LevelFatal}

	for _, level := range levels {
		t.Run(level.String(), func(t *testing.T) {
			event := NewMockEvent(t)
			option := WithEventLevel(level)

			event.On("SetLevel", level).Return()

			// Act & Assert - should not panic
			assert.NotPanics(t, func() {
				option.ApplyEvent(event)
			})

			event.AssertExpectations(t)
		})
	}
}

func TestWithEventTags_WhenAppliedToEvent_ShouldNotPanic(t *testing.T) {
	// Arrange
	event := NewMockEvent(t)
	tags := map[string]string{
		"component": "user-service",
		"operation": "create-user",
		"layer":     LayerService,
	}
	option := WithEventTags(tags)

	event.On("SetTags", tags).Return()

	// Act & Assert - should not panic
	assert.NotPanics(t, func() {
		option.ApplyEvent(event)
	})

	event.AssertExpectations(t)
}

func TestWithEventTags_WhenAppliedWithEmptyTags_ShouldNotPanic(t *testing.T) {
	// Arrange
	event := NewMockEvent(t)
	tags := map[string]string{}
	option := WithEventTags(tags)

	event.On("SetTags", tags).Return()

	// Act & Assert - should not panic
	assert.NotPanics(t, func() {
		option.ApplyEvent(event)
	})

	event.AssertExpectations(t)
}

func TestWithEventTags_WhenAppliedWithNilTags_ShouldNotPanic(t *testing.T) {
	// Arrange
	event := NewMockEvent(t)
	option := WithEventTags(nil)

	event.On("SetTags", map[string]string(nil)).Return()

	// Act & Assert - should not panic
	assert.NotPanics(t, func() {
		option.ApplyEvent(event)
	})

	event.AssertExpectations(t)
}

func TestWithEventTag_WhenAppliedToEvent_ShouldNotPanic(t *testing.T) {
	// Arrange
	event := NewMockEvent(t)
	option := WithEventTag("component", "user-service")

	event.On("SetTag", "component", "user-service").Return()

	// Act & Assert - should not panic
	assert.NotPanics(t, func() {
		option.ApplyEvent(event)
	})

	event.AssertExpectations(t)
}

func TestWithEventTag_WhenAppliedWithEmptyKey_ShouldNotPanic(t *testing.T) {
	// Arrange
	event := NewMockEvent(t)
	option := WithEventTag("", "value")

	event.On("SetTag", "", "value").Return()

	// Act & Assert - should not panic
	assert.NotPanics(t, func() {
		option.ApplyEvent(event)
	})

	event.AssertExpectations(t)
}

func TestWithEventTag_WhenAppliedWithEmptyValue_ShouldNotPanic(t *testing.T) {
	// Arrange
	event := NewMockEvent(t)
	option := WithEventTag("key", "")

	event.On("SetTag", "key", "").Return()

	// Act & Assert - should not panic
	assert.NotPanics(t, func() {
		option.ApplyEvent(event)
	})

	event.AssertExpectations(t)
}

func TestWithEventUser_WhenAppliedToEvent_ShouldNotPanic(t *testing.T) {
	// Arrange
	event := NewMockEvent(t)
	option := WithEventUser("user123", "test@example.com", "testuser")

	event.On("SetUser", "user123", "test@example.com", "testuser").Return()

	// Act & Assert - should not panic
	assert.NotPanics(t, func() {
		option.ApplyEvent(event)
	})

	event.AssertExpectations(t)
}

func TestWithEventUser_WhenAppliedWithEmptyValues_ShouldNotPanic(t *testing.T) {
	// Arrange
	event := NewMockEvent(t)
	option := WithEventUser("", "", "")

	event.On("SetUser", "", "", "").Return()

	// Act & Assert - should not panic
	assert.NotPanics(t, func() {
		option.ApplyEvent(event)
	})

	event.AssertExpectations(t)
}

func TestWithEventFingerprint_WhenAppliedToEvent_ShouldNotPanic(t *testing.T) {
	// Arrange
	event := NewMockEvent(t)
	fingerprint := []string{"error", "database", "connection"}
	option := WithEventFingerprint(fingerprint)

	event.On("SetFingerprint", fingerprint).Return()

	// Act & Assert - should not panic
	assert.NotPanics(t, func() {
		option.ApplyEvent(event)
	})

	event.AssertExpectations(t)
}

func TestWithEventFingerprint_WhenAppliedWithEmptySlice_ShouldNotPanic(t *testing.T) {
	// Arrange
	event := NewMockEvent(t)
	fingerprint := []string{}
	option := WithEventFingerprint(fingerprint)

	event.On("SetFingerprint", fingerprint).Return()

	// Act & Assert - should not panic
	assert.NotPanics(t, func() {
		option.ApplyEvent(event)
	})

	event.AssertExpectations(t)
}

func TestWithEventFingerprint_WhenAppliedWithNilSlice_ShouldNotPanic(t *testing.T) {
	// Arrange
	event := NewMockEvent(t)
	option := WithEventFingerprint(nil)

	event.On("SetFingerprint", []string(nil)).Return()

	// Act & Assert - should not panic
	assert.NotPanics(t, func() {
		option.ApplyEvent(event)
	})

	event.AssertExpectations(t)
}

func TestWithEventError_WhenAppliedToEvent_ShouldNotPanic(t *testing.T) {
	// Arrange
	event := NewMockEvent(t)
	err := errors.New("test error")
	option := WithEventError(err)

	event.On("SetError", err).Return()

	// Act & Assert - should not panic
	assert.NotPanics(t, func() {
		option.ApplyEvent(event)
	})

	event.AssertExpectations(t)
}

func TestWithEventError_WhenAppliedWithNilError_ShouldNotPanic(t *testing.T) {
	// Arrange
	event := NewMockEvent(t)
	option := WithEventError(nil)

	event.On("SetError", error(nil)).Return()

	// Act & Assert - should not panic
	assert.NotPanics(t, func() {
		option.ApplyEvent(event)
	})

	event.AssertExpectations(t)
}

func TestEventOptionFunc_WhenImplementsEventOption_ShouldBeValidInterface(t *testing.T) {
	// Arrange
	var option EventOption = eventOptionFunc(func(e Event) {
		e.SetLevel(LevelInfo)
	})

	// Assert
	assert.NotNil(t, option)
	assert.Implements(t, (*EventOption)(nil), option)
}

func TestEventOptionFunc_WhenAppliedToEvent_ShouldExecuteFunction(t *testing.T) {
	// Arrange
	event := NewMockEvent(t)
	executed := false

	option := eventOptionFunc(func(e Event) {
		executed = true
		e.SetLevel(LevelInfo)
	})

	event.On("SetLevel", LevelInfo).Return()

	// Act
	option.ApplyEvent(event)

	// Assert
	assert.True(t, executed)
	event.AssertExpectations(t)
}

func TestEventOptions_WhenMultipleOptionsApplied_ShouldNotPanic(t *testing.T) {
	// Arrange
	event := NewMockEvent(t)
	err := errors.New("test error")
	tags := map[string]string{"component": "test"}
	fingerprint := []string{"error", "test"}

	options := []EventOption{
		WithEventLevel(LevelError),
		WithEventTags(tags),
		WithEventTag("additional", "tag"),
		WithEventUser("user123", "test@example.com", "testuser"),
		WithEventFingerprint(fingerprint),
		WithEventError(err),
	}

	// Set up expectations
	event.On("SetLevel", LevelError).Return()
	event.On("SetTags", tags).Return()
	event.On("SetTag", "additional", "tag").Return()
	event.On("SetUser", "user123", "test@example.com", "testuser").Return()
	event.On("SetFingerprint", fingerprint).Return()
	event.On("SetError", err).Return()

	// Act & Assert - should not panic
	for _, option := range options {
		assert.NotPanics(t, func() {
			option.ApplyEvent(event)
		})
	}

	event.AssertExpectations(t)
}

func TestEventOptions_WhenUsedWithCommonTags_ShouldNotPanic(t *testing.T) {
	// Arrange
	event := NewMockEvent(t)
	commonTags := map[string]string{
		TagOperation: "create-user",
		TagService:   "user-service",
		TagComponent: "user-repository",
		TagLayer:     LayerService,
		TagMethod:    "CreateUser",
	}

	var options []EventOption
	for key, value := range commonTags {
		options = append(options, WithEventTag(key, value))
		event.On("SetTag", key, value).Return()
	}

	// Act & Assert - should not panic
	for _, option := range options {
		assert.NotPanics(t, func() {
			option.ApplyEvent(event)
		})
	}

	event.AssertExpectations(t)
}

func TestEventOption_WhenImplementedAsInterface_ShouldWork(t *testing.T) {
	// Test that all option types implement the interface correctly
	options := []EventOption{
		WithEventLevel(LevelInfo),
		WithEventTags(map[string]string{"key": "value"}),
		WithEventTag("key", "value"),
		WithEventUser("id", "email", "username"),
		WithEventFingerprint([]string{"test"}),
		WithEventError(errors.New("error")),
	}

	// Act & Assert
	for i, option := range options {
		t.Run(string(rune('A'+i)), func(t *testing.T) {
			assert.Implements(t, (*EventOption)(nil), option)
			assert.NotNil(t, option)
		})
	}
}

func TestEventOptions_WhenUsedInRealScenario_ShouldWorkCorrectly(t *testing.T) {
	// Arrange
	observer := NewNoopObserver()
	ctx := context.Background()
	err := errors.New("database connection failed")

	// Act - simulate real usage scenario with CaptureException
	assert.NotPanics(t, func() {
		observer.CaptureException(ctx, err,
			WithEventLevel(LevelError),
			WithEventTag(TagComponent, "database"),
			WithEventTag(TagOperation, "connect"),
			WithEventUser("user123", "test@example.com", "testuser"),
			WithEventFingerprint([]string{"database", "connection", "error"}),
		)
	})

	// Act - simulate real usage scenario with CaptureMessage
	assert.NotPanics(t, func() {
		observer.CaptureMessage(ctx, "User created successfully", LevelInfo,
			WithEventTag(TagComponent, "user-service"),
			WithEventTag(TagOperation, "create-user"),
			WithEventUser("user123", "test@example.com", "testuser"),
		)
	})
}
