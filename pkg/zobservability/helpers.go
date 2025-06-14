package zobservability

import "context"

// SpanBuilder helps build spans with common patterns
type SpanBuilder struct {
	observer  Observer
	operation string
	options   []SpanOption
}

// NewSpanBuilder creates a new span builder
func NewSpanBuilder(observer Observer, operation string) *SpanBuilder {
	return &SpanBuilder{
		observer:  observer,
		operation: operation,
		options:   make([]SpanOption, 0),
	}
}

// WithLayer adds a layer tag to the span
func (sb *SpanBuilder) WithLayer(layer string) *SpanBuilder {
	sb.options = append(sb.options, WithSpanTag(TagLayer, layer))
	return sb
}

// WithService adds a service tag to the span
func (sb *SpanBuilder) WithService(service string) *SpanBuilder {
	sb.options = append(sb.options, WithSpanTag(TagService, service))
	return sb
}

// WithComponent adds a component tag to the span
func (sb *SpanBuilder) WithComponent(component string) *SpanBuilder {
	sb.options = append(sb.options, WithSpanTag(TagComponent, component))
	return sb
}

// WithOperation adds an operation tag to the span
func (sb *SpanBuilder) WithOperation(operation string) *SpanBuilder {
	sb.options = append(sb.options, WithSpanTag(TagOperation, operation))
	return sb
}

// WithTag adds a custom tag to the span
func (sb *SpanBuilder) WithTag(key, value string) *SpanBuilder {
	sb.options = append(sb.options, WithSpanTag(key, value))
	return sb
}

// Start creates and starts the span with all configured options
func (sb *SpanBuilder) Start(ctx context.Context) (context.Context, Span) {
	return sb.observer.StartSpan(ctx, sb.operation, sb.options...)
}

// Convenience functions for common layer patterns

// StartServiceSpan creates a span for service layer with common tags
func StartServiceSpan(ctx context.Context, observer Observer, operation, service string) (context.Context, Span) {
	return NewSpanBuilder(observer, operation).
		WithLayer(LayerService).
		WithService(service).
		WithOperation(operation).
		Start(ctx)
}

// EventSpanBuilder creates spans with layer prefix in operation name for better visibility
type EventSpanBuilder struct {
	*SpanBuilder
	layer string
}

// NewEventSpanBuilder creates a new event span builder with layer prefix in operation name
func NewEventSpanBuilder(observer Observer, operation, layer string) *EventSpanBuilder {
	operationWithLayer := layer + "." + operation
	return &EventSpanBuilder{
		SpanBuilder: NewSpanBuilder(observer, operationWithLayer).
			WithLayer(layer).
			WithOperation(operation),
		layer: layer,
	}
}
