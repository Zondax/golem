package otel

import (
	"go.uber.org/zap/zapcore"
)

// levelFilterCore wraps a zapcore.Core and filters logs by level
// This is a simple implementation that respects the minimum level configuration
type levelFilterCore struct {
	zapcore.Core
	minLevel zapcore.Level
}

// Enabled returns true if the given level is at or above the minimum level
func (c *levelFilterCore) Enabled(level zapcore.Level) bool {
	// First check if the wrapped core enables this level
	if !c.Core.Enabled(level) {
		return false
	}
	// Then check our minimum level
	return level >= c.minLevel
}

// Check determines whether the supplied Entry should be logged
func (c *levelFilterCore) Check(entry zapcore.Entry, checkedEntry *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if c.Enabled(entry.Level) {
		return c.Core.Check(entry, checkedEntry)
	}
	return checkedEntry
}

// With adds structured context to the Core
func (c *levelFilterCore) With(fields []zapcore.Field) zapcore.Core {
	return &levelFilterCore{
		Core:     c.Core.With(fields),
		minLevel: c.minLevel,
	}
}
