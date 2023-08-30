package zdbconnector

import (
	"github.com/zondax/golem/pkg/zdb/zdbconfig"
	"gorm.io/gorm"
)

const (
	DBTypeClickhouse = "clickhouse"
)

var Connectors = map[string]DBConnector{
	DBTypeClickhouse: &ClickHouseConnector{},
}

type DBConnector interface {
	Connect(config *zdbconfig.Config) (*gorm.DB, error)
}