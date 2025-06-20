package sentry

import (
	"context"

	"github.com/getsentry/sentry-go"
	"github.com/zondax/golem/pkg/zobservability"
	"go.opentelemetry.io/otel/trace"
)

type sentryTransaction struct {
	hub *sentry.Hub
	tx  *sentry.Span
	ctx context.Context
}

func (t *sentryTransaction) Context() context.Context {
	return t.ctx
}

func (t *sentryTransaction) SetName(name string) {
	t.tx.Name = name
}

func (t *sentryTransaction) SetTag(key, value string) {
	t.tx.SetTag(key, value)
}

func (t *sentryTransaction) SetData(key string, value interface{}) {
	t.tx.SetData(key, value)
}

func (t *sentryTransaction) StartChild(operation string, opts ...zobservability.SpanOption) zobservability.Span {
	childSpan := t.tx.StartChild(operation)
	span := &sentrySpan{
		hub:       t.hub,
		operation: operation,
		span:      childSpan,
	}

	for _, opt := range opts {
		opt.ApplySpan(span)
	}

	return span
}

func (t *sentryTransaction) Finish(status zobservability.TransactionStatus) {
	switch status {
	case zobservability.TransactionOK:
		t.tx.Status = sentry.SpanStatusOK
	case zobservability.TransactionError:
		t.tx.Status = sentry.SpanStatusInternalError
	case zobservability.TransactionCancelled:
		t.tx.Status = sentry.SpanStatusCanceled
	default:
		t.tx.Status = sentry.SpanStatusUnknown
	}
	t.tx.Finish()
}

type sentrySpan struct {
	hub       *sentry.Hub
	operation string
	span      *sentry.Span
}

func (s *sentrySpan) SetTag(key, value string) {
	s.span.SetTag(key, value)
}

func (s *sentrySpan) SetData(key string, value interface{}) {
	s.span.SetData(key, value)
}

func (s *sentrySpan) SetError(err error, opts ...trace.EventOption) {
	if err != nil {
		s.span.Status = sentry.SpanStatusInternalError
		s.span.SetTag("error", err.Error())
	}
}

func (s *sentrySpan) Finish() {
	s.span.Finish()
}

type sentryEvent struct {
	hub   *sentry.Hub
	event *sentry.Event
}

func (e *sentryEvent) SetLevel(level zobservability.Level) {
	e.event.Level = convertLevel(level)
}

func (e *sentryEvent) SetTags(tags map[string]string) {
	if e.event.Tags == nil {
		e.event.Tags = make(map[string]string)
	}
	for k, v := range tags {
		e.event.Tags[k] = v
	}
}

func (e *sentryEvent) SetTag(key, value string) {
	if e.event.Tags == nil {
		e.event.Tags = make(map[string]string)
	}
	e.event.Tags[key] = value
}

func (e *sentryEvent) SetUser(id, email, username string) {
	e.event.User = sentry.User{
		ID:       id,
		Email:    email,
		Username: username,
	}
}

func (e *sentryEvent) SetFingerprint(fingerprint []string) {
	e.event.Fingerprint = fingerprint
}

func (e *sentryEvent) SetError(err error) {
	if err != nil {
		e.event.Exception = []sentry.Exception{{
			Value:      err.Error(),
			Type:       "error",
			Stacktrace: sentry.NewStacktrace(),
		}}
	}
}

func (e *sentryEvent) Capture() {
	e.hub.CaptureEvent(e.event)
}
