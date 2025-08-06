package signoz

import (
	"os"
	"strconv"
	"time"

	"github.com/zondax/golem/pkg/zobservability"
)

// Config holds the configuration for the SigNoz observer
type Config struct {
	Endpoint    string            `yaml:"endpoint" mapstructure:"endpoint"`
	ServiceName string            `yaml:"service_name" mapstructure:"service_name"`
	Environment string            `yaml:"environment" mapstructure:"environment"`
	Release     string            `yaml:"release" mapstructure:"release"`
	Debug       bool              `yaml:"debug" mapstructure:"debug"`
	Insecure    bool              `yaml:"insecure" mapstructure:"insecure"`
	Headers     map[string]string `yaml:"headers" mapstructure:"headers"`
	SampleRate  float64           `yaml:"sample_rate" mapstructure:"sample_rate"`

	// IgnoreParentSampling forces sampling decisions to be made locally,
	// ignoring parent trace sampling decisions from headers (like traceparent).
	// This is ESSENTIAL for Google Cloud Run deployments where GCP automatically
	// injects trace headers with sampling decisions that can cause traces to be dropped.
	// Set to true when deploying to Cloud Run or other GCP services.
	IgnoreParentSampling bool `yaml:"ignore_parent_sampling" mapstructure:"ignore_parent_sampling"`

	// Metrics configuration
	Metrics zobservability.MetricsConfig `yaml:"metrics" mapstructure:"metrics"`

	// Advanced configuration
	// BatchConfig is a POINTER (*BatchConfig) for these reasons:
	// 1. OPTIONAL: nil means "use defaults", non-nil means "use custom values"
	// 2. MEMORY EFFICIENT: Only allocates memory when actually configured
	// 3. YAML FLEXIBILITY: Can be omitted entirely from config files
	// 4. PARTIAL OVERRIDE: Can set only some fields, others use defaults
	BatchConfig *BatchConfig `yaml:"batch_config,omitempty" mapstructure:"batch_config"`

	// ResourceConfig is also a POINTER for the same reasons as BatchConfig
	// Allows optional metadata configuration without forcing all users to specify it
	ResourceConfig *ResourceConfig `yaml:"resource_config,omitempty" mapstructure:"resource_config"`

	// UseSimpleSpan enables immediate span export without batching
	// When true, spans are exported immediately when they finish instead of being batched
	// This can increase network overhead but provides real-time visibility
	UseSimpleSpan bool `yaml:"use_simple_span" mapstructure:"use_simple_span"`

	// Propagation configuration
	Propagation zobservability.PropagationConfig `yaml:"propagation" mapstructure:"propagation"`

	// TracingExclusions contains the list of gRPC methods to exclude from tracing
	TracingExclusions []string `yaml:"tracing_exclusions" mapstructure:"tracing_exclusions"`
}

// BatchConfig controls how spans are batched and sent to SigNoz
// This configuration directly affects:
// - PERFORMANCE: How efficiently data is sent to SigNoz
// - MEMORY USAGE: How much memory is used for buffering
// - LATENCY: How quickly traces appear in SigNoz UI
// - RELIABILITY: How much data can be lost during traffic spikes
type BatchConfig struct {
	// BatchTimeout is how often to send batches (default: 5s)
	// Lower values = more real-time visibility, higher network overhead
	// Higher values = more efficient batching, delayed visibility
	// Example: "1s" for development, "5s" for production
	BatchTimeout time.Duration `yaml:"batch_timeout,omitempty" mapstructure:"batch_timeout"`

	// ExportTimeout is timeout for individual exports (default: 30s)
	// Should be higher than your network latency to SigNoz
	// Too low = failed exports, too high = hanging connections
	ExportTimeout time.Duration `yaml:"export_timeout,omitempty" mapstructure:"export_timeout"`

	// MaxExportBatch is maximum spans per batch (default: 512)
	// Higher values = more efficient network usage, more memory per batch
	// Lower values = less memory usage, more network requests
	// SigNoz recommendation: 512 for most cases
	MaxExportBatch int `yaml:"max_export_batch,omitempty" mapstructure:"max_export_batch"`

	// MaxQueueSize is maximum spans in queue (default: 2048)
	// Higher values = less data loss during traffic spikes, more memory usage
	// Lower values = less memory usage, potential data loss under load
	// Should be 4x MaxExportBatch for optimal performance
	MaxQueueSize int `yaml:"max_queue_size,omitempty" mapstructure:"max_queue_size"`
}

// ResourceConfig controls what metadata is attached to traces
type ResourceConfig struct {
	// IncludeHostname adds host.name to resource attributes
	IncludeHostname bool `yaml:"include_hostname,omitempty" mapstructure:"include_hostname"`
	// IncludeProcessID adds process.pid to resource attributes
	IncludeProcessID bool `yaml:"include_process_id,omitempty" mapstructure:"include_process_id"`
	// CustomAttributes allows adding custom resource attributes
	CustomAttributes map[string]string `yaml:"custom_attributes,omitempty" mapstructure:"custom_attributes"`
}

// Validate validates the SigNoz configuration
func (c *Config) Validate() error {
	if c.Endpoint == "" {
		return ErrMissingEndpoint
	}
	if c.ServiceName == "" {
		return ErrMissingServiceName
	}
	if c.SampleRate < 0 || c.SampleRate > 1 {
		return ErrInvalidSampleRate
	}
	return nil
}

// HasHeaders returns true if custom headers are configured
func (c *Config) HasHeaders() bool {
	return len(c.Headers) > 0
}

// IsInsecure returns true if insecure mode is enabled
func (c *Config) IsInsecure() bool {
	return c.Insecure
}

// GetSampleRate returns the sample rate, defaulting to 0.1 if not set or invalid
func (c *Config) GetSampleRate() float64 {
	// Only negative values should default to 0.1 (conservative approach)
	if c.SampleRate < 0 {
		return 0.1
	}
	// If SampleRate is 0, it means "no sampling" (0%)
	// If SampleRate is 0.5, it means "50% sampling"
	// If SampleRate is 1.0, it means "100% sampling"
	return c.SampleRate
}

// GetBatchConfig returns batch configuration with defaults
func (c *Config) GetBatchConfig() *BatchConfig {
	if c.BatchConfig == nil {
		return &BatchConfig{
			BatchTimeout:   DefaultBatchTimeout,
			ExportTimeout:  DefaultExportTimeout,
			MaxExportBatch: DefaultMaxExportBatch,
			MaxQueueSize:   DefaultMaxQueueSize,
		}
	}

	// Apply defaults for unset values
	batch := *c.BatchConfig
	if batch.BatchTimeout == 0 {
		batch.BatchTimeout = DefaultBatchTimeout
	}
	if batch.ExportTimeout == 0 {
		batch.ExportTimeout = DefaultExportTimeout
	}
	if batch.MaxExportBatch == 0 {
		batch.MaxExportBatch = DefaultMaxExportBatch
	}
	if batch.MaxQueueSize == 0 {
		batch.MaxQueueSize = DefaultMaxQueueSize
	}

	return &batch
}

// GetResourceConfig returns resource configuration with defaults
func (c *Config) GetResourceConfig() *ResourceConfig {
	if c.ResourceConfig == nil {
		return &ResourceConfig{
			IncludeHostname:  true,  // CHANGED: hostname is MANDATORY by default (We can have many previews)
			IncludeProcessID: false, // Default: don't include PID for security
			CustomAttributes: make(map[string]string),
		}
	}

	return c.ResourceConfig
}

// GetHostname returns the hostname using the generic zobservability hostname detection
func (c *Config) GetHostname() string {
	return zobservability.GetHostname()
}

// GetProcessID returns the process ID if configured to include it
// Process ID (PID) is useful for:
// - DEBUGGING: Identifying which specific process handled a request
// - MULTI-PROCESS APPS: When running multiple instances on same server
// - MEMORY ANALYSIS: Correlating traces with memory/CPU usage by process
// - RESTART DETECTION: Knowing when a process was restarted (PID changes)
// - CONTAINER DEBUGGING: Identifying processes within containers
//
// Security considerations:
// - PIDs can reveal system information
// - Generally safe but some security policies prohibit exposing them
// - Recommended: Enable only in development/staging, disable in production
func (c *Config) GetProcessID() string {
	if c.GetResourceConfig().IncludeProcessID {
		return strconv.Itoa(os.Getpid())
	}
	return ""
}

// ShouldIgnoreParentSampling returns true if parent sampling decisions should be ignored.
// DEFAULT BEHAVIOR: Returns true by default to prevent trace loss in cloud environments.
//
// This fixes a common issue where cloud platforms (Google Cloud Run, Cloud Functions,
// App Engine, etc.) automatically inject trace headers with sampling decisions that
// cause traces to be dropped, making distributed tracing nearly useless.
//
// The method respects explicit configuration:
// - ignore_parent_sampling: true (explicit enable)
// - ignore_parent_sampling: false (explicit disable - use with caution in cloud environments)
//
// When not explicitly configured, defaults to true to ensure traces are not lost.
func (c *Config) ShouldIgnoreParentSampling() bool {
	// DEFAULT: true for cloud environments to prevent trace loss
	// Return the configured value, with true as default
	return c.IgnoreParentSampling
}

// GetBatchProfileConfig returns a predefined batch configuration for the specified profile
// This provides standardized, performance-optimized configurations for different scenarios
func GetBatchProfileConfig(profile string) *BatchConfig {
	switch profile {
	case BatchProfileDevelopment:
		return &BatchConfig{
			BatchTimeout:   DevBatchTimeout,
			ExportTimeout:  DevExportTimeout,
			MaxExportBatch: DevMaxExportBatch,
			MaxQueueSize:   DevMaxQueueSize,
		}
	case BatchProfileProduction:
		return &BatchConfig{
			BatchTimeout:   ProdBatchTimeout,
			ExportTimeout:  ProdExportTimeout,
			MaxExportBatch: ProdMaxExportBatch,
			MaxQueueSize:   ProdMaxQueueSize,
		}
	case BatchProfileHighVolume:
		return &BatchConfig{
			BatchTimeout:   HighVolBatchTimeout,
			ExportTimeout:  HighVolExportTimeout,
			MaxExportBatch: HighVolMaxExportBatch,
			MaxQueueSize:   HighVolMaxQueueSize,
		}
	case BatchProfileLowLatency:
		return &BatchConfig{
			BatchTimeout:   LowLatBatchTimeout,
			ExportTimeout:  LowLatExportTimeout,
			MaxExportBatch: LowLatMaxExportBatch,
			MaxQueueSize:   LowLatMaxQueueSize,
		}
	default:
		// Default to production profile for unknown profiles
		return GetBatchProfileConfig(BatchProfileProduction)
	}
}

// GetMetricsConfig returns the metrics configuration with defaults applied
func (c *Config) GetMetricsConfig() zobservability.MetricsConfig {
	metrics := zobservability.DefaultMetricsConfig()

	// Configure OpenTelemetry metrics with SigNoz defaults
	metrics.OpenTelemetry.Endpoint = c.Endpoint
	metrics.OpenTelemetry.ServiceName = c.ServiceName
	metrics.OpenTelemetry.ServiceVersion = c.Release
	metrics.OpenTelemetry.Environment = c.Environment
	metrics.OpenTelemetry.Hostname = c.GetHostname()
	metrics.OpenTelemetry.Insecure = c.IsInsecure()

	// Copy headers if present
	if c.HasHeaders() {
		metrics.OpenTelemetry.Headers = make(map[string]string)
		for k, v := range c.Headers {
			metrics.OpenTelemetry.Headers[k] = v
		}
	}

	return metrics
}

// GetPropagationConfig returns the propagation configuration with defaults
func (c *Config) GetPropagationConfig() zobservability.PropagationConfig {
	if len(c.Propagation.Formats) == 0 {
		// Default to B3 because is the only one supported by GCP+Signoz
		return zobservability.PropagationConfig{
			Formats: []string{zobservability.PropagationB3},
		}
	}
	return c.Propagation
}
