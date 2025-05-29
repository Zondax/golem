package otel

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/sdk/log"

	"github.com/zondax/golem/pkg/logger"
)

func TestProvider_createLoggerProvider(t *testing.T) {
	provider := NewProvider()

	t.Run("creates provider successfully with HTTP", func(t *testing.T) {
		config := &logger.OpenTelemetryConfig{
			ServiceName: "test-service",
			Endpoint:    "localhost:4318",
			Protocol:    ProtocolHTTP,
			Insecure:    true,
		}

		loggerProvider, err := provider.createLoggerProvider(config)
		require.NoError(t, err)
		require.NotNil(t, loggerProvider)
		assert.IsType(t, &log.LoggerProvider{}, loggerProvider)
	})

	t.Run("creates provider successfully with gRPC", func(t *testing.T) {
		config := &logger.OpenTelemetryConfig{
			ServiceName: "test-service",
			Endpoint:    "localhost:4317",
			Protocol:    ProtocolGRPC,
			Insecure:    true,
		}

		loggerProvider, err := provider.createLoggerProvider(config)
		require.NoError(t, err)
		require.NotNil(t, loggerProvider)
		assert.IsType(t, &log.LoggerProvider{}, loggerProvider)
	})

	t.Run("fails with invalid protocol", func(t *testing.T) {
		config := &logger.OpenTelemetryConfig{
			ServiceName: "test-service",
			Endpoint:    "localhost:4318",
			Protocol:    "invalid",
			Insecure:    true,
		}

		loggerProvider, err := provider.createLoggerProvider(config)
		require.Error(t, err)
		assert.Nil(t, loggerProvider)
		assert.Contains(t, err.Error(), "unsupported protocol")
	})

	t.Run("fails with empty service name", func(t *testing.T) {
		config := &logger.OpenTelemetryConfig{
			ServiceName: "", // Empty service name should cause error
			Endpoint:    "localhost:4318",
			Protocol:    ProtocolHTTP,
			Insecure:    true,
		}

		loggerProvider, err := provider.createLoggerProvider(config)
		require.Error(t, err)
		assert.Nil(t, loggerProvider)
		assert.Contains(t, err.Error(), "service name is required")
	})
}

func TestCreateLoggerProviderEdgeCases(t *testing.T) {
	provider := NewProvider()

	t.Run("handles empty endpoint", func(t *testing.T) {
		config := &logger.OpenTelemetryConfig{
			ServiceName: "test-service",
			Endpoint:    "",
			Protocol:    ProtocolHTTP,
			Insecure:    true,
		}

		loggerProvider, err := provider.createLoggerProvider(config)
		require.Error(t, err)
		assert.Nil(t, loggerProvider)
	})

	t.Run("handles nil config", func(t *testing.T) {
		loggerProvider, err := provider.createLoggerProvider(nil)
		require.Error(t, err)
		assert.Nil(t, loggerProvider)
	})

	t.Run("handles config with headers", func(t *testing.T) {
		config := &logger.OpenTelemetryConfig{
			ServiceName: "test-service",
			Endpoint:    "localhost:4318",
			Protocol:    ProtocolHTTP,
			Insecure:    true,
			Headers: map[string]string{
				"Authorization": "Bearer token",
				"X-API-Key":     "api-key",
			},
		}

		loggerProvider, err := provider.createLoggerProvider(config)
		require.NoError(t, err)
		require.NotNil(t, loggerProvider)
		assert.IsType(t, &log.LoggerProvider{}, loggerProvider)
	})

	t.Run("handles secure connection", func(t *testing.T) {
		config := &logger.OpenTelemetryConfig{
			ServiceName: "test-service",
			Endpoint:    "otel-collector.example.com:4318",
			Protocol:    ProtocolHTTP,
			Insecure:    false,
		}

		loggerProvider, err := provider.createLoggerProvider(config)
		require.NoError(t, err)
		require.NotNil(t, loggerProvider)
		assert.IsType(t, &log.LoggerProvider{}, loggerProvider)
	})

	t.Run("handles all optional fields", func(t *testing.T) {
		config := &logger.OpenTelemetryConfig{
			ServiceName:    "full-service",
			ServiceVersion: "1.2.3",
			Environment:    "production",
			Hostname:       "server-01",
			Endpoint:       "localhost:4318",
			Protocol:       ProtocolHTTP,
			Insecure:       true,
			Headers: map[string]string{
				"Custom-Header": "custom-value",
			},
		}

		loggerProvider, err := provider.createLoggerProvider(config)
		require.NoError(t, err)
		require.NotNil(t, loggerProvider)
		assert.IsType(t, &log.LoggerProvider{}, loggerProvider)
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

	t.Run("returns error with empty service name", func(t *testing.T) {
		config := &logger.OpenTelemetryConfig{
			ServiceName:    "", // Empty service name should return error
			ServiceVersion: "1.0.0",
			Environment:    "test",
		}

		attrs, err := provider.buildResourceAttributes(config)
		require.Error(t, err)
		assert.Nil(t, attrs)
		assert.Contains(t, err.Error(), "service name is required")
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

	t.Run("returns error with empty service name", func(t *testing.T) {
		config := &logger.OpenTelemetryConfig{
			ServiceName: "", // Empty service name should return error
		}

		resource, err := provider.createResource(config)
		require.Error(t, err)
		assert.Nil(t, resource)
		assert.Contains(t, err.Error(), "service name is required")
	})
}
