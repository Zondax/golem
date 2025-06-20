package zobservability

import (
	"context"

	"go.opentelemetry.io/otel/trace"
)

type noopObserver struct {
	metrics MetricsProvider
}

// NewNoopObserver creates a new no-op observer that does nothing
func NewNoopObserver() Observer {
	return &noopObserver{
		metrics: NewNoopMetricsProvider("noop"),
	}
}

func (n *noopObserver) StartTransaction(ctx context.Context, name string, opts ...TransactionOption) Transaction {
	return &noopTransaction{}
}

func (n *noopObserver) StartSpan(ctx context.Context, operation string, opts ...SpanOption) (context.Context, Span) {
	return ctx, &noopSpan{}
}

func (n *noopObserver) CaptureException(ctx context.Context, err error, opts ...EventOption) {
	// Do nothing
}

func (n *noopObserver) CaptureMessage(ctx context.Context, message string, level Level, opts ...EventOption) {
	// Do nothing
}

func (n *noopObserver) GetMetrics() MetricsProvider {
	return n.metrics
}

func (n *noopObserver) Close() error {
	return nil
}

func (n *noopObserver) GetConfig() Config {
	return Config{
		Provider:    "noop",
		Enabled:     false,
		Environment: "development",
		Debug:       false,
		SampleRate:  0,
		Middleware: MiddlewareConfig{
			CaptureErrors: false,
		},
		Metrics: MetricsConfig{
			Enabled:       false,
			Provider:      "noop",
			Path:          "/metrics",
			Port:          9090,
			OpenTelemetry: DefaultOpenTelemetryMetricsConfig(),
		},
	}
}

type noopTransaction struct{}

func (t *noopTransaction) Context() context.Context {
	return context.Background()
}

func (t *noopTransaction) SetName(name string) {}

func (t *noopTransaction) SetTag(key, value string) {}

func (t *noopTransaction) SetData(key string, value interface{}) {}

func (t *noopTransaction) StartChild(operation string, opts ...SpanOption) Span {
	return &noopSpan{}
}

func (t *noopTransaction) Finish(status TransactionStatus) {}

type noopSpan struct{}

func (s *noopSpan) SetTag(key, value string) {}

func (s *noopSpan) SetData(key string, value interface{}) {}

func (s *noopSpan) SetError(err error, opts ...trace.EventOption) {}

func (s *noopSpan) Finish() {}
