package zdb

import (
	"database/sql"
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

func (z *zDatabase) Scan(dest interface{}) ZDatabase {
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

func (z *zDatabase) Delete(value interface{}, where ...interface{}) ZDatabase {
	return wrap(z.db.Delete(value, where...))
}

func (z *zDatabase) Raw(sql string, values ...interface{}) ZDatabase {
	return wrap(z.db.Raw(sql, values...))
}

func (z *zDatabase) Table(name string) ZDatabase {
	return wrap(z.db.Table(name))
}

func (z *zDatabase) Clauses(conds ...clause.Expression) ZDatabase {
	return wrap(z.db.Clauses(conds...))
}

func (z *zDatabase) Transaction(fc func(tx ZDatabase) error, opts ...*sql.TxOptions) (err error) {
	return z.db.Transaction(func(tx *gorm.DB) error {
		return fc(wrap(tx))
	}, opts...)
}

func (z *zDatabase) RowsAffected() int64 {
	return z.db.RowsAffected
}

func (z *zDatabase) Error() error {
	return z.db.Error
}
