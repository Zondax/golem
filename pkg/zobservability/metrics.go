package zobservability

import (
	"context"
	"time"
)

// =============================================================================
// CORE INTERFACES AND TYPES
// =============================================================================

// MetricsProvider defines the interface for metrics operations
// Trace correlation is handled transparently when context is available
type MetricsProvider interface {
	// Counter operations
	IncrementCounter(ctx context.Context, name string, labels map[string]string) error
	AddCounter(ctx context.Context, name string, value float64, labels map[string]string) error

	// Gauge operations
	SetGauge(ctx context.Context, name string, value float64, labels map[string]string) error
	AddGauge(ctx context.Context, name string, value float64, labels map[string]string) error

	// Histogram operations
	RecordHistogram(ctx context.Context, name string, value float64, labels map[string]string) error

	// Timer operations
	RecordDuration(ctx context.Context, name string, duration time.Duration, labels map[string]string) error

	// Metric registration
	RegisterCounter(name, help string, labelNames []string) error
	RegisterGauge(name, help string, labelNames []string) error
	RegisterHistogram(name, help string, labelNames []string, buckets []float64) error

	// Lifecycle
	Start() error
	Stop() error
	Name() string
}

// MetricType represents the type of metric
type MetricType int

const (
	MetricTypeCounter MetricType = iota
	MetricTypeGauge
	MetricTypeHistogram
)

// MetricDefinition defines a metric to be registered
type MetricDefinition struct {
	Name       string
	Help       string
	Type       MetricType
	LabelNames []string
	Buckets    []float64 // Only used for histograms
}
