package zdbconnector

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	"cloud.google.com/go/cloudsqlconn"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/zondax/golem/pkg/zdb/zdbconfig"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const (
	CloudSQLDriverName = "pgx"

	// DSN parameter keys
	dsnKeyUser     = "user"
	dsnKeyDatabase = "database"
	dsnKeyPassword = "password"
	dsnKeySSLMode  = "sslmode"
)

type CloudSQLPostgresConnector struct{}

func (c *CloudSQLPostgresConnector) Connect(config *zdbconfig.Config) (*gorm.DB, error) {
	if !config.ConnectionParams.CloudSQL.Enabled {
		return nil, fmt.Errorf("Cloud SQL is not enabled in configuration")
	}

	if config.ConnectionParams.CloudSQL.InstanceName == "" {
		return nil, fmt.Errorf("Cloud SQL instance name is required")
	}

	ctx := context.Background()

	// Build dialer options
	var dialerOpts []cloudsqlconn.Option

	// Add credentials file if specified
	if config.ConnectionParams.CloudSQL.CredentialsFile != "" {
		dialerOpts = append(dialerOpts, cloudsqlconn.WithCredentialsFile(config.ConnectionParams.CloudSQL.CredentialsFile))
	}

	// Enable IAM authentication if specified
	if config.ConnectionParams.CloudSQL.UseIAMAuth {
		dialerOpts = append(dialerOpts, cloudsqlconn.WithIAMAuthN())
	}

	// Set refresh timeout if specified
	if config.ConnectionParams.CloudSQL.RefreshTimeout > 0 {
		timeout := time.Duration(config.ConnectionParams.CloudSQL.RefreshTimeout) * time.Second
		dialerOpts = append(dialerOpts, cloudsqlconn.WithRefreshTimeout(timeout))
	}

	// Build dial options
	var dialOpts []cloudsqlconn.DialOption
	if config.ConnectionParams.CloudSQL.UsePrivateIP {
		dialOpts = append(dialOpts, cloudsqlconn.WithPrivateIP())
	}

	// Set default dial options if any
	if len(dialOpts) > 0 {
		dialerOpts = append(dialerOpts, cloudsqlconn.WithDefaultDialOptions(dialOpts...))
	}

	// Create the dialer
	dialer, err := cloudsqlconn.NewDialer(ctx, dialerOpts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create Cloud SQL dialer: %w", err)
	}

	// Build DSN for Cloud SQL
	dsn := buildCloudSQLPostgresDSN(config.ConnectionParams)

	// Parse the DSN using pgx
	pgxConfig, err := pgx.ParseConfig(dsn)
	if err != nil {
		dialer.Close()
		return nil, fmt.Errorf("failed to parse DSN: %w", err)
	}

	// Configure the DialFunc to use Cloud SQL connector
	pgxConfig.DialFunc = func(ctx context.Context, network, addr string) (net.Conn, error) {
		return dialer.Dial(ctx, config.ConnectionParams.CloudSQL.InstanceName)
	}

	// Register the config and get the connection string
	connStr := stdlib.RegisterConnConfig(pgxConfig)

	// Build GORM config
	gormConfig := zdbconfig.BuildGormConfig(config.LogConfig)

	// Open GORM connection directly using the registered driver name
	dbConn, err := gorm.Open(postgres.New(postgres.Config{
		DriverName: CloudSQLDriverName,
		DSN:        connStr,
	}), gormConfig)
	if err != nil {
		dialer.Close()
		return nil, fmt.Errorf("failed to connect to Cloud SQL PostgreSQL: %w", err)
	}

	// Configure connection pool settings
	sqlDB, err := dbConn.DB()
	if err != nil {
		dialer.Close()
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	if config.MaxIdleConns > 0 {
		sqlDB.SetMaxIdleConns(config.MaxIdleConns)
	}
	if config.MaxOpenConns > 0 {
		sqlDB.SetMaxOpenConns(config.MaxOpenConns)
	}
	if config.ConnMaxLifetime > 0 {
		sqlDB.SetConnMaxLifetime(config.ConnMaxLifetime)
	}

	return dbConn, nil
}

func (c *CloudSQLPostgresConnector) VerifyConnection(db *gorm.DB) error {
	return db.Exec("SELECT 1").Error
}

func buildCloudSQLPostgresDSN(params zdbconfig.ConnectionParams) string {
	// Build DSN parameters as key-value pairs
	dsnParams := make(map[string]string)

	// Add required parameters if they exist
	addParamIfNotEmpty(dsnParams, dsnKeyUser, params.User)
	addParamIfNotEmpty(dsnParams, dsnKeyDatabase, params.Name)

	// Add password only if not using IAM authentication
	if !params.CloudSQL.UseIAMAuth {
		addParamIfNotEmpty(dsnParams, dsnKeyPassword, params.Password)
	}

	// Cloud SQL connector handles SSL automatically, so we disable it in the DSN
	dsnParams[dsnKeySSLMode] = "disable"

	// Parse and add additional parameters if specified
	if params.Params != "" {
		additionalParams := parseConnectionParams(params.Params)
		for key, value := range additionalParams {
			dsnParams[key] = value
		}
	}

	// Build the final DSN string
	return buildDSNString(dsnParams)
}

// addParamIfNotEmpty adds a parameter to the DSN map only if the value is not empty
func addParamIfNotEmpty(dsnParams map[string]string, key, value string) {
	if value != "" {
		dsnParams[key] = value
	}
}

// buildDSNString constructs the final DSN string from the parameters map
func buildDSNString(dsnParams map[string]string) string {
	var dsnParts []string
	for key, value := range dsnParams {
		dsnParts = append(dsnParts, fmt.Sprintf("%s=%s", key, value))
	}
	return strings.Join(dsnParts, " ")
}

// parseConnectionParams parses a connection parameter string into key-value pairs
func parseConnectionParams(params string) map[string]string {
	result := make(map[string]string)

	// Split by spaces to get individual parameters
	parts := strings.Fields(params)

	for _, part := range parts {
		// Split each part by '=' to get key-value pairs
		if kv := strings.SplitN(part, "=", 2); len(kv) == 2 {
			result[kv[0]] = kv[1]
		}
	}

	return result
}
