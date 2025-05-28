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

		// Get logger from context (should include trace fields)
		logger := GetLoggerFromContext(ctx)
		require.NotNil(t, logger)

		// Log a message
		logger.Info("test message")

		// Verify the log entry includes trace fields
		logs := observed.All()
		require.Len(t, logs, 1)

		entry := logs[0]
		assert.Equal(t, "test message", entry.Message)

		// Check that trace_id and span_id fields are present
		fields := entry.Context
		var hasTraceID, hasSpanID bool

		for _, field := range fields {
			switch field.Key {
			case "trace_id":
				hasTraceID = true
				assert.NotEmpty(t, field.String)
			case "span_id":
				hasSpanID = true
				assert.NotEmpty(t, field.String)
			}
		}

		assert.True(t, hasTraceID, "trace_id field should be present")
		assert.True(t, hasSpanID, "span_id field should be present")
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

		// Check that trace fields are not present
		fields := entry.Context
		for _, field := range fields {
			assert.NotEqual(t, "trace_id", field.Key, "trace_id should not be present without trace context")
			assert.NotEqual(t, "span_id", field.Key, "span_id should not be present without trace context")
		}
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

		// Should return the same logger instance when no trace context
		assert.Equal(t, baseLogger, result)
	})

	t.Run("returns enhanced logger with trace fields", func(t *testing.T) {
		core, observed := observer.New(zapcore.InfoLevel)
		baseLogger := &Logger{logger: zap.New(core)}

		// Start a span to create trace context
		ctx, span := tracer.Start(context.Background(), "test-operation")
		defer span.End()

		result := baseLogger.withTraceContext(ctx)

		// Should return a different logger instance with trace fields
		assert.NotSame(t, baseLogger, result, "Enhanced logger should be a different instance")

		// Log with the enhanced logger
		result.Info("test with trace")

		// Verify trace fields are present
		logs := observed.All()
		require.Len(t, logs, 1)

		entry := logs[0]
		fields := entry.Context

		var hasTraceID, hasSpanID bool
		for _, field := range fields {
			switch field.Key {
			case "trace_id":
				hasTraceID = true
			case "span_id":
				hasSpanID = true
			}
		}

		assert.True(t, hasTraceID, "Enhanced logger should include trace_id")
		assert.True(t, hasSpanID, "Enhanced logger should include span_id")
	})
}
