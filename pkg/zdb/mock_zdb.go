package zdb

import (
	"database/sql"
	"github.com/stretchr/testify/mock"
	"github.com/zondax/golem/pkg/zdb/zdbconfig"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type MockZDatabase struct {
	mock.Mock
}

func (m *MockZDatabase) GetDbConnection() *gorm.DB {
	args := m.Called()
	return args.Get(0).(*gorm.DB)
}

func (m *MockZDatabase) Find(dest interface{}, conds ...interface{}) ZDatabase {
	args := m.Called(dest, conds[0], conds[1])
	return args.Get(0).(ZDatabase)
}

func (m *MockZDatabase) Scan(dest interface{}) ZDatabase {
	m.Called(dest)
	return m
}

func (m *MockZDatabase) Rows() (*sql.Rows, error) {
	args := m.Called()
	return args.Get(0).(*sql.Rows), args.Error(1)
}

func (m *MockZDatabase) ScanRows(rows *sql.Rows, result interface{}) error {
	args := m.Called(rows, result)
	return args.Error(0)
}

func (m *MockZDatabase) Create(value interface{}) ZDatabase {
	m.Called(value)
	return m
}

func (m *MockZDatabase) Delete(value interface{}, conds ...interface{}) ZDatabase {
	args := m.Called(value, conds[0], conds[1])
	return args.Get(0).(ZDatabase)
}

func (m *MockZDatabase) Raw(sql string, values ...interface{}) ZDatabase {
	args := m.Called(sql, values[0])
	return args.Get(0).(ZDatabase)
}

func (m *MockZDatabase) Select(query interface{}, values ...interface{}) ZDatabase {
	args := m.Called(query, values[0])
	return args.Get(0).(ZDatabase)
}

func (m *MockZDatabase) Where(query interface{}, values ...interface{}) ZDatabase {
	args := m.Called(query, values[0])
	return args.Get(0).(ZDatabase)
}

func (m *MockZDatabase) Limit(limit int) ZDatabase {
	args := m.Called(limit)
	return args.Get(0).(ZDatabase)
}

func (m *MockZDatabase) Exec(sql string, values ...interface{}) ZDatabase {
	m.Called(sql, values)
	return m
}

func (m *MockZDatabase) Table(name string) ZDatabase {
	m.Called(name)
	return m
}

func (m *MockZDatabase) Transaction(fc func(tx ZDatabase) error, opts ...*sql.TxOptions) (err error) {
	args := m.Called(fc, opts)
	return args.Error(0)
}

func (m *MockZDatabase) Clauses(conds ...clause.Expression) ZDatabase {
	args := m.Called(conds[0])
	return args.Get(0).(ZDatabase)
}

func (m *MockZDatabase) Error() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockZDatabase) RowsAffected() int64 {
	args := m.Called()
	return args.Get(0).(int64)
}

// MockDBConnector

type MockDBConnector struct {
	mock.Mock
}

func (m *MockDBConnector) NewInstance(dbType string, config *zdbconfig.Config) (ZDatabase, error) {
	args := m.Called(dbType, config)
	return args.Get(0).(ZDatabase), args.Error(1)
}

func (m *MockDBConnector) Connect(config *zdbconfig.Config) (*gorm.DB, error) {
	args := m.Called(config)
	return args.Get(0).(*gorm.DB), args.Error(1)
}

func (m *MockDBConnector) VerifyConnection(db *gorm.DB) error {
	args := m.Called(db)
	return args.Error(0)
}
