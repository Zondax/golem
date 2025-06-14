package zobservability

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLevel_WhenDebugLevel_ShouldReturnDebugString(t *testing.T) {
	// Arrange
	level := LevelDebug

	// Act
	result := level.String()

	// Assert
	assert.Equal(t, "debug", result)
}

func TestLevel_WhenInfoLevel_ShouldReturnInfoString(t *testing.T) {
	// Arrange
	level := LevelInfo

	// Act
	result := level.String()

	// Assert
	assert.Equal(t, "info", result)
}

func TestLevel_WhenWarningLevel_ShouldReturnWarningString(t *testing.T) {
	// Arrange
	level := LevelWarning

	// Act
	result := level.String()

	// Assert
	assert.Equal(t, "warning", result)
}

func TestLevel_WhenErrorLevel_ShouldReturnErrorString(t *testing.T) {
	// Arrange
	level := LevelError

	// Act
	result := level.String()

	// Assert
	assert.Equal(t, "error", result)
}

func TestLevel_WhenFatalLevel_ShouldReturnFatalString(t *testing.T) {
	// Arrange
	level := LevelFatal

	// Act
	result := level.String()

	// Assert
	assert.Equal(t, "fatal", result)
}

func TestLevel_WhenUnknownLevel_ShouldReturnUnknownString(t *testing.T) {
	// Arrange
	level := Level(999) // Invalid level

	// Act
	result := level.String()

	// Assert
	assert.Equal(t, "unknown", result)
}

func TestLevel_WhenNegativeLevel_ShouldReturnUnknownString(t *testing.T) {
	// Arrange
	level := Level(-1) // Invalid negative level

	// Act
	result := level.String()

	// Assert
	assert.Equal(t, "unknown", result)
}

func TestLevel_WhenAllValidLevels_ShouldReturnCorrectStrings(t *testing.T) {
	// Arrange
	testCases := []struct {
		name     string
		level    Level
		expected string
	}{
		{"debug_level", LevelDebug, "debug"},
		{"info_level", LevelInfo, "info"},
		{"warning_level", LevelWarning, "warning"},
		{"error_level", LevelError, "error"},
		{"fatal_level", LevelFatal, "fatal"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			result := tc.level.String()

			// Assert
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestLevel_WhenLevelConstants_ShouldHaveCorrectValues(t *testing.T) {
	// Assert that level constants have expected integer values
	assert.Equal(t, 0, int(LevelDebug))
	assert.Equal(t, 1, int(LevelInfo))
	assert.Equal(t, 2, int(LevelWarning))
	assert.Equal(t, 3, int(LevelError))
	assert.Equal(t, 4, int(LevelFatal))
}

func TestLevel_WhenComparingLevels_ShouldMaintainOrder(t *testing.T) {
	// Assert that levels maintain their hierarchical order
	assert.True(t, LevelDebug < LevelInfo)
	assert.True(t, LevelInfo < LevelWarning)
	assert.True(t, LevelWarning < LevelError)
	assert.True(t, LevelError < LevelFatal)
}
