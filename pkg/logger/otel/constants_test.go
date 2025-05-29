package otel

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConstants(t *testing.T) {
	t.Run("Protocol constants are defined", func(t *testing.T) {
		// Test that all protocol constants have expected values
		assert.Equal(t, "http", ProtocolHTTP)
		assert.Equal(t, "grpc", ProtocolGRPC)
	})

	t.Run("Protocol constants are distinct", func(t *testing.T) {
		// Ensure protocols are not empty and different from each other
		assert.NotEmpty(t, ProtocolHTTP)
		assert.NotEmpty(t, ProtocolGRPC)
		assert.NotEqual(t, ProtocolHTTP, ProtocolGRPC)
	})

	t.Run("Protocol constants are lowercase", func(t *testing.T) {
		// Ensure consistency in protocol naming
		assert.Equal(t, ProtocolHTTP, "http")
		assert.Equal(t, ProtocolGRPC, "grpc")
	})
}
