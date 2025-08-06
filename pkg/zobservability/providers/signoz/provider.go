package signoz

import (
	"context"
	"fmt"

	"go.opentelemetry.io/contrib/propagators/b3"
	"go.opentelemetry.io/contrib/propagators/jaeger"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/credentials"

	"github.com/zondax/golem/pkg/logger"
	"github.com/zondax/golem/pkg/zobservability"
)

// signozObserver implements the Observer interface using OpenTelemetry
type signozObserver struct {
	tracer         trace.Tracer
	config         *Config
	tracerProvider *sdktrace.TracerProvider
	metrics        zobservability.MetricsProvider
}

// NewObserver creates a new SigNoz observer using OpenTelemetry
func NewObserver(cfg *Config) (zobservability.Observer, error) {
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	// Initialize traces
	tracerProvider, tracer, err := createTracerProvider(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create tracer provider: %w", err)
	}

	// Initialize metrics provider
	metricsConfig := cfg.GetMetricsConfig()
	// Configure OpenTelemetry metrics with SigNoz settings
	metricsConfig.OpenTelemetry.Endpoint = cfg.Endpoint
	metricsConfig.OpenTelemetry.ServiceName = cfg.ServiceName
	metricsConfig.OpenTelemetry.ServiceVersion = cfg.Release
	metricsConfig.OpenTelemetry.Environment = cfg.Environment
	metricsConfig.OpenTelemetry.Hostname = cfg.GetHostname()
	metricsConfig.OpenTelemetry.Insecure = cfg.IsInsecure()

	// Pass SigNoz headers to metrics config (including access token if present)
	if cfg.HasHeaders() {
		if metricsConfig.OpenTelemetry.Headers == nil {
			metricsConfig.OpenTelemetry.Headers = make(map[string]string)
		}
		for key, value := range cfg.Headers {
			metricsConfig.OpenTelemetry.Headers[key] = value
		}
	}

	metricsProvider, err := zobservability.NewMetricsProvider(cfg.ServiceName, metricsConfig)
	if err != nil {
		if shutdownErr := tracerProvider.Shutdown(context.Background()); shutdownErr != nil {
			return nil, fmt.Errorf("failed to create metrics provider: %w (also failed to cleanup tracer provider: %v)", err, shutdownErr)
		}
		return nil, fmt.Errorf("failed to create metrics provider: %w", err)
	}

	// Start metrics provider
	if err := metricsProvider.Start(); err != nil {
		if shutdownErr := tracerProvider.Shutdown(context.Background()); shutdownErr != nil {
			return nil, fmt.Errorf("failed to start metrics provider: %w (also failed to cleanup tracer provider: %v)", err, shutdownErr)
		}
		return nil, fmt.Errorf("failed to start metrics provider: %w", err)
	}

	return &signozObserver{
		tracer:         tracer,
		config:         cfg,
		tracerProvider: tracerProvider,
		metrics:        metricsProvider,
	}, nil
}

// createTracerProvider creates and configures the OpenTelemetry tracer provider
func createTracerProvider(cfg *Config) (*sdktrace.TracerProvider, trace.Tracer, error) {
	secureOption := otlptracegrpc.WithTLSCredentials(credentials.NewClientTLSFromCert(nil, ""))
	if cfg.IsInsecure() {
		secureOption = otlptracegrpc.WithInsecure()
	}

	exporter, err := otlptrace.New(
		context.Background(),
		otlptracegrpc.NewClient(
			secureOption,
			otlptracegrpc.WithEndpoint(cfg.Endpoint),
			otlptracegrpc.WithHeaders(cfg.Headers),
		),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create OTLP exporter: %w", err)
	}

	// Create resource with service information
	// Resource defines WHO is generating the telemetry data - it's like a "business card" for the service
	// This metadata helps SigNoz identify and group traces by service, version, environment, etc.
	// These attributes appear in SigNoz UI and help with filtering and service mapping
	resources, err := createTracingResource(cfg)
	if err != nil {
		if shutdownErr := exporter.Shutdown(context.Background()); shutdownErr != nil {
			return nil, nil, fmt.Errorf("failed to create resource: %w (also failed to cleanup exporter: %v)", err, shutdownErr)
		}
		return nil, nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Configure sampler
	// Sampler controls HOW MUCH telemetry data to collect (performance vs completeness trade-off)
	// AlwaysSample = collect everything (good for dev, expensive for prod)
	// TraceIDRatioBased = collect only a percentage (cost-effective for high-traffic prod)
	sampler := sdktrace.AlwaysSample()
	sampleRate := cfg.GetSampleRate()
	if sampleRate > 0 && sampleRate < 1 {
		sampler = sdktrace.TraceIDRatioBased(sampleRate)
	}

	// CRITICAL FIX FOR GOOGLE CLOUD RUN:
	// Google Cloud Platform (Cloud Run, Cloud Functions, App Engine) automatically injects
	// trace headers (traceparent) with sampling decisions that can cause traces to be dropped.
	// When ShouldIgnoreParentSampling() returns true, we use our local sampling decision
	// instead of respecting the parent's sampling decision from GCP headers.
	//
	// This fixes the issue described in:
	// https://anecdotes.dev/opentelemetry-on-google-cloud-unraveling-the-mystery-f61f044c18be
	//
	// Without this fix, most traces in Cloud Run would be created as NonRecordingSpan
	// and never exported to SigNoz, making distributed tracing nearly useless.
	if !cfg.ShouldIgnoreParentSampling() {
		// Normal behavior: respect parent sampling decisions (for non-GCP environments)
		// This creates a ParentBased sampler that:
		// - Uses parent's sampling decision if present
		// - Falls back to our local sampler if no parent
		logger.GetLoggerFromContext(context.Background()).Infof("[GCP-SAMPLER] Using parent based sampler (ignore_parent_sampling=false)")
		sampler = sdktrace.ParentBased(sampler)
	} else {
		logger.GetLoggerFromContext(context.Background()).Infof("[GCP-SAMPLER] Using direct sampler (ignore_parent_sampling=true)")
	}
	// If ShouldIgnoreParentSampling() is true, we keep the direct sampler (AlwaysSample or TraceIDRatioBased)
	// This ensures our application makes its own sampling decisions regardless of GCP headers

	// Create tracer provider - this is the "engine" that creates and manages traces
	// TracerProvider is responsible for:
	// 1. Creating individual tracers for different components
	// 2. Applying sampling decisions (what to collect)
	// 3. Batching and sending data to SigNoz via the exporter
	// 4. Managing trace lifecycle and resource cleanup
	var tracerProviderOpts []sdktrace.TracerProviderOption

	// Add sampler and resource options
	tracerProviderOpts = append(tracerProviderOpts,
		sdktrace.WithSampler(sampler),
		sdktrace.WithResource(resources),
	)

	// Choose between SimpleSpan (immediate export) or Batch processor
	var baseProcessor sdktrace.SpanProcessor
	if cfg.UseSimpleSpan {
		// Use OpenTelemetry's native SimpleSpanProcessor for immediate export without batching
		// This processor exports spans immediately when they finish, providing real-time visibility
		// at the cost of increased network overhead (one request per span)
		baseProcessor = sdktrace.NewSimpleSpanProcessor(exporter)
	} else {
		// Get batch configuration for performance tuning
		batchConfig := cfg.GetBatchConfig()

		// Batch processor - groups spans before sending (more efficient than one-by-one)
		// Configurable batching improves performance and reduces network overhead
		baseProcessor = sdktrace.NewBatchSpanProcessor(
			exporter,
			// How often to send batches (lower = more real-time, higher = more efficient)
			sdktrace.WithBatchTimeout(batchConfig.BatchTimeout),
			// Timeout for individual export operations
			sdktrace.WithExportTimeout(batchConfig.ExportTimeout),
			// Maximum spans per batch (higher = more efficient, but more memory)
			sdktrace.WithMaxExportBatchSize(batchConfig.MaxExportBatch),
			// Maximum spans in queue (higher = less data loss, but more memory)
			sdktrace.WithMaxQueueSize(batchConfig.MaxQueueSize),
		)
	}

	// Use the base processor directly
	finalProcessor := baseProcessor

	// Debug configuration values
	logger.GetLoggerFromContext(context.Background()).Debugf("DEBUG: SigNoz Config - IgnoreParentSampling: %v, SampleRate: %f, UseSimpleSpan: %v",
		cfg.IgnoreParentSampling, cfg.SampleRate, cfg.UseSimpleSpan)

	tracerProviderOpts = append(tracerProviderOpts, sdktrace.WithSpanProcessor(finalProcessor))

	tracerProvider := sdktrace.NewTracerProvider(tracerProviderOpts...)

	// Set global tracer provider
	otel.SetTracerProvider(tracerProvider)

	// Configure propagators based on configuration
	propagator := createPropagator(cfg)
	otel.SetTextMapPropagator(propagator)

	// Log which propagator is being used
	logger.GetLoggerFromContext(context.Background()).Infof("OpenTelemetry propagator configured: type=%T, formats=%v",
		propagator, cfg.GetPropagationConfig().Formats)

	// Create tracer
	tracer := otel.Tracer(cfg.ServiceName)

	return tracerProvider, tracer, nil
}

// createPropagator creates a composite propagator based on the configuration
func createPropagator(cfg *Config) propagation.TextMapPropagator {
	propagationConfig := cfg.GetPropagationConfig()
	formats := propagationConfig.Formats

	// Default to W3C
	if len(formats) == 0 {
		return createW3CPropagator()
	}

	var propagators []propagation.TextMapPropagator
	for _, format := range formats {
		if prop := createPropagatorByFormat(format); prop != nil {
			propagators = append(propagators, prop...)
		}
	}

	// Fallback to W3C if no valid propagators were created
	if len(propagators) == 0 {
		return createW3CPropagator()
	}

	return propagation.NewCompositeTextMapPropagator(propagators...)
}

// createW3CPropagator creates the default W3C propagator
func createW3CPropagator() propagation.TextMapPropagator {
	return propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
}

// createPropagatorByFormat creates propagators for a specific format
func createPropagatorByFormat(format string) []propagation.TextMapPropagator {
	switch format {
	case zobservability.PropagationW3C:
		return []propagation.TextMapPropagator{
			propagation.TraceContext{},
			propagation.Baggage{},
		}
	case zobservability.PropagationB3:
		return []propagation.TextMapPropagator{b3.New()}
	case zobservability.PropagationB3Single:
		return []propagation.TextMapPropagator{
			b3.New(b3.WithInjectEncoding(b3.B3SingleHeader)),
		}
	case zobservability.PropagationJaeger:
		return []propagation.TextMapPropagator{jaeger.Jaeger{}}
	default:
		return nil
	}
}

// createTracingResource creates a resource with service information for tracing
func createTracingResource(cfg *Config) (*resource.Resource, error) {
	resourceConfig := cfg.GetResourceConfig()
	resourceAttributes := []attribute.KeyValue{
		attribute.String(zobservability.ResourceServiceName, cfg.ServiceName),
		attribute.String(zobservability.ResourceServiceVersion, cfg.Release),
		attribute.String(zobservability.ResourceEnvironment, cfg.Environment),
		attribute.String(zobservability.ResourceLanguage, zobservability.ResourceLanguageGo),
		attribute.String(zobservability.ResourceHostName, cfg.GetHostname()),
	}

	// Add optional process ID if configured (useful for debugging)
	if pid := cfg.GetProcessID(); pid != "" {
		resourceAttributes = append(resourceAttributes, attribute.String(zobservability.ResourceProcessPID, pid))
	}

	// Add custom attributes from configuration
	for key, value := range resourceConfig.CustomAttributes {
		resourceAttributes = append(resourceAttributes, attribute.String(key, value))
	}

	return resource.New(
		context.Background(),
		resource.WithAttributes(resourceAttributes...),
	)
}

func (s *signozObserver) StartTransaction(ctx context.Context, name string, opts ...zobservability.TransactionOption) zobservability.Transaction {
	ctx, span := s.tracer.Start(ctx, name)

	signozTx := &signozTransaction{
		ctx:  ctx,
		span: span,
		name: name,
	}

	for _, opt := range opts {
		opt.ApplyTransaction(signozTx)
	}

	return signozTx
}

func (s *signozObserver) StartSpan(ctx context.Context, operation string, opts ...zobservability.SpanOption) (context.Context, zobservability.Span) {
	ctx, span := s.tracer.Start(ctx, operation)

	signozSpan := &signozSpan{
		span:      span,
		operation: operation,
	}

	for _, opt := range opts {
		opt.ApplySpan(signozSpan)
	}

	return ctx, signozSpan
}

// CaptureException captures ERROR events with full error context
// Use this for: Exceptions, failures, critical errors that need investigation
// What it does:
// - Records the error with full stack trace information
// - Sets span status to ERROR (affects service health metrics)
// - Marks the entire trace as having an error
// - Provides structured error data for debugging
// - Appears in SigNoz as "Error" events with red indicators
func (s *signozObserver) CaptureException(ctx context.Context, err error, opts ...zobservability.EventOption) {
	span := trace.SpanFromContext(ctx)
	if span != nil {
		span.RecordError(err, trace.WithStackTrace(true))
		span.SetStatus(codes.Error, err.Error())

		// Apply additional event options
		event := &signozEvent{
			span: span,
		}
		for _, opt := range opts {
			opt.ApplyEvent(event)
		}
		event.Capture()
	}
}

// CaptureMessage captures INFORMATIONAL events with custom severity levels
// Use this for: Logs, debug info, warnings, business events, audit trails
// What it does:
// - Records a message with specified level (Debug, Info, Warning, etc.)
// - Does NOT affect span status (span remains successful)
// - Does NOT mark trace as failed
// - Provides contextual information for debugging
// - Appears in SigNoz as "Event" logs with level-based colors
func (s *signozObserver) CaptureMessage(ctx context.Context, message string, level zobservability.Level, opts ...zobservability.EventOption) {
	span := trace.SpanFromContext(ctx)
	if span != nil {
		span.AddEvent(message, trace.WithAttributes(
			attribute.String(zobservability.SpanAttributeLevel, level.String()),
		))

		// Apply additional event options
		event := &signozEvent{
			span: span,
		}
		for _, opt := range opts {
			opt.ApplyEvent(event)
		}
		event.Capture()
	}
}

// GetMetrics returns the metrics provider
func (s *signozObserver) GetMetrics() zobservability.MetricsProvider {
	return s.metrics
}

// ForceFlush forces the immediate export of all spans that have not yet been exported.
// This is particularly important for Cloud Run environments where containers can be
// terminated without warning, potentially causing spans to be lost if they remain buffered.
//
// ForceFlush ensures that all manually created spans (via StartSpan or NewEventSpanBuilder)
// are immediately sent to SigNoz, complementing the automatic gRPC interceptor spans.
func (s *signozObserver) ForceFlush(ctx context.Context) error {
	if s.tracerProvider == nil {
		return nil // Nothing to flush if tracer provider is not initialized
	}

	// Create timeout context for the flush operation
	// Use a generous timeout to ensure spans have time to export, especially important for:
	// - Slow network connections to SigNoz
	// - Large batches of spans waiting to be exported
	// - Cloud Run cold starts where export might take longer
	flushCtx, cancel := context.WithTimeout(ctx, DefaultForceFlushTimeout)
	defer cancel()

	// ForceFlush immediately exports all spans that have not yet been exported
	// for all the registered span processors. This is critical for ensuring
	// that manually instrumented business logic spans are not lost when containers terminate.
	if err := s.tracerProvider.ForceFlush(flushCtx); err != nil {
		return fmt.Errorf("failed to force flush spans: %w", err)
	}

	return nil
}

// Close shuts down the observer and all its providers
// CRITICAL: ForceFlush is called before Shutdown to ensure all pending spans are exported.
// This prevents data loss in Cloud Run environments where containers can be terminated abruptly.
func (s *signozObserver) Close() error {
	// Create a shutdown context with sufficient timeout for both flush and shutdown operations
	ctx, cancel := context.WithTimeout(context.Background(), DefaultShutdownTimeout+DefaultForceFlushTimeout)
	defer cancel()

	// 1. Force flush all pending spans before shutting down
	// This ensures manually instrumented spans (StartSpan, NewEventSpanBuilder) are exported
	// before the tracer provider shuts down and stops accepting new export operations
	if err := s.ForceFlush(ctx); err != nil {
		// Log the error but continue with shutdown - we don't want to block cleanup
		// if ForceFlush fails, but we still need to properly close resources
		// Note: This could be improved with better logging once logger is available in context
		logger.GetLoggerFromContext(ctx).Errorf("Warning: failed to force flush spans during close: %v\n", err)
	}

	// 2. Close metrics provider
	if s.metrics != nil {
		if err := s.metrics.Stop(); err != nil {
			return fmt.Errorf("failed to stop metrics provider: %w", err)
		}
	}

	// 3. Close tracer provider (includes final flush as per OpenTelemetry spec)
	// Shutdown() automatically includes the effects of ForceFlush(), but we've already
	// called it explicitly above with better error handling and timeout management
	if s.tracerProvider != nil {
		shutdownCtx, shutdownCancel := context.WithTimeout(ctx, DefaultShutdownTimeout)
		defer shutdownCancel()
		if err := s.tracerProvider.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("failed to shutdown tracer provider: %w", err)
		}
	}

	return nil
}

func (s *signozObserver) GetConfig() zobservability.Config {
	return zobservability.Config{
		Provider:    zobservability.ProviderSigNoz,
		Enabled:     true,
		Environment: s.config.Environment,
		Release:     s.config.Release,
		Debug:       s.config.Debug,
		Address:     s.config.Endpoint,
		SampleRate:  s.config.SampleRate,
		Middleware: zobservability.MiddlewareConfig{
			CaptureErrors: true,
		},
		TracingExclusions: s.config.TracingExclusions,
	}
}
