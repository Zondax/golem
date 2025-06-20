package zobservability

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWithSpanTag_WhenAppliedToSpan_ShouldNotPanic(t *testing.T) {
	// Arrange
	span := &noopSpan{}
	option := WithSpanTag("key", "value")

	// Act & Assert - should not panic
	assert.NotPanics(t, func() {
		option.ApplySpan(span)
	})
}

func TestWithSpanTag_WhenAppliedWithEmptyKey_ShouldNotPanic(t *testing.T) {
	// Arrange
	span := &noopSpan{}
	option := WithSpanTag("", "value")

	// Act & Assert - should not panic
	assert.NotPanics(t, func() {
		option.ApplySpan(span)
	})
}

func TestWithSpanTag_WhenAppliedWithEmptyValue_ShouldNotPanic(t *testing.T) {
	// Arrange
	span := &noopSpan{}
	option := WithSpanTag("key", "")

	// Act & Assert - should not panic
	assert.NotPanics(t, func() {
		option.ApplySpan(span)
	})
}

func TestWithSpanData_WhenAppliedToSpan_ShouldNotPanic(t *testing.T) {
	// Arrange
	span := &noopSpan{}
	data := map[string]interface{}{"test": "data"}
	option := WithSpanData("key", data)

	// Act & Assert - should not panic
	assert.NotPanics(t, func() {
		option.ApplySpan(span)
	})
}

func TestWithSpanData_WhenAppliedWithNilData_ShouldNotPanic(t *testing.T) {
	// Arrange
	span := &noopSpan{}
	option := WithSpanData("key", nil)

	// Act & Assert - should not panic
	assert.NotPanics(t, func() {
		option.ApplySpan(span)
	})
}

func TestWithSpanData_WhenAppliedWithComplexData_ShouldNotPanic(t *testing.T) {
	// Arrange
	span := &noopSpan{}
	complexData := map[string]interface{}{
		"string":  "value",
		"number":  42,
		"boolean": true,
		"array":   []string{"a", "b", "c"},
		"nested": map[string]interface{}{
			"inner": "value",
		},
	}
	option := WithSpanData("complex", complexData)

	// Act & Assert - should not panic
	assert.NotPanics(t, func() {
		option.ApplySpan(span)
	})
}

func TestWithSpanError_WhenAppliedToSpan_ShouldNotPanic(t *testing.T) {
	// Arrange
	span := &noopSpan{}
	err := errors.New("test error")
	option := WithSpanError(err)

	// Act & Assert - should not panic
	assert.NotPanics(t, func() {
		option.ApplySpan(span)
	})
}

func TestWithSpanError_WhenAppliedWithNilError_ShouldNotPanic(t *testing.T) {
	// Arrange
	span := &noopSpan{}
	option := WithSpanError(nil)

	// Act & Assert - should not panic
	assert.NotPanics(t, func() {
		option.ApplySpan(span)
	})
}

func TestSpanOptionFunc_WhenImplementsSpanOption_ShouldBeValidInterface(t *testing.T) {
	// Arrange
	var option SpanOption = spanOptionFunc(func(s Span) {
		s.SetTag("test", "value")
	})

	// Assert
	assert.NotNil(t, option)
	assert.Implements(t, (*SpanOption)(nil), option)
}

func TestSpanOptionFunc_WhenAppliedToSpan_ShouldExecuteFunction(t *testing.T) {
	// Arrange
	span := &noopSpan{}
	executed := false

	option := spanOptionFunc(func(s Span) {
		executed = true
		s.SetTag("test", "value")
	})

	// Act
	option.ApplySpan(span)

	// Assert
	assert.True(t, executed)
}

func TestSpanOptions_WhenMultipleOptionsApplied_ShouldNotPanic(t *testing.T) {
	// Arrange
	span := &noopSpan{}
	err := errors.New("test error")
	data := map[string]interface{}{"key": "value"}

	options := []SpanOption{
		WithSpanTag("tag1", "value1"),
		WithSpanTag("tag2", "value2"),
		WithSpanData("data1", data),
		WithSpanError(err),
	}

	// Act & Assert - should not panic
	for _, option := range options {
		assert.NotPanics(t, func() {
			option.ApplySpan(span)
		})
	}
}

func TestSpanOptions_WhenAppliedToNoopSpan_ShouldNotPanic(t *testing.T) {
	// Arrange
	span := &noopSpan{}
	err := errors.New("test error")
	data := map[string]interface{}{"key": "value"}

	options := []SpanOption{
		WithSpanTag("tag1", "value1"),
		WithSpanData("data1", data),
		WithSpanError(err),
	}

	// Act & Assert - should not panic
	for _, option := range options {
		assert.NotPanics(t, func() {
			option.ApplySpan(span)
		})
	}
}

func TestWithSpanTag_WhenUsedWithCommonTags_ShouldNotPanic(t *testing.T) {
	// Arrange
	span := &noopSpan{}
	commonTags := map[string]string{
		TagOperation: "create-user",
		TagService:   "user-service",
		TagComponent: "user-repository",
		TagLayer:     LayerService,
		TagMethod:    "CreateUser",
	}

	var options []SpanOption
	for key, value := range commonTags {
		options = append(options, WithSpanTag(key, value))
	}

	// Act & Assert - should not panic
	for _, option := range options {
		assert.NotPanics(t, func() {
			option.ApplySpan(span)
		})
	}
}

func TestWithSpanData_WhenUsedWithDifferentDataTypes_ShouldNotPanic(t *testing.T) {
	// Arrange
	span := &noopSpan{}

	testCases := []struct {
		name string
		key  string
		data interface{}
	}{
		{"string_data", "string_key", "string_value"},
		{"int_data", "int_key", 42},
		{"bool_data", "bool_key", true},
		{"float_data", "float_key", 3.14},
		{"slice_data", "slice_key", []string{"a", "b", "c"}},
		{"map_data", "map_key", map[string]string{"nested": "value"}},
		{"nil_data", "nil_key", nil},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			option := WithSpanData(tc.key, tc.data)

			// Act & Assert - should not panic
			assert.NotPanics(t, func() {
				option.ApplySpan(span)
			})
		})
	}
}

func TestSpanOptions_WhenChainedWithBuilder_ShouldAccumulateCorrectly(t *testing.T) {
	// Arrange
	observer := NewNoopObserver()
	builder := NewSpanBuilder(observer, "test-operation")

	// Act
	result := builder.
		WithTag("custom1", "value1").
		WithTag("custom2", "value2")

	// Assert
	assert.Equal(t, builder, result)
	assert.Len(t, builder.options, 2)
}

func TestWithSpanError_WhenUsedWithDifferentErrorTypes_ShouldNotPanic(t *testing.T) {
	// Arrange
	span := &noopSpan{}

	testCases := []struct {
		name string
		err  error
	}{
		{"simple_error", errors.New("simple error")},
		{"formatted_error", errors.New("formatted error: value")},
		{"nil_error", nil},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			option := WithSpanError(tc.err)

			// Act & Assert - should not panic
			assert.NotPanics(t, func() {
				option.ApplySpan(span)
			})
		})
	}
}

func TestSpanOption_WhenImplementedAsInterface_ShouldWork(t *testing.T) {
	// Arrange
	span := &noopSpan{}

	// Test that all option types implement the interface correctly
	options := []SpanOption{
		WithSpanTag("key", "value"),
		WithSpanData("data", "value"),
		WithSpanError(errors.New("error")),
	}

	// Act & Assert
	for i, option := range options {
		t.Run(string(rune('A'+i)), func(t *testing.T) {
			assert.Implements(t, (*SpanOption)(nil), option)
			assert.NotPanics(t, func() {
				option.ApplySpan(span)
			})
		})
	}
}

func TestSpanOptions_WhenUsedInRealScenario_ShouldWorkCorrectly(t *testing.T) {
	// Arrange
	observer := NewNoopObserver()
	ctx := context.Background()

	// Act - simulate real usage scenario
	newCtx, span := observer.StartSpan(ctx, "test-operation",
		WithSpanTag(TagLayer, LayerService),
		WithSpanTag(TagComponent, "user-service"),
		WithSpanData("user_id", "123"),
		WithSpanError(errors.New("test error")),
	)

	// Assert
	assert.NotNil(t, newCtx)
	assert.NotNil(t, span)
	assert.Implements(t, (*Span)(nil), span)

	// Should not panic when finishing
	assert.NotPanics(t, func() {
		span.Finish()
	})
}
