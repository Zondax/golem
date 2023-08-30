package zdb

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/zondax/golem/pkg/zdb/zdbconfig"
	"github.com/zondax/golem/pkg/zdb/zdbconnector"
	"gorm.io/gorm"
	"testing"
)

func TestNewInstanceForClickHouse(t *testing.T) {
	mockConnector := new(MockDBConnector)

	config := &zdbconfig.Config{}

	mockConnector.On("Connect", config).Return(&gorm.DB{}, nil)
	mockConnector.On("VerifyConnection", mock.Anything).Return(nil)

	zdbconnector.Connectors["clickhouse"] = mockConnector
	_, err := NewInstance("clickhouse", config)

	assert.NoError(t, err)
}
