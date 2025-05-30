package zdb

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/zondax/golem/pkg/logger"
	"github.com/zondax/golem/pkg/zdb/zdbconfig"
	"github.com/zondax/golem/pkg/zdb/zdbconnector"
	"gorm.io/gorm/clause"

	"gorm.io/gorm"
)

const (
	retryDefault       = 3
	maxAttemptsDefault = 5
)

type ZDatabase interface {
	Find(out interface{}, where ...interface{}) ZDatabase
	First(dest interface{}, where ...interface{}) ZDatabase
	FirstOrCreate(dest interface{}, where ...interface{}) ZDatabase
	Scan(dest interface{}) ZDatabase
	Rows() (*sql.Rows, error)
	ScanRows(rows *sql.Rows, result interface{}) error
	Select(query interface{}, args ...interface{}) ZDatabase
	Where(query interface{}, args ...interface{}) ZDatabase
	Joins(query string, args ...interface{}) ZDatabase
	UnionAll(subQuery1 ZDatabase, subQuery2 ZDatabase) ZDatabase
	UnionDistinct(subQuery1 ZDatabase, subQuery2 ZDatabase) ZDatabase
	Limit(limit int) ZDatabase
	Offset(offset int) ZDatabase
	Order(value interface{}) ZDatabase
	Distinct(args ...interface{}) ZDatabase
	Count(count *int64) ZDatabase
	Group(name string) ZDatabase
	Create(value interface{}) ZDatabase
	Updates(value interface{}) ZDatabase
	Update(column string, value interface{}) ZDatabase
	Delete(value interface{}, where ...interface{}) ZDatabase
	Raw(sql string, values ...interface{}) ZDatabase
	Exec(sql string, values ...interface{}) ZDatabase
	Table(name string, args ...interface{}) ZDatabase
	Transaction(fc func(tx ZDatabase) error, opts ...*sql.TxOptions) (err error)
	Clauses(conds ...clause.Expression) ZDatabase
	WithContext(ctx context.Context) ZDatabase
	Error() error
	Scopes(funcs ...func(ZDatabase) ZDatabase) ZDatabase
	RowsAffected() int64
	GetDbConnection() *gorm.DB
	GetDBStats() (sql.DBStats, error)
}

type zDatabase struct {
	db *gorm.DB
}

func NewInstance(dbType string, config *zdbconfig.Config) (ZDatabase, error) {
	if config.RetryInterval == 0 {
		config.RetryInterval = retryDefault
	}

	if config.MaxAttempts == 0 {
		config.MaxAttempts = maxAttemptsDefault
	}

	connector, ok := zdbconnector.Connectors[dbType]
	if !ok {
		return nil, fmt.Errorf("unsupported database type %s", dbType)
	}

	var dbConn *gorm.DB
	var err error

	for i := 0; i < config.MaxAttempts; i++ {
		dbConn, err = connector.Connect(config)
		if err == nil {
			verifyErr := connector.VerifyConnection(dbConn)
			if verifyErr == nil {
				// Setup OpenTelemetry instrumentation with configuration
				if err := setupOpenTelemetryInstrumentation(dbConn, &config.OpenTelemetry); err != nil {
					return nil, fmt.Errorf("failed to setup OpenTelemetry instrumentation: %w", err)
				}

				return &zDatabase{db: dbConn}, nil
			}

			err = verifyErr
		}

		logger.GetLoggerFromContext(context.Background()).Infof("Failed to establish database connection: %v. Attempt %d/%d. Retrying in %d seconds...", err, i+1, config.MaxAttempts, config.RetryInterval)
		time.Sleep(time.Duration(config.RetryInterval) * time.Second)
	}

	logger.GetLoggerFromContext(context.Background()).Infof("Unable to establish database connection after %d attempts.", config.MaxAttempts)
	return nil, err
}
