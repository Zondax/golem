package otel

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/zondax/golem/pkg/logger"
)

// TestHelpers contains helper functions for testing
// Helper function to create a standard logger for testing
func createTestStandardLogger(config logger.Config) *zap.Logger {
	// We need to access the internal function, but since it's not exported,
	// we'll use configureAndBuildLogger with OpenTelemetry disabled
	testConfig := config
	testConfig.OpenTelemetry = nil
	return logger.NewLogger(testConfig).GetZapLogger()
}

// Helper function to check if an error is a connection error
func isConnectionError(err error) bool {
	errStr := err.Error()
	return strings.Contains(errStr, "connection refused") ||
		strings.Contains(errStr, "dial tcp") ||
		strings.Contains(errStr, "no such host")
}

// TestProvider tests the basic provider creation and initialization
func TestNewProvider(t *testing.T) {
	t.Run("creates new provider successfully", func(t *testing.T) {
		provider := NewProvider()

		require.NotNil(t, provider, "expected provider to be created, got nil")
		assert.Nil(t, provider.loggerProvider, "expected loggerProvider to be nil initially")
	})

	t.Run("multiple providers are independent", func(t *testing.T) {
		provider1 := NewProvider()
		provider2 := NewProvider()

		require.NotNil(t, provider1)
		require.NotNil(t, provider2)
		assert.NotSame(t, provider1, provider2, "providers should be different instances")
	})
}

func TestProvider_CreateLogger(t *testing.T) {
	provider := NewProvider()

	t.Run("nil OpenTelemetry config", func(t *testing.T) {
		config := logger.Config{
			Level:         "info",
			Encoding:      "json",
			OpenTelemetry: nil,
		}
		standardLogger := createTestStandardLogger(config)

		_, err := provider.CreateLogger(config, standardLogger)
		if err == nil {
			t.Error("expected error when OpenTelemetry config is nil")
		}

		expectedError := "OpenTelemetry configuration is nil"
		if err.Error() != expectedError {
			t.Errorf("expected error '%s', got '%s'", expectedError, err.Error())
		}
	})

	t.Run("missing service name", func(t *testing.T) {
		config := logger.Config{
			Level:    "info",
			Encoding: "json",
			OpenTelemetry: &logger.OpenTelemetryConfig{
				Enabled:  true,
				Endpoint: "localhost:4318",
				Protocol: ProtocolHTTP,
				Insecure: true,
			},
		}
		standardLogger := createTestStandardLogger(config)

		_, err := provider.CreateLogger(config, standardLogger)
		if err == nil {
			t.Error("expected error when service name is missing")
		}

		expectedError := "service name is required"
		if err.Error() != expectedError {
			t.Errorf("expected error '%s', got '%s'", expectedError, err.Error())
		}
	})

	t.Run("missing endpoint", func(t *testing.T) {
		config := logger.Config{
			Level:    "info",
			Encoding: "json",
			OpenTelemetry: &logger.OpenTelemetryConfig{
				Enabled:     true,
				ServiceName: "test-service",
				Protocol:    ProtocolHTTP,
				Insecure:    true,
			},
		}
		standardLogger := createTestStandardLogger(config)

		_, err := provider.CreateLogger(config, standardLogger)
		if err == nil {
			t.Error("expected error when endpoint is missing")
		}

		expectedError := "endpoint is required"
		if err.Error() != expectedError {
			t.Errorf("expected error '%s', got '%s'", expectedError, err.Error())
		}
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
				Headers: map[string]string{
					"Authorization": "Bearer test-token",
				},
			},
		}
		standardLogger := createTestStandardLogger(config)

		enhancedLogger, err := provider.CreateLogger(config, standardLogger)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if enhancedLogger == nil {
			t.Fatal("expected enhanced logger to be created, got nil")
		}

		if provider.loggerProvider == nil {
			t.Error("expected loggerProvider to be set after creating logger")
		}

		// Test that the enhanced logger can log
		enhancedLogger.Info("test message from enhanced logger")
	})

	t.Run("valid config with gRPC protocol", func(t *testing.T) {
		config := logger.Config{
			Level:    "info",
			Encoding: "json",
			OpenTelemetry: &logger.OpenTelemetryConfig{
				Enabled:     true,
				ServiceName: "test-service-grpc",
				Endpoint:    "localhost:4317",
				Protocol:    ProtocolGRPC,
				Insecure:    true,
			},
		}
		standardLogger := createTestStandardLogger(config)

		enhancedLogger, err := provider.CreateLogger(config, standardLogger)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if enhancedLogger == nil {
			t.Fatal("expected enhanced logger to be created, got nil")
		}

		// Test that the enhanced logger can log
		enhancedLogger.Info("test message from gRPC enhanced logger")
	})

	t.Run("default protocol (should use HTTP)", func(t *testing.T) {
		config := logger.Config{
			Level:    "info",
			Encoding: "json",
			OpenTelemetry: &logger.OpenTelemetryConfig{
				Enabled:     true,
				ServiceName: "test-service-default",
				Endpoint:    "localhost:4318",
				Insecure:    true,
			},
		}
		standardLogger := createTestStandardLogger(config)

		enhancedLogger, err := provider.CreateLogger(config, standardLogger)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if enhancedLogger == nil {
			t.Fatal("expected enhanced logger to be created, got nil")
		}

		// Test that the enhanced logger can log
		enhancedLogger.Info("test message with default protocol")
	})

	t.Run("unsupported protocol", func(t *testing.T) {
		config := logger.Config{
			Level:    "info",
			Encoding: "json",
			OpenTelemetry: &logger.OpenTelemetryConfig{
				Enabled:     true,
				ServiceName: "test-service",
				Endpoint:    "localhost:4317",
				Protocol:    "websocket", // unsupported
				Insecure:    true,
			},
		}
		standardLogger := createTestStandardLogger(config)

		_, err := provider.CreateLogger(config, standardLogger)
		if err == nil {
			t.Error("expected error for unsupported protocol")
		}

		expectedError := "unsupported protocol: websocket"
		if err.Error() != expectedError {
			t.Errorf("expected error '%s', got '%s'", expectedError, err.Error())
		}
	})
}

func TestProvider_Shutdown(t *testing.T) {
	ctx := context.Background()
	provider := NewProvider()

	t.Run("shutdown without logger provider", func(t *testing.T) {
		err := provider.Shutdown(ctx)
		if err != nil {
			t.Errorf("expected no error when shutting down without logger provider, got %v", err)
		}
	})

	t.Run("shutdown with logger provider", func(t *testing.T) {
		// First create a logger to initialize the provider
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
		standardLogger := createTestStandardLogger(config)

		_, err := provider.CreateLogger(config, standardLogger)
		if err != nil {
			t.Fatalf("failed to create logger: %v", err)
		}

		// Now test shutdown
		err = provider.Shutdown(ctx)
		if err != nil {
			t.Errorf("expected no error during shutdown, got %v", err)
		}

		// Verify provider is cleaned up
		if provider.loggerProvider != nil {
			t.Error("expected loggerProvider to be nil after shutdown")
		}
	})

	t.Run("multiple shutdowns should not error", func(t *testing.T) {
		err1 := provider.Shutdown(ctx)
		err2 := provider.Shutdown(ctx)

		if err1 != nil {
			t.Errorf("expected no error on first shutdown, got %v", err1)
		}

		if err2 != nil {
			t.Errorf("expected no error on second shutdown, got %v", err2)
		}
	})
}

func TestProvider_Integration(t *testing.T) {
	t.Run("full integration test", func(t *testing.T) {
		provider := NewProvider()

		// Register the provider
		logger.RegisterOpenTelemetryProvider(provider)
		defer logger.RegisterOpenTelemetryProvider(nil) // cleanup

		// Create config with OpenTelemetry enabled
		config := logger.Config{
			Level:    "info",
			Encoding: "json",
			OpenTelemetry: &logger.OpenTelemetryConfig{
				Enabled:     true,
				ServiceName: "integration-test",
				Endpoint:    "localhost:4318",
				Protocol:    ProtocolHTTP,
				Insecure:    true,
			},
		}

		// This should use the OpenTelemetry provider through configureAndBuildLogger
		enhancedLogger := logger.NewLogger(config).GetZapLogger()

		if enhancedLogger == nil {
			t.Fatal("expected enhanced logger to be created, got nil")
		}

		// Test logging
		enhancedLogger.Info("integration test message")
		enhancedLogger.Warn("integration test warning")
		enhancedLogger.Error("integration test error")

		// Test shutdown - ignore connection errors since we don't have a real OpenTelemetry server
		ctx := context.Background()
		err := logger.ShutdownOpenTelemetryLogger(ctx)
		// Don't fail the test if there's a connection error - this is expected in unit tests
		if err != nil && !isConnectionError(err) {
			t.Errorf("unexpected error during shutdown: %v", err)
		}
	})
}

// Benchmark tests
func BenchmarkProvider_CreateLogger(b *testing.B) {
	provider := NewProvider()
	config := logger.Config{
		Level:    "info",
		Encoding: "json",
		OpenTelemetry: &logger.OpenTelemetryConfig{
			Enabled:     true,
			ServiceName: "benchmark-test",
			Endpoint:    "localhost:4318",
			Protocol:    ProtocolHTTP,
			Insecure:    true,
		},
	}
	standardLogger := createTestStandardLogger(config)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		enhancedLogger, err := provider.CreateLogger(config, standardLogger)
		if err != nil {
			b.Fatalf("unexpected error: %v", err)
		}
		_ = enhancedLogger

		// Clean up for next iteration
		err = provider.Shutdown(context.Background())
		if err != nil {
			b.Fatalf("unexpected error during shutdown: %v", err)
		}
	}
}

func TestProvider_CreateLogger_LevelFiltering(t *testing.T) {
	tests := []struct {
		name         string
		logLevel     string
		messageLevel zapcore.Level
		shouldLog    bool
	}{
		{
			name:         "info level allows info messages",
			logLevel:     "info",
			messageLevel: zapcore.InfoLevel,
			shouldLog:    true,
		},
		{
			name:         "info level blocks debug messages",
			logLevel:     "info",
			messageLevel: zapcore.DebugLevel,
			shouldLog:    false,
		},
		{
			name:         "warn level blocks info messages",
			logLevel:     "warn",
			messageLevel: zapcore.InfoLevel,
			shouldLog:    false,
		},
		{
			name:         "warn level allows warn messages",
			logLevel:     "warn",
			messageLevel: zapcore.WarnLevel,
			shouldLog:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create config with the test log level
			config := logger.Config{
				Level: tt.logLevel,
				OpenTelemetry: &logger.OpenTelemetryConfig{
					ServiceName: "test-service",
					Endpoint:    "localhost:4318",
					Protocol:    ProtocolHTTP,
					Insecure:    true,
				},
			}

			// Create provider
			provider := NewProvider()

			// Create standard logger with the specified level
			standardLogger := createTestStandardLogger(config)

			// Create enhanced logger
			enhancedLogger, err := provider.CreateLogger(config, standardLogger)
			require.NoError(t, err)
			require.NotNil(t, enhancedLogger)

			// Test that the enhanced logger respects the level configuration
			// The Tee core should have the same level behavior as the standard logger
			enabled := enhancedLogger.Core().Enabled(tt.messageLevel)

			if tt.shouldLog {
				assert.True(t, enabled, "Expected level %v to be enabled for log level %s", tt.messageLevel, tt.logLevel)
			} else {
				assert.False(t, enabled, "Expected level %v to be disabled for log level %s", tt.messageLevel, tt.logLevel)
			}

			// Cleanup
			_ = provider.Shutdown(context.Background())
		})
	}
}
