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

func TestProvider_buildResourceAttributes(t *testing.T) {
	provider := NewProvider()

	t.Run("builds attributes with all config fields", func(t *testing.T) {
		config := &logger.OpenTelemetryConfig{
			ServiceName:    "test-service",
			ServiceVersion: "2.1.0",
			Environment:    "production",
			Hostname:       "app-server-01",
		}

		attrs, err := provider.buildResourceAttributes(config)
		require.NoError(t, err)

		// Convert to map for easier testing
		attrMap := make(map[string]string)
		for _, attr := range attrs {
			attrMap[string(attr.Key)] = attr.Value.AsString()
		}

		assert.Equal(t, "test-service", attrMap["service.name"])
		assert.Equal(t, "2.1.0", attrMap["service.version"])
		assert.Equal(t, "production", attrMap["deployment.environment"])
		assert.Equal(t, "app-server-01", attrMap["host.name"])
		assert.Len(t, attrs, 4) // Should have exactly 4 attributes
	})

	t.Run("builds attributes with minimal config", func(t *testing.T) {
		config := &logger.OpenTelemetryConfig{
			ServiceName: "minimal-service",
		}

		attrs, err := provider.buildResourceAttributes(config)
		require.NoError(t, err)

		// Convert to map for easier testing
		attrMap := make(map[string]string)
		for _, attr := range attrs {
			attrMap[string(attr.Key)] = attr.Value.AsString()
		}

		assert.Equal(t, "minimal-service", attrMap["service.name"])
		assert.Equal(t, "unknown", attrMap["service.version"]) // Default when not specified
		assert.NotContains(t, attrMap, "deployment.environment")
		assert.NotContains(t, attrMap, "host.name")
		assert.Len(t, attrs, 2) // Should have service.name and service.version only
	})

	t.Run("builds attributes with empty service version", func(t *testing.T) {
		config := &logger.OpenTelemetryConfig{
			ServiceName:    "test-service",
			ServiceVersion: "", // Empty version should default to "unknown"
			Environment:    "development",
		}

		attrs, err := provider.buildResourceAttributes(config)
		require.NoError(t, err)

		// Convert to map for easier testing
		attrMap := make(map[string]string)
		for _, attr := range attrs {
			attrMap[string(attr.Key)] = attr.Value.AsString()
		}

		assert.Equal(t, "test-service", attrMap["service.name"])
		assert.Equal(t, "unknown", attrMap["service.version"])
		assert.Equal(t, "development", attrMap["deployment.environment"])
		assert.Len(t, attrs, 3)
	})

	t.Run("builds attributes with empty optional fields", func(t *testing.T) {
		config := &logger.OpenTelemetryConfig{
			ServiceName:    "test-service",
			ServiceVersion: "1.0.0",
			Environment:    "", // Empty environment should be omitted
			Hostname:       "", // Empty hostname should be omitted
		}

		attrs, err := provider.buildResourceAttributes(config)
		require.NoError(t, err)

		// Convert to map for easier testing
		attrMap := make(map[string]string)
		for _, attr := range attrs {
			attrMap[string(attr.Key)] = attr.Value.AsString()
		}

		assert.Equal(t, "test-service", attrMap["service.name"])
		assert.Equal(t, "1.0.0", attrMap["service.version"])
		assert.NotContains(t, attrMap, "deployment.environment")
		assert.NotContains(t, attrMap, "host.name")
		assert.Len(t, attrs, 2)
	})

	t.Run("builds attributes with special characters", func(t *testing.T) {
		config := &logger.OpenTelemetryConfig{
			ServiceName:    "service-with_special.chars-123",
			ServiceVersion: "1.0.0-beta.1+build.123",
			Environment:    "staging-eu-west-1",
			Hostname:       "server-01.example.com",
		}

		attrs, err := provider.buildResourceAttributes(config)
		require.NoError(t, err)

		// Convert to map for easier testing
		attrMap := make(map[string]string)
		for _, attr := range attrs {
			attrMap[string(attr.Key)] = attr.Value.AsString()
		}

		assert.Equal(t, "service-with_special.chars-123", attrMap["service.name"])
		assert.Equal(t, "1.0.0-beta.1+build.123", attrMap["service.version"])
		assert.Equal(t, "staging-eu-west-1", attrMap["deployment.environment"])
		assert.Equal(t, "server-01.example.com", attrMap["host.name"])
	})

	t.Run("builds attributes with empty service name", func(t *testing.T) {
		config := &logger.OpenTelemetryConfig{
			ServiceName:    "", // Empty service name should be omitted
			ServiceVersion: "1.0.0",
			Environment:    "test",
		}

		attrs, err := provider.buildResourceAttributes(config)
		require.NoError(t, err)

		// Convert to map for easier testing
		attrMap := make(map[string]string)
		for _, attr := range attrs {
			attrMap[string(attr.Key)] = attr.Value.AsString()
		}

		assert.NotContains(t, attrMap, "service.name")
		assert.Equal(t, "1.0.0", attrMap["service.version"])
		assert.Equal(t, "test", attrMap["deployment.environment"])
		assert.Len(t, attrs, 2)
	})
}

func TestProvider_createResource(t *testing.T) {
	provider := NewProvider()

	t.Run("creates resource successfully", func(t *testing.T) {
		config := &logger.OpenTelemetryConfig{
			ServiceName:    "test-service",
			ServiceVersion: "1.0.0",
			Environment:    "production",
			Hostname:       "app-01",
		}

		resource, err := provider.createResource(config)
		require.NoError(t, err)
		require.NotNil(t, resource)

		// Verify that resource has the expected attributes
		attrs := resource.Attributes()
		require.NotNil(t, attrs)
	})

	t.Run("creates resource with minimal config", func(t *testing.T) {
		config := &logger.OpenTelemetryConfig{
			ServiceName: "minimal-service",
		}

		resource, err := provider.createResource(config)
		require.NoError(t, err)
		require.NotNil(t, resource)

		attrs := resource.Attributes()
		require.NotNil(t, attrs)
	})
}
