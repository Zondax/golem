package zobservability

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	testCreateUserOperation = "create-user"
)

func TestNewSpanBuilder_WhenCalled_ShouldReturnValidBuilder(t *testing.T) {
	// Arrange
	observer := NewNoopObserver()
	operation := "test-operation"

	// Act
	builder := NewSpanBuilder(observer, operation)

	// Assert
	assert.NotNil(t, builder)
	assert.Equal(t, observer, builder.observer)
	assert.Equal(t, operation, builder.operation)
	assert.Empty(t, builder.options)
}

func TestSpanBuilder_WhenWithLayer_ShouldAddLayerOption(t *testing.T) {
	// Arrange
	observer := NewNoopObserver()
	builder := NewSpanBuilder(observer, "test-operation")
	layer := LayerService

	// Act
	result := builder.WithLayer(layer)

	// Assert
	assert.Equal(t, builder, result) // Should return same builder for chaining
	assert.Len(t, builder.options, 1)
}

func TestSpanBuilder_WhenWithService_ShouldAddServiceOption(t *testing.T) {
	// Arrange
	observer := NewNoopObserver()
	builder := NewSpanBuilder(observer, "test-operation")
	service := "user-service"

	// Act
	result := builder.WithService(service)

	// Assert
	assert.Equal(t, builder, result) // Should return same builder for chaining
	assert.Len(t, builder.options, 1)
}

func TestSpanBuilder_WhenWithComponent_ShouldAddComponentOption(t *testing.T) {
	// Arrange
	observer := NewNoopObserver()
	builder := NewSpanBuilder(observer, "test-operation")
	component := "user-repository"

	// Act
	result := builder.WithComponent(component)

	// Assert
	assert.Equal(t, builder, result) // Should return same builder for chaining
	assert.Len(t, builder.options, 1)
}

func TestSpanBuilder_WhenWithOperation_ShouldAddOperationOption(t *testing.T) {
	// Arrange
	observer := NewNoopObserver()
	builder := NewSpanBuilder(observer, "test-operation")
	operation := "custom-operation"

	// Act
	result := builder.WithOperation(operation)

	// Assert
	assert.Equal(t, builder, result) // Should return same builder for chaining
	assert.Len(t, builder.options, 1)
}

func TestSpanBuilder_WhenWithTag_ShouldAddTagOption(t *testing.T) {
	// Arrange
	observer := NewNoopObserver()
	builder := NewSpanBuilder(observer, "test-operation")
	key := "request-id"
	value := "123"

	// Act
	result := builder.WithTag(key, value)

	// Assert
	assert.Equal(t, builder, result) // Should return same builder for chaining
	assert.Len(t, builder.options, 1)
}

func TestSpanBuilder_WhenChainedCalls_ShouldAccumulateOptions(t *testing.T) {
	// Arrange
	observer := NewNoopObserver()
	builder := NewSpanBuilder(observer, "test-operation")

	// Act
	result := builder.
		WithLayer(LayerService).
		WithService("user-service").
		WithComponent("user-repository").
		WithOperation(testCreateUserOperation).
		WithTag("request-id", "123")

	// Assert
	assert.Equal(t, builder, result) // Should return same builder for chaining
	assert.Len(t, builder.options, 5)
}

func TestSpanBuilder_WhenStart_ShouldCallObserverStartSpan(t *testing.T) {
	// Arrange
	observer := NewNoopObserver()
	ctx := context.Background()
	builder := NewSpanBuilder(observer, "test-operation")

	// Act
	resultCtx, resultSpan := builder.Start(ctx)

	// Assert
	assert.NotNil(t, resultCtx)
	assert.NotNil(t, resultSpan)
	assert.Implements(t, (*Span)(nil), resultSpan)
}

func TestSpanBuilder_WhenStartWithOptions_ShouldPassOptionsToObserver(t *testing.T) {
	// Arrange
	observer := NewNoopObserver()
	ctx := context.Background()
	builder := NewSpanBuilder(observer, "test-operation").WithLayer(LayerService)

	// Act
	resultCtx, resultSpan := builder.Start(ctx)

	// Assert
	assert.NotNil(t, resultCtx)
	assert.NotNil(t, resultSpan)
	assert.Implements(t, (*Span)(nil), resultSpan)
}

func TestStartServiceSpan_WhenCalled_ShouldCreateSpanWithCorrectTags(t *testing.T) {
	// Arrange
	observer := NewNoopObserver()
	ctx := context.Background()
	operation := testCreateUserOperation
	service := "user-service"

	// Act
	resultCtx, resultSpan := StartServiceSpan(ctx, observer, operation, service)

	// Assert
	assert.NotNil(t, resultCtx)
	assert.NotNil(t, resultSpan)
	assert.Implements(t, (*Span)(nil), resultSpan)
}

func TestNewEventSpanBuilder_WhenCalled_ShouldReturnValidBuilder(t *testing.T) {
	// Arrange
	observer := NewNoopObserver()
	operation := testCreateUserOperation
	layer := LayerService

	// Act
	builder := NewEventSpanBuilder(observer, operation, layer)

	// Assert
	assert.NotNil(t, builder)
	assert.NotNil(t, builder.SpanBuilder)
	assert.Equal(t, layer, builder.layer)
	assert.Equal(t, observer, builder.observer)
	assert.Equal(t, layer+"."+operation, builder.operation)
}

func TestNewEventSpanBuilder_WhenStart_ShouldCreateSpanWithLayerPrefix(t *testing.T) {
	// Arrange
	observer := NewNoopObserver()
	ctx := context.Background()
	operation := testCreateUserOperation
	layer := LayerService
	builder := NewEventSpanBuilder(observer, operation, layer)

	// Act
	resultCtx, resultSpan := builder.Start(ctx)

	// Assert
	assert.NotNil(t, resultCtx)
	assert.NotNil(t, resultSpan)
	assert.Implements(t, (*Span)(nil), resultSpan)
	assert.Equal(t, layer+"."+operation, builder.operation)
}

func TestEventSpanBuilder_WhenChainedWithAdditionalOptions_ShouldAccumulateOptions(t *testing.T) {
	// Arrange
	observer := NewNoopObserver()
	ctx := context.Background()
	operation := testCreateUserOperation
	layer := LayerService
	builder := NewEventSpanBuilder(observer, operation, layer)

	// Act
	resultCtx, resultSpan := builder.WithTag("request-id", "123").Start(ctx)

	// Assert
	assert.NotNil(t, resultCtx)
	assert.NotNil(t, resultSpan)
	assert.Implements(t, (*Span)(nil), resultSpan)
	assert.Len(t, builder.options, 3) // layer + operation + custom tag
}

func TestSpanBuilder_WhenMultipleBuilders_ShouldBeIndependent(t *testing.T) {
	// Arrange
	observer := NewNoopObserver()

	// Act
	builder1 := NewSpanBuilder(observer, "operation1").WithLayer(LayerService)
	builder2 := NewSpanBuilder(observer, "operation2").WithLayer(LayerRepository)

	// Assert
	assert.NotEqual(t, builder1, builder2)
	assert.Equal(t, "operation1", builder1.operation)
	assert.Equal(t, "operation2", builder2.operation)
	assert.Len(t, builder1.options, 1)
	assert.Len(t, builder2.options, 1)
}

func TestSpanBuilder_WhenEmptyOperation_ShouldStillWork(t *testing.T) {
	// Arrange
	observer := NewNoopObserver()
	ctx := context.Background()
	builder := NewSpanBuilder(observer, "")

	// Act
	resultCtx, resultSpan := builder.Start(ctx)

	// Assert
	assert.NotNil(t, resultCtx)
	assert.NotNil(t, resultSpan)
	assert.Implements(t, (*Span)(nil), resultSpan)
}

func TestSpanBuilder_WhenNilObserver_ShouldPanicOnStart(t *testing.T) {
	// Arrange
	builder := NewSpanBuilder(nil, "test-operation")
	ctx := context.Background()

	// Act & Assert
	assert.Panics(t, func() {
		builder.Start(ctx)
	})
}

func TestEventSpanBuilder_WhenInheritedMethods_ShouldWorkCorrectly(t *testing.T) {
	// Arrange
	observer := NewNoopObserver()
	ctx := context.Background()
	operation := testCreateUserOperation
	layer := LayerService
	builder := NewEventSpanBuilder(observer, operation, layer)

	// Act - test that inherited methods work
	resultCtx, resultSpan := builder.
		WithService("user-service").
		WithComponent("user-repository").
		Start(ctx)

	// Assert
	assert.NotNil(t, resultCtx)
	assert.NotNil(t, resultSpan)
	assert.Implements(t, (*Span)(nil), resultSpan)
	assert.Len(t, builder.options, 4) // layer + operation + service + component
}

func TestSpanBuilder_WhenWithAllTagTypes_ShouldAccumulateCorrectly(t *testing.T) {
	// Arrange
	observer := NewNoopObserver()
	builder := NewSpanBuilder(observer, "test-operation")

	// Act
	result := builder.
		WithLayer("custom-layer").
		WithService("custom-service").
		WithComponent("custom-component").
		WithOperation("custom-operation").
		WithTag("key1", "value1").
		WithTag("key2", "value2")

	// Assert
	assert.Equal(t, builder, result)
	assert.Len(t, builder.options, 6) // 4 predefined + 2 custom tags
}
