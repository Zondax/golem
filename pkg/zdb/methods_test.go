package zdb

import (
	"database/sql"
	"github.com/stretchr/testify/mock"
	"testing"

	"github.com/stretchr/testify/suite"
	"gorm.io/gorm/clause"
)

type TestStruct struct {
	Name string
	Age  int
}

type ZDatabaseSuite struct {
	suite.Suite
	db ZDatabase
}

func (suite *ZDatabaseSuite) SetupTest() {
	mockDb := new(MockZDatabase)
	suite.db = mockDb
}

func (suite *ZDatabaseSuite) TestFind() {
	suite.db.(*MockZDatabase).On("Find", &TestStruct{}, "Name = ?", "Messi").Return(suite.db)
	newDb := suite.db.Find(&TestStruct{}, "Name = ?", "Messi")
	suite.NotNil(newDb)
	suite.db.(*MockZDatabase).AssertExpectations(suite.T())
}

func (suite *ZDatabaseSuite) TestScan() {
	suite.db.(*MockZDatabase).On("Scan", &TestStruct{}).Return(suite.db)
	newDb := suite.db.Scan(&TestStruct{})
	suite.NotNil(newDb)
	suite.db.(*MockZDatabase).AssertExpectations(suite.T())
}

func (suite *ZDatabaseSuite) TestRows() {
	mockRows := new(sql.Rows)
	suite.db.(*MockZDatabase).On("Rows").Return(mockRows, nil)
	rows, err := suite.db.Rows()
	suite.Nil(err)
	suite.Equal(mockRows, rows)
	suite.db.(*MockZDatabase).AssertExpectations(suite.T())
}

func (suite *ZDatabaseSuite) TestScanRows() {
	mockRows := new(sql.Rows)
	suite.db.(*MockZDatabase).On("ScanRows", mockRows, &TestStruct{}).Return(nil)
	err := suite.db.ScanRows(mockRows, &TestStruct{})
	suite.Nil(err)
	suite.db.(*MockZDatabase).AssertExpectations(suite.T())
}

func (suite *ZDatabaseSuite) TestCreate() {
	suite.db.(*MockZDatabase).On("Create", &TestStruct{}).Return(suite.db)
	newDb := suite.db.Create(&TestStruct{})
	suite.NotNil(newDb)
	suite.db.(*MockZDatabase).AssertExpectations(suite.T())
}

func (suite *ZDatabaseSuite) TestDelete() {
	suite.db.(*MockZDatabase).On("Delete", &TestStruct{}, "Name = ?", "Messi").Return(suite.db)
	newDb := suite.db.Delete(&TestStruct{}, "Name = ?", "Messi")
	suite.NotNil(newDb)
	suite.db.(*MockZDatabase).AssertExpectations(suite.T())
}

func (suite *ZDatabaseSuite) TestRaw() {
	suite.db.(*MockZDatabase).On("Raw", "SELECT * FROM tests WHERE name = ?", "Messi").Return(suite.db)
	newDb := suite.db.Raw("SELECT * FROM tests WHERE name = ?", "Messi")
	suite.NotNil(newDb)
	suite.db.(*MockZDatabase).AssertExpectations(suite.T())
}

func (suite *ZDatabaseSuite) TestExec() {
	suite.db.(*MockZDatabase).On("Exec", "UPDATE test SET name = ?", []interface{}{"Messi"}).Return(suite.db)
	newDb := suite.db.Exec("UPDATE test SET name = ?", "Messi")
	suite.NotNil(newDb)
	suite.db.(*MockZDatabase).AssertExpectations(suite.T())
}

func (suite *ZDatabaseSuite) TestSelect() {
	suite.db.(*MockZDatabase).On("Select", "name", []interface{}{"Messi"}).Return(suite.db)
	newDb := suite.db.Select("name", []interface{}{"Messi"})
	suite.NotNil(newDb)
	suite.db.(*MockZDatabase).AssertExpectations(suite.T())
}

func (suite *ZDatabaseSuite) TestWhere() {
	suite.db.(*MockZDatabase).On("Where", "name = ?", []interface{}{"Messi"}).Return(suite.db)
	newDb := suite.db.Where("name = ?", []interface{}{"Messi"})
	suite.NotNil(newDb)
	suite.db.(*MockZDatabase).AssertExpectations(suite.T())
}

func (suite *ZDatabaseSuite) TestLimit() {
	suite.db.(*MockZDatabase).On("Limit", 10).Return(suite.db)
	newDb := suite.db.Limit(10)
	suite.NotNil(newDb)
	suite.db.(*MockZDatabase).AssertExpectations(suite.T())
}

func (suite *ZDatabaseSuite) TestTable() {
	suite.db.(*MockZDatabase).On("Table", "test_table").Return(suite.db)
	newDb := suite.db.Table("test_table")
	suite.NotNil(newDb)
	suite.db.(*MockZDatabase).AssertExpectations(suite.T())
}

func (suite *ZDatabaseSuite) TestClauses() {
	suite.db.(*MockZDatabase).On("Clauses", mock.Anything).Return(suite.db)
	newDb := suite.db.Clauses(clause.OnConflict{})
	suite.NotNil(newDb)
	suite.db.(*MockZDatabase).AssertExpectations(suite.T())
}

func (suite *ZDatabaseSuite) TestError() {
	suite.db.(*MockZDatabase).On("Error").Return(nil)
	err := suite.db.Error()
	suite.Nil(err)
	suite.db.(*MockZDatabase).AssertExpectations(suite.T())
}

func (suite *ZDatabaseSuite) TestRowsAffected() {
	suite.db.(*MockZDatabase).On("RowsAffected").Return(int64(1))
	rows := suite.db.RowsAffected()
	suite.Equal(int64(1), rows)
	suite.db.(*MockZDatabase).AssertExpectations(suite.T())
}

func TestZDatabaseSuite(t *testing.T) {
	suite.Run(t, new(ZDatabaseSuite))
}
