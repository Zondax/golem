package zdbconnector

import (
	"fmt"
	"github.com/zondax/golem/pkg/zdb/zdbconfig"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type PostgresConnector struct{}

func (c *PostgresConnector) Connect(config *zdbconfig.Config) (*gorm.DB, error) {
	dsn := buildPostgresDSN(config.ConnectionParams)
	gormConfig := zdbconfig.BuildGormConfig(config.LogConfig)

	dbConn, err := gorm.Open(postgres.Open(dsn), gormConfig)
	if err != nil {
		return nil, err
	}
	return dbConn, nil
}

func (c *PostgresConnector) VerifyConnection(db *gorm.DB) error {
	return db.Exec("SELECT 1").Error
}

func buildPostgresDSN(params zdbconfig.ConnectionParams) string {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%v %s",
		params.Host,
		params.User,
		params.Password,
		params.Name,
		params.Port,
		params.Params,
	)
	return dsn
}
