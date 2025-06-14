package zobservability

import "go.opentelemetry.io/otel/trace"

// Span represents a unit of work or operation within a transaction
type Span interface {
	SetTag(key, value string)
	SetData(key string, value interface{})
	SetError(err error, opts ...trace.EventOption)
	Finish()
}

// SpanOption configures a span
type SpanOption interface {
	ApplySpan(Span)
}

type spanOptionFunc func(Span)

func (f spanOptionFunc) ApplySpan(s Span) {
	f(s)
}

// WithSpanTag adds a tag to the span
func WithSpanTag(key, value string) SpanOption {
	return spanOptionFunc(func(s Span) {
		s.SetTag(key, value)
	})
}

// WithSpanData adds data to the span
func WithSpanData(key string, value interface{}) SpanOption {
	return spanOptionFunc(func(s Span) {
		s.SetData(key, value)
	})
}

// WithSpanError marks the span as failed with an error
func WithSpanError(err error) SpanOption {
	return spanOptionFunc(func(s Span) {
		s.SetError(err)
	})
}
