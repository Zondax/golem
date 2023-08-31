package zdbconnector

import (
	"github.com/zondax/golem/pkg/zdb/zdbconfig"
	"gorm.io/gorm"
)

const (
	DBTypeClickhouse = "clickhouse"
	DBTypePostgres   = "postgres"
)

var Connectors = map[string]DBConnector{
	DBTypeClickhouse: &ClickHouseConnector{},
	DBTypePostgres:   &PostgresConnector{},
}

type DBConnector interface {
	Connect(config *zdbconfig.Config) (*gorm.DB, error)
	VerifyConnection(db *gorm.DB) error
}
