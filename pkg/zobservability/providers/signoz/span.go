package signoz

import (
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// signozSpan implements the Span interface
type signozSpan struct {
	span      trace.Span
	operation string
}

func (s *signozSpan) SetTag(key, value string) {
	s.span.SetAttributes(attribute.String(key, value))
}

func (s *signozSpan) SetData(key string, value interface{}) {
	setSpanAttribute(s.span, key, value)
}

func (s *signozSpan) SetError(err error, opts ...trace.EventOption) {
	opts = append(opts, trace.WithStackTrace(true)) // We want to capture the stack trace always
	s.span.RecordError(err, opts...)
	s.span.SetStatus(codes.Error, err.Error())
}

func (s *signozSpan) Finish() {
	s.span.End()
}
