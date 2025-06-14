package signoz

import (
	"context"
	"fmt"

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

	// Get batch configuration for performance tuning
	batchConfig := cfg.GetBatchConfig()

	// Create tracer provider - this is the "engine" that creates and manages traces
	// TracerProvider is responsible for:
	// 1. Creating individual tracers for different components
	// 2. Applying sampling decisions (what to collect)
	// 3. Batching and sending data to SigNoz via the exporter
	// 4. Managing trace lifecycle and resource cleanup
	tracerProvider := sdktrace.NewTracerProvider(
		// Sampling strategy - controls data volume and costs
		sdktrace.WithSampler(sampler),
		// Batch processor - groups spans before sending (more efficient than one-by-one)
		// Configurable batching improves performance and reduces network overhead
		sdktrace.WithBatcher(
			exporter,
			// How often to send batches (lower = more real-time, higher = more efficient)
			sdktrace.WithBatchTimeout(batchConfig.BatchTimeout),
			// Timeout for individual export operations
			sdktrace.WithExportTimeout(batchConfig.ExportTimeout),
			// Maximum spans per batch (higher = more efficient, but more memory)
			sdktrace.WithMaxExportBatchSize(batchConfig.MaxExportBatch),
			// Maximum spans in queue (higher = less data loss, but more memory)
			sdktrace.WithMaxQueueSize(batchConfig.MaxQueueSize),
		),
		// Resource metadata - attaches service info to all traces
		sdktrace.WithResource(resources),
	)

	// Set global tracer provider
	otel.SetTracerProvider(tracerProvider)

	// Configure text map propagator for distributed tracing
	// This is CRITICAL for distributed tracing to work across services
	// It tells OpenTelemetry how to inject/extract trace context in HTTP/gRPC headers
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{}, // W3C Trace Context (standard)
		propagation.Baggage{},      // W3C Baggage (for custom attributes)
	))

	// Create tracer
	tracer := otel.Tracer(cfg.ServiceName)

	return tracerProvider, tracer, nil
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

// Close shuts down the observer and all its providers
func (s *signozObserver) Close() error {
	// Close metrics provider
	if s.metrics != nil {
		if err := s.metrics.Stop(); err != nil {
			return fmt.Errorf("failed to stop metrics provider: %w", err)
		}
	}

	// Close tracer provider
	if s.tracerProvider != nil {
		ctx, cancel := context.WithTimeout(context.Background(), DefaultShutdownTimeout)
		defer cancel()
		if err := s.tracerProvider.Shutdown(ctx); err != nil {
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
	}
}
