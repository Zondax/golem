package zcache

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRemoteConfig_SetAddr(t *testing.T) {
	tests := []struct {
		name     string
		host     string
		port     int
		expected string
	}{
		{"standard localhost", "localhost", 6379, "localhost:6379"},
		{"ip address", "192.168.1.1", 6379, "192.168.1.1:6379"},
		{"ipv6 loopback", "::1", 6379, "[::1]:6379"},
		{"ipv6 full address", "2001:db8::1", 6380, "[2001:db8::1]:6380"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &RemoteConfig{}
			config.SetAddr(tt.host, tt.port)
			assert.Equal(t, tt.expected, config.Addr)
		})
	}
}

func TestRemoteConfig_GetHost(t *testing.T) {
	tests := []struct {
		name     string
		addr     string
		expected string
	}{
		{"standard address", "localhost:6379", "localhost"},
		{"ip address", "192.168.1.1:6379", "192.168.1.1"},
		{"empty address", "", ""},
		{"ipv6 loopback", "[::1]:6379", "::1"},
		{"ipv6 full address", "[2001:db8::1]:6379", "2001:db8::1"},
		{"host only (no port)", "localhost", "localhost"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &RemoteConfig{Addr: tt.addr}
			assert.Equal(t, tt.expected, config.GetHost())
		})
	}
}

func TestRemoteConfig_GetPort(t *testing.T) {
	tests := []struct {
		name     string
		addr     string
		expected int
	}{
		{"standard address", "localhost:6379", 6379},
		{"different port", "localhost:6380", 6380},
		{"empty address", "", 0},
		{"no port", "localhost", 0},
		{"invalid port", "localhost:abc", 0},
		{"ipv6 loopback", "[::1]:6379", 6379},
		{"ipv6 full address", "[2001:db8::1]:6380", 6380},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &RemoteConfig{Addr: tt.addr}
			assert.Equal(t, tt.expected, config.GetPort())
		})
	}
}

func TestRemoteConfig_ToRedisConfig_WithTLS(t *testing.T) {
	config := &RemoteConfig{
		Addr:               "localhost:6379",
		TLSEnabled:         true,
		InsecureSkipVerify: true,
	}

	redisOpts, err := config.ToRedisConfig()
	assert.NoError(t, err)
	assert.NotNil(t, redisOpts.TLSConfig)
	assert.True(t, redisOpts.TLSConfig.InsecureSkipVerify)
}

func TestRemoteConfig_ToRedisConfig_WithoutTLS(t *testing.T) {
	config := &RemoteConfig{
		Addr:       "localhost:6379",
		TLSEnabled: false,
	}

	redisOpts, err := config.ToRedisConfig()
	assert.NoError(t, err)
	assert.Nil(t, redisOpts.TLSConfig)
}

func TestRemoteConfig_ToRedisConfig_InvalidCertPath(t *testing.T) {
	config := &RemoteConfig{
		Addr:        "localhost:6379",
		TLSEnabled:  true,
		TLSCertPath: "/nonexistent/cert.pem",
		TLSKeyPath:  "/nonexistent/key.pem",
	}

	_, err := config.ToRedisConfig()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to load client certificate")
}

func TestRemoteConfig_ToRedisConfig_InvalidCAPath(t *testing.T) {
	config := &RemoteConfig{
		Addr:       "localhost:6379",
		TLSEnabled: true,
		TLSCAPath:  "/nonexistent/ca.pem",
	}

	_, err := config.ToRedisConfig()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read CA certificate")
}

func TestRemoteConfig_ToRedisConfig_PartialMTLSConfig(t *testing.T) {
	tests := []struct {
		name        string
		certPath    string
		keyPath     string
		expectedErr string
	}{
		{
			name:        "cert without key",
			certPath:    "/path/to/cert.pem",
			keyPath:     "",
			expectedErr: "TLSCertPath provided without TLSKeyPath",
		},
		{
			name:        "key without cert",
			certPath:    "",
			keyPath:     "/path/to/key.pem",
			expectedErr: "TLSKeyPath provided without TLSCertPath",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &RemoteConfig{
				Addr:        "localhost:6379",
				TLSEnabled:  true,
				TLSCertPath: tt.certPath,
				TLSKeyPath:  tt.keyPath,
			}

			_, err := config.ToRedisConfig()
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedErr)
		})
	}
}
