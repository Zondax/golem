package zobservability

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// =============================================================================
// OPENTELEMETRY METRICS CONFIG TESTS
// =============================================================================

func TestDefaultOpenTelemetryMetricsConfig_WhenCalled_ShouldReturnValidDefaults(t *testing.T) {
	// Act
	config := DefaultOpenTelemetryMetricsConfig()

	// Assert
	assert.Equal(t, "localhost:4317", config.Endpoint)
	assert.True(t, config.Insecure)
	assert.Equal(t, "unknown-service", config.ServiceName)
	assert.Equal(t, "1.0.0", config.ServiceVersion)
	assert.Equal(t, "development", config.Environment)
	assert.Equal(t, "localhost", config.Hostname)
	assert.NotNil(t, config.Headers)
	assert.Empty(t, config.Headers)
	assert.Equal(t, OTelExportModePush, config.ExportMode)
	assert.Equal(t, DefaultPushInterval, config.PushInterval)
	assert.Equal(t, DefaultBatchTimeout, config.BatchTimeout)
	assert.Equal(t, DefaultExportTimeout, config.ExportTimeout)
}

func TestOpenTelemetryMetricsConfig_WhenValidateWithValidConfig_ShouldReturnNil(t *testing.T) {
	// Arrange
	config := OpenTelemetryMetricsConfig{
		Endpoint:      "localhost:4317",
		ServiceName:   "test-service",
		ExportMode:    OTelExportModePush,
		PushInterval:  30 * time.Second,
		BatchTimeout:  5 * time.Second,
		ExportTimeout: 30 * time.Second,
	}

	// Act
	err := config.Validate()

	// Assert
	assert.NoError(t, err)
}

func TestOpenTelemetryMetricsConfig_WhenValidateWithEmptyEndpoint_ShouldReturnError(t *testing.T) {
	// Arrange
	config := OpenTelemetryMetricsConfig{
		Endpoint:     "",
		ServiceName:  "test-service",
		ExportMode:   OTelExportModePush,
		PushInterval: 30 * time.Second,
	}

	// Act
	err := config.Validate()

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "endpoint is required")
}

func TestOpenTelemetryMetricsConfig_WhenValidateWithEmptyServiceName_ShouldReturnError(t *testing.T) {
	// Arrange
	config := OpenTelemetryMetricsConfig{
		Endpoint:     "localhost:4317",
		ServiceName:  "",
		ExportMode:   OTelExportModePush,
		PushInterval: 30 * time.Second,
	}

	// Act
	err := config.Validate()

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "service name is required")
}

func TestOpenTelemetryMetricsConfig_WhenValidateWithInvalidExportMode_ShouldReturnError(t *testing.T) {
	// Arrange
	config := OpenTelemetryMetricsConfig{
		Endpoint:     "localhost:4317",
		ServiceName:  "test-service",
		ExportMode:   "invalid-mode",
		PushInterval: 30 * time.Second,
	}

	// Act
	err := config.Validate()

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "export_mode must be either")
	assert.Contains(t, err.Error(), OTelExportModePush)
	assert.Contains(t, err.Error(), OTelExportModeEndpoint)
}

func TestOpenTelemetryMetricsConfig_WhenValidateWithZeroPushInterval_ShouldReturnError(t *testing.T) {
	// Arrange
	config := OpenTelemetryMetricsConfig{
		Endpoint:     "localhost:4317",
		ServiceName:  "test-service",
		ExportMode:   OTelExportModePush,
		PushInterval: 0,
	}

	// Act
	err := config.Validate()

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "push_interval must be greater than 0")
}

func TestOpenTelemetryMetricsConfig_WhenValidateWithNegativePushInterval_ShouldReturnError(t *testing.T) {
	// Arrange
	config := OpenTelemetryMetricsConfig{
		Endpoint:     "localhost:4317",
		ServiceName:  "test-service",
		ExportMode:   OTelExportModePush,
		PushInterval: -5 * time.Second,
	}

	// Act
	err := config.Validate()

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "push_interval must be greater than 0")
}

func TestOpenTelemetryMetricsConfig_WhenValidateWithEndpointMode_ShouldReturnNil(t *testing.T) {
	// Arrange
	config := OpenTelemetryMetricsConfig{
		Endpoint:     "localhost:4317",
		ServiceName:  "test-service",
		ExportMode:   OTelExportModeEndpoint,
		PushInterval: 30 * time.Second,
	}

	// Act
	err := config.Validate()

	// Assert
	assert.NoError(t, err)
}

func TestOpenTelemetryMetricsConfig_WhenValidateWithCompleteConfig_ShouldReturnNil(t *testing.T) {
	// Arrange
	config := OpenTelemetryMetricsConfig{
		Endpoint:       "ingest.signoz.io:443",
		Insecure:       false,
		ServiceName:    "production-service",
		ServiceVersion: "2.1.0",
		Environment:    "production",
		Hostname:       "prod-server-01",
		Headers: map[string]string{
			"signoz-access-token": "token123",
			"x-api-key":           "key456",
		},
		ExportMode:    OTelExportModePush,
		PushInterval:  60 * time.Second,
		BatchTimeout:  10 * time.Second,
		ExportTimeout: 45 * time.Second,
	}

	// Act
	err := config.Validate()

	// Assert
	assert.NoError(t, err)
}

// =============================================================================
// METRICS CONFIG TESTS
// =============================================================================

func TestDefaultMetricsConfig_WhenCalled_ShouldReturnValidDefaults(t *testing.T) {
	// Act
	config := DefaultMetricsConfig()

	// Assert
	assert.True(t, config.Enabled)
	assert.Equal(t, string(MetricsProviderOpenTelemetry), config.Provider)
	assert.Equal(t, DefaultOpenTelemetryMetricsConfig(), config.OpenTelemetry)
}

func TestMetricsConfig_WhenValidateWithDisabledMetrics_ShouldReturnNil(t *testing.T) {
	// Arrange
	config := MetricsConfig{
		Enabled: false,
	}

	// Act
	err := config.Validate()

	// Assert
	assert.NoError(t, err)
}

func TestMetricsConfig_WhenValidateWithEnabledButEmptyProvider_ShouldReturnError(t *testing.T) {
	// Arrange
	config := MetricsConfig{
		Enabled:  true,
		Provider: "",
	}

	// Act
	err := config.Validate()

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "metrics provider is required when enabled")
}

func TestMetricsConfig_WhenValidateWithOpenTelemetryProvider_ShouldValidateOpenTelemetryConfig(t *testing.T) {
	// Arrange
	config := MetricsConfig{
		Enabled:  true,
		Provider: string(MetricsProviderOpenTelemetry),
		OpenTelemetry: OpenTelemetryMetricsConfig{
			Endpoint:     "localhost:4317",
			ServiceName:  "test-service",
			ExportMode:   OTelExportModePush,
			PushInterval: 30 * time.Second,
		},
	}

	// Act
	err := config.Validate()

	// Assert
	assert.NoError(t, err)
}

func TestMetricsConfig_WhenValidateWithInvalidOpenTelemetryConfig_ShouldReturnError(t *testing.T) {
	// Arrange
	config := MetricsConfig{
		Enabled:  true,
		Provider: string(MetricsProviderOpenTelemetry),
		OpenTelemetry: OpenTelemetryMetricsConfig{
			Endpoint:    "", // Invalid: empty endpoint
			ServiceName: "test-service",
			ExportMode:  OTelExportModePush,
		},
	}

	// Act
	err := config.Validate()

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "endpoint is required")
}

func TestMetricsConfig_WhenValidateWithNoopProvider_ShouldReturnNil(t *testing.T) {
	// Arrange
	config := MetricsConfig{
		Enabled:  true,
		Provider: string(MetricsProviderNoop),
	}

	// Act
	err := config.Validate()

	// Assert
	assert.NoError(t, err)
}

func TestMetricsConfig_WhenValidateWithUnsupportedProvider_ShouldReturnError(t *testing.T) {
	// Arrange
	config := MetricsConfig{
		Enabled:  true,
		Provider: "unsupported-provider",
	}

	// Act
	err := config.Validate()

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported metrics provider")
	assert.Contains(t, err.Error(), "unsupported-provider")
}

func TestMetricsConfig_WhenValidateWithCompleteValidConfig_ShouldReturnNil(t *testing.T) {
	// Arrange
	config := MetricsConfig{
		Enabled:  true,
		Provider: string(MetricsProviderOpenTelemetry),
		Path:     "/metrics",
		Port:     9090,
		OpenTelemetry: OpenTelemetryMetricsConfig{
			Endpoint:       "ingest.signoz.io:443",
			Insecure:       false,
			ServiceName:    "production-service",
			ServiceVersion: "2.1.0",
			Environment:    "production",
			Hostname:       "prod-server-01",
			Headers: map[string]string{
				"signoz-access-token": "token123",
			},
			ExportMode:    OTelExportModePush,
			PushInterval:  60 * time.Second,
			BatchTimeout:  10 * time.Second,
			ExportTimeout: 45 * time.Second,
		},
	}

	// Act
	err := config.Validate()

	// Assert
	assert.NoError(t, err)
}

// =============================================================================
// CONSTANTS TESTS
// =============================================================================

func TestOTelExportModeConstants_WhenAccessed_ShouldHaveExpectedValues(t *testing.T) {
	// Assert
	assert.Equal(t, "push", OTelExportModePush)
	assert.Equal(t, "endpoint", OTelExportModeEndpoint)
}

func TestDefaultIntervalConstants_WhenAccessed_ShouldHaveExpectedValues(t *testing.T) {
	// Assert
	assert.Equal(t, 30*time.Second, DefaultPushInterval)
	assert.Equal(t, 5*time.Second, DefaultBatchTimeout)
	assert.Equal(t, 30*time.Second, DefaultExportTimeout)
}

// =============================================================================
// EDGE CASES AND COMPLEX SCENARIOS
// =============================================================================

func TestOpenTelemetryMetricsConfig_WhenValidateWithMinimalValidConfig_ShouldReturnNil(t *testing.T) {
	// Arrange
	config := OpenTelemetryMetricsConfig{
		Endpoint:     "localhost:4317",
		ServiceName:  "minimal-service",
		ExportMode:   OTelExportModePush,
		PushInterval: 1 * time.Nanosecond, // Minimal positive duration
	}

	// Act
	err := config.Validate()

	// Assert
	assert.NoError(t, err)
}

func TestOpenTelemetryMetricsConfig_WhenValidateWithLargeHeaders_ShouldReturnNil(t *testing.T) {
	// Arrange
	largeHeaders := make(map[string]string)
	for i := 0; i < 100; i++ {
		largeHeaders[fmt.Sprintf("header-%d", i)] = fmt.Sprintf("value-%d", i)
	}

	config := OpenTelemetryMetricsConfig{
		Endpoint:     "localhost:4317",
		ServiceName:  "test-service",
		Headers:      largeHeaders,
		ExportMode:   OTelExportModePush,
		PushInterval: 30 * time.Second,
	}

	// Act
	err := config.Validate()

	// Assert
	assert.NoError(t, err)
}

func TestMetricsConfig_WhenValidateWithAllProviderTypes_ShouldHandleCorrectly(t *testing.T) {
	testCases := []struct {
		name     string
		provider MetricsProviderType
		config   MetricsConfig
		wantErr  bool
	}{
		{
			name:     "opentelemetry_provider",
			provider: MetricsProviderOpenTelemetry,
			config: MetricsConfig{
				Enabled:  true,
				Provider: string(MetricsProviderOpenTelemetry),
				OpenTelemetry: OpenTelemetryMetricsConfig{
					Endpoint:     "localhost:4317",
					ServiceName:  "test-service",
					ExportMode:   OTelExportModePush,
					PushInterval: 30 * time.Second,
				},
			},
			wantErr: false,
		},
		{
			name:     "noop_provider",
			provider: MetricsProviderNoop,
			config: MetricsConfig{
				Enabled:  true,
				Provider: string(MetricsProviderNoop),
			},
			wantErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			err := tc.config.Validate()

			// Assert
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestMetricsConfig_WhenValidateWithLegacyFields_ShouldStillWork(t *testing.T) {
	// Arrange
	config := MetricsConfig{
		Enabled:  true,
		Provider: string(MetricsProviderNoop),
		Path:     "/legacy-metrics",
		Port:     8080,
	}

	// Act
	err := config.Validate()

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, "/legacy-metrics", config.Path)
	assert.Equal(t, 8080, config.Port)
}
