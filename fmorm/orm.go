package fmorm

import (
	"database/sql"
	"github.com/go-farmyard/farmyard/fmutil"
	"github.com/jmoiron/sqlx"
	"github.com/jmoiron/sqlx/reflectx"
)

type TableNameProvider interface {
	TableName() string
}

type DbExecutor interface {
	Exec(query string, args ...any) (sql.Result, error)
	Select(dest any, query string, args ...any) error
	Get(dest any, query string, args ...any) error
}

type Orm struct {
	db          *sqlx.DB
	dialect     Dialect
	dbExecutor  DbExecutor
	fieldMapper *reflectx.Mapper
}

func (d *Orm) Exec(query string, args ...any) (sql.Result, error) {
	return d.db.Exec(query, args...)
}

func (d *Orm) Select(dest any, query string, args ...any) error {
	return d.db.Select(dest, query, args...)
}

func (d *Orm) Get(dest any, query string, args ...any) error {
	return d.db.Get(dest, query, args...)
}

func NewOrm(DB *sqlx.DB) *Orm {
	orm := &Orm{
		db:          DB,
		fieldMapper: reflectx.NewMapper("db"),
	}
	orm.dbExecutor = orm

	if orm.db != nil {
		switch orm.db.DriverName() {
		case "mysql":
			orm.dialect = &DialectMySQL{}
		default:
			fmutil.Panic("unsupported driver dialet: %s", orm.db.DriverName())
		}
	} else {
		orm.dialect = &DialectNop{}
	}
	return orm
}

func (d *Orm) Cmd() *CmdBuilder {
	return NewCmdBuilder(d)
}

func (d *Orm) Table(t any) *CmdBuilder {
	return d.Cmd().Table(t)
}

func (d *Orm) Columns(cols ...string) *CmdBuilder {
	return d.Cmd().Columns(cols...)
}

func (d *Orm) Where(clause any, args ...any) *CmdBuilder {
	return d.Cmd().Where(clause, args...)
}

func (d *Orm) WhereModel(model TableNameProvider) *CmdBuilder {
	return d.Cmd().WhereModel(model)
}
