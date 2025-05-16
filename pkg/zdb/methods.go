package zdb

import (
	"database/sql"
	"reflect"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func wrap(db *gorm.DB) ZDatabase {
	return &zDatabase{db}
}

func (z *zDatabase) GetDbConnection() *gorm.DB {
	return z.db
}

func (z *zDatabase) Exec(query string, values ...interface{}) ZDatabase {
	return wrap(z.db.Exec(query, values...))
}

func (z *zDatabase) Find(out interface{}, where ...interface{}) ZDatabase {
	return wrap(z.db.Find(out, where...))
}

func (z *zDatabase) First(dest interface{}, where ...interface{}) ZDatabase {
	return wrap(z.db.First(dest, where...))
}

func (z *zDatabase) FirstOrCreate(dest interface{}, where ...interface{}) ZDatabase {
	return wrap(z.db.FirstOrCreate(dest, where...))
}

func (z *zDatabase) Scan(dest interface{}) ZDatabase {
	// Avoid scanning if destination is nil
	if dest == nil {
		return z
	}

	// Avoid scanning if destination is an empty slice
	v := reflect.ValueOf(dest)
	if v.Kind() == reflect.Ptr && v.Elem().Kind() == reflect.Slice && v.Elem().Len() == 0 {
		return z
	}

	return wrap(z.db.Scan(dest))
}

func (z *zDatabase) Row() *sql.Row {
	return z.db.Row()
}

func (z *zDatabase) Rows() (*sql.Rows, error) {
	return z.db.Rows()
}

func (z *zDatabase) ScanRows(rows *sql.Rows, result interface{}) error {
	return z.db.ScanRows(rows, result)
}

func (z *zDatabase) Create(value interface{}) ZDatabase {
	return wrap(z.db.Create(value))
}

func (z *zDatabase) Updates(value interface{}) ZDatabase {
	return wrap(z.db.Updates(value))
}

func (z *zDatabase) Update(column string, value interface{}) ZDatabase {
	return wrap(z.db.Update(column, value))
}

func (z *zDatabase) Delete(value interface{}, where ...interface{}) ZDatabase {
	return wrap(z.db.Delete(value, where...))
}

func (z *zDatabase) Raw(sql string, values ...interface{}) ZDatabase {
	return wrap(z.db.Raw(sql, values...))
}

func (z *zDatabase) Table(name string, args ...interface{}) ZDatabase {
	return wrap(z.db.Table(name, args...))
}

func (z *zDatabase) Clauses(conds ...clause.Expression) ZDatabase {
	return wrap(z.db.Clauses(conds...))
}

func (z *zDatabase) Select(query interface{}, args ...interface{}) ZDatabase {
	return wrap(z.db.Select(query, args...))
}

func (z *zDatabase) Where(query interface{}, args ...interface{}) ZDatabase {
	return wrap(z.db.Where(query, args...))
}

func (z *zDatabase) Joins(query string, args ...interface{}) ZDatabase {
	return wrap(z.db.Joins(query, args...))
}

// Gorm doesn't have a UnionAll clause, so we need to build a workaround, which was found in this issue: https://github.com/go-gorm/gorm/issues/3781.
func (z *zDatabase) UnionAll(subQuery1 ZDatabase, subQuery2 ZDatabase) ZDatabase {
	unionAll := z.db.
		Table("(? ", subQuery1.GetDbConnection()).
		Joins("UNION ALL ?)", subQuery2.GetDbConnection())

	return wrap(unionAll)
}

// Gorm doesn't have a UnionDistinct clause, so we need to build a workaround, which was found in this issue: https://github.com/go-gorm/gorm/issues/3781.
func (z *zDatabase) UnionDistinct(subQuery1 ZDatabase, subQuery2 ZDatabase) ZDatabase {
	unionAll := z.db.
		Table("(? ", subQuery1.GetDbConnection()).
		Joins("UNION DISTINCT ?)", subQuery2.GetDbConnection())

	return wrap(unionAll)
}

func (z *zDatabase) Limit(limit int) ZDatabase {
	return wrap(z.db.Limit(limit))
}

func (z *zDatabase) Offset(offset int) ZDatabase {
	return wrap(z.db.Offset(offset))
}

func (z *zDatabase) Order(value interface{}) ZDatabase {
	return wrap(z.db.Order(value))
}

func (z *zDatabase) Distinct(args ...interface{}) ZDatabase {
	return wrap(z.db.Distinct(args...))
}

func (z *zDatabase) Count(count *int64) ZDatabase {
	return wrap(z.db.Count(count))
}

func (z *zDatabase) Group(name string) ZDatabase {
	return wrap(z.db.Group(name))
}

func (z *zDatabase) Transaction(fc func(tx ZDatabase) error, opts ...*sql.TxOptions) (err error) {
	return z.db.Transaction(func(tx *gorm.DB) error {
		return fc(wrap(tx))
	}, opts...)
}

func (z *zDatabase) RowsAffected() int64 {
	return z.db.RowsAffected
}

func (z *zDatabase) Scopes(funcs ...func(ZDatabase) ZDatabase) ZDatabase {
	gormFuncs := make([]func(*gorm.DB) *gorm.DB, len(funcs))
	for i, f := range funcs {
		gormFuncs[i] = func(db *gorm.DB) *gorm.DB {
			return f(wrap(db)).GetDbConnection()
		}
	}
	return wrap(z.db.Scopes(gormFuncs...))
}

func (z *zDatabase) GetDBStats() (sql.DBStats, error) {
	sqlDB, err := z.db.DB()
	return sqlDB.Stats(), err
}

func (z *zDatabase) Error() error {
	return z.db.Error
}
