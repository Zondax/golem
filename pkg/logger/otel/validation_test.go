package otel

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/zondax/golem/pkg/logger"
)

func TestProvider_validateConfig(t *testing.T) {
	provider := NewProvider()

	t.Run("nil OpenTelemetry config", func(t *testing.T) {
		config := logger.Config{
			Level:         "info",
			Encoding:      "json",
			OpenTelemetry: nil,
		}

		err := provider.validateConfig(config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "OpenTelemetry configuration is nil")
	})

	t.Run("missing service name", func(t *testing.T) {
		config := logger.Config{
			Level:    "info",
			Encoding: "json",
			OpenTelemetry: &logger.OpenTelemetryConfig{
				Enabled:  true,
				Endpoint: "localhost:4318",
				Protocol: ProtocolHTTP,
			},
		}

		err := provider.validateConfig(config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "service name is required")
	})

	t.Run("empty service name", func(t *testing.T) {
		config := logger.Config{
			Level:    "info",
			Encoding: "json",
			OpenTelemetry: &logger.OpenTelemetryConfig{
				Enabled:     true,
				ServiceName: "",
				Endpoint:    "localhost:4318",
				Protocol:    ProtocolHTTP,
			},
		}

		err := provider.validateConfig(config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "service name is required")
	})

	t.Run("missing endpoint", func(t *testing.T) {
		config := logger.Config{
			Level:    "info",
			Encoding: "json",
			OpenTelemetry: &logger.OpenTelemetryConfig{
				Enabled:     true,
				ServiceName: "test-service",
				Protocol:    ProtocolHTTP,
			},
		}

		err := provider.validateConfig(config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "endpoint is required")
	})

	t.Run("empty endpoint", func(t *testing.T) {
		config := logger.Config{
			Level:    "info",
			Encoding: "json",
			OpenTelemetry: &logger.OpenTelemetryConfig{
				Enabled:     true,
				ServiceName: "test-service",
				Endpoint:    "",
				Protocol:    ProtocolHTTP,
			},
		}

		err := provider.validateConfig(config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "endpoint is required")
	})

	t.Run("unsupported protocol", func(t *testing.T) {
		config := logger.Config{
			Level:    "info",
			Encoding: "json",
			OpenTelemetry: &logger.OpenTelemetryConfig{
				Enabled:     true,
				ServiceName: "test-service",
				Endpoint:    "localhost:4318",
				Protocol:    "websocket", // unsupported protocol
			},
		}

		err := provider.validateConfig(config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported protocol: websocket")
	})

	t.Run("valid config with HTTP protocol", func(t *testing.T) {
		config := logger.Config{
			Level:    "info",
			Encoding: "json",
			OpenTelemetry: &logger.OpenTelemetryConfig{
				Enabled:     true,
				ServiceName: "test-service",
				Endpoint:    "localhost:4318",
				Protocol:    ProtocolHTTP,
				Insecure:    true,
			},
		}

		err := provider.validateConfig(config)
		assert.NoError(t, err)
	})

	t.Run("valid config with gRPC protocol", func(t *testing.T) {
		config := logger.Config{
			Level:    "info",
			Encoding: "json",
			OpenTelemetry: &logger.OpenTelemetryConfig{
				Enabled:     true,
				ServiceName: "test-service",
				Endpoint:    "localhost:4317",
				Protocol:    ProtocolGRPC,
				Insecure:    true,
			},
		}

		err := provider.validateConfig(config)
		assert.NoError(t, err)
	})

	t.Run("valid config with empty protocol (should default)", func(t *testing.T) {
		config := logger.Config{
			Level:    "info",
			Encoding: "json",
			OpenTelemetry: &logger.OpenTelemetryConfig{
				Enabled:     true,
				ServiceName: "test-service",
				Endpoint:    "localhost:4318",
				Protocol:    "", // empty protocol should be valid (defaults to HTTP)
				Insecure:    true,
			},
		}

		err := provider.validateConfig(config)
		assert.NoError(t, err)
	})

	t.Run("valid config with headers", func(t *testing.T) {
		config := logger.Config{
			Level:    "info",
			Encoding: "json",
			OpenTelemetry: &logger.OpenTelemetryConfig{
				Enabled:     true,
				ServiceName: "test-service",
				Endpoint:    "localhost:4318",
				Protocol:    ProtocolHTTP,
				Insecure:    true,
				Headers: map[string]string{
					"Authorization": "Bearer token",
					"X-Custom":      "value",
				},
			},
		}

		err := provider.validateConfig(config)
		assert.NoError(t, err)
	})
}

func TestProvider_isSupportedProtocol(t *testing.T) {
	provider := NewProvider()

	t.Run("supported protocols", func(t *testing.T) {
		supportedProtocols := []string{
			ProtocolHTTP,
			ProtocolGRPC,
		}

		for _, protocol := range supportedProtocols {
			t.Run(protocol, func(t *testing.T) {
				assert.True(t, provider.isSupportedProtocol(protocol))
			})
		}
	})

	t.Run("unsupported protocols", func(t *testing.T) {
		unsupportedProtocols := []string{
			"websocket",
			"tcp",
			"udp",
			"mqtt",
			"amqp",
			"kafka",
			"HTTP", // uppercase should not be supported
			"GRPC", // uppercase should not be supported
			"Http", // mixed case should not be supported
			"",     // empty string should not be supported
		}

		for _, protocol := range unsupportedProtocols {
			t.Run(protocol+"_should_be_unsupported", func(t *testing.T) {
				assert.False(t, provider.isSupportedProtocol(protocol))
			})
		}
	})

	t.Run("case sensitivity", func(t *testing.T) {
		// Test that protocol validation is case-sensitive
		testCases := []struct {
			protocol  string
			supported bool
		}{
			{ProtocolHTTP, true}, // "http"
			{ProtocolGRPC, true}, // "grpc"
			{"HTTP", false},      // uppercase
			{"GRPC", false},      // uppercase
			{"Http", false},      // mixed case
			{"Grpc", false},      // mixed case
			{"hTTP", false},      // mixed case
			{"gRPC", false},      // mixed case
		}

		for _, tc := range testCases {
			t.Run("protocol_"+tc.protocol, func(t *testing.T) {
				result := provider.isSupportedProtocol(tc.protocol)
				assert.Equal(t, tc.supported, result,
					"Protocol %q should have supported=%v", tc.protocol, tc.supported)
			})
		}
	})
}

func TestValidationEdgeCases(t *testing.T) {
	provider := NewProvider()

	t.Run("config with special characters in service name", func(t *testing.T) {
		config := logger.Config{
			OpenTelemetry: &logger.OpenTelemetryConfig{
				ServiceName: "test-service_with.special-chars",
				Endpoint:    "localhost:4318",
				Protocol:    ProtocolHTTP,
			},
		}

		err := provider.validateConfig(config)
		assert.NoError(t, err, "Service names with special characters should be valid")
	})

	t.Run("config with unicode in service name", func(t *testing.T) {
		config := logger.Config{
			OpenTelemetry: &logger.OpenTelemetryConfig{
				ServiceName: "test-service-ñáéíóú",
				Endpoint:    "localhost:4318",
				Protocol:    ProtocolHTTP,
			},
		}

		err := provider.validateConfig(config)
		assert.NoError(t, err, "Service names with unicode should be valid")
	})

	t.Run("config with long service name", func(t *testing.T) {
		longServiceName := "test-service-with-a-very-long-name-that-exceeds-normal-expectations-for-service-names-but-should-still-be-valid"
		config := logger.Config{
			OpenTelemetry: &logger.OpenTelemetryConfig{
				ServiceName: longServiceName,
				Endpoint:    "localhost:4318",
				Protocol:    ProtocolHTTP,
			},
		}

		err := provider.validateConfig(config)
		assert.NoError(t, err, "Long service names should be valid")
	})

	t.Run("config with whitespace in service name", func(t *testing.T) {
		config := logger.Config{
			OpenTelemetry: &logger.OpenTelemetryConfig{
				ServiceName: "  test service  ",
				Endpoint:    "localhost:4318",
				Protocol:    ProtocolHTTP,
			},
		}

		err := provider.validateConfig(config)
		assert.NoError(t, err, "Service names with whitespace should be valid")
	})

	t.Run("config with various endpoint formats", func(t *testing.T) {
		validEndpoints := []string{
			"localhost:4318",
			"127.0.0.1:4318",
			"https://example.com:4318",
			"http://localhost:4318",
			"otel-collector:4318",
			"otel-collector.namespace.svc.cluster.local:4318",
		}

		for _, endpoint := range validEndpoints {
			t.Run("endpoint_"+endpoint, func(t *testing.T) {
				config := logger.Config{
					OpenTelemetry: &logger.OpenTelemetryConfig{
						ServiceName: "test-service",
						Endpoint:    endpoint,
						Protocol:    ProtocolHTTP,
					},
				}

				err := provider.validateConfig(config)
				assert.NoError(t, err, "Endpoint %q should be valid", endpoint)
			})
		}
	})
}
