package zobservability

import (
	"fmt"
)

// PropagationConfig controls trace propagation formats
type PropagationConfig struct {
	Formats []string `yaml:"formats" mapstructure:"formats"` // ["w3c", "b3", "b3-single", "jaeger"]
}

// Config holds configuration for all observability features (tracing, logging, metrics)
type Config struct {
	Provider                        string            `yaml:"provider" mapstructure:"provider"`
	Enabled                         bool              `yaml:"enabled" mapstructure:"enabled"` // Enable/disable observability
	Environment                     string            `yaml:"environment" mapstructure:"environment"`
	Release                         string            `yaml:"release" mapstructure:"release"`
	Debug                           bool              `yaml:"debug" mapstructure:"debug"`
	Address                         string            `yaml:"address" mapstructure:"address"`         // Common endpoint/address/dsn for providers
	SampleRate                      float64           `yaml:"sample_rate" mapstructure:"sample_rate"` // Common sampling rate
	Middleware                      MiddlewareConfig  `yaml:"middleware" mapstructure:"middleware"`
	Metrics                         MetricsConfig     `yaml:"metrics" mapstructure:"metrics"`             // Metrics configuration
	Propagation                     PropagationConfig `yaml:"propagation" mapstructure:"propagation"`     // Trace propagation configuration
	CustomConfig                    map[string]string `yaml:"custom_config" mapstructure:"custom_config"` // Provider-specific configuration
	InterceptorTracingExcludeMethods []string         `yaml:"interceptor_tracing_exclude_methods" mapstructure:"interceptor_tracing_exclude_methods"` // Methods to exclude from tracing
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

	// Set propagation defaults
	if len(c.Propagation.Formats) == 0 {
		c.Propagation.Formats = []string{PropagationB3} // Default to B3 because is the only one supported by GCP+Signoz
	}

	// InterceptorTracingExcludeMethods will be configured via YAML or environment variables
	// No hardcoded defaults here
}
