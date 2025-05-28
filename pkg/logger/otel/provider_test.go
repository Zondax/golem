package otel

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/sdk/log"

	"github.com/zondax/golem/pkg/logger"
)

func TestProvider_createLoggerProvider(t *testing.T) {
	provider := NewProvider()

	t.Run("creates logger provider with HTTP exporter", func(t *testing.T) {
		config := &logger.OpenTelemetryConfig{
			ServiceName: "test-service",
			Endpoint:    "localhost:4318",
			Protocol:    ProtocolHTTP,
			Insecure:    true,
		}

		loggerProvider, err := provider.createLoggerProvider(config)
		require.NoError(t, err)
		require.NotNil(t, loggerProvider)

		// Verify it's the correct type
		assert.IsType(t, &log.LoggerProvider{}, loggerProvider)
	})

	t.Run("creates logger provider with gRPC exporter", func(t *testing.T) {
		config := &logger.OpenTelemetryConfig{
			ServiceName: "test-service",
			Endpoint:    "localhost:4317",
			Protocol:    ProtocolGRPC,
			Insecure:    true,
		}

		loggerProvider, err := provider.createLoggerProvider(config)
		require.NoError(t, err)
		require.NotNil(t, loggerProvider)

		// Verify it's the correct type
		assert.IsType(t, &log.LoggerProvider{}, loggerProvider)
	})

	t.Run("creates logger provider with default protocol", func(t *testing.T) {
		config := &logger.OpenTelemetryConfig{
			ServiceName: "test-service",
			Endpoint:    "localhost:4318",
			Protocol:    "", // empty protocol should default to HTTP
			Insecure:    true,
		}

		loggerProvider, err := provider.createLoggerProvider(config)
		require.NoError(t, err)
		require.NotNil(t, loggerProvider)

		// Verify it's the correct type
		assert.IsType(t, &log.LoggerProvider{}, loggerProvider)
	})

	t.Run("creates logger provider with headers", func(t *testing.T) {
		config := &logger.OpenTelemetryConfig{
			ServiceName: "test-service",
			Endpoint:    "localhost:4318",
			Protocol:    ProtocolHTTP,
			Insecure:    true,
			Headers: map[string]string{
				"Authorization": "Bearer test-token",
				"X-Custom":      "value",
			},
		}

		loggerProvider, err := provider.createLoggerProvider(config)
		require.NoError(t, err)
		require.NotNil(t, loggerProvider)

		// Verify it's the correct type
		assert.IsType(t, &log.LoggerProvider{}, loggerProvider)
	})

	t.Run("creates logger provider with secure connection", func(t *testing.T) {
		config := &logger.OpenTelemetryConfig{
			ServiceName: "test-service",
			Endpoint:    "example.com:4318",
			Protocol:    ProtocolHTTP,
			Insecure:    false, // secure connection
		}

		loggerProvider, err := provider.createLoggerProvider(config)
		require.NoError(t, err)
		require.NotNil(t, loggerProvider)

		// Verify it's the correct type
		assert.IsType(t, &log.LoggerProvider{}, loggerProvider)
	})

	t.Run("returns error for unsupported protocol", func(t *testing.T) {
		config := &logger.OpenTelemetryConfig{
			ServiceName: "test-service",
			Endpoint:    "localhost:4318",
			Protocol:    "websocket", // unsupported protocol
			Insecure:    true,
		}

		loggerProvider, err := provider.createLoggerProvider(config)
		assert.Error(t, err)
		assert.Nil(t, loggerProvider)
		assert.Contains(t, err.Error(), "failed to create OTLP log exporter")
		assert.Contains(t, err.Error(), "unsupported protocol: websocket")
	})
}

func TestCreateLoggerProviderEdgeCases(t *testing.T) {
	provider := NewProvider()

	t.Run("creates provider with minimal valid config", func(t *testing.T) {
		config := &logger.OpenTelemetryConfig{
			ServiceName: "minimal-service",
			Endpoint:    "localhost:4318",
			Insecure:    true,
		}

		loggerProvider, err := provider.createLoggerProvider(config)
		require.NoError(t, err)
		require.NotNil(t, loggerProvider)
		assert.IsType(t, &log.LoggerProvider{}, loggerProvider)
	})

	t.Run("creates provider with complex service name", func(t *testing.T) {
		config := &logger.OpenTelemetryConfig{
			ServiceName: "complex-service-name_with.special-chars-123",
			Endpoint:    "localhost:4318",
			Protocol:    ProtocolHTTP,
			Insecure:    true,
		}

		loggerProvider, err := provider.createLoggerProvider(config)
		require.NoError(t, err)
		require.NotNil(t, loggerProvider)
		assert.IsType(t, &log.LoggerProvider{}, loggerProvider)
	})

	t.Run("creates provider with unicode service name", func(t *testing.T) {
		config := &logger.OpenTelemetryConfig{
			ServiceName: "service-ñáéíóú-测试",
			Endpoint:    "localhost:4318",
			Protocol:    ProtocolHTTP,
			Insecure:    true,
		}

		loggerProvider, err := provider.createLoggerProvider(config)
		require.NoError(t, err)
		require.NotNil(t, loggerProvider)
		assert.IsType(t, &log.LoggerProvider{}, loggerProvider)
	})

	t.Run("creates provider with various endpoint formats", func(t *testing.T) {
		validEndpoints := []string{
			"localhost:4318",
			"127.0.0.1:4318",
			"example.com:4318",
			"otel-collector:4318",
			"otel-collector.namespace.svc.cluster.local:4318",
		}

		for _, endpoint := range validEndpoints {
			t.Run("endpoint_"+endpoint, func(t *testing.T) {
				config := &logger.OpenTelemetryConfig{
					ServiceName: "test-service",
					Endpoint:    endpoint,
					Protocol:    ProtocolHTTP,
					Insecure:    true,
				}

				loggerProvider, err := provider.createLoggerProvider(config)
				require.NoError(t, err, "Should create provider for endpoint: %s", endpoint)
				require.NotNil(t, loggerProvider)
				assert.IsType(t, &log.LoggerProvider{}, loggerProvider)
			})
		}
	})

	t.Run("creates provider with many headers", func(t *testing.T) {
		headers := make(map[string]string)
		for i := 0; i < 20; i++ {
			headers[fmt.Sprintf("X-Custom-Header-%d", i)] = fmt.Sprintf("value-%d", i)
		}
		headers["Authorization"] = "Bearer long-token-value-here"
		headers["User-Agent"] = "test-agent/1.0.0"

		config := &logger.OpenTelemetryConfig{
			ServiceName: "test-service",
			Endpoint:    "localhost:4318",
			Protocol:    ProtocolHTTP,
			Insecure:    true,
			Headers:     headers,
		}

		loggerProvider, err := provider.createLoggerProvider(config)
		require.NoError(t, err)
		require.NotNil(t, loggerProvider)
		assert.IsType(t, &log.LoggerProvider{}, loggerProvider)
	})

	t.Run("creates provider multiple times with same config", func(t *testing.T) {
		config := &logger.OpenTelemetryConfig{
			ServiceName: "test-service",
			Endpoint:    "localhost:4318",
			Protocol:    ProtocolHTTP,
			Insecure:    true,
		}

		// Create multiple providers
		provider1, err1 := provider.createLoggerProvider(config)
		provider2, err2 := provider.createLoggerProvider(config)
		provider3, err3 := provider.createLoggerProvider(config)

		// All should succeed
		require.NoError(t, err1)
		require.NoError(t, err2)
		require.NoError(t, err3)

		require.NotNil(t, provider1)
		require.NotNil(t, provider2)
		require.NotNil(t, provider3)

		// All should be the correct type
		assert.IsType(t, &log.LoggerProvider{}, provider1)
		assert.IsType(t, &log.LoggerProvider{}, provider2)
		assert.IsType(t, &log.LoggerProvider{}, provider3)
	})
}

func TestCreateLoggerProviderBothProtocols(t *testing.T) {
	provider := NewProvider()

	t.Run("creates providers for both HTTP and gRPC", func(t *testing.T) {
		// Test HTTP
		httpConfig := &logger.OpenTelemetryConfig{
			ServiceName: "http-service",
			Endpoint:    "localhost:4318",
			Protocol:    ProtocolHTTP,
			Insecure:    true,
		}

		httpProvider, err := provider.createLoggerProvider(httpConfig)
		require.NoError(t, err)
		require.NotNil(t, httpProvider)
		assert.IsType(t, &log.LoggerProvider{}, httpProvider)

		// Test gRPC
		grpcConfig := &logger.OpenTelemetryConfig{
			ServiceName: "grpc-service",
			Endpoint:    "localhost:4317",
			Protocol:    ProtocolGRPC,
			Insecure:    true,
		}

		grpcProvider, err := provider.createLoggerProvider(grpcConfig)
		require.NoError(t, err)
		require.NotNil(t, grpcProvider)
		assert.IsType(t, &log.LoggerProvider{}, grpcProvider)

		// Both should be valid LoggerProvider instances
		// They are different instances but same type
		assert.NotSame(t, httpProvider, grpcProvider, "Providers should be different instances")
	})
}
