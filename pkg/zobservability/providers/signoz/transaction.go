package signoz

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/zondax/golem/pkg/zobservability"
)

// signozTransaction implements the Transaction interface
type signozTransaction struct {
	ctx  context.Context
	span trace.Span
	name string
}

func (t *signozTransaction) Context() context.Context {
	return t.ctx
}

func (t *signozTransaction) SetName(name string) {
	t.span.SetName(name)
	t.name = name
}

func (t *signozTransaction) SetTag(key, value string) {
	t.span.SetAttributes(attribute.String(key, value))
}

func (t *signozTransaction) SetData(key string, value interface{}) {
	setSpanAttribute(t.span, key, value)
}

func (t *signozTransaction) StartChild(operation string, opts ...zobservability.SpanOption) zobservability.Span {
	_, span := t.span.TracerProvider().Tracer(TracerName).Start(t.ctx, operation)

	signozSpan := &signozSpan{
		span:      span,
		operation: operation,
	}

	for _, opt := range opts {
		opt.ApplySpan(signozSpan)
	}

	return signozSpan
}

func (t *signozTransaction) Finish(status zobservability.TransactionStatus) {
	switch status {
	case zobservability.TransactionOK:
		t.span.SetStatus(codes.Ok, zobservability.TransactionSuccessMessage)
	case zobservability.TransactionError:
		t.span.SetStatus(codes.Error, zobservability.TransactionFailureMessage)
	case zobservability.TransactionCancelled:
		t.span.SetStatus(codes.Error, zobservability.TransactionCancelledMessage)
	default:
		t.span.SetStatus(codes.Unset, zobservability.TransactionSuccessMessage)
	}
	t.span.End()
}
