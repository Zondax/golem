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
