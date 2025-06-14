package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultString(t *testing.T) {
	tests := []struct {
		name         string
		value        string
		defaultValue string
		expected     string
	}{
		{
			name:         "WhenValueIsEmpty_ShouldReturnDefault",
			value:        "",
			defaultValue: "default",
			expected:     "default",
		},
		{
			name:         "WhenValueIsNotEmpty_ShouldReturnValue",
			value:        "actual",
			defaultValue: "default",
			expected:     "actual",
		},
		{
			name:         "WhenValueIsWhitespace_ShouldReturnValue",
			value:        " ",
			defaultValue: "default",
			expected:     " ",
		},
		{
			name:         "WhenValueIsTab_ShouldReturnValue",
			value:        "\t",
			defaultValue: "default",
			expected:     "\t",
		},
		{
			name:         "WhenValueIsNewline_ShouldReturnValue",
			value:        "\n",
			defaultValue: "default",
			expected:     "\n",
		},
		{
			name:         "WhenBothEmpty_ShouldReturnEmptyDefault",
			value:        "",
			defaultValue: "",
			expected:     "",
		},
		{
			name:         "WhenValueIsZero_ShouldReturnValue",
			value:        "0",
			defaultValue: "default",
			expected:     "0",
		},
		{
			name:         "WhenValueIsFalse_ShouldReturnValue",
			value:        "false",
			defaultValue: "default",
			expected:     "false",
		},
		{
			name:         "WhenValueIsUnicode_ShouldReturnValue",
			value:        "🚀",
			defaultValue: "default",
			expected:     "🚀",
		},
		{
			name:         "WhenDefaultIsUnicode_ShouldReturnDefault",
			value:        "",
			defaultValue: "🔥",
			expected:     "🔥",
		},
		{
			name:         "WhenValueIsLongString_ShouldReturnValue",
			value:        "this is a very long string with multiple words and spaces",
			defaultValue: "default",
			expected:     "this is a very long string with multiple words and spaces",
		},
		{
			name:         "WhenDefaultIsLongString_ShouldReturnDefault",
			value:        "",
			defaultValue: "this is a very long default string with multiple words and spaces",
			expected:     "this is a very long default string with multiple words and spaces",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DefaultString(tt.value, tt.defaultValue)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Benchmark tests for performance
func BenchmarkDefaultString_WithValue(b *testing.B) {
	value := "actual_value"
	defaultValue := "default_value"
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		DefaultString(value, defaultValue)
	}
}

func BenchmarkDefaultString_WithEmpty(b *testing.B) {
	value := ""
	defaultValue := "default_value"
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		DefaultString(value, defaultValue)
	}
}

func BenchmarkDefaultString_WithLongStrings(b *testing.B) {
	value := ""
	defaultValue := "this is a very long default string that might be used in real world scenarios with lots of text and content"
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		DefaultString(value, defaultValue)
	}
}
