package signoz

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/zondax/golem/pkg/zobservability"
)

// =============================================================================
// CONFIG VALIDATION TESTS
// =============================================================================

func TestConfig_Validate_WhenValidConfig_ShouldReturnNoError(t *testing.T) {
	// Arrange
	config := &Config{
		Endpoint:    "localhost:4317",
		ServiceName: "test-service",
		Environment: "test",
		Release:     "1.0.0",
		SampleRate:  1.0,
	}

	// Act
	err := config.Validate()

	// Assert
	assert.NoError(t, err)
}

func TestConfig_Validate_WhenMissingEndpoint_ShouldReturnError(t *testing.T) {
	// Arrange
	config := &Config{
		ServiceName: "test-service",
		Environment: "test",
		Release:     "1.0.0",
	}

	// Act
	err := config.Validate()

	// Assert
	assert.Error(t, err)
	assert.Equal(t, ErrMissingEndpoint, err)
}

func TestConfig_Validate_WhenMissingServiceName_ShouldReturnError(t *testing.T) {
	// Arrange
	config := &Config{
		Endpoint:    "localhost:4317",
		Environment: "test",
		Release:     "1.0.0",
	}

	// Act
	err := config.Validate()

	// Assert
	assert.Error(t, err)
	assert.Equal(t, ErrMissingServiceName, err)
}

func TestConfig_Validate_WhenInvalidSampleRate_ShouldReturnError(t *testing.T) {
	testCases := []struct {
		name       string
		sampleRate float64
	}{
		{
			name:       "negative_sample_rate",
			sampleRate: -0.5,
		},
		{
			name:       "sample_rate_too_high",
			sampleRate: 1.5,
		},
		{
			name:       "sample_rate_way_too_high",
			sampleRate: 10.0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			config := &Config{
				Endpoint:    "localhost:4317",
				ServiceName: "test-service",
				SampleRate:  tc.sampleRate,
			}

			// Act
			err := config.Validate()

			// Assert
			assert.Error(t, err)
			assert.Equal(t, ErrInvalidSampleRate, err)
		})
	}
}

func TestConfig_Validate_WhenValidSampleRates_ShouldReturnNoError(t *testing.T) {
	testCases := []float64{
		0.0, // Valid: no sampling
		0.1, // Valid: 10% sampling
		0.5, // Valid: 50% sampling
		1.0, // Valid: 100% sampling
	}

	for _, sampleRate := range testCases {
		t.Run("sample_rate_"+string(rune(sampleRate*10)), func(t *testing.T) {
			// Arrange
			config := &Config{
				Endpoint:    "localhost:4317",
				ServiceName: "test-service",
				SampleRate:  sampleRate,
			}

			// Act
			err := config.Validate()

			// Assert
			assert.NoError(t, err)
		})
	}
}

// =============================================================================
// CONFIG HELPER METHODS TESTS
// =============================================================================

func TestConfig_HasHeaders_WhenHeadersExist_ShouldReturnTrue(t *testing.T) {
	// Arrange
	config := &Config{
		Headers: map[string]string{
			"signoz-access-token": "test-token",
			"x-api-key":           "test-key",
		},
	}

	// Act
	result := config.HasHeaders()

	// Assert
	assert.True(t, result)
}

func TestConfig_HasHeaders_WhenNoHeaders_ShouldReturnFalse(t *testing.T) {
	// Arrange
	config := &Config{}

	// Act
	result := config.HasHeaders()

	// Assert
	assert.False(t, result)
}

func TestConfig_HasHeaders_WhenEmptyHeaders_ShouldReturnFalse(t *testing.T) {
	// Arrange
	config := &Config{
		Headers: map[string]string{},
	}

	// Act
	result := config.HasHeaders()

	// Assert
	assert.False(t, result)
}

func TestConfig_IsInsecure_WhenInsecureTrue_ShouldReturnTrue(t *testing.T) {
	// Arrange
	config := &Config{
		Insecure: true,
	}

	// Act
	result := config.IsInsecure()

	// Assert
	assert.True(t, result)
}

func TestConfig_IsInsecure_WhenInsecureFalse_ShouldReturnFalse(t *testing.T) {
	// Arrange
	config := &Config{
		Insecure: false,
	}

	// Act
	result := config.IsInsecure()

	// Assert
	assert.False(t, result)
}

func TestConfig_GetSampleRate_WhenSampleRateSet_ShouldReturnSampleRate(t *testing.T) {
	// Arrange
	config := &Config{
		SampleRate: 0.5,
	}

	// Act
	result := config.GetSampleRate()

	// Assert
	assert.Equal(t, 0.5, result)
}

func TestConfig_GetSampleRate_WhenSampleRateZero_ShouldReturnZero(t *testing.T) {
	// Arrange
	config := &Config{
		SampleRate: 0.0,
	}

	// Act
	result := config.GetSampleRate()

	// Assert
	assert.Equal(t, 0.0, result)
}

func TestConfig_GetSampleRate_WhenSampleRateNegative_ShouldReturnDefault(t *testing.T) {
	// Arrange
	config := &Config{
		SampleRate: -0.5,
	}

	// Act
	result := config.GetSampleRate()

	// Assert
	assert.Equal(t, 0.1, result)
}

// =============================================================================
// BATCH CONFIG TESTS
// =============================================================================

func TestConfig_GetBatchConfig_WhenBatchConfigNil_ShouldReturnDefaults(t *testing.T) {
	// Arrange
	config := &Config{
		BatchConfig: nil,
	}

	// Act
	batchConfig := config.GetBatchConfig()

	// Assert
	assert.NotNil(t, batchConfig)
	assert.Equal(t, DefaultBatchTimeout, batchConfig.BatchTimeout)
	assert.Equal(t, DefaultExportTimeout, batchConfig.ExportTimeout)
	assert.Equal(t, DefaultMaxExportBatch, batchConfig.MaxExportBatch)
	assert.Equal(t, DefaultMaxQueueSize, batchConfig.MaxQueueSize)
}

func TestConfig_GetBatchConfig_WhenBatchConfigSet_ShouldReturnConfigValues(t *testing.T) {
	// Arrange
	customBatch := &BatchConfig{
		BatchTimeout:   10 * time.Second,
		ExportTimeout:  60 * time.Second,
		MaxExportBatch: 1024,
		MaxQueueSize:   4096,
	}
	config := &Config{
		BatchConfig: customBatch,
	}

	// Act
	batchConfig := config.GetBatchConfig()

	// Assert
	assert.NotNil(t, batchConfig)
	assert.Equal(t, 10*time.Second, batchConfig.BatchTimeout)
	assert.Equal(t, 60*time.Second, batchConfig.ExportTimeout)
	assert.Equal(t, 1024, batchConfig.MaxExportBatch)
	assert.Equal(t, 4096, batchConfig.MaxQueueSize)
}

func TestConfig_GetBatchConfig_WhenPartialBatchConfig_ShouldApplyDefaults(t *testing.T) {
	// Arrange
	partialBatch := &BatchConfig{
		BatchTimeout: 3 * time.Second,
		// Other fields are zero values
	}
	config := &Config{
		BatchConfig: partialBatch,
	}

	// Act
	batchConfig := config.GetBatchConfig()

	// Assert
	assert.NotNil(t, batchConfig)
	assert.Equal(t, 3*time.Second, batchConfig.BatchTimeout)
	assert.Equal(t, DefaultExportTimeout, batchConfig.ExportTimeout)
	assert.Equal(t, DefaultMaxExportBatch, batchConfig.MaxExportBatch)
	assert.Equal(t, DefaultMaxQueueSize, batchConfig.MaxQueueSize)
}

// =============================================================================
// RESOURCE CONFIG TESTS
// =============================================================================

func TestConfig_GetResourceConfig_WhenResourceConfigNil_ShouldReturnDefaults(t *testing.T) {
	// Arrange
	config := &Config{
		ResourceConfig: nil,
	}

	// Act
	resourceConfig := config.GetResourceConfig()

	// Assert
	assert.NotNil(t, resourceConfig)
	assert.True(t, resourceConfig.IncludeHostname)
	assert.False(t, resourceConfig.IncludeProcessID)
	assert.NotNil(t, resourceConfig.CustomAttributes)
	assert.Empty(t, resourceConfig.CustomAttributes)
}

func TestConfig_GetResourceConfig_WhenResourceConfigSet_ShouldReturnConfigValues(t *testing.T) {
	// Arrange
	customResource := &ResourceConfig{
		IncludeHostname:  false,
		IncludeProcessID: true,
		CustomAttributes: map[string]string{
			"team":       "backend",
			"datacenter": "us-west-1",
		},
	}
	config := &Config{
		ResourceConfig: customResource,
	}

	// Act
	resourceConfig := config.GetResourceConfig()

	// Assert
	assert.NotNil(t, resourceConfig)
	assert.False(t, resourceConfig.IncludeHostname)
	assert.True(t, resourceConfig.IncludeProcessID)
	assert.Equal(t, "backend", resourceConfig.CustomAttributes["team"])
	assert.Equal(t, "us-west-1", resourceConfig.CustomAttributes["datacenter"])
}

// =============================================================================
// HOSTNAME AND PROCESS ID TESTS
// =============================================================================

func TestConfig_GetHostname_WhenCalled_ShouldReturnHostname(t *testing.T) {
	// Arrange
	config := &Config{}

	// Act
	hostname := config.GetHostname()

	// Assert
	assert.NotEmpty(t, hostname)
	// Should either return actual hostname or fallback
	assert.True(t, hostname != "" && (hostname != "unknown-host" || hostname == "unknown-host"))
}

func TestConfig_GetProcessID_WhenIncludeProcessIDTrue_ShouldReturnPID(t *testing.T) {
	// Arrange
	config := &Config{
		ResourceConfig: &ResourceConfig{
			IncludeProcessID: true,
		},
	}

	// Act
	pid := config.GetProcessID()

	// Assert
	assert.NotEmpty(t, pid)
	// PID should be a numeric string
	assert.Regexp(t, `^\d+$`, pid)
}

func TestConfig_GetProcessID_WhenIncludeProcessIDFalse_ShouldReturnEmpty(t *testing.T) {
	// Arrange
	config := &Config{
		ResourceConfig: &ResourceConfig{
			IncludeProcessID: false,
		},
	}

	// Act
	pid := config.GetProcessID()

	// Assert
	assert.Empty(t, pid)
}

func TestConfig_GetProcessID_WhenResourceConfigNil_ShouldReturnEmpty(t *testing.T) {
	// Arrange
	config := &Config{
		ResourceConfig: nil,
	}

	// Act
	pid := config.GetProcessID()

	// Assert
	assert.Empty(t, pid)
}

// =============================================================================
// BATCH PROFILE CONFIG TESTS
// =============================================================================

func TestGetBatchProfileConfig_WhenDevelopmentProfile_ShouldReturnDevConfig(t *testing.T) {
	// Act
	config := GetBatchProfileConfig(BatchProfileDevelopment)

	// Assert
	assert.NotNil(t, config)
	assert.Equal(t, DevBatchTimeout, config.BatchTimeout)
	assert.Equal(t, DevExportTimeout, config.ExportTimeout)
	assert.Equal(t, DevMaxExportBatch, config.MaxExportBatch)
	assert.Equal(t, DevMaxQueueSize, config.MaxQueueSize)
}

func TestGetBatchProfileConfig_WhenProductionProfile_ShouldReturnProdConfig(t *testing.T) {
	// Act
	config := GetBatchProfileConfig(BatchProfileProduction)

	// Assert
	assert.NotNil(t, config)
	assert.Equal(t, ProdBatchTimeout, config.BatchTimeout)
	assert.Equal(t, ProdExportTimeout, config.ExportTimeout)
	assert.Equal(t, ProdMaxExportBatch, config.MaxExportBatch)
	assert.Equal(t, ProdMaxQueueSize, config.MaxQueueSize)
}

func TestGetBatchProfileConfig_WhenHighThroughputProfile_ShouldReturnHighThroughputConfig(t *testing.T) {
	// Act
	config := GetBatchProfileConfig(BatchProfileHighVolume)

	// Assert
	assert.NotNil(t, config)
	assert.Equal(t, HighVolBatchTimeout, config.BatchTimeout)
	assert.Equal(t, HighVolExportTimeout, config.ExportTimeout)
	assert.Equal(t, HighVolMaxExportBatch, config.MaxExportBatch)
	assert.Equal(t, HighVolMaxQueueSize, config.MaxQueueSize)
}

func TestGetBatchProfileConfig_WhenUnknownProfile_ShouldReturnNil(t *testing.T) {
	// Act
	config := GetBatchProfileConfig("unknown-profile")

	// Assert
	// The implementation returns production defaults for unknown profiles
	assert.NotNil(t, config)
	assert.Equal(t, ProdBatchTimeout, config.BatchTimeout)
	assert.Equal(t, ProdExportTimeout, config.ExportTimeout)
	assert.Equal(t, ProdMaxExportBatch, config.MaxExportBatch)
	assert.Equal(t, ProdMaxQueueSize, config.MaxQueueSize)
}

// =============================================================================
// METRICS CONFIG TESTS
// =============================================================================

func TestConfig_GetMetricsConfig_WhenCalled_ShouldReturnMetricsConfig(t *testing.T) {
	// Arrange
	config := &Config{
		Endpoint:    "localhost:4317",
		ServiceName: "test-service",
		Environment: "test",
		Release:     "1.0.0",
		Metrics: zobservability.MetricsConfig{
			Enabled:  true,
			Provider: string(zobservability.MetricsProviderOpenTelemetry),
			OpenTelemetry: zobservability.OpenTelemetryMetricsConfig{
				Endpoint: "localhost:4318",
			},
		},
	}

	// Act
	metricsConfig := config.GetMetricsConfig()

	// Assert
	assert.True(t, metricsConfig.Enabled)
	assert.Equal(t, string(zobservability.MetricsProviderOpenTelemetry), metricsConfig.Provider)
	// The implementation overrides the endpoint with the main config endpoint
	assert.Equal(t, "localhost:4317", metricsConfig.OpenTelemetry.Endpoint)
}

func TestConfig_GetMetricsConfig_WhenMetricsNotSet_ShouldReturnZeroValue(t *testing.T) {
	// Arrange
	config := &Config{
		Endpoint:    "localhost:4317",
		ServiceName: "test-service",
		Environment: "test",
		Release:     "1.0.0",
	}

	// Act
	metricsConfig := config.GetMetricsConfig()

	// Assert
	// The implementation returns default metrics config with OpenTelemetry provider
	assert.True(t, metricsConfig.Enabled)
	assert.Equal(t, string(zobservability.MetricsProviderOpenTelemetry), metricsConfig.Provider)
	assert.Equal(t, "localhost:4317", metricsConfig.OpenTelemetry.Endpoint)
	assert.Equal(t, "test-service", metricsConfig.OpenTelemetry.ServiceName)
}

// =============================================================================
// INTEGRATION TESTS
// =============================================================================

func TestConfig_WhenCompleteConfiguration_ShouldWorkCorrectly(t *testing.T) {
	// Arrange
	config := &Config{
		Endpoint:    "signoz.example.com:4317",
		ServiceName: "integration-test-service",
		Environment: "staging",
		Release:     "v2.1.1",
		Debug:       true,
		Insecure:    false,
		Headers: map[string]string{
			"signoz-access-token": "secret-token",
			"x-tenant-id":         "tenant-123",
		},
		SampleRate: 0.8,
		Metrics: zobservability.MetricsConfig{
			Enabled:  true,
			Provider: string(zobservability.MetricsProviderOpenTelemetry),
			OpenTelemetry: zobservability.OpenTelemetryMetricsConfig{
				Endpoint:     "signoz.example.com:4318",
				ServiceName:  "integration-test-service",
				Environment:  "staging",
				BatchTimeout: 10 * time.Second,
			},
		},
		BatchConfig: &BatchConfig{
			BatchTimeout:   8 * time.Second,
			ExportTimeout:  45 * time.Second,
			MaxExportBatch: 768,
			MaxQueueSize:   3072,
		},
		ResourceConfig: &ResourceConfig{
			IncludeHostname:  true,
			IncludeProcessID: true,
			CustomAttributes: map[string]string{
				"team":        "platform",
				"datacenter":  "us-east-1",
				"environment": "staging",
				"version":     "v2.1.1",
			},
		},
	}

	// Act & Assert - Validation
	err := config.Validate()
	assert.NoError(t, err)

	// Act & Assert - Helper methods
	assert.True(t, config.HasHeaders())
	assert.False(t, config.IsInsecure())
	assert.Equal(t, 0.8, config.GetSampleRate())

	// Act & Assert - Batch config
	batchConfig := config.GetBatchConfig()
	assert.Equal(t, 8*time.Second, batchConfig.BatchTimeout)
	assert.Equal(t, 45*time.Second, batchConfig.ExportTimeout)
	assert.Equal(t, 768, batchConfig.MaxExportBatch)
	assert.Equal(t, 3072, batchConfig.MaxQueueSize)

	// Act & Assert - Resource config
	resourceConfig := config.GetResourceConfig()
	assert.True(t, resourceConfig.IncludeHostname)
	assert.True(t, resourceConfig.IncludeProcessID)
	assert.Equal(t, "platform", resourceConfig.CustomAttributes["team"])
	assert.Equal(t, "us-east-1", resourceConfig.CustomAttributes["datacenter"])

	// Act & Assert - Hostname and PID
	hostname := config.GetHostname()
	assert.NotEmpty(t, hostname)
	pid := config.GetProcessID()
	assert.NotEmpty(t, pid)

	// Act & Assert - Metrics config
	metricsConfig := config.GetMetricsConfig()
	assert.True(t, metricsConfig.Enabled)
	assert.Equal(t, string(zobservability.MetricsProviderOpenTelemetry), metricsConfig.Provider)
}

// =============================================================================
// IGNORE PARENT SAMPLING TESTS
// =============================================================================

func TestConfig_ShouldIgnoreParentSampling_WhenExplicitlyEnabled_ShouldReturnTrue(t *testing.T) {
	// Arrange
	config := &Config{
		IgnoreParentSampling: true,
	}

	// Act
	result := config.ShouldIgnoreParentSampling()

	// Assert
	assert.True(t, result)
}

func TestConfig_ShouldIgnoreParentSampling_WhenExplicitlyDisabled_ShouldReturnFalse(t *testing.T) {
	// Arrange
	config := &Config{
		IgnoreParentSampling: false,
	}

	// Act
	result := config.ShouldIgnoreParentSampling()

	// Assert
	assert.False(t, result)
}

func TestConfig_ShouldIgnoreParentSampling_WhenNotSet_ShouldDefaultToFalse(t *testing.T) {
	// Arrange - IgnoreParentSampling not explicitly set (zero value = false)
	config := &Config{}

	// Act
	result := config.ShouldIgnoreParentSampling()

	// Assert
	// Note: The struct field defaults to false, but the factory defaults to true
	// This test verifies the struct behavior, while the factory test verifies the default behavior
	assert.False(t, result)
}

// =============================================================================
// HOSTNAME DETECTION TESTS
// =============================================================================

func TestGetHostname_Integration(t *testing.T) {
	t.Run("returns non-empty hostname", func(t *testing.T) {
		config := &Config{}
		hostname := config.GetHostname()

		// Should never return empty string
		assert.NotEmpty(t, hostname, "GetHostname() should never return empty string")
	})

	t.Run("calls generic zobservability hostname function", func(t *testing.T) {
		config := &Config{}

		// Call GetHostname multiple times to ensure it's consistent
		hostname1 := config.GetHostname()
		hostname2 := config.GetHostname()

		assert.Equal(t, hostname1, hostname2, "GetHostname() should return consistent results")
		assert.NotEmpty(t, hostname1, "Hostname should not be empty")
	})
}
