package fmorm

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/go-farmyard/farmyard/fmutil"
	"reflect"
)

var columnsFull = []string{"*"}
var columnsModel = []string{"__Model"}

type CmdBuilder struct {
	orm *Orm

	sqlCmdPrefix string
	sqlCmd       SqlCmd

	tableReference string

	columns    []string
	whereCond  ClauseProvider
	groupBy    string
	havingCond ClauseProvider
	orderBy    string

	limitOffset, limitCount int
}

func NewCmdBuilder(orm *Orm) *CmdBuilder {
	return &CmdBuilder{orm: orm}
}

func (cb *CmdBuilder) Table(v any) *CmdBuilder {
	switch v := v.(type) {
	case string:
		cb.tableReference = v
	case TableNameProvider:
		cb.tableReference = v.TableName()
	default:
		fmutil.Panic("unsupported table name type: %T", v)
	}
	return cb
}

func (cb *CmdBuilder) setColumns(cols []string) {
	fmutil.MustTrue(cb.columns == nil, "columns can only be set once")
	cb.columns = cols
}

// Columns is only for UPDATE/INSERT
func (cb *CmdBuilder) Columns(cols ...string) *CmdBuilder {
	cb.setColumns(cols)
	return cb
}

func toClauseProvider(clause any, args ...any) ClauseProvider {
	switch c := clause.(type) {
	case ClauseProvider:
		fmutil.MustTrue(len(args) == 0, "clause should not have args")
		return c
	case string:
		return CondArgs(c, args...)
	case TableNameProvider:
		fmutil.Panic("you should convert a model to a cond (CondModel)")
		return nil
	default:
		fmutil.Panic("unsupported clause type: %T", clause)
		return nil
	}
}

func (cb *CmdBuilder) Where(clause any, args ...any) *CmdBuilder {
	cb.whereCond = toClauseProvider(clause, args...)
	return cb
}

func (cb *CmdBuilder) WhereModel(model TableNameProvider) *CmdBuilder {
	cb.tryToUseTable(model)
	return cb.Where(CondModel(model))
}

func (cb *CmdBuilder) GroupBy(groupBy string) *CmdBuilder {
	cb.groupBy = groupBy
	return cb
}

func (cb *CmdBuilder) Having(clause any, args ...any) *CmdBuilder {
	cb.havingCond = toClauseProvider(clause, args...)
	return cb
}

func (cb *CmdBuilder) OrderBy(orderBy string) *CmdBuilder {
	cb.orderBy = orderBy
	return cb
}

// Limit demo SELECT * FROM tbl LIMIT 5,10;  # Retrieve rows 6-15
func (cb *CmdBuilder) Limit(n ...int) *CmdBuilder {
	if len(n) == 1 {
		cb.limitOffset = 0
		cb.limitCount = n[0]
	} else if len(n) == 2 {
		cb.limitOffset = n[0]
		cb.limitCount = n[1]
	} else {
		fmutil.Panic("invalid limit args")
	}
	return cb
}

func (cb *CmdBuilder) startCommand(cmd string) error {
	fmutil.MustTrue(cb.sqlCmdPrefix == "", "sql command should be only set once")
	cb.sqlCmdPrefix = cmd
	cb.sqlCmd.Append(cmd, " ")
	return nil
}

func (cb *CmdBuilder) prepareTableRowByColumns(row any) ModelStructFields {
	cb.tryToUseTable(row)
	fmutil.MustTrue(cb.tableReference != "", "sql command lacks table reference")
	fmutil.MustTrue(len(cb.columns) != 0, "columns must be set")

	if len(cb.columns) == 1 {
		if cb.columns[0] == columnsFull[0] {
			fields, err := modelStructFieldValues(cb.orm.fieldMapper, row, true)
			fmutil.MustNoError(err, "row must be a map or model struct")
			return fields
		} else if cb.columns[0] == columnsModel[0] {
			fields, err := modelStructFieldValues(cb.orm.fieldMapper, row, false)
			fmutil.MustNoError(err, "row must be a map or model struct")
			return fields
		}
	}

	var fields ModelStructFields
	err := modelStructFieldTraversal(cb.orm.fieldMapper, row, cb.columns, func(name string, val reflect.Value) {
		fields = append(fields, &ModelStructField{Name: name, Value: val.Interface()})
	})
	fmutil.MustNoError(err, "can not get row columns")
	return fields
}

func (cb *CmdBuilder) UpdateRow(row any) (sql.Result, error) {
	if err := cb.startCommand("UPDATE"); err != nil {
		return nil, err
	}
	rowFields := cb.prepareTableRowByColumns(row)

	cb.sqlCmd.Append(cb.tableReference, " SET ")
	for i, field := range rowFields {
		cb.sqlCmd.Append(field.Name, "=?")
		if i != len(rowFields)-1 {
			cb.sqlCmd.Append(",")
		}
		cb.sqlCmd.Args = append(cb.sqlCmd.Args, field.Value)
	}

	err := cb.sqlCmd.Append(" WHERE ").AppendClause(cb.orm, cb.whereCond)
	fmutil.MustNoError(err)

	cb.sqlCmdAppendOrderLimit()
	return cb.orm.dbExecutor.Exec(cb.sqlCmd.Query.String(), cb.sqlCmd.Args...)
}

func (cb *CmdBuilder) UpdateFull(row any) (sql.Result, error) {
	cb.setColumns(columnsFull)
	return cb.UpdateRow(row)
}

func (cb *CmdBuilder) UpdateModel(row any) (sql.Result, error) {
	cb.setColumns(columnsModel)
	return cb.UpdateRow(row)
}

func (cb *CmdBuilder) InsertRow(row any) (sql.Result, error) {
	if err := cb.startCommand("INSERT"); err != nil {
		return nil, err
	}
	rowFields := cb.prepareTableRowByColumns(row)

	cb.sqlCmd.Append("INTO ", cb.tableReference, " (")
	for i, field := range rowFields {
		cb.sqlCmd.Append(field.Name)
		if i != len(rowFields)-1 {
			cb.sqlCmd.Append(",")
		}
	}
	cb.sqlCmd.Append(") VALUES (")
	for i, field := range rowFields {
		cb.sqlCmd.Append("?")
		if i != len(rowFields)-1 {
			cb.sqlCmd.Append(",")
		}
		cb.sqlCmd.Args = append(cb.sqlCmd.Args, field.Value)
	}
	cb.sqlCmd.Append(")")

	return cb.orm.dbExecutor.Exec(cb.sqlCmd.Query.String(), cb.sqlCmd.Args...)
}

func (cb *CmdBuilder) InsertFull(row any) (sql.Result, error) {
	cb.setColumns(columnsFull)
	return cb.InsertRow(row)
}

func (cb *CmdBuilder) InsertModel(row any) (sql.Result, error) {
	cb.setColumns(columnsModel)
	return cb.InsertRow(row)
}

func (cb *CmdBuilder) Delete() (sql.Result, error) {
	if err := cb.startCommand("DELETE"); err != nil {
		return nil, err
	}
	fmutil.MustTrue(cb.tableReference != "", "sql command lacks table reference")

	err := cb.sqlCmd.Append("FROM ", cb.tableReference, " WHERE ").AppendClause(cb.orm, cb.whereCond)
	fmutil.MustNoError(err)

	cb.sqlCmdAppendOrderLimit()
	return cb.orm.dbExecutor.Exec(cb.sqlCmd.Query.String(), cb.sqlCmd.Args...)
}

func (cb *CmdBuilder) prepareSelectQuery() error {
	if err := cb.startCommand("SELECT"); err != nil {
		return err
	}
	if len(cb.columns) != 0 {
		fmutil.MustTrue(!(len(cb.columns) == 1 && cb.columns[0] == columnsModel[0]), "select can not work with ColumnsModel")
		cb.sqlCmd.AppendStrings(cb.columns, ",")
	} else {
		cb.sqlCmd.Append("*")
	}

	fmutil.MustTrue(cb.tableReference != "", "sql command lacks table reference")

	err := cb.sqlCmd.Append(" FROM ", cb.tableReference, " WHERE ").AppendClause(cb.orm, cb.whereCond)
	fmutil.MustNoError(err)

	if cb.groupBy != "" {
		cb.sqlCmd.Append(" GROUP BY ", cb.groupBy)
	}

	if cb.havingCond != nil {
		if err := cb.sqlCmd.Append(" HAVING ").AppendClause(cb.orm, cb.havingCond); err != nil {
			return err
		}
	}
	cb.sqlCmdAppendOrderLimit()
	return nil
}

func (cb *CmdBuilder) tryToUseTable(table any) {
	if cb.tableReference == "" {
		if tableName, ok := getTableFromSlice(table); ok && tableName != "" {
			cb.Table(tableName)
		} else {
			cb.Table(table)
		}
	}
}

func (cb *CmdBuilder) sqlCmdAppendOrderLimit() {
	if cb.orderBy != "" {
		cb.sqlCmd.Append(" ORDER BY ", cb.orderBy)
	}

	if cb.limitCount != 0 {
		if cb.limitOffset != 0 {
			cb.sqlCmd.Append(fmt.Sprintf(" LIMIT %d,%d", cb.limitOffset, cb.limitCount))
		} else {
			cb.sqlCmd.Append(fmt.Sprintf(" LIMIT %d", cb.limitCount))
		}
	}
}

func (cb *CmdBuilder) SelectOne(result any) (found bool, err error) {
	cb.tryToUseTable(result)
	err = cb.Limit(1).prepareSelectQuery()
	fmutil.MustNoError(err)

	sqlQuery := cb.sqlCmd.Query.String()
	err = cb.orm.dbExecutor.Get(result, sqlQuery, cb.sqlCmd.Args...)
	if err != nil {
		resetToZeroValue(result)
		if errors.Is(err, sql.ErrNoRows) {
			err = nil
		}
	} else {
		found = true
	}
	return found, err
}

func (cb *CmdBuilder) SelectAll(rowsPtr any) error {
	cb.tryToUseTable(rowsPtr)
	err := cb.prepareSelectQuery()
	fmutil.MustNoError(err)
	sqlQuery := cb.sqlCmd.Query.String()
	return cb.orm.dbExecutor.Select(rowsPtr, sqlQuery, cb.sqlCmd.Args...)
}
