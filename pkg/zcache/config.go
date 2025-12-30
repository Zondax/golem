package zcache

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/dgraph-io/ristretto"
	"github.com/go-redis/redis/v8"
	"github.com/zondax/golem/pkg/logger"
	"github.com/zondax/golem/pkg/metrics"
)

const (
	// Default Ristretto cache config
	DefaultNumCounters = int64(1e7)  // 10M keys
	DefaultMaxCostMB   = int64(1024) // 1GB
	DefaultBufferItems = int64(64)
)

type StatsMetrics struct {
	Enable         bool
	UpdateInterval time.Duration
}

type RemoteConfig struct {
	Network            string
	Addr               string
	Password           string
	DB                 int
	DialTimeout        time.Duration
	ReadTimeout        time.Duration
	WriteTimeout       time.Duration
	PoolSize           int
	MinIdleConns       int
	MaxConnAge         time.Duration
	PoolTimeout        time.Duration
	IdleTimeout        time.Duration
	IdleCheckFrequency time.Duration
	Prefix             string
	Logger             *logger.Logger
	MetricServer       metrics.TaskMetrics
	StatsMetrics       StatsMetrics

	// TLS Configuration
	TLSEnabled         bool   // Enable TLS connection
	TLSCertPath        string // Path to client TLS certificate (optional, for mTLS)
	TLSKeyPath         string // Path to client TLS key (optional, for mTLS)
	TLSCAPath          string // Path to CA certificate (optional)
	InsecureSkipVerify bool   // Skip TLS verification (dev only)
}

type LocalConfig struct {
	Prefix       string
	Logger       *logger.Logger
	MetricServer metrics.TaskMetrics
	StatsMetrics StatsMetrics

	// Add Ristretto cache configuration
	NumCounters int64 `json:"num_counters"` // default: 1e7
	MaxCostMB   int64 `json:"max_cost_mb"`  // in MB, default: 1024 (1GB)
	BufferItems int64 `json:"buffer_items"` // default: 64
}

// SetAddr sets Addr from separate host and port values
func (c *RemoteConfig) SetAddr(host string, port int) {
	c.Addr = fmt.Sprintf("%s:%d", host, port)
}

// GetHost returns the host portion of Addr.
// Properly handles IPv6 addresses (e.g., "[::1]:6379").
func (c *RemoteConfig) GetHost() string {
	if c.Addr == "" {
		return ""
	}
	host, _, err := net.SplitHostPort(c.Addr)
	if err != nil {
		// If SplitHostPort fails, the address might be host-only without port
		return c.Addr
	}
	return host
}

// GetPort returns the port portion of Addr, or 0 if not set or invalid.
// Properly handles IPv6 addresses (e.g., "[::1]:6379").
func (c *RemoteConfig) GetPort() int {
	if c.Addr == "" {
		return 0
	}
	_, portStr, err := net.SplitHostPort(c.Addr)
	if err != nil {
		return 0
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return 0
	}
	return port
}

// buildTLSConfig creates a TLS configuration based on RemoteConfig settings
func (c *RemoteConfig) buildTLSConfig() (*tls.Config, error) {
	if !c.TLSEnabled {
		return nil, nil
	}

	tlsConfig := &tls.Config{
		InsecureSkipVerify: c.InsecureSkipVerify, //nolint:gosec // User explicitly opted in
	}

	// Load CA certificate if provided
	if c.TLSCAPath != "" {
		caCert, err := os.ReadFile(c.TLSCAPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read CA certificate: %w", err)
		}
		caCertPool := x509.NewCertPool()
		if !caCertPool.AppendCertsFromPEM(caCert) {
			return nil, fmt.Errorf("failed to parse CA certificate")
		}
		tlsConfig.RootCAs = caCertPool
	}

	// Load client certificate and key if provided (for mTLS)
	if c.TLSCertPath != "" && c.TLSKeyPath != "" {
		cert, err := tls.LoadX509KeyPair(c.TLSCertPath, c.TLSKeyPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load client certificate: %w", err)
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
	}

	return tlsConfig, nil
}

func (c *RemoteConfig) ToRedisConfig() (*redis.Options, error) {
	tlsConfig, err := c.buildTLSConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to build TLS config: %w", err)
	}

	return &redis.Options{
		Network:            c.Network,
		Addr:               c.Addr,
		Password:           c.Password,
		DB:                 c.DB,
		DialTimeout:        c.DialTimeout,
		ReadTimeout:        c.ReadTimeout,
		WriteTimeout:       c.WriteTimeout,
		PoolSize:           c.PoolSize,
		MinIdleConns:       c.MinIdleConns,
		MaxConnAge:         c.MaxConnAge,
		PoolTimeout:        c.PoolTimeout,
		IdleTimeout:        c.IdleTimeout,
		IdleCheckFrequency: c.IdleCheckFrequency,
		TLSConfig:          tlsConfig,
	}, nil
}

func (c *LocalConfig) ToRistrettoConfig() *ristretto.Config {
	numCounters := c.NumCounters
	if numCounters == 0 {
		numCounters = DefaultNumCounters
	}

	maxCostMB := c.MaxCostMB
	if maxCostMB == 0 {
		maxCostMB = DefaultMaxCostMB
	}
	// Convert MB to bytes
	maxCost := maxCostMB << 20

	bufferItems := c.BufferItems
	if bufferItems == 0 {
		bufferItems = DefaultBufferItems
	}

	return &ristretto.Config{
		NumCounters: numCounters,
		MaxCost:     maxCost,
		BufferItems: bufferItems,
	}
}

type CombinedConfig struct {
	Local              *LocalConfig
	Remote             *RemoteConfig
	GlobalLogger       *logger.Logger
	GlobalPrefix       string
	GlobalMetricServer metrics.TaskMetrics
	GlobalStatsMetrics StatsMetrics
	IsRemoteBestEffort bool
}
