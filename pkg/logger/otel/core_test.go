package otel

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func TestLevelFilterCore(t *testing.T) {
	t.Run("Enabled method respects minimum level", func(t *testing.T) {
		tests := []struct {
			name         string
			minLevel     zapcore.Level
			testLevel    zapcore.Level
			shouldEnable bool
		}{
			{
				name:         "info min level allows info",
				minLevel:     zapcore.InfoLevel,
				testLevel:    zapcore.InfoLevel,
				shouldEnable: true,
			},
			{
				name:         "info min level blocks debug",
				minLevel:     zapcore.InfoLevel,
				testLevel:    zapcore.DebugLevel,
				shouldEnable: false,
			},
			{
				name:         "warn min level allows error",
				minLevel:     zapcore.WarnLevel,
				testLevel:    zapcore.ErrorLevel,
				shouldEnable: true,
			},
			{
				name:         "warn min level blocks info",
				minLevel:     zapcore.WarnLevel,
				testLevel:    zapcore.InfoLevel,
				shouldEnable: false,
			},
			{
				name:         "error min level allows error",
				minLevel:     zapcore.ErrorLevel,
				testLevel:    zapcore.ErrorLevel,
				shouldEnable: true,
			},
			{
				name:         "error min level blocks warn",
				minLevel:     zapcore.ErrorLevel,
				testLevel:    zapcore.WarnLevel,
				shouldEnable: false,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				// Create an observer core that logs everything
				observedZapCore, _ := observer.New(zapcore.DebugLevel)

				// Create level filter core with our minimum level
				filterCore := &levelFilterCore{
					Core:     observedZapCore,
					minLevel: tt.minLevel,
				}

				// Test if the level is enabled
				enabled := filterCore.Enabled(tt.testLevel)
				assert.Equal(t, tt.shouldEnable, enabled)
			})
		}
	})

	t.Run("Enabled method respects wrapped core level", func(t *testing.T) {
		// Create an observer core that only allows warn and above
		observedZapCore, _ := observer.New(zapcore.WarnLevel)

		// Create level filter core with debug minimum level (should be overridden by wrapped core)
		filterCore := &levelFilterCore{
			Core:     observedZapCore,
			minLevel: zapcore.DebugLevel,
		}

		// Even though our minLevel is debug, the wrapped core should block debug and info
		assert.False(t, filterCore.Enabled(zapcore.DebugLevel))
		assert.False(t, filterCore.Enabled(zapcore.InfoLevel))
		assert.True(t, filterCore.Enabled(zapcore.WarnLevel))
		assert.True(t, filterCore.Enabled(zapcore.ErrorLevel))
	})

	t.Run("Check method filters log entries", func(t *testing.T) {
		// Create an observer core
		observedZapCore, observedLogs := observer.New(zapcore.DebugLevel)

		// Create level filter core with info minimum level
		filterCore := &levelFilterCore{
			Core:     observedZapCore,
			minLevel: zapcore.InfoLevel,
		}

		// Create log entries
		debugEntry := zapcore.Entry{
			Level:   zapcore.DebugLevel,
			Message: "debug message",
		}
		infoEntry := zapcore.Entry{
			Level:   zapcore.InfoLevel,
			Message: "info message",
		}

		// Test debug entry (should be filtered out)
		debugChecked := filterCore.Check(debugEntry, nil)
		assert.Nil(t, debugChecked)

		// Test info entry (should pass through)
		infoChecked := filterCore.Check(infoEntry, nil)
		assert.NotNil(t, infoChecked)

		// Log the info entry and verify it was recorded
		if infoChecked != nil {
			infoChecked.Write()
		}

		// Verify only the info message was logged
		assert.Equal(t, 1, observedLogs.Len())
		assert.Equal(t, "info message", observedLogs.All()[0].Message)
	})

	t.Run("With method preserves filtering", func(t *testing.T) {
		// Create an observer core
		observedZapCore, _ := observer.New(zapcore.DebugLevel)

		// Create level filter core with warn minimum level
		filterCore := &levelFilterCore{
			Core:     observedZapCore,
			minLevel: zapcore.WarnLevel,
		}

		// Add fields using With method
		fields := []zapcore.Field{
			zap.String("key", "value"),
			zap.Int("count", 42),
		}
		newCore := filterCore.With(fields)

		// Ensure the returned core is still a levelFilterCore
		newFilterCore, ok := newCore.(*levelFilterCore)
		require.True(t, ok, "With method should return a levelFilterCore")

		// Ensure the minimum level is preserved
		assert.Equal(t, zapcore.WarnLevel, newFilterCore.minLevel)

		// Ensure filtering still works
		assert.False(t, newFilterCore.Enabled(zapcore.InfoLevel))
		assert.True(t, newFilterCore.Enabled(zapcore.WarnLevel))
	})

	t.Run("Sync method delegates to wrapped core", func(t *testing.T) {
		// Create a mock core that tracks sync calls
		mockCore := &mockSyncCore{synced: false}

		// Create level filter core
		filterCore := &levelFilterCore{
			Core:     mockCore,
			minLevel: zapcore.InfoLevel,
		}

		// Call sync
		err := filterCore.Sync()

		// Verify sync was called on wrapped core and no error occurred
		assert.NoError(t, err)
		assert.True(t, mockCore.synced)
	})
}

// mockSyncCore is a mock implementation of zapcore.Core for testing sync functionality
type mockSyncCore struct {
	synced bool
}

func (m *mockSyncCore) Enabled(level zapcore.Level) bool {
	return true
}

func (m *mockSyncCore) With(fields []zapcore.Field) zapcore.Core {
	return m
}

func (m *mockSyncCore) Check(entry zapcore.Entry, checkedEntry *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	return checkedEntry.AddCore(entry, m)
}

func (m *mockSyncCore) Write(entry zapcore.Entry, fields []zapcore.Field) error {
	return nil
}

func (m *mockSyncCore) Sync() error {
	m.synced = true
	return nil
}

func TestLevelFilterCoreIntegration(t *testing.T) {
	t.Run("Integration with real zap logger", func(t *testing.T) {
		// Create an observer core that logs everything
		observedZapCore, observedLogs := observer.New(zapcore.DebugLevel)

		// Create level filter core with info minimum level
		filterCore := &levelFilterCore{
			Core:     observedZapCore,
			minLevel: zapcore.InfoLevel,
		}

		// Create a logger with the filter core
		logger := zap.New(filterCore)

		// Log messages at different levels
		logger.Debug("debug message - should be filtered")
		logger.Info("info message - should pass")
		logger.Warn("warn message - should pass")
		logger.Error("error message - should pass")

		// Verify only info, warn, and error messages were logged
		logs := observedLogs.All()
		assert.Equal(t, 3, len(logs))
		assert.Equal(t, "info message - should pass", logs[0].Message)
		assert.Equal(t, "warn message - should pass", logs[1].Message)
		assert.Equal(t, "error message - should pass", logs[2].Message)
	})
}
