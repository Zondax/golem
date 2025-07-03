package zobservability

import (
	"fmt"
)

// SpanCountingConfig controls span counting functionality
type SpanCountingConfig struct {
	Enabled       bool `yaml:"enabled" mapstructure:"enabled"`
	LogSpanCounts bool `yaml:"log_span_counts" mapstructure:"log_span_counts"`
}

// Config holds configuration for all observability features (tracing, logging, metrics)
type Config struct {
	Provider     string             `yaml:"provider" mapstructure:"provider"`
	Enabled      bool               `yaml:"enabled" mapstructure:"enabled"` // Enable/disable observability
	Environment  string             `yaml:"environment" mapstructure:"environment"`
	Release      string             `yaml:"release" mapstructure:"release"`
	Debug        bool               `yaml:"debug" mapstructure:"debug"`
	Address      string             `yaml:"address" mapstructure:"address"`         // Common endpoint/address/dsn for providers
	SampleRate   float64            `yaml:"sample_rate" mapstructure:"sample_rate"` // Common sampling rate
	Middleware   MiddlewareConfig   `yaml:"middleware" mapstructure:"middleware"`
	Metrics      MetricsConfig      `yaml:"metrics" mapstructure:"metrics"`                     // Metrics configuration
	SpanCounting *SpanCountingConfig `yaml:"span_counting,omitempty" mapstructure:"span_counting"` // Span counting configuration
	CustomConfig map[string]string  `yaml:"custom_config" mapstructure:"custom_config"`        // Provider-specific configuration
}

func (c Config) Validate() error {
	if !c.Enabled {
		return nil // Skip validation if observability is disabled
	}

	if c.Provider == "" {
		return fmt.Errorf("observability provider is required when enabled")
	}

	if c.Environment == "" {
		return fmt.Errorf("observability environment is required when enabled")
	}
	if c.Address == "" {
		return fmt.Errorf("observability address is required when enabled")
	}

	// Validate metrics configuration
	if err := c.Metrics.Validate(); err != nil {
		return fmt.Errorf("invalid metrics configuration: %w", err)
	}

	return nil
}

type MiddlewareConfig struct {
	CaptureErrors bool `yaml:"capture_errors" mapstructure:"capture_errors"`
}

func (c *Config) SetDefaults() {
	if c.Environment == "" {
		c.Environment = "development"
	}
	if c.SampleRate == 0 {
		c.SampleRate = 0.1
	}

	c.Middleware.CaptureErrors = true // Enabled by default

	// Set metrics defaults
	if c.Metrics.Provider == "" {
		c.Metrics = DefaultMetricsConfig()
	}
}
