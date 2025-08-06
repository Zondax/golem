package sentry

import (
	"context"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/zondax/golem/pkg/zobservability"
)

// Config holds the configuration for the Sentry observer
type Config struct {
	DSN           string
	Environment   string
	Release       string
	Debug         bool
	ServiceName   string
	SampleRate    float64
	CaptureErrors bool
}

type sentryObserver struct {
	client *sentry.Client
	hub    *sentry.Hub
	config *Config
}

// NewObserver creates a new Sentry observer
func NewObserver(cfg *Config) (zobservability.Observer, error) {
	client, err := sentry.NewClient(sentry.ClientOptions{
		Dsn:              cfg.DSN,
		Environment:      cfg.Environment,
		Release:          cfg.Release,
		Debug:            cfg.Debug,
		AttachStacktrace: true,
		SampleRate:       cfg.SampleRate,
	})
	if err != nil {
		return nil, err
	}

	hub := sentry.NewHub(client, sentry.NewScope())
	return &sentryObserver{
		client: client,
		hub:    hub,
		config: cfg,
	}, nil
}

func (s *sentryObserver) StartTransaction(ctx context.Context, name string, opts ...zobservability.TransactionOption) zobservability.Transaction {
	tx := sentry.StartTransaction(ctx, name)
	sentryTx := &sentryTransaction{
		hub: s.hub,
		tx:  tx,
		ctx: ctx,
	}

	for _, opt := range opts {
		opt.ApplyTransaction(sentryTx)
	}

	return sentryTx
}

func (s *sentryObserver) StartSpan(ctx context.Context, operation string, opts ...zobservability.SpanOption) (context.Context, zobservability.Span) {
	parentSpan := sentry.SpanFromContext(ctx)
	var span *sentry.Span
	if parentSpan != nil {
		span = parentSpan.StartChild(operation)
	} else {
		span = sentry.StartSpan(ctx, operation)
	}

	sentrySpan := &sentrySpan{
		hub:       s.hub,
		operation: operation,
		span:      span,
	}

	for _, opt := range opts {
		opt.ApplySpan(sentrySpan)
	}

	ctx = sentry.SetHubOnContext(ctx, s.hub)
	return ctx, sentrySpan
}

func (s *sentryObserver) CaptureException(ctx context.Context, err error, opts ...zobservability.EventOption) {
	event := &sentryEvent{
		hub:   s.hub,
		event: sentry.NewEvent(),
	}

	event.SetError(err)
	for _, opt := range opts {
		opt.ApplyEvent(event)
	}

	event.Capture()
}

func (s *sentryObserver) CaptureMessage(ctx context.Context, message string, level zobservability.Level, opts ...zobservability.EventOption) {
	event := &sentryEvent{
		hub:   s.hub,
		event: sentry.NewEvent(),
	}

	event.event.Message = message
	event.event.Level = convertLevel(level)

	for _, opt := range opts {
		opt.ApplyEvent(event)
	}

	event.Capture()
}

// GetMetrics returns a no-op metrics provider since Sentry doesn't support metrics
func (s *sentryObserver) GetMetrics() zobservability.MetricsProvider {
	return zobservability.NewNoopMetricsProvider("sentry")
}

func (s *sentryObserver) ForceFlush(ctx context.Context) error {
	if s.client != nil {
		s.client.Flush(2 * time.Second)
	}
	return nil
}

func (s *sentryObserver) Close() error {
	if s.client != nil {
		s.client.Flush(2 * time.Second)
	}
	return nil
}

func (s *sentryObserver) GetConfig() zobservability.Config {
	return zobservability.Config{
		Provider:    zobservability.ProviderSentry,
		Enabled:     true,
		Environment: s.config.Environment,
		Release:     s.config.Release,
		Debug:       s.config.Debug,
		Address:     s.config.DSN,
		SampleRate:  s.config.SampleRate,
		Middleware: zobservability.MiddlewareConfig{
			CaptureErrors: s.config.CaptureErrors,
		},
		InterceptorTracingExcludeMethods: []string{},
	}
}

func convertLevel(level zobservability.Level) sentry.Level {
	switch level {
	case zobservability.LevelDebug:
		return sentry.LevelDebug
	case zobservability.LevelInfo:
		return sentry.LevelInfo
	case zobservability.LevelWarning:
		return sentry.LevelWarning
	case zobservability.LevelError:
		return sentry.LevelError
	case zobservability.LevelFatal:
		return sentry.LevelFatal
	default:
		return sentry.LevelError
	}
}
