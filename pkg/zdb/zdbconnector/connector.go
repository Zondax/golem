package zdbconnector

import (
	"github.com/zondax/golem/pkg/zdb/zdbconfig"
	"gorm.io/gorm"
)

const (
	DBTypeClickhouse       = "clickhouse"
	DBTypePostgres         = "postgres"
	DBTypeCloudSQLPostgres = "cloudsql-postgres"
)

var Connectors = map[string]DBConnector{
	DBTypeClickhouse:       &ClickHouseConnector{},
	DBTypePostgres:         &PostgresConnector{},
	DBTypeCloudSQLPostgres: &CloudSQLPostgresConnector{},
}

type DBConnector interface {
	Connect(config *zdbconfig.Config) (*gorm.DB, error)
	VerifyConnection(db *gorm.DB) error
}
