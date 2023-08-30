package zdbconnector

import (
	"fmt"
	"github.com/zondax/golem/pkg/zdb/zdbconfig"
	"gorm.io/driver/clickhouse"
	"gorm.io/gorm"
)

type ClickHouseConnector struct{}

func (c *ClickHouseConnector) Connect(config *zdbconfig.Config) (*gorm.DB, error) {
	dsn := buildClickhouseDSN(config.ConnectionParams)
	gormConfig := zdbconfig.BuildGormConfig(config.LogConfig)

	var dbConn *gorm.DB
	var err error
	dbConn, err = gorm.Open(clickhouse.Open(dsn), gormConfig)
	if err != nil {
		return nil, err
	}
	return dbConn, nil
}

func buildClickhouseDSN(params zdbconfig.ConnectionParams) string {
	dsn := fmt.Sprintf(
		"clickhouse://%s:%s@%s:%v/%s",
		params.User,
		params.Password,
		params.Host,
		params.Port,
		params.Name,
	)

	if params.Params != "" {
		dsn = fmt.Sprintf("%s?%s", dsn, params.Params)
	}

	return dsn
}

func (c *ClickHouseConnector) VerifyConnection(db *gorm.DB) error {
	return db.Exec("SELECT 1").Error
}
