package zobservability

import (
	"context"
)

// Observer defines the main interface for observability operations
type Observer interface {
	// Tracing operations
	StartTransaction(ctx context.Context, name string, opts ...TransactionOption) Transaction
	StartSpan(ctx context.Context, operation string, opts ...SpanOption) (context.Context, Span)
	CaptureException(ctx context.Context, err error, opts ...EventOption)
	CaptureMessage(ctx context.Context, message string, level Level, opts ...EventOption)

	// Metrics operations
	GetMetrics() MetricsProvider

	// Configuration and lifecycle
	GetConfig() Config
	Close() error
}

// Level represents the severity level of an event
type Level int

const (
	// LevelDebug represents debug level events
	LevelDebug Level = iota
	// LevelInfo represents informational events
	LevelInfo
	// LevelWarning represents warning events
	LevelWarning
	// LevelError represents error events
	LevelError
	// LevelFatal represents fatal events
	LevelFatal
)

// String implements the Stringer interface
func (l Level) String() string {
	switch l {
	case LevelDebug:
		return "debug"
	case LevelInfo:
		return "info"
	case LevelWarning:
		return "warning"
	case LevelError:
		return "error"
	case LevelFatal:
		return "fatal"
	default:
		return "unknown"
	}
}
