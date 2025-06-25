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

	// Cloud SQL specific configuration
	CloudSQL CloudSQLConfig `yaml:"cloud_sql" mapstructure:"cloud_sql"`
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

// CloudSQLConfig represents Cloud SQL specific configuration
type CloudSQLConfig struct {
	// Enabled controls whether to use Cloud SQL connector
	Enabled bool `yaml:"enabled" mapstructure:"enabled"`

	// InstanceName is the Cloud SQL instance connection name (project:region:instance)
	InstanceName string `yaml:"instance_name" mapstructure:"instance_name"`

	// UsePrivateIP controls whether to use private IP for connection
	UsePrivateIP bool `yaml:"use_private_ip" mapstructure:"use_private_ip"`

	// UseIAMAuth controls whether to use IAM authentication
	UseIAMAuth bool `yaml:"use_iam_auth" mapstructure:"use_iam_auth"`

	// CredentialsFile path to service account credentials file
	CredentialsFile string `yaml:"credentials_file" mapstructure:"credentials_file"`

	// RefreshTimeout for connection refresh (optional)
	RefreshTimeout int `yaml:"refresh_timeout" mapstructure:"refresh_timeout"`
}

func BuildGormConfig(logConfig LogConfig) *gorm.Config {
	newLogger := getDBLogger(logConfig)
	return &gorm.Config{Logger: newLogger}
}
