package zobservability

import (
	"fmt"
	"time"
)

// =============================================================================
// CONFIGURATION CONSTANTS
// =============================================================================

// Constants for OpenTelemetry metrics configuration
const (
	// Export modes
	OTelExportModePush     = "push"
	OTelExportModeEndpoint = "endpoint"

	// Default intervals
	DefaultPushInterval  = 30 * time.Second
	DefaultBatchTimeout  = 5 * time.Second
	DefaultExportTimeout = 30 * time.Second
)

// =============================================================================
// CONFIGURATION TYPES
// =============================================================================

// OpenTelemetryMetricsConfig holds OpenTelemetry-specific metrics configuration
type OpenTelemetryMetricsConfig struct {
	Endpoint       string            `yaml:"endpoint" mapstructure:"endpoint"`
	Insecure       bool              `yaml:"insecure" mapstructure:"insecure"`
	ServiceName    string            `yaml:"service_name" mapstructure:"service_name"`
	ServiceVersion string            `yaml:"service_version" mapstructure:"service_version"`
	Environment    string            `yaml:"environment" mapstructure:"environment"`
	Hostname       string            `yaml:"hostname" mapstructure:"hostname"`
	Headers        map[string]string `yaml:"headers" mapstructure:"headers"`

	// Export configuration
	ExportMode    string        `yaml:"export_mode" mapstructure:"export_mode"`       // "push" or "endpoint"
	PushInterval  time.Duration `yaml:"push_interval" mapstructure:"push_interval"`   // For push mode
	BatchTimeout  time.Duration `yaml:"batch_timeout" mapstructure:"batch_timeout"`   // Batch timeout
	ExportTimeout time.Duration `yaml:"export_timeout" mapstructure:"export_timeout"` // Export timeout
}

// MetricsConfig holds configuration for metrics
type MetricsConfig struct {
	Enabled       bool                       `yaml:"enabled" mapstructure:"enabled"`
	Provider      string                     `yaml:"provider" mapstructure:"provider"`           // Use MetricsProviderType constants
	Path          string                     `yaml:"path" mapstructure:"path"`                   // Metrics endpoint path (legacy)
	Port          int                        `yaml:"port" mapstructure:"port"`                   // Metrics server port (legacy)
	OpenTelemetry OpenTelemetryMetricsConfig `yaml:"opentelemetry" mapstructure:"opentelemetry"` // OpenTelemetry specific config
}

// =============================================================================
// CONFIGURATION DEFAULTS
// =============================================================================

// DefaultOpenTelemetryMetricsConfig returns default OpenTelemetry metrics configuration
func DefaultOpenTelemetryMetricsConfig() OpenTelemetryMetricsConfig {
	return OpenTelemetryMetricsConfig{
		Endpoint:       "localhost:4317",
		Insecure:       true,
		ServiceName:    "unknown-service",
		ServiceVersion: "1.0.0",
		Environment:    "development",
		Hostname:       "localhost",
		Headers:        make(map[string]string),
		ExportMode:     OTelExportModePush,
		PushInterval:   DefaultPushInterval,
		BatchTimeout:   DefaultBatchTimeout,
		ExportTimeout:  DefaultExportTimeout,
	}
}

// DefaultMetricsConfig returns default metrics configuration
func DefaultMetricsConfig() MetricsConfig {
	return MetricsConfig{
		Enabled:       true,
		Provider:      string(MetricsProviderOpenTelemetry),
		OpenTelemetry: DefaultOpenTelemetryMetricsConfig(),
	}
}

// =============================================================================
// CONFIGURATION VALIDATION
// =============================================================================

// Validate validates the OpenTelemetry metrics configuration
func (c OpenTelemetryMetricsConfig) Validate() error {
	if c.Endpoint == "" {
		return fmt.Errorf("endpoint is required")
	}
	if c.ServiceName == "" {
		return fmt.Errorf("service name is required")
	}
	if c.ExportMode != OTelExportModePush && c.ExportMode != OTelExportModeEndpoint {
		return fmt.Errorf("export_mode must be either '%s' or '%s'", OTelExportModePush, OTelExportModeEndpoint)
	}
	if c.PushInterval <= 0 {
		return fmt.Errorf("push_interval must be greater than 0")
	}
	return nil
}

// Validate validates the metrics configuration
func (c MetricsConfig) Validate() error {
	if !c.Enabled {
		return nil
	}

	if c.Provider == "" {
		return fmt.Errorf("metrics provider is required when enabled")
	}

	// Validate provider-specific configuration using the typed constants
	providerType := MetricsProviderType(c.Provider)
	switch providerType {
	case MetricsProviderOpenTelemetry:
		return c.OpenTelemetry.Validate()
	case MetricsProviderNoop:
		// No validation needed for noop provider
		return nil
	default:
		return fmt.Errorf("unsupported metrics provider: %s", c.Provider)
	}
}
