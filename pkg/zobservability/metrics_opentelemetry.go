package zobservability

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
)

// =============================================================================
// OPENTELEMETRY METRICS PROVIDER IMPLEMENTATION
// =============================================================================

// openTelemetryMetricsProvider implements MetricsProvider using OpenTelemetry
type openTelemetryMetricsProvider struct {
	name          string
	config        MetricsConfig
	meterProvider *sdkmetric.MeterProvider
	meter         metric.Meter

	// Metric storage
	counters   map[string]metric.Int64Counter
	gauges     map[string]metric.Float64Gauge
	histograms map[string]metric.Float64Histogram
	mutex      sync.RWMutex
}

// NewOpenTelemetryMetricsProvider creates a new OpenTelemetry metrics provider
func NewOpenTelemetryMetricsProvider(name string, config MetricsConfig) (MetricsProvider, error) {
	if err := config.OpenTelemetry.Validate(); err != nil {
		return nil, fmt.Errorf("invalid OpenTelemetry configuration: %w", err)
	}

	exporter, err := createOTLPExporter(config.OpenTelemetry)
	if err != nil {
		return nil, fmt.Errorf("failed to create OTLP exporter: %w", err)
	}

	resource, err := createResource(config.OpenTelemetry)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	meterProvider, err := createMeterProvider(config.OpenTelemetry, exporter, resource)
	if err != nil {
		return nil, fmt.Errorf("failed to create meter provider: %w", err)
	}

	// Set global meter provider
	otel.SetMeterProvider(meterProvider)

	// Create meter
	meter := meterProvider.Meter(config.OpenTelemetry.ServiceName)

	return &openTelemetryMetricsProvider{
		name:          name,
		config:        config,
		meterProvider: meterProvider,
		meter:         meter,
		counters:      make(map[string]metric.Int64Counter),
		gauges:        make(map[string]metric.Float64Gauge),
		histograms:    make(map[string]metric.Float64Histogram),
	}, nil
}

// =============================================================================
// HELPER FUNCTIONS FOR SETUP
// =============================================================================

// createOTLPExporter creates an OTLP metrics exporter with the given configuration
func createOTLPExporter(config OpenTelemetryMetricsConfig) (sdkmetric.Exporter, error) {
	opts := []otlpmetricgrpc.Option{
		otlpmetricgrpc.WithEndpoint(config.Endpoint),
	}

	// Configure security
	if config.Insecure {
		opts = append(opts, otlpmetricgrpc.WithInsecure())
	}

	// Add custom headers
	if len(config.Headers) > 0 {
		opts = append(opts, otlpmetricgrpc.WithHeaders(config.Headers))
	}

	return otlpmetricgrpc.New(context.Background(), opts...)
}

// createResource creates an OpenTelemetry resource with service metadata
func createResource(config OpenTelemetryMetricsConfig) (*resource.Resource, error) {
	return resource.New(
		context.Background(),
		resource.WithAttributes(
			attribute.String(ResourceServiceName, config.ServiceName),
			attribute.String(ResourceServiceVersion, config.ServiceVersion),
			attribute.String(ResourceEnvironment, config.Environment),
			attribute.String(ResourceHostName, config.Hostname),
			attribute.String(ResourceLanguage, ResourceLanguageGo),
		),
	)
}

// createMeterProvider creates a meter provider with the appropriate reader based on export mode
func createMeterProvider(config OpenTelemetryMetricsConfig, exporter sdkmetric.Exporter, res *resource.Resource) (*sdkmetric.MeterProvider, error) {
	reader, err := createMetricReader(config, exporter)
	if err != nil {
		return nil, err
	}

	return sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(reader),
		sdkmetric.WithResource(res),
	), nil
}

// createMetricReader creates the appropriate metric reader based on export mode
func createMetricReader(config OpenTelemetryMetricsConfig, exporter sdkmetric.Exporter) (sdkmetric.Reader, error) {
	switch config.ExportMode {
	case OTelExportModePush:
		return sdkmetric.NewPeriodicReader(
			exporter,
			sdkmetric.WithInterval(config.PushInterval),
			sdkmetric.WithTimeout(config.ExportTimeout),
		), nil
	case OTelExportModeEndpoint:
		return sdkmetric.NewManualReader(), nil
	default:
		return nil, fmt.Errorf("unsupported export mode: %s", config.ExportMode)
	}
}

// labelsToAttributes converts a map of labels to OpenTelemetry attributes
func labelsToAttributes(labels map[string]string) []attribute.KeyValue {
	attrs := make([]attribute.KeyValue, 0, len(labels))
	for k, v := range labels {
		attrs = append(attrs, attribute.String(k, v))
	}
	return attrs
}

// =============================================================================
// METRICS OPERATIONS IMPLEMENTATION
// =============================================================================

// Counter operations
func (o *openTelemetryMetricsProvider) IncrementCounter(ctx context.Context, name string, labels map[string]string) error {
	if ctx == nil {
		return fmt.Errorf("context cannot be nil")
	}
	if name == "" {
		return fmt.Errorf("metric name cannot be empty")
	}
	return o.AddCounter(ctx, name, 1, labels)
}

func (o *openTelemetryMetricsProvider) AddCounter(ctx context.Context, name string, value float64, labels map[string]string) error {
	if ctx == nil {
		return fmt.Errorf("context cannot be nil")
	}
	if name == "" {
		return fmt.Errorf("metric name cannot be empty")
	}
	if value < 0 {
		return fmt.Errorf("counter value must be non-negative, got %f", value)
	}

	counter, err := o.getCounter(name)
	if err != nil {
		return err
	}

	attrs := labelsToAttributes(labels)
	counter.Add(ctx, int64(value), metric.WithAttributes(attrs...))
	return nil
}

// Gauge operations
func (o *openTelemetryMetricsProvider) SetGauge(ctx context.Context, name string, value float64, labels map[string]string) error {
	if ctx == nil {
		return fmt.Errorf("context cannot be nil")
	}
	if name == "" {
		return fmt.Errorf("metric name cannot be empty")
	}

	gauge, err := o.getGauge(name)
	if err != nil {
		return err
	}

	attrs := labelsToAttributes(labels)
	gauge.Record(ctx, value, metric.WithAttributes(attrs...))
	return nil
}

func (o *openTelemetryMetricsProvider) AddGauge(ctx context.Context, name string, value float64, labels map[string]string) error {
	if ctx == nil {
		return fmt.Errorf("context cannot be nil")
	}
	if name == "" {
		return fmt.Errorf("metric name cannot be empty")
	}
	// OpenTelemetry gauges don't support Add operation directly
	return fmt.Errorf("AddGauge not supported for OpenTelemetry gauges, use SetGauge instead")
}

// Histogram operations
func (o *openTelemetryMetricsProvider) RecordHistogram(ctx context.Context, name string, value float64, labels map[string]string) error {
	if ctx == nil {
		return fmt.Errorf("context cannot be nil")
	}
	if name == "" {
		return fmt.Errorf("metric name cannot be empty")
	}
	if value < 0 {
		return fmt.Errorf("histogram value must be non-negative, got %f", value)
	}

	histogram, err := o.getHistogram(name)
	if err != nil {
		return err
	}

	attrs := labelsToAttributes(labels)
	histogram.Record(ctx, value, metric.WithAttributes(attrs...))
	return nil
}

// Timer operations
func (o *openTelemetryMetricsProvider) RecordDuration(ctx context.Context, name string, duration time.Duration, labels map[string]string) error {
	if ctx == nil {
		return fmt.Errorf("context cannot be nil")
	}
	if name == "" {
		return fmt.Errorf("metric name cannot be empty")
	}
	if duration < 0 {
		return fmt.Errorf("duration must be non-negative, got %v", duration)
	}
	return o.RecordHistogram(ctx, name, duration.Seconds(), labels)
}

// =============================================================================
// METRIC REGISTRATION
// =============================================================================

func (o *openTelemetryMetricsProvider) RegisterCounter(name, help string, labelNames []string) error {
	if name == "" {
		return fmt.Errorf("metric name cannot be empty")
	}

	o.mutex.Lock()
	defer o.mutex.Unlock()

	if _, exists := o.counters[name]; exists {
		return nil // Already registered, ignore
	}

	counter, err := o.meter.Int64Counter(name, metric.WithDescription(help))
	if err != nil {
		return fmt.Errorf("failed to create counter %s: %w", name, err)
	}

	o.counters[name] = counter
	return nil
}

func (o *openTelemetryMetricsProvider) RegisterGauge(name, help string, labelNames []string) error {
	if name == "" {
		return fmt.Errorf("metric name cannot be empty")
	}

	o.mutex.Lock()
	defer o.mutex.Unlock()

	if _, exists := o.gauges[name]; exists {
		return nil // Already registered, ignore
	}

	gauge, err := o.meter.Float64Gauge(name, metric.WithDescription(help))
	if err != nil {
		return fmt.Errorf("failed to create gauge %s: %w", name, err)
	}

	o.gauges[name] = gauge
	return nil
}

func (o *openTelemetryMetricsProvider) RegisterHistogram(name, help string, labelNames []string, buckets []float64) error {
	if name == "" {
		return fmt.Errorf("metric name cannot be empty")
	}

	o.mutex.Lock()
	defer o.mutex.Unlock()

	if _, exists := o.histograms[name]; exists {
		return nil // Already registered, ignore
	}

	opts := []metric.Float64HistogramOption{
		metric.WithDescription(help),
	}

	if buckets != nil {
		opts = append(opts, metric.WithExplicitBucketBoundaries(buckets...))
	}

	histogram, err := o.meter.Float64Histogram(name, opts...)
	if err != nil {
		return fmt.Errorf("failed to create histogram %s: %w", name, err)
	}

	o.histograms[name] = histogram
	return nil
}

// =============================================================================
// HELPER METHODS FOR METRIC RETRIEVAL
// =============================================================================

func (o *openTelemetryMetricsProvider) getCounter(name string) (metric.Int64Counter, error) {
	o.mutex.RLock()
	defer o.mutex.RUnlock()

	counter, exists := o.counters[name]
	if !exists {
		return nil, fmt.Errorf("counter %s not registered", name)
	}
	return counter, nil
}

func (o *openTelemetryMetricsProvider) getGauge(name string) (metric.Float64Gauge, error) {
	o.mutex.RLock()
	defer o.mutex.RUnlock()

	gauge, exists := o.gauges[name]
	if !exists {
		return nil, fmt.Errorf("gauge %s not registered", name)
	}
	return gauge, nil
}

func (o *openTelemetryMetricsProvider) getHistogram(name string) (metric.Float64Histogram, error) {
	o.mutex.RLock()
	defer o.mutex.RUnlock()

	histogram, exists := o.histograms[name]
	if !exists {
		return nil, fmt.Errorf("histogram %s not registered", name)
	}
	return histogram, nil
}

// =============================================================================
// LIFECYCLE MANAGEMENT
// =============================================================================

func (o *openTelemetryMetricsProvider) Start() error {
	// OpenTelemetry metrics start automatically when created
	return nil
}

func (o *openTelemetryMetricsProvider) Stop() error {
	if o.meterProvider != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return o.meterProvider.Shutdown(ctx)
	}
	return nil
}

func (o *openTelemetryMetricsProvider) Name() string {
	return o.name
}
