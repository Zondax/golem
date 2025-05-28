package zdbconfig

import (
	"time"

	"gorm.io/gorm"
)

// OpenTelemetry Query Formatter Constants
const (
	// QueryFormatterDefault uses OpenTelemetry's default query formatting
	QueryFormatterDefault = "default"

	// QueryFormatterUpper converts SQL queries to uppercase
	QueryFormatterUpper = "upper"

	// QueryFormatterLower converts SQL queries to lowercase
	QueryFormatterLower = "lower"

	// QueryFormatterNone hides SQL queries for security
	QueryFormatterNone = "none"
)

type ConnectionParams struct {
	User     string
	Password string
	Name     string
	Host     string
	Port     uint
	Params   string
	Protocol string
}

type Config struct {
	RetryInterval    int
	MaxAttempts      int
	ConnectionParams ConnectionParams
	LogConfig        LogConfig
	MaxIdleConns     int
	MaxOpenConns     int
	ConnMaxLifetime  time.Duration

	// OpenTelemetry configuration
	OpenTelemetry OpenTelemetryConfig
}

type LogConfig struct {
	Prefix                    string
	LogLevel                  string
	SlowThreshold             int
	IgnoreRecordNotFoundError bool
	ParameterizedQuery        bool
	Colorful                  bool
}

// OpenTelemetryConfig represents OpenTelemetry instrumentation configuration
type OpenTelemetryConfig struct {
	// Enabled controls whether OpenTelemetry instrumentation is active
	Enabled bool

	// IncludeQueryParameters controls whether SQL query parameters are included in spans
	IncludeQueryParameters bool

	// QueryFormatter controls how SQL queries are formatted in spans
	// Options: QueryFormatterDefault, QueryFormatterUpper, QueryFormatterLower, QueryFormatterNone
	QueryFormatter string

	// DefaultAttributes are custom attributes to add to all database spans
	DefaultAttributes map[string]string

	// DisableMetrics controls whether to disable OpenTelemetry metrics collection
	DisableMetrics bool

	// DBStatsEnabled controls whether to collect database connection pool stats
	DBStatsEnabled bool
}

func BuildGormConfig(logConfig LogConfig) *gorm.Config {
	newLogger := getDBLogger(logConfig)
	return &gorm.Config{Logger: newLogger}
}
