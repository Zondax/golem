package zdbconnector

import (
	"fmt"
	clickhouse2 "github.com/ClickHouse/clickhouse-go/v2"
	"github.com/zondax/golem/pkg/zdb/zdbconfig"
	"gorm.io/driver/clickhouse"
	"gorm.io/gorm"
	"strings"
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
	protocol := clickhouse2.Native.String()
	if strings.EqualFold(params.Protocol, clickhouse2.HTTP.String()) {
		protocol = clickhouse2.HTTP.String()
	}

	dsn := fmt.Sprintf(
		"%s://%s:%s@%s:%v/%s",
		protocol,
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
