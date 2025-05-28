package logger

import (
	"context"
	"errors"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Mock OpenTelemetry provider for testing
type mockOTelProvider struct {
	shouldFail     bool
	createCalled   bool
	shutdownCalled bool
	config         Config
	standardLogger *zap.Logger
}

func (m *mockOTelProvider) CreateLogger(config Config, standardLogger *zap.Logger) (*zap.Logger, error) {
	m.createCalled = true
	m.config = config
	m.standardLogger = standardLogger

	if m.shouldFail {
		return nil, errors.New("mock OpenTelemetry error")
	}

	// Return a logger with a different core to verify it was called
	core := zapcore.NewNopCore()
	return zap.New(core), nil
}

func (m *mockOTelProvider) Shutdown(ctx context.Context) error {
	m.shutdownCalled = true
	return nil
}

func TestCreateStandardLoggerInternal(t *testing.T) {
	tests := []struct {
		name     string
		config   Config
		wantJSON bool
	}{
		{
			name: "default production config",
			config: Config{
				Level:    "info",
				Encoding: "json",
			},
			wantJSON: true,
		},
		{
			name: "console encoding",
			config: Config{
				Level:    "debug",
				Encoding: "console",
			},
			wantJSON: false,
		},
		{
			name: "invalid level defaults to info",
			config: Config{
				Level:    "invalid",
				Encoding: "json",
			},
			wantJSON: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := createStandardLoggerInternal(tt.config)

			if logger == nil {
				t.Fatal("expected logger to be created, got nil")
			}

			// Test that logger can log without panicking
			logger.Info("test message")
		})
	}
}

func TestConfigureAndBuildLogger_WithoutOpenTelemetry(t *testing.T) {
	tests := []struct {
		name   string
		config Config
	}{
		{
			name: "nil OpenTelemetry config",
			config: Config{
				Level:         "info",
				Encoding:      "json",
				OpenTelemetry: nil,
			},
		},
		{
			name: "disabled OpenTelemetry",
			config: Config{
				Level:    "info",
				Encoding: "json",
				OpenTelemetry: &OpenTelemetryConfig{
					Enabled: false,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := configureAndBuildLogger(tt.config)

			if logger == nil {
				t.Fatal("expected logger to be created, got nil")
			}

			// Should return standard logger
			logger.Info("test message")
		})
	}
}

func TestConfigureAndBuildLogger_WithOpenTelemetry(t *testing.T) {
	// Clean up any existing provider
	RegisterOpenTelemetryProvider(nil)

	t.Run("no provider registered", func(t *testing.T) {
		config := Config{
			Level:    "info",
			Encoding: "json",
			OpenTelemetry: &OpenTelemetryConfig{
				Enabled:     true,
				ServiceName: "test-service",
			},
		}

		logger := configureAndBuildLogger(config)

		if logger == nil {
			t.Fatal("expected logger to be created, got nil")
		}

		// Should return standard logger when no provider is registered
		logger.Info("test message")
	})

	t.Run("provider registered and succeeds", func(t *testing.T) {
		mockProvider := &mockOTelProvider{shouldFail: false}
		RegisterOpenTelemetryProvider(mockProvider)
		defer RegisterOpenTelemetryProvider(nil) // cleanup

		config := Config{
			Level:    "info",
			Encoding: "json",
			OpenTelemetry: &OpenTelemetryConfig{
				Enabled:     true,
				ServiceName: "test-service",
				Endpoint:    "http://localhost:4317",
			},
		}

		logger := configureAndBuildLogger(config)

		if logger == nil {
			t.Fatal("expected logger to be created, got nil")
		}

		if !mockProvider.createCalled {
			t.Error("expected CreateLogger to be called on provider")
		}

		if mockProvider.config.OpenTelemetry.ServiceName != "test-service" {
			t.Errorf("expected service name 'test-service', got %s", mockProvider.config.OpenTelemetry.ServiceName)
		}

		if mockProvider.standardLogger == nil {
			t.Error("expected standard logger to be passed to provider")
		}
	})

	t.Run("provider registered but fails", func(t *testing.T) {
		mockProvider := &mockOTelProvider{shouldFail: true}
		RegisterOpenTelemetryProvider(mockProvider)
		defer RegisterOpenTelemetryProvider(nil) // cleanup

		config := Config{
			Level:    "info",
			Encoding: "json",
			OpenTelemetry: &OpenTelemetryConfig{
				Enabled:     true,
				ServiceName: "test-service",
			},
		}

		// Test the actual function behavior
		logger := configureAndBuildLogger(config)

		if logger == nil {
			t.Fatal("expected logger to be created, got nil")
		}

		if !mockProvider.createCalled {
			t.Error("expected CreateLogger to be called on provider")
		}

		// The function should have logged a warning and returned a fallback logger
		logger.Info("test message after fallback")
	})
}

func TestNewLogger(t *testing.T) {
	t.Run("with config", func(t *testing.T) {
		config := Config{
			Level:    "debug",
			Encoding: "console",
		}

		logger := NewLogger(config)

		if logger == nil {
			t.Fatal("expected logger to be created, got nil")
		}

		if logger.logger == nil {
			t.Fatal("expected internal zap logger to be set")
		}
	})

	t.Run("with fields", func(t *testing.T) {
		field1 := Field{Key: "service", Value: "test"}
		field2 := Field{Key: "version", Value: "1.0.0"}

		logger := NewLogger(field1, field2)

		if logger == nil {
			t.Fatal("expected logger to be created, got nil")
		}

		// Test that logger can log without panicking
		logger.logger.Info("test message")
	})

	t.Run("with config and fields", func(t *testing.T) {
		config := Config{
			Level:    "warn",
			Encoding: "json",
		}
		field := Field{Key: "component", Value: "test"}

		logger := NewLogger(config, field)

		if logger == nil {
			t.Fatal("expected logger to be created, got nil")
		}

		logger.logger.Info("test message")
	})
}

func TestNewNopLogger(t *testing.T) {
	logger := NewNopLogger()

	if logger == nil {
		t.Fatal("expected logger to be created, got nil")
	}

	if logger.logger == nil {
		t.Fatal("expected internal zap logger to be set")
	}

	// Nop logger should not panic when logging
	logger.logger.Info("this should be ignored")
	logger.logger.Error("this should also be ignored")
}

func TestNewDevelopmentLogger(t *testing.T) {
	t.Run("without fields", func(t *testing.T) {
		logger := NewDevelopmentLogger()

		if logger == nil {
			t.Fatal("expected logger to be created, got nil")
		}

		logger.logger.Info("development log message")
	})

	t.Run("with fields", func(t *testing.T) {
		field1 := Field{Key: "env", Value: "development"}
		field2 := Field{Key: "debug", Value: true}

		logger := NewDevelopmentLogger(field1, field2)

		if logger == nil {
			t.Fatal("expected logger to be created, got nil")
		}

		logger.logger.Info("development log message with fields")
	})
}

func TestInitLogger(t *testing.T) {
	// Save original global logger
	originalLogger := zap.L()
	defer zap.ReplaceGlobals(originalLogger)

	config := Config{
		Level:    "debug",
		Encoding: "json",
	}

	InitLogger(config)

	// Verify global logger was set
	globalLogger := zap.L()
	if globalLogger == nil {
		t.Fatal("expected global logger to be set")
	}

	// Test that global logger works
	globalLogger.Info("test global logger")
}

func TestOpenTelemetryProviderRegistration(t *testing.T) {
	// Clean up
	RegisterOpenTelemetryProvider(nil)

	t.Run("register and get provider", func(t *testing.T) {
		mockProvider := &mockOTelProvider{}

		RegisterOpenTelemetryProvider(mockProvider)

		retrieved := getOpenTelemetryProvider()
		if retrieved != mockProvider {
			t.Error("expected to retrieve the same provider that was registered")
		}
	})

	t.Run("register nil provider", func(t *testing.T) {
		RegisterOpenTelemetryProvider(nil)

		retrieved := getOpenTelemetryProvider()
		if retrieved != nil {
			t.Error("expected nil provider after registering nil")
		}
	})
}

func TestShutdownOpenTelemetryLogger(t *testing.T) {
	ctx := context.Background()

	t.Run("no provider registered", func(t *testing.T) {
		RegisterOpenTelemetryProvider(nil)

		err := ShutdownOpenTelemetryLogger(ctx)
		if err != nil {
			t.Errorf("expected no error when no provider is registered, got %v", err)
		}
	})

	t.Run("provider registered", func(t *testing.T) {
		mockProvider := &mockOTelProvider{}
		RegisterOpenTelemetryProvider(mockProvider)
		defer RegisterOpenTelemetryProvider(nil) // cleanup

		err := ShutdownOpenTelemetryLogger(ctx)
		if err != nil {
			t.Errorf("expected no error from mock provider, got %v", err)
		}

		if !mockProvider.shutdownCalled {
			t.Error("expected Shutdown to be called on provider")
		}
	})
}

// Benchmark tests
func BenchmarkCreateStandardLoggerInternal(b *testing.B) {
	config := Config{
		Level:    "info",
		Encoding: "json",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger := createStandardLoggerInternal(config)
		_ = logger
	}
}

func BenchmarkConfigureAndBuildLogger(b *testing.B) {
	config := Config{
		Level:    "info",
		Encoding: "json",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger := configureAndBuildLogger(config)
		_ = logger
	}
}
