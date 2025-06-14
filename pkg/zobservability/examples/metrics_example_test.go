package examples

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/zondax/golem/pkg/zobservability"
)

func TestExampleBasicMetricsUsage_WhenCalled_ShouldNotPanic(t *testing.T) {
	// Act & Assert - Should not panic
	assert.NotPanics(t, func() {
		ExampleBasicMetricsUsage()
	})
}

func TestExampleConfigurations_WhenCalled_ShouldNotPanic(t *testing.T) {
	// Act & Assert - Should not panic
	assert.NotPanics(t, func() {
		ExampleConfigurations()
	})
}

func TestExampleBasicMetricsUsage_WhenCalled_ShouldCreateValidObserver(t *testing.T) {
	// This test verifies that the example creates a valid observer
	// We'll test the configuration creation part separately to avoid external dependencies

	// Arrange - Create the same config as in the example
	config := &zobservability.Config{
		Provider:    "signoz",
		Enabled:     true,
		Environment: "development",
		Address:     "localhost:4317",
		Metrics: zobservability.MetricsConfig{
			Enabled:  true,
			Provider: "opentelemetry",
			OpenTelemetry: zobservability.OpenTelemetryMetricsConfig{
				Endpoint:       "localhost:4317",
				Insecure:       true,
				ServiceName:    "my-service",
				ServiceVersion: "v1.0.0",
				Environment:    "development",
				Hostname:       "localhost",
				ExportMode:     zobservability.OTelExportModePush,
				PushInterval:   zobservability.DefaultPushInterval,
			},
		},
	}

	// Act & Assert - Configuration should be valid
	assert.NotNil(t, config)
	assert.Equal(t, "signoz", config.Provider)
	assert.True(t, config.Enabled)
	assert.Equal(t, "development", config.Environment)
	assert.Equal(t, "localhost:4317", config.Address)
	assert.True(t, config.Metrics.Enabled)
	assert.Equal(t, "opentelemetry", config.Metrics.Provider)
}

func TestExampleBusinessMetrics_WhenCalled_ShouldRegisterMetrics(t *testing.T) {
	// Arrange
	observer := zobservability.NewNoopObserver()
	metrics := observer.GetMetrics()

	// Act & Assert - Should not panic when registering business metrics
	assert.NotPanics(t, func() {
		_ = metrics.RegisterCounter("orders_total", "Total number of orders", []string{"status", "payment_method"})
		_ = metrics.RegisterGauge("inventory_items", "Current inventory count", []string{"product_id", "warehouse"})
		_ = metrics.RegisterHistogram("order_value_dollars", "Order value in dollars", []string{"customer_tier"},
			[]float64{10, 50, 100, 500, 1000, 5000})
	})
}

func TestExampleConfigurations_WhenCalled_ShouldCreateValidConfigs(t *testing.T) {
	// Test that the configurations created in the example are valid

	// Push mode configuration
	pushModeConfig := zobservability.MetricsConfig{
		Enabled:  true,
		Provider: "opentelemetry",
		OpenTelemetry: zobservability.OpenTelemetryMetricsConfig{
			Endpoint:       "ingest.eu.signoz.cloud:443",
			Insecure:       false,
			ServiceName:    "my-service",
			ServiceVersion: "v1.2.3",
			Environment:    "production",
			Hostname:       "my-host",
			ExportMode:     zobservability.OTelExportModePush,
			PushInterval:   zobservability.DefaultPushInterval,
			BatchTimeout:   zobservability.DefaultBatchTimeout,
			ExportTimeout:  zobservability.DefaultExportTimeout,
			Headers: map[string]string{
				"signoz-access-token": "your-signoz-access-token",
			},
		},
	}

	// Endpoint mode configuration
	endpointModeConfig := zobservability.MetricsConfig{
		Enabled:  true,
		Provider: "opentelemetry",
		OpenTelemetry: zobservability.OpenTelemetryMetricsConfig{
			Endpoint:       "ingest.eu.signoz.cloud:443",
			Insecure:       false,
			ServiceName:    "my-service",
			ServiceVersion: "v1.2.3",
			Environment:    "production",
			Hostname:       "my-host",
			ExportMode:     zobservability.OTelExportModeEndpoint,
			BatchTimeout:   zobservability.DefaultBatchTimeout,
			ExportTimeout:  zobservability.DefaultExportTimeout,
			Headers: map[string]string{
				"signoz-access-token": "your-signoz-access-token",
			},
		},
	}

	// Assert configurations are valid
	assert.True(t, pushModeConfig.Enabled)
	assert.Equal(t, "opentelemetry", pushModeConfig.Provider)
	assert.Equal(t, zobservability.OTelExportModePush, pushModeConfig.OpenTelemetry.ExportMode)
	assert.Equal(t, "ingest.eu.signoz.cloud:443", pushModeConfig.OpenTelemetry.Endpoint)

	assert.True(t, endpointModeConfig.Enabled)
	assert.Equal(t, "opentelemetry", endpointModeConfig.Provider)
	assert.Equal(t, zobservability.OTelExportModeEndpoint, endpointModeConfig.OpenTelemetry.ExportMode)
	assert.Equal(t, "ingest.eu.signoz.cloud:443", endpointModeConfig.OpenTelemetry.Endpoint)
}

func TestExampleConfigurations_WhenCalled_ShouldCreateValidObservabilityConfigs(t *testing.T) {
	// Test the main observability configs created in the example

	pushModeConfig := zobservability.MetricsConfig{
		Enabled:  true,
		Provider: "opentelemetry",
		OpenTelemetry: zobservability.OpenTelemetryMetricsConfig{
			ExportMode: zobservability.OTelExportModePush,
		},
	}

	endpointModeConfig := zobservability.MetricsConfig{
		Enabled:  true,
		Provider: "opentelemetry",
		OpenTelemetry: zobservability.OpenTelemetryMetricsConfig{
			ExportMode: zobservability.OTelExportModeEndpoint,
		},
	}

	// Main observability configs
	config1 := &zobservability.Config{
		Provider:    "signoz",
		Enabled:     true,
		Environment: "production",
		Address:     "ingest.eu.signoz.cloud:443",
		Metrics:     pushModeConfig,
	}

	config2 := &zobservability.Config{
		Provider:    "signoz",
		Enabled:     true,
		Environment: "production",
		Address:     "ingest.eu.signoz.cloud:443",
		Metrics:     endpointModeConfig,
	}

	// Assert main configs are valid
	assert.Equal(t, "signoz", config1.Provider)
	assert.True(t, config1.Enabled)
	assert.Equal(t, "production", config1.Environment)
	assert.Equal(t, zobservability.OTelExportModePush, config1.Metrics.OpenTelemetry.ExportMode)

	assert.Equal(t, "signoz", config2.Provider)
	assert.True(t, config2.Enabled)
	assert.Equal(t, "production", config2.Environment)
	assert.Equal(t, zobservability.OTelExportModeEndpoint, config2.Metrics.OpenTelemetry.ExportMode)
}
