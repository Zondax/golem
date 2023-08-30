package zdatabase

import (
	"database/sql"
	"fmt"
	"github.com/zondax/golem/pkg/zdatabase/zdbconfig"
	"github.com/zondax/golem/pkg/zdatabase/zdbconnector"
	"go.uber.org/zap"
	"gorm.io/gorm/clause"
	"time"

	"gorm.io/gorm"
)

const (
	retryDefault       = 3
	maxAttemptsDefault = 5
)

type ZDatabase interface {
	Find(out interface{}, where ...interface{}) ZDatabase
	Scan(dest interface{}) ZDatabase
	Rows() (*sql.Rows, error)
	ScanRows(rows *sql.Rows, result interface{}) error
	Create(value interface{}) ZDatabase
	Delete(value interface{}, where ...interface{}) ZDatabase
	Raw(sql string, values ...interface{}) ZDatabase
	Exec(sql string, values ...interface{}) ZDatabase
	Table(name string) ZDatabase
	Transaction(fc func(tx ZDatabase) error, opts ...*sql.TxOptions) (err error)
	Clauses(conds ...clause.Expression) ZDatabase
	Error() error
	RowsAffected() int64
	GetDbConnection() *gorm.DB
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
			verifyErr := verifyDBConnection(dbConn)
			if verifyErr == nil {
				return &zDatabase{db: dbConn}, nil
			}

			err = verifyErr
		}

		zap.S().Infof("Failed to establish database connection: %v. Attempt %d/%d. Retrying in %d seconds...", err, i+1, config.MaxAttempts, config.RetryInterval)
		time.Sleep(time.Duration(config.RetryInterval) * time.Second)
	}

	zap.S().Infof("Unable to establish database connection after %d attempts.", config.MaxAttempts)
	return nil, err
}

func verifyDBConnection(dbConn *gorm.DB) error {
	return dbConn.Exec("SELECT 1").Error
}