package zdbconnector

import (
	"github.com/zondax/golem/pkg/zdatabase/zdbconfig"
	"gorm.io/gorm"
)

var Connectors = map[string]DBConnector{
	"clickhouse": &ClickHouseConnector{},
}

type DBConnector interface {
	Connect(config *zdbconfig.Config) (*gorm.DB, error)
}
