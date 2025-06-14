package examples

import (
	"context"

	"github.com/zondax/golem/pkg/zobservability"
	"github.com/zondax/golem/pkg/zobservability/factory"
)

// This file demonstrates how to use the observability metrics system

// Example 1: Basic metrics usage with Observer
func ExampleBasicMetricsUsage() {
	// Create observer configuration
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

	// Create observer (this would typically be done in your application startup)
	observer, err := factory.NewObserver(config, "my-service")
	if err != nil {
		panic(err)
	}
	defer func() {
		_ = observer.Close()
	}()

	// Get metrics provider
	metrics := observer.GetMetrics()

	// Register custom metrics
	err = metrics.RegisterCounter("my_custom_counter", "A custom counter metric", []string{"label1", "label2"})
	if err != nil {
		panic(err)
	}

	// Use the metrics with context
	ctx := context.Background()
	labels := map[string]string{
		"label1": "value1",
		"label2": "value2",
	}
	_ = metrics.IncrementCounter(ctx, "my_custom_counter", labels)
}

// Example 2: Configuration examples
func ExampleConfigurations() {
	// OpenTelemetry configuration with SigNoz - Push mode (every 30 seconds)
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
			PushInterval:   zobservability.DefaultPushInterval, // 30 seconds
			BatchTimeout:   zobservability.DefaultBatchTimeout,
			ExportTimeout:  zobservability.DefaultExportTimeout,
			Headers: map[string]string{
				"signoz-access-token": "your-signoz-access-token",
			},
		},
	}

	// OpenTelemetry configuration with endpoint mode
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

	// Use configurations in main observability config
	_ = &zobservability.Config{
		Provider:    "signoz",
		Enabled:     true,
		Environment: "production",
		Address:     "ingest.eu.signoz.cloud:443",
		Metrics:     pushModeConfig, // or endpointModeConfig
	}

	// Example of using endpoint mode config
	_ = &zobservability.Config{
		Provider:    "signoz",
		Enabled:     true,
		Environment: "production",
		Address:     "ingest.eu.signoz.cloud:443",
		Metrics:     endpointModeConfig,
	}
}
