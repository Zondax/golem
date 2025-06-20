package signoz

import (
	"strings"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/zondax/golem/pkg/zobservability"
)

// signozEvent implements event handling for SigNoz
type signozEvent struct {
	span trace.Span
}

func (e *signozEvent) SetLevel(level zobservability.Level) {
	e.span.SetAttributes(attribute.String(zobservability.SpanAttributeLevel, level.String()))
}

func (e *signozEvent) SetTags(tags map[string]string) {
	for key, value := range tags {
		e.span.SetAttributes(attribute.String(key, value))
	}
}

func (e *signozEvent) SetTag(key, value string) {
	e.span.SetAttributes(attribute.String(key, value))
}

func (e *signozEvent) SetUser(id, email, username string) {
	e.span.SetAttributes(
		attribute.String(zobservability.UserAttributeID, id),
		attribute.String(zobservability.UserAttributeEmail, email),
		attribute.String(zobservability.UserAttributeUsername, username),
	)
}

func (e *signozEvent) SetFingerprint(fingerprint []string) {
	fingerprintStr := strings.Join(fingerprint, zobservability.FingerprintSeparator)
	e.span.SetAttributes(attribute.String(zobservability.FingerprintAttribute, fingerprintStr))
}

func (e *signozEvent) SetError(err error) {
	if err != nil {
		e.span.RecordError(err, trace.WithStackTrace(true))
		e.span.SetStatus(codes.Error, err.Error())
	}
}

func (e *signozEvent) Capture() {
	// In OpenTelemetry, events are captured automatically when added to spans
	// This method is here for interface compliance but doesn't need to do anything
	// as the span events are already recorded when SetError, SetTag, etc. are called
}

func (e *signozEvent) SetData(key string, value interface{}) {
	setSpanAttribute(e.span, key, value)
}
