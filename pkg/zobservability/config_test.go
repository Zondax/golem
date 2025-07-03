package zobservability

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig_WhenValidateWithDisabledObservability_ShouldReturnNil(t *testing.T) {
	// Arrange
	config := Config{
		Enabled: false,
	}

	// Act
	err := config.Validate()

	// Assert
	assert.NoError(t, err)
}

func TestConfig_WhenValidateWithEnabledButMissingProvider_ShouldReturnError(t *testing.T) {
	// Arrange
	config := Config{
		Enabled:     true,
		Provider:    "",
		Environment: "production",
		Address:     "http://localhost:8080",
	}

	// Act
	err := config.Validate()

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "observability provider is required when enabled")
}

func TestConfig_WhenValidateWithEnabledButMissingEnvironment_ShouldReturnError(t *testing.T) {
	// Arrange
	config := Config{
		Enabled:     true,
		Provider:    ProviderSentry,
		Environment: "",
		Address:     "http://localhost:8080",
	}

	// Act
	err := config.Validate()

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "observability environment is required when enabled")
}

func TestConfig_WhenValidateWithEnabledButMissingAddress_ShouldReturnError(t *testing.T) {
	// Arrange
	config := Config{
		Enabled:     true,
		Provider:    ProviderSentry,
		Environment: EnvironmentProduction,
		Address:     "",
	}

	// Act
	err := config.Validate()

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "observability address is required when enabled")
}

func TestConfig_WhenValidateWithValidConfiguration_ShouldReturnNil(t *testing.T) {
	// Arrange
	config := Config{
		Enabled:     true,
		Provider:    ProviderSentry,
		Environment: EnvironmentProduction,
		Address:     "http://localhost:8080",
		Metrics:     DefaultMetricsConfig(),
	}

	// Act
	err := config.Validate()

	// Assert
	assert.NoError(t, err)
}

func TestConfig_WhenValidateWithInvalidMetricsConfig_ShouldReturnError(t *testing.T) {
	// Arrange
	config := Config{
		Enabled:     true,
		Provider:    ProviderSentry,
		Environment: EnvironmentProduction,
		Address:     "http://localhost:8080",
		Metrics: MetricsConfig{
			Enabled:  true,
			Provider: "", // Invalid - missing provider when enabled
		},
	}

	// Act
	err := config.Validate()

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid metrics configuration")
}

func TestConfig_WhenSetDefaultsWithEmptyConfig_ShouldSetDefaults(t *testing.T) {
	// Arrange
	config := &Config{}

	// Act
	config.SetDefaults()

	// Assert
	assert.Equal(t, "development", config.Environment)
	assert.Equal(t, 0.1, config.SampleRate)
	assert.True(t, config.Middleware.CaptureErrors)
	assert.NotEmpty(t, config.Metrics.Provider) // Should have default metrics config
}

func TestConfig_WhenSetDefaultsWithPartialConfig_ShouldPreserveExistingValues(t *testing.T) {
	// Arrange
	config := &Config{
		Environment: EnvironmentProduction,
		SampleRate:  0.5,
	}

	// Act
	config.SetDefaults()

	// Assert
	assert.Equal(t, EnvironmentProduction, config.Environment) // Should preserve existing value
	assert.Equal(t, 0.5, config.SampleRate)                    // Should preserve existing value
	assert.True(t, config.Middleware.CaptureErrors)            // Should set default
}

func TestConfig_WhenSetDefaultsWithZeroSampleRate_ShouldSetDefault(t *testing.T) {
	// Arrange
	config := &Config{
		SampleRate: 0,
	}

	// Act
	config.SetDefaults()

	// Assert
	assert.Equal(t, 0.1, config.SampleRate)
}

func TestConfig_WhenSetDefaultsWithNonZeroSampleRate_ShouldPreserveValue(t *testing.T) {
	// Arrange
	config := &Config{
		SampleRate: 0.8,
	}

	// Act
	config.SetDefaults()

	// Assert
	assert.Equal(t, 0.8, config.SampleRate)
}

func TestConfig_WhenSetDefaultsWithExistingMetricsProvider_ShouldPreserveMetrics(t *testing.T) {
	// Arrange
	existingMetrics := MetricsConfig{
		Provider: "custom-provider",
		Enabled:  true,
	}
	config := &Config{
		Metrics: existingMetrics,
	}

	// Act
	config.SetDefaults()

	// Assert
	assert.Equal(t, existingMetrics, config.Metrics) // Should preserve existing metrics config
}

func TestConfig_WhenSetDefaultsWithEmptyMetricsProvider_ShouldSetDefaultMetrics(t *testing.T) {
	// Arrange
	config := &Config{
		Metrics: MetricsConfig{
			Provider: "",
		},
	}

	// Act
	config.SetDefaults()

	// Assert
	assert.NotEmpty(t, config.Metrics.Provider) // Should set default metrics config
	defaultMetrics := DefaultMetricsConfig()
	assert.Equal(t, defaultMetrics, config.Metrics)
}

func TestConfig_WhenValidateWithAllProviders_ShouldWork(t *testing.T) {
	// Arrange
	providers := []string{ProviderSentry, ProviderSigNoz}

	for _, provider := range providers {
		t.Run(provider, func(t *testing.T) {
			config := Config{
				Enabled:     true,
				Provider:    provider,
				Environment: EnvironmentProduction,
				Address:     "http://localhost:8080",
				Metrics:     DefaultMetricsConfig(),
			}

			// Act
			err := config.Validate()

			// Assert
			assert.NoError(t, err)
		})
	}
}

func TestConfig_WhenValidateWithAllEnvironments_ShouldWork(t *testing.T) {
	// Arrange
	environments := []string{
		EnvironmentProduction,
		EnvironmentDevelopment,
		EnvironmentStaging,
		EnvironmentLocal,
	}

	for _, env := range environments {
		t.Run(env, func(t *testing.T) {
			config := Config{
				Enabled:     true,
				Provider:    ProviderSentry,
				Environment: env,
				Address:     "http://localhost:8080",
				Metrics:     DefaultMetricsConfig(),
			}

			// Act
			err := config.Validate()

			// Assert
			assert.NoError(t, err)
		})
	}
}

func TestConfig_WhenCompleteConfiguration_ShouldValidateSuccessfully(t *testing.T) {
	// Arrange
	config := Config{
		Provider:    ProviderSentry,
		Enabled:     true,
		Environment: EnvironmentProduction,
		Release:     "v1.0.0",
		Debug:       false,
		Address:     "https://sentry.example.com",
		SampleRate:  0.1,
		Middleware: MiddlewareConfig{
			CaptureErrors: true,
		},
		Metrics: MetricsConfig{
			Enabled:  true,
			Provider: "noop",
			Path:     "/metrics",
			Port:     9090,
		},
		CustomConfig: map[string]string{
			"custom_key": "custom_value",
		},
	}

	// Act
	err := config.Validate()

	// Assert
	assert.NoError(t, err)
}

func TestMiddlewareConfig_WhenDefaultValues_ShouldBeCorrect(t *testing.T) {
	// Arrange
	config := &Config{}

	// Act
	config.SetDefaults()

	// Assert
	assert.True(t, config.Middleware.CaptureErrors)
}

func TestConfig_WhenSetDefaultsMultipleTimes_ShouldBeIdempotent(t *testing.T) {
	// Arrange
	config := &Config{}

	// Act
	config.SetDefaults()
	firstEnvironment := config.Environment
	firstSampleRate := config.SampleRate
	firstCaptureErrors := config.Middleware.CaptureErrors

	config.SetDefaults() // Call again

	// Assert
	assert.Equal(t, firstEnvironment, config.Environment)
	assert.Equal(t, firstSampleRate, config.SampleRate)
	assert.Equal(t, firstCaptureErrors, config.Middleware.CaptureErrors)
}

// Tests for PropagationConfig

func TestPropagationConfig(t *testing.T) {
	tests := []struct {
		name           string
		config         PropagationConfig
		expectedLength int
		expectedFirst  string
	}{
		{
			name: "single format",
			config: PropagationConfig{
				Formats: []string{PropagationB3},
			},
			expectedLength: 1,
			expectedFirst:  PropagationB3,
		},
		{
			name: "multiple formats",
			config: PropagationConfig{
				Formats: []string{PropagationB3, PropagationW3C, PropagationJaeger},
			},
			expectedLength: 3,
			expectedFirst:  PropagationB3,
		},
		{
			name: "empty formats",
			config: PropagationConfig{
				Formats: []string{},
			},
			expectedLength: 0,
		},
		{
			name: "nil formats",
			config: PropagationConfig{
				Formats: nil,
			},
			expectedLength: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Len(t, tt.config.Formats, tt.expectedLength)
			if tt.expectedLength > 0 {
				assert.Equal(t, tt.expectedFirst, tt.config.Formats[0])
			}
		})
	}
}

func TestConfig_WhenSetDefaultsPropagationFormats_ShouldSetW3CDefault(t *testing.T) {
	tests := []struct {
		name           string
		initialConfig  Config
		expectedLength int
		expectedFirst  string
	}{
		{
			name: "empty propagation formats get W3C default",
			initialConfig: Config{
				Propagation: PropagationConfig{
					Formats: []string{},
				},
			},
			expectedLength: 1,
			expectedFirst:  PropagationW3C,
		},
		{
			name: "nil propagation formats get W3C default",
			initialConfig: Config{
				Propagation: PropagationConfig{
					Formats: nil,
				},
			},
			expectedLength: 1,
			expectedFirst:  PropagationW3C,
		},
		{
			name: "existing formats are preserved",
			initialConfig: Config{
				Propagation: PropagationConfig{
					Formats: []string{PropagationB3, PropagationJaeger},
				},
			},
			expectedLength: 2,
			expectedFirst:  PropagationB3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := tt.initialConfig
			config.SetDefaults()

			assert.Len(t, config.Propagation.Formats, tt.expectedLength)
			if tt.expectedLength > 0 {
				assert.Equal(t, tt.expectedFirst, config.Propagation.Formats[0])
			}
		})
	}
}

func TestConfig_WhenValidateWithPropagation_ShouldSucceed(t *testing.T) {
	tests := []struct {
		name        string
		config      Config
		expectError bool
	}{
		{
			name: "valid config with propagation",
			config: Config{
				Provider:    ProviderSigNoz,
				Enabled:     true,
				Environment: "test",
				Address:     "localhost:4317",
				Propagation: PropagationConfig{
					Formats: []string{PropagationB3},
				},
			},
			expectError: false,
		},
		{
			name: "valid config with multiple propagation formats",
			config: Config{
				Provider:    ProviderSigNoz,
				Enabled:     true,
				Environment: "test",
				Address:     "localhost:4317",
				Propagation: PropagationConfig{
					Formats: []string{PropagationB3, PropagationW3C, PropagationJaeger},
				},
			},
			expectError: false,
		},
		{
			name: "valid config with empty propagation formats",
			config: Config{
				Provider:    ProviderSigNoz,
				Enabled:     true,
				Environment: "test",
				Address:     "localhost:4317",
				Propagation: PropagationConfig{
					Formats: []string{},
				},
			},
			expectError: false,
		},
		{
			name: "disabled config with propagation doesn't validate",
			config: Config{
				Provider:    ProviderSigNoz,
				Enabled:     false,
				Environment: "",
				Address:     "",
				Propagation: PropagationConfig{
					Formats: []string{PropagationB3},
				},
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestConfig_WhenPropagationIntegration_ShouldWorkEndToEnd(t *testing.T) {
	// Test that a full configuration works end-to-end
	config := Config{
		Provider:    ProviderSigNoz,
		Enabled:     true,
		Environment: "test",
		Release:     "1.0.0",
		Address:     "localhost:4317",
		SampleRate:  1.0,
		Propagation: PropagationConfig{
			Formats: []string{PropagationB3, PropagationW3C},
		},
	}

	// Set defaults and validate
	config.SetDefaults()
	err := config.Validate()
	assert.NoError(t, err)

	// Check that propagation formats are preserved
	assert.Len(t, config.Propagation.Formats, 2)
	assert.Equal(t, PropagationB3, config.Propagation.Formats[0])
	assert.Equal(t, PropagationW3C, config.Propagation.Formats[1])
}
