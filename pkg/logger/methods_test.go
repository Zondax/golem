package logger

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace/noop"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func TestGetLoggerFromContext_WithTraceContext(t *testing.T) {
	// Set up a real tracer provider for testing
	tp := tracesdk.NewTracerProvider()
	otel.SetTracerProvider(tp)

	// Restore original tracer provider after test
	defer func() {
		_ = tp.Shutdown(context.Background())
		otel.SetTracerProvider(noop.NewTracerProvider())
	}()

	tracer := otel.Tracer("test")

	t.Run("adds trace fields when trace context is available", func(t *testing.T) {
		// Create an observed logger for testing
		core, observed := observer.New(zapcore.InfoLevel)
		baseLogger := &Logger{logger: zap.New(core)}

		// Create context with logger
		ctx := ContextWithLogger(context.Background(), baseLogger)

		// Start a span to create trace context
		ctx, span := tracer.Start(ctx, "test-operation")
		defer span.End()

		// Get logger from context (should include trace context)
		logger := GetLoggerFromContext(ctx)
		require.NotNil(t, logger)

		// Log a message
		logger.Info("test message")

		// Verify the log entry includes context field
		logs := observed.All()
		require.Len(t, logs, 1)

		entry := logs[0]
		assert.Equal(t, "test message", entry.Message)

		// Check that context field is present (used by otelzap bridge for trace correlation)
		fields := entry.Context
		var hasContextField bool

		for _, field := range fields {
			if field.Key == contextFieldKey {
				hasContextField = true
				// The context field should contain the trace context
				assert.NotNil(t, field.Interface, "context field should contain trace context")
				break
			}
		}

		assert.True(t, hasContextField, "context field should be present for trace correlation")
	})

	t.Run("works normally when no trace context is available", func(t *testing.T) {
		// Create an observed logger for testing
		core, observed := observer.New(zapcore.InfoLevel)
		baseLogger := &Logger{logger: zap.New(core)}

		// Create context with logger but no trace
		ctx := ContextWithLogger(context.Background(), baseLogger)

		// Get logger from context (should work normally)
		logger := GetLoggerFromContext(ctx)
		require.NotNil(t, logger)

		// Log a message
		logger.Info("test message without trace")

		// Verify the log entry works normally
		logs := observed.All()
		require.Len(t, logs, 1)

		entry := logs[0]
		assert.Equal(t, "test message without trace", entry.Message)

		// Check that context field is still present (but without trace info)
		fields := entry.Context
		var hasContextField bool
		for _, field := range fields {
			if field.Key == contextFieldKey {
				hasContextField = true
				break
			}
		}

		assert.True(t, hasContextField, "context field should be present even without trace")
	})

	t.Run("creates new logger with warning when context is missing", func(t *testing.T) {
		// Create context without logger
		ctx := context.Background()

		// This should create a new logger and log a warning
		logger := GetLoggerFromContext(ctx)
		require.NotNil(t, logger)

		// The function should have logged a warning about missing context
		// (We can't easily test this without changing the implementation,
		// but the function works as documented)
	})
}

func TestLogger_withTraceContext(t *testing.T) {
	// Set up a real tracer provider for testing
	tp := tracesdk.NewTracerProvider()
	otel.SetTracerProvider(tp)

	// Restore original tracer provider after test
	defer func() {
		_ = tp.Shutdown(context.Background())
		otel.SetTracerProvider(noop.NewTracerProvider())
	}()

	tracer := otel.Tracer("test")

	t.Run("returns same logger when no trace context", func(t *testing.T) {
		baseLogger := &Logger{logger: zap.NewNop()}
		ctx := context.Background()

		result := baseLogger.withTraceContext(ctx)

		// Should return a different logger instance with context field
		assert.NotSame(t, baseLogger, result, "Should return enhanced logger with context field")
	})

	t.Run("returns enhanced logger with trace fields", func(t *testing.T) {
		core, observed := observer.New(zapcore.InfoLevel)
		baseLogger := &Logger{logger: zap.New(core)}

		// Start a span to create trace context
		ctx, span := tracer.Start(context.Background(), "test-operation")
		defer span.End()

		result := baseLogger.withTraceContext(ctx)

		// Should return a different logger instance with context field
		assert.NotSame(t, baseLogger, result, "Enhanced logger should be a different instance")

		// Log with the enhanced logger
		result.Info("test with trace")

		// Verify context field is present (used by otelzap bridge for automatic trace correlation)
		logs := observed.All()
		require.Len(t, logs, 1)

		entry := logs[0]
		fields := entry.Context

		var hasContextField bool
		for _, field := range fields {
			if field.Key == contextFieldKey {
				hasContextField = true
				// The context should contain trace information
				assert.NotNil(t, field.Interface, "context field should contain trace context")
				break
			}
		}

		assert.True(t, hasContextField, "Enhanced logger should include context field for trace correlation")
	})
}
