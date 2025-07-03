package signoz

import "time"

const (
	// TracerName is the name used for the OpenTelemetry tracer
	TracerName = "signoz"

	// DefaultShutdownTimeout is the default timeout for shutting down the tracer provider
	DefaultShutdownTimeout = 5 * time.Second

	// DefaultForceFlushTimeout is the default timeout for forcing span exports
	// Optimized for Cloud Run environments where containers can be terminated quickly
	DefaultForceFlushTimeout = 10 * time.Second

	// Batch processor configuration
	DefaultBatchTimeout   = 5 * time.Second  // How often to send batches
	DefaultExportTimeout  = 30 * time.Second // Timeout for individual exports
	DefaultMaxExportBatch = 512              // Maximum spans per batch
	DefaultMaxQueueSize   = 2048             // Maximum spans in queue

	// Factory configuration keys - used in factory.go to parse custom_config
	ConfigKeyHeaderPrefix = "header_"  // Prefix for header configuration keys
	ConfigKeyInsecure     = "insecure" // Key for insecure connection setting
	ConfigKeyTrueValue    = "true"     // String value representing boolean true

	// SignOz Token Configuration Keys
	HeaderSignOzAccessToken     = "signoz-access-token"        // Header key for SignOz access token
	ConfigKeyHeaderSignOzToken  = "header_signoz-access-token" // nolint:gosec // Config key for SignOz token in custom_config
	GCPObservabilityAPIKeyField = "gcp_observability_api_key"  // nolint:gosec // GCP field name for observability API key

	// Advanced configuration keys
	ConfigKeyBatchConfig    = "batch_config"    // Key for batch configuration
	ConfigKeyResourceConfig = "resource_config" // Key for resource configuration
	ConfigKeyBatchProfile   = "batch_profile"   // Key for predefined batch profile

	// Sampling configuration keys
	ConfigKeyIgnoreParentSampling = "ignore_parent_sampling" // Key for ignoring parent sampling decisions

	// SimpleSpan configuration keys
	ConfigKeyUseSimpleSpan = "use_simple_span" // Key for enabling SimpleSpan immediate export

	// Standardized Batch Profiles - optimized configurations for different scenarios
	BatchProfileDevelopment = "development" // Real-time visibility, low latency
	BatchProfileProduction  = "production"  // Balanced performance and efficiency
	BatchProfileHighVolume  = "high_volume" // Maximum efficiency for high traffic
	BatchProfileLowLatency  = "low_latency" // Minimal delay, higher overhead

	// Development Profile - optimized for debugging and real-time feedback
	DevBatchTimeout   = 1 * time.Second  // Very fast batching for immediate visibility
	DevExportTimeout  = 10 * time.Second // Quick timeout for fast feedback
	DevMaxExportBatch = 100              // Small batches for low memory usage
	DevMaxQueueSize   = 500              // Small queue for development

	// Production Profile - balanced performance and resource usage
	ProdBatchTimeout   = 5 * time.Second  // Standard batching interval
	ProdExportTimeout  = 30 * time.Second // Reliable timeout for production
	ProdMaxExportBatch = 512              // Optimal batch size for most cases
	ProdMaxQueueSize   = 2048             // Good balance of memory and reliability

	// High Volume Profile - maximum efficiency for high-traffic applications
	HighVolBatchTimeout   = 10 * time.Second // Larger batches for efficiency
	HighVolExportTimeout  = 60 * time.Second // Longer timeout for large batches
	HighVolMaxExportBatch = 1000             // Large batches for maximum efficiency
	HighVolMaxQueueSize   = 5000             // Large queue to handle traffic spikes

	// Low Latency Profile - minimal delay for real-time monitoring
	LowLatBatchTimeout   = 500 * time.Millisecond // Very fast batching
	LowLatExportTimeout  = 5 * time.Second        // Quick timeout
	LowLatMaxExportBatch = 50                     // Very small batches
	LowLatMaxQueueSize   = 200                    // Small queue for minimal delay
)
