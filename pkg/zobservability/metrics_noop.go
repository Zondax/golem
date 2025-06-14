package zobservability

import (
	"context"
	"fmt"
	"time"
)

// noopMetricsProvider is a no-operation implementation of MetricsProvider
type noopMetricsProvider struct {
	name string
}

// NewNoopMetricsProvider creates a new no-operation metrics provider
func NewNoopMetricsProvider(name string) MetricsProvider {
	return &noopMetricsProvider{name: name}
}

// Counter operations
func (n *noopMetricsProvider) IncrementCounter(ctx context.Context, name string, labels map[string]string) error {
	if ctx == nil {
		return fmt.Errorf("context cannot be nil")
	}
	if name == "" {
		return fmt.Errorf("metric name cannot be empty")
	}
	return nil
}

func (n *noopMetricsProvider) AddCounter(ctx context.Context, name string, value float64, labels map[string]string) error {
	if ctx == nil {
		return fmt.Errorf("context cannot be nil")
	}
	if name == "" {
		return fmt.Errorf("metric name cannot be empty")
	}
	if value < 0 {
		return fmt.Errorf("counter value must be non-negative, got %f", value)
	}
	return nil
}

// Gauge operations
func (n *noopMetricsProvider) SetGauge(ctx context.Context, name string, value float64, labels map[string]string) error {
	if ctx == nil {
		return fmt.Errorf("context cannot be nil")
	}
	if name == "" {
		return fmt.Errorf("metric name cannot be empty")
	}
	return nil
}

func (n *noopMetricsProvider) AddGauge(ctx context.Context, name string, value float64, labels map[string]string) error {
	if ctx == nil {
		return fmt.Errorf("context cannot be nil")
	}
	if name == "" {
		return fmt.Errorf("metric name cannot be empty")
	}
	return nil
}

// Histogram operations
func (n *noopMetricsProvider) RecordHistogram(ctx context.Context, name string, value float64, labels map[string]string) error {
	if ctx == nil {
		return fmt.Errorf("context cannot be nil")
	}
	if name == "" {
		return fmt.Errorf("metric name cannot be empty")
	}
	if value < 0 {
		return fmt.Errorf("histogram value must be non-negative, got %f", value)
	}
	return nil
}

// Timer operations
func (n *noopMetricsProvider) RecordDuration(ctx context.Context, name string, duration time.Duration, labels map[string]string) error {
	if ctx == nil {
		return fmt.Errorf("context cannot be nil")
	}
	if name == "" {
		return fmt.Errorf("metric name cannot be empty")
	}
	if duration < 0 {
		return fmt.Errorf("duration must be non-negative, got %v", duration)
	}
	return nil
}

// Metric registration
func (n *noopMetricsProvider) RegisterCounter(name, help string, labelNames []string) error {
	if name == "" {
		return fmt.Errorf("metric name cannot be empty")
	}
	return nil
}

func (n *noopMetricsProvider) RegisterGauge(name, help string, labelNames []string) error {
	if name == "" {
		return fmt.Errorf("metric name cannot be empty")
	}
	return nil
}

func (n *noopMetricsProvider) RegisterHistogram(name, help string, labelNames []string, buckets []float64) error {
	if name == "" {
		return fmt.Errorf("metric name cannot be empty")
	}
	return nil
}

// Lifecycle
func (n *noopMetricsProvider) Start() error {
	return nil
}

func (n *noopMetricsProvider) Stop() error {
	return nil
}

func (n *noopMetricsProvider) Name() string {
	return n.name
}
