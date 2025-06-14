package signoz

import (
	"fmt"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// setSpanAttribute is a helper function to set span attributes with proper type handling
func setSpanAttribute(span trace.Span, key string, value interface{}) {
	switch v := value.(type) {
	case string:
		span.SetAttributes(attribute.String(key, v))
	case int:
		span.SetAttributes(attribute.Int(key, v))
	case int64:
		span.SetAttributes(attribute.Int64(key, v))
	case float64:
		span.SetAttributes(attribute.Float64(key, v))
	case bool:
		span.SetAttributes(attribute.Bool(key, v))
	default:
		span.SetAttributes(attribute.String(key, fmt.Sprintf("%v", v)))
	}
}
