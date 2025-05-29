package otel

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"

	"github.com/zondax/golem/pkg/logger"
)

func TestProvider_getProtocol(t *testing.T) {
	provider := NewProvider()

	t.Run("defaults to HTTP when protocol is empty", func(t *testing.T) {
		config := &logger.OpenTelemetryConfig{
			Protocol: "",
		}

		protocol := provider.getProtocol(config)
		assert.Equal(t, ProtocolHTTP, protocol)
	})

	t.Run("returns configured protocol when specified", func(t *testing.T) {
		testCases := []struct {
			configProtocol   string
			expectedProtocol string
		}{
			{ProtocolHTTP, ProtocolHTTP},
			{ProtocolGRPC, ProtocolGRPC},
		}

		for _, tc := range testCases {
			t.Run("protocol_"+tc.configProtocol, func(t *testing.T) {
				config := &logger.OpenTelemetryConfig{
					Protocol: tc.configProtocol,
				}

				protocol := provider.getProtocol(config)
				assert.Equal(t, tc.expectedProtocol, protocol)
			})
		}
	})
}

func TestProvider_createExporter(t *testing.T) {
	provider := NewProvider()

	t.Run("creates HTTP exporter when protocol is HTTP", func(t *testing.T) {
		config := &logger.OpenTelemetryConfig{
			ServiceName: "test-service",
			Endpoint:    "localhost:4318",
			Protocol:    ProtocolHTTP,
			Insecure:    true,
		}

		exporter, err := provider.createExporter(config)
		require.NoError(t, err)
		require.NotNil(t, exporter)

		// Verify the exporter is the expected type
		// Note: We can't directly assert the type since it's an interface,
		// but we can test that it was created successfully
		assert.IsType(t, &otlploghttp.Exporter{}, exporter)
	})

	t.Run("creates gRPC exporter when protocol is gRPC", func(t *testing.T) {
		config := &logger.OpenTelemetryConfig{
			ServiceName: "test-service",
			Endpoint:    "localhost:4317",
			Protocol:    ProtocolGRPC,
			Insecure:    true,
		}

		exporter, err := provider.createExporter(config)
		require.NoError(t, err)
		require.NotNil(t, exporter)

		// Verify the exporter is the expected type
		assert.IsType(t, &otlploggrpc.Exporter{}, exporter)
	})

	t.Run("creates HTTP exporter when protocol is empty (default)", func(t *testing.T) {
		config := &logger.OpenTelemetryConfig{
			ServiceName: "test-service",
			Endpoint:    "localhost:4318",
			Protocol:    "", // empty protocol should default to HTTP
			Insecure:    true,
		}

		exporter, err := provider.createExporter(config)
		require.NoError(t, err)
		require.NotNil(t, exporter)

		assert.IsType(t, &otlploghttp.Exporter{}, exporter)
	})

	t.Run("returns error for unsupported protocol", func(t *testing.T) {
		config := &logger.OpenTelemetryConfig{
			ServiceName: "test-service",
			Endpoint:    "localhost:4318",
			Protocol:    "websocket", // unsupported protocol
			Insecure:    true,
		}

		exporter, err := provider.createExporter(config)
		assert.Error(t, err)
		assert.Nil(t, exporter)
		assert.Contains(t, err.Error(), "unsupported protocol: websocket")
	})
}

func TestProvider_createHTTPExporter(t *testing.T) {
	provider := NewProvider()

	t.Run("creates HTTP exporter with basic config", func(t *testing.T) {
		config := &logger.OpenTelemetryConfig{
			Endpoint: "localhost:4318",
			Insecure: true,
		}

		exporter, err := provider.createHTTPExporter(config)
		require.NoError(t, err)
		require.NotNil(t, exporter)
		assert.IsType(t, &otlploghttp.Exporter{}, exporter)
	})

	t.Run("creates HTTP exporter with secure connection", func(t *testing.T) {
		config := &logger.OpenTelemetryConfig{
			Endpoint: "example.com:4318",
			Insecure: false,
		}

		exporter, err := provider.createHTTPExporter(config)
		require.NoError(t, err)
		require.NotNil(t, exporter)
		assert.IsType(t, &otlploghttp.Exporter{}, exporter)
	})

	t.Run("creates HTTP exporter with headers", func(t *testing.T) {
		config := &logger.OpenTelemetryConfig{
			Endpoint: "localhost:4318",
			Insecure: true,
			Headers: map[string]string{
				"Authorization": "Bearer test-token",
				"X-Custom":      "value",
				"User-Agent":    "test-agent",
			},
		}

		exporter, err := provider.createHTTPExporter(config)
		require.NoError(t, err)
		require.NotNil(t, exporter)
		assert.IsType(t, &otlploghttp.Exporter{}, exporter)
	})

	t.Run("creates HTTP exporter with empty headers", func(t *testing.T) {
		config := &logger.OpenTelemetryConfig{
			Endpoint: "localhost:4318",
			Insecure: true,
			Headers:  map[string]string{},
		}

		exporter, err := provider.createHTTPExporter(config)
		require.NoError(t, err)
		require.NotNil(t, exporter)
		assert.IsType(t, &otlploghttp.Exporter{}, exporter)
	})

	t.Run("creates HTTP exporter with nil headers", func(t *testing.T) {
		config := &logger.OpenTelemetryConfig{
			Endpoint: "localhost:4318",
			Insecure: true,
			Headers:  nil,
		}

		exporter, err := provider.createHTTPExporter(config)
		require.NoError(t, err)
		require.NotNil(t, exporter)
		assert.IsType(t, &otlploghttp.Exporter{}, exporter)
	})
}

func TestProvider_createGRPCExporter(t *testing.T) {
	provider := NewProvider()

	t.Run("creates gRPC exporter with basic config", func(t *testing.T) {
		config := &logger.OpenTelemetryConfig{
			Endpoint: "localhost:4317",
			Insecure: true,
		}

		exporter, err := provider.createGRPCExporter(config)
		require.NoError(t, err)
		require.NotNil(t, exporter)
		assert.IsType(t, &otlploggrpc.Exporter{}, exporter)
	})

	t.Run("creates gRPC exporter with secure connection", func(t *testing.T) {
		config := &logger.OpenTelemetryConfig{
			Endpoint: "example.com:4317",
			Insecure: false,
		}

		exporter, err := provider.createGRPCExporter(config)
		require.NoError(t, err)
		require.NotNil(t, exporter)
		assert.IsType(t, &otlploggrpc.Exporter{}, exporter)
	})

	t.Run("creates gRPC exporter with headers", func(t *testing.T) {
		config := &logger.OpenTelemetryConfig{
			Endpoint: "localhost:4317",
			Insecure: true,
			Headers: map[string]string{
				"Authorization": "Bearer test-token",
				"X-Custom":      "value",
				"User-Agent":    "test-agent",
			},
		}

		exporter, err := provider.createGRPCExporter(config)
		require.NoError(t, err)
		require.NotNil(t, exporter)
		assert.IsType(t, &otlploggrpc.Exporter{}, exporter)
	})

	t.Run("creates gRPC exporter with empty headers", func(t *testing.T) {
		config := &logger.OpenTelemetryConfig{
			Endpoint: "localhost:4317",
			Insecure: true,
			Headers:  map[string]string{},
		}

		exporter, err := provider.createGRPCExporter(config)
		require.NoError(t, err)
		require.NotNil(t, exporter)
		assert.IsType(t, &otlploggrpc.Exporter{}, exporter)
	})

	t.Run("creates gRPC exporter with nil headers", func(t *testing.T) {
		config := &logger.OpenTelemetryConfig{
			Endpoint: "localhost:4317",
			Insecure: true,
			Headers:  nil,
		}

		exporter, err := provider.createGRPCExporter(config)
		require.NoError(t, err)
		require.NotNil(t, exporter)
		assert.IsType(t, &otlploggrpc.Exporter{}, exporter)
	})
}

func TestExporterCreationEdgeCases(t *testing.T) {
	provider := NewProvider()

	t.Run("HTTP exporter with various endpoint formats", func(t *testing.T) {
		validEndpoints := []string{
			"localhost:4318",
			"127.0.0.1:4318",
			"example.com:4318",
			"otel-collector:4318",
		}

		for _, endpoint := range validEndpoints {
			t.Run("endpoint_"+endpoint, func(t *testing.T) {
				config := &logger.OpenTelemetryConfig{
					Endpoint: endpoint,
					Insecure: true,
				}

				exporter, err := provider.createHTTPExporter(config)
				require.NoError(t, err, "Should create HTTP exporter for endpoint: %s", endpoint)
				require.NotNil(t, exporter)
				assert.IsType(t, &otlploghttp.Exporter{}, exporter)
			})
		}
	})

	t.Run("gRPC exporter with various endpoint formats", func(t *testing.T) {
		validEndpoints := []string{
			"localhost:4317",
			"127.0.0.1:4317",
			"example.com:4317",
			"otel-collector:4317",
		}

		for _, endpoint := range validEndpoints {
			t.Run("endpoint_"+endpoint, func(t *testing.T) {
				config := &logger.OpenTelemetryConfig{
					Endpoint: endpoint,
					Insecure: true,
				}

				exporter, err := provider.createGRPCExporter(config)
				require.NoError(t, err, "Should create gRPC exporter for endpoint: %s", endpoint)
				require.NotNil(t, exporter)
				assert.IsType(t, &otlploggrpc.Exporter{}, exporter)
			})
		}
	})

	t.Run("HTTP exporter with special headers", func(t *testing.T) {
		config := &logger.OpenTelemetryConfig{
			Endpoint: "localhost:4318",
			Insecure: true,
			Headers: map[string]string{
				"Content-Type":    "application/json",
				"Accept-Encoding": "gzip",
				"X-Trace-Id":      "123456789",
				"X-Request-ID":    "req-abc-123",
				"Authorization":   "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9",
			},
		}

		exporter, err := provider.createHTTPExporter(config)
		require.NoError(t, err)
		require.NotNil(t, exporter)
		assert.IsType(t, &otlploghttp.Exporter{}, exporter)
	})

	t.Run("gRPC exporter with special headers", func(t *testing.T) {
		config := &logger.OpenTelemetryConfig{
			Endpoint: "localhost:4317",
			Insecure: true,
			Headers: map[string]string{
				"grpc-timeout":  "30s",
				"user-agent":    "otel-go/1.0.0",
				"authorization": "bearer token",
				"x-request-id":  "req-grpc-123",
			},
		}

		exporter, err := provider.createGRPCExporter(config)
		require.NoError(t, err)
		require.NotNil(t, exporter)
		assert.IsType(t, &otlploggrpc.Exporter{}, exporter)
	})
}

func TestExporterConfigurationConsistency(t *testing.T) {
	provider := NewProvider()

	t.Run("same config produces consistent exporters", func(t *testing.T) {
		config := &logger.OpenTelemetryConfig{
			Endpoint: "localhost:4318",
			Protocol: ProtocolHTTP,
			Insecure: true,
			Headers: map[string]string{
				"Authorization": "Bearer test",
			},
		}

		// Create multiple exporters with the same config
		exporter1, err1 := provider.createExporter(config)
		exporter2, err2 := provider.createExporter(config)

		// Both should succeed
		require.NoError(t, err1)
		require.NoError(t, err2)
		require.NotNil(t, exporter1)
		require.NotNil(t, exporter2)

		// Both should be the same type
		assert.IsType(t, &otlploghttp.Exporter{}, exporter1)
		assert.IsType(t, &otlploghttp.Exporter{}, exporter2)
	})

	t.Run("different protocols produce different exporter types", func(t *testing.T) {
		baseConfig := &logger.OpenTelemetryConfig{
			Endpoint: "localhost:4318",
			Insecure: true,
		}

		// Create HTTP exporter
		httpConfig := *baseConfig
		httpConfig.Protocol = ProtocolHTTP
		httpExporter, err := provider.createExporter(&httpConfig)
		require.NoError(t, err)
		require.NotNil(t, httpExporter)

		// Create gRPC exporter
		grpcConfig := *baseConfig
		grpcConfig.Protocol = ProtocolGRPC
		grpcExporter, err := provider.createExporter(&grpcConfig)
		require.NoError(t, err)
		require.NotNil(t, grpcExporter)

		// They should be different types
		assert.IsType(t, &otlploghttp.Exporter{}, httpExporter)
		assert.IsType(t, &otlploggrpc.Exporter{}, grpcExporter)
	})
}
