package providers

import (
	"context"
	"testing"
)

func TestGcpProvider_IsSecretKey(t *testing.T) {
	provider := GcpProvider{}

	tests := []struct {
		name     string
		key      string
		expected bool
	}{
		// Top-level keys (direct matches)
		{
			name:     "simple key with prefix",
			key:      "gcp_secret",
			expected: true,
		},
		{
			name:     "simple key with prefix and suffix",
			key:      "gcp_secret_value",
			expected: true,
		},

		// Nested keys (with dots)
		{
			name:     "nested key with prefix in last segment",
			key:      "database.gcp_secret",
			expected: true,
		},
		{
			name:     "deeply nested key with prefix in last segment",
			key:      "config.database.credentials.gcp_password",
			expected: true,
		},

		// Non-matching keys
		{
			name:     "key without prefix",
			key:      "secret",
			expected: false,
		},
		{
			name:     "nested key without prefix in last segment",
			key:      "database.secret",
			expected: false,
		},
		{
			name:     "nested key with prefix in middle segment",
			key:      "gcp_section.password",
			expected: false,
		},
		{
			name:     "empty key",
			key:      "",
			expected: false,
		},

		// Edge cases
		{
			name:     "key with prefix at wrong position",
			key:      "database.gcpsecret", // missing underscore
			expected: false,
		},
		{
			name:     "just prefix",
			key:      "gcp_",
			expected: true,
		},
		{
			name:     "key with dot at the end",
			key:      "database.gcp_secret.",
			expected: false, // last segment is empty
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := provider.IsSecretKey(context.Background(), tc.key)
			if result != tc.expected {
				t.Errorf("IsSecretKey(%q) = %v, want %v", tc.key, result, tc.expected)
			}
		})
	}
}

// Test to ensure we're safely handling edge cases like empty strings or keys with multiple dots
func TestGcpProvider_IsSecretKey_EdgeCases(t *testing.T) {
	provider := GcpProvider{}

	// Test with a series of dots
	result := provider.IsSecretKey(context.Background(), "...")
	if result {
		t.Error("Expected '...' to not be detected as a secret key")
	}

	// Test with only dots and prefix
	result = provider.IsSecretKey(context.Background(), "..gcp_")
	if !result {
		t.Error("Expected '..gcp_' to be detected as a secret key (last segment is 'gcp_')")
	}
}
