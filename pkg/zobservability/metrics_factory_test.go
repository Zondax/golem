package zobservability

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const (
	testServiceName = "test-service"
)

// =============================================================================
// METRICS PROVIDER TYPE TESTS
// =============================================================================

func TestMetricsProviderType_WhenConstants_ShouldHaveExpectedValues(t *testing.T) {
	// Assert
	assert.Equal(t, MetricsProviderType("opentelemetry"), MetricsProviderOpenTelemetry)
	assert.Equal(t, MetricsProviderType("noop"), MetricsProviderNoop)
}

func TestMetricsProviderType_WhenStringConversion_ShouldReturnCorrectValues(t *testing.T) {
	// Assert
	assert.Equal(t, "opentelemetry", string(MetricsProviderOpenTelemetry))
	assert.Equal(t, "noop", string(MetricsProviderNoop))
}

// =============================================================================
// NEW METRICS PROVIDER TESTS
// =============================================================================

func TestNewMetricsProvider_WhenDisabledConfig_ShouldReturnNoopProvider(t *testing.T) {
	// Arrange
	name := testServiceName
	config := MetricsConfig{
		Enabled: false,
	}

	// Act
	provider, err := NewMetricsProvider(name, config)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, provider)
	assert.Equal(t, name, provider.Name())
	assert.Implements(t, (*MetricsProvider)(nil), provider)
}

func TestNewMetricsProvider_WhenNoopProvider_ShouldReturnNoopProvider(t *testing.T) {
	// Arrange
	name := testServiceName
	config := MetricsConfig{
		Enabled:  true,
		Provider: string(MetricsProviderNoop),
	}

	// Act
	provider, err := NewMetricsProvider(name, config)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, provider)
	assert.Equal(t, name, provider.Name())
	assert.Implements(t, (*MetricsProvider)(nil), provider)
}

func TestNewMetricsProvider_WhenOpenTelemetryProviderWithValidConfig_ShouldReturnOpenTelemetryProvider(t *testing.T) {
	// Arrange
	name := testServiceName
	config := MetricsConfig{
		Enabled:  true,
		Provider: string(MetricsProviderOpenTelemetry),
		OpenTelemetry: OpenTelemetryMetricsConfig{
			Endpoint:     "localhost:4317",
			ServiceName:  testServiceName,
			ExportMode:   OTelExportModePush,
			PushInterval: 30 * time.Second,
		},
	}

	// Act
	provider, err := NewMetricsProvider(name, config)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, provider)
	assert.Equal(t, name, provider.Name())
	assert.Implements(t, (*MetricsProvider)(nil), provider)
}

func TestNewMetricsProvider_WhenOpenTelemetryProviderWithInvalidConfig_ShouldReturnError(t *testing.T) {
	// Arrange
	name := testServiceName
	config := MetricsConfig{
		Enabled:  true,
		Provider: string(MetricsProviderOpenTelemetry),
		OpenTelemetry: OpenTelemetryMetricsConfig{
			Endpoint:    "", // Invalid: empty endpoint
			ServiceName: testServiceName,
			ExportMode:  OTelExportModePush,
		},
	}

	// Act
	provider, err := NewMetricsProvider(name, config)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, provider)
	assert.Contains(t, err.Error(), "endpoint is required")
}

func TestNewMetricsProvider_WhenUnsupportedProvider_ShouldReturnError(t *testing.T) {
	// Arrange
	name := testServiceName
	config := MetricsConfig{
		Enabled:  true,
		Provider: "unsupported-provider",
	}

	// Act
	provider, err := NewMetricsProvider(name, config)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, provider)
	assert.Contains(t, err.Error(), "unsupported metrics provider")
	assert.Contains(t, err.Error(), "unsupported-provider")
}

func TestNewMetricsProvider_WhenEmptyName_ShouldStillWork(t *testing.T) {
	// Arrange
	name := ""
	config := MetricsConfig{
		Enabled:  true,
		Provider: string(MetricsProviderNoop),
	}

	// Act
	provider, err := NewMetricsProvider(name, config)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, provider)
	assert.Equal(t, name, provider.Name())
}

func TestNewMetricsProvider_WhenCaseInsensitiveProvider_ShouldWork(t *testing.T) {
	testCases := []struct {
		name     string
		provider string
		wantErr  bool
	}{
		{
			name:     "lowercase_opentelemetry",
			provider: "opentelemetry",
			wantErr:  false,
		},
		{
			name:     "uppercase_opentelemetry",
			provider: "OPENTELEMETRY",
			wantErr:  false,
		},
		{
			name:     "mixed_case_opentelemetry",
			provider: "OpenTelemetry",
			wantErr:  false,
		},
		{
			name:     "lowercase_noop",
			provider: "noop",
			wantErr:  false,
		},
		{
			name:     "uppercase_noop",
			provider: "NOOP",
			wantErr:  false,
		},
		{
			name:     "mixed_case_noop",
			provider: "NoOp",
			wantErr:  false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			name := testServiceName
			config := MetricsConfig{
				Enabled:  true,
				Provider: tc.provider,
			}

			// For OpenTelemetry, add valid config
			if tc.provider != "noop" && tc.provider != "NOOP" && tc.provider != "NoOp" {
				config.OpenTelemetry = OpenTelemetryMetricsConfig{
					Endpoint:     "localhost:4317",
					ServiceName:  testServiceName,
					ExportMode:   OTelExportModePush,
					PushInterval: 30 * time.Second,
				}
			}

			// Act
			provider, err := NewMetricsProvider(name, config)

			// Assert
			if tc.wantErr {
				assert.Error(t, err)
				assert.Nil(t, provider)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, provider)
				assert.Equal(t, name, provider.Name())
			}
		})
	}
}

// =============================================================================
// EDGE CASES AND COMPLEX SCENARIOS
// =============================================================================

func TestNewMetricsProvider_WhenComplexOpenTelemetryConfig_ShouldWork(t *testing.T) {
	// Arrange
	name := "complex-service"
	config := MetricsConfig{
		Enabled:  true,
		Provider: string(MetricsProviderOpenTelemetry),
		OpenTelemetry: OpenTelemetryMetricsConfig{
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
			ExportMode:    OTelExportModeEndpoint,
			PushInterval:  60 * time.Second,
			BatchTimeout:  10 * time.Second,
			ExportTimeout: 45 * time.Second,
		},
	}

	// Act
	provider, err := NewMetricsProvider(name, config)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, provider)
	assert.Equal(t, name, provider.Name())
	assert.Implements(t, (*MetricsProvider)(nil), provider)
}

func TestNewMetricsProvider_WhenMultipleInstances_ShouldCreateIndependentProviders(t *testing.T) {
	// Arrange
	config1 := MetricsConfig{
		Enabled:  true,
		Provider: string(MetricsProviderNoop),
	}
	config2 := MetricsConfig{
		Enabled:  true,
		Provider: string(MetricsProviderNoop),
	}

	// Act
	provider1, err1 := NewMetricsProvider("service-1", config1)
	provider2, err2 := NewMetricsProvider("service-2", config2)

	// Assert
	assert.NoError(t, err1)
	assert.NoError(t, err2)
	assert.NotNil(t, provider1)
	assert.NotNil(t, provider2)
	assert.NotEqual(t, provider1, provider2) // Should be different instances
	assert.Equal(t, "service-1", provider1.Name())
	assert.Equal(t, "service-2", provider2.Name())
}

func TestNewMetricsProvider_WhenProviderWithSpecialCharacters_ShouldHandleCorrectly(t *testing.T) {
	// Arrange
	name := "service-with-special-chars_123"
	config := MetricsConfig{
		Enabled:  true,
		Provider: string(MetricsProviderNoop),
	}

	// Act
	provider, err := NewMetricsProvider(name, config)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, provider)
	assert.Equal(t, name, provider.Name())
}

func TestNewMetricsProvider_WhenProviderWithEmptyStringProvider_ShouldReturnError(t *testing.T) {
	// Arrange
	name := testServiceName
	config := MetricsConfig{
		Enabled:  true,
		Provider: "",
	}

	// Act
	provider, err := NewMetricsProvider(name, config)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, provider)
	assert.Contains(t, err.Error(), "unsupported metrics provider")
}

func TestNewMetricsProvider_WhenProviderWithWhitespaceProvider_ShouldReturnError(t *testing.T) {
	// Arrange
	name := testServiceName
	config := MetricsConfig{
		Enabled:  true,
		Provider: "   ",
	}

	// Act
	provider, err := NewMetricsProvider(name, config)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, provider)
	assert.Contains(t, err.Error(), "unsupported metrics provider")
}

func TestNewMetricsProvider_WhenAllSupportedProviders_ShouldWork(t *testing.T) {
	supportedProviders := []struct {
		name     string
		provider MetricsProviderType
		config   MetricsConfig
	}{
		{
			name:     "noop_provider",
			provider: MetricsProviderNoop,
			config: MetricsConfig{
				Enabled:  true,
				Provider: string(MetricsProviderNoop),
			},
		},
		{
			name:     "opentelemetry_provider",
			provider: MetricsProviderOpenTelemetry,
			config: MetricsConfig{
				Enabled:  true,
				Provider: string(MetricsProviderOpenTelemetry),
				OpenTelemetry: OpenTelemetryMetricsConfig{
					Endpoint:     "localhost:4317",
					ServiceName:  testServiceName,
					ExportMode:   OTelExportModePush,
					PushInterval: 30 * time.Second,
				},
			},
		},
	}

	for _, sp := range supportedProviders {
		t.Run(sp.name, func(t *testing.T) {
			// Act
			provider, err := NewMetricsProvider(testServiceName, sp.config)

			// Assert
			assert.NoError(t, err)
			assert.NotNil(t, provider)
			assert.Equal(t, testServiceName, provider.Name())
			assert.Implements(t, (*MetricsProvider)(nil), provider)
		})
	}
}
