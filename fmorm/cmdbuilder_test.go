package fmorm

import (
	"database/sql"
	"github.com/stretchr/testify/assert"
	"testing"
)

type MockedDbExecutor struct {
	sqlText string
	sqlArgs []any
	sqlErr  error
}

func (m *MockedDbExecutor) Select(dest any, query string, args ...any) error {
	m.sqlText, m.sqlArgs = query, args
	return nil
}

func (m *MockedDbExecutor) Get(dest any, query string, args ...any) error {
	m.sqlText, m.sqlArgs = query, args
	return nil
}

func (m *MockedDbExecutor) Exec(query string, args ...any) (sql.Result, error) {
	m.sqlText, m.sqlArgs = query, args
	return nil, nil
}

type testModel struct {
	ID int `db:"id"`
	K1 int `db:"k1"`
}

func (t testModel) TableName() string {
	return "test_table"
}

func TestCmdBuilder(t *testing.T) {
	orm := NewOrm(nil)
	e := &MockedDbExecutor{}
	orm.dbExecutor = e

	tests := []struct {
		name      string
		fn        func()
		sqlText   string
		sqlArgs   []any
		sqlErrMsg string
	}{
		{
			name: "DeleteByModel-empty",
			fn: func() {
				_, e.sqlErr = orm.Cmd().WhereModel(&testModel{}).Delete()
			},
			sqlText: "DELETE FROM test_table WHERE (1=0)",
		},
		{
			name: "DeleteByModel-k1",
			fn: func() {
				_, e.sqlErr = orm.Cmd().WhereModel(&testModel{K1: 1}).Delete()
			},
			sqlText: "DELETE FROM test_table WHERE (k1=?)",
			sqlArgs: []any{1},
		},
		{
			name: "Delete-Order-Limit",
			fn: func() {
				_, e.sqlErr = orm.Cmd().WhereModel(&testModel{K1: 1}).OrderBy("O").Limit(3, 4).Delete()
			},
			sqlText: "DELETE FROM test_table WHERE (k1=?) ORDER BY O LIMIT 3,4",
			sqlArgs: []any{1},
		},
		{
			name: "Update-Error",
			fn: func() {
				_, e.sqlErr = orm.Cmd().UpdateModel(&testModel{ID: 0, K1: 1})
			},
			sqlErrMsg: "clause is nil",
		},
		{
			name: "Update-Order-Limit",
			fn: func() {
				_, e.sqlErr = orm.Cmd().WhereModel(&testModel{K1: 1}).OrderBy("O").Limit(3, 4).UpdateModel(&testModel{K1: 123})
			},
			sqlText: "UPDATE test_table SET k1=? WHERE (k1=?) ORDER BY O LIMIT 3,4",
			sqlArgs: []any{123, 1},
		},
		{
			name: "Update-Columns",
			fn: func() {
				_, e.sqlErr = orm.Cmd().Where(CondMap{}).Columns("id").UpdateRow(&testModel{ID: 0, K1: 1})
			},
			sqlText: "UPDATE test_table SET id=? WHERE 1=0",
			sqlArgs: []any{0},
		},
		{
			name: "Update-Full",
			fn: func() {
				_, e.sqlErr = orm.Cmd().Where(CondMap{}).UpdateFull(&testModel{ID: 1, K1: 2})
			},
			sqlText: "UPDATE test_table SET id=?,k1=? WHERE 1=0",
			sqlArgs: []any{1, 2},
		},
		{
			name: "Insert-Model",
			fn: func() {
				_, e.sqlErr = orm.Cmd().InsertModel(&testModel{K1: 1})
			},
			sqlText: "INSERT INTO test_table (k1) VALUES (?)",
			sqlArgs: []any{1},
		},
		{
			name: "Insert-Full",
			fn: func() {
				_, e.sqlErr = orm.Cmd().InsertFull(&testModel{K1: 1})
			},
			sqlText: "INSERT INTO test_table (id,k1) VALUES (?,?)",
			sqlArgs: []any{0, 1},
		},
		{
			name: "Insert-Error",
			fn: func() {
				_, e.sqlErr = orm.Cmd().InsertRow(&testModel{K1: 1})
			},
			sqlErrMsg: "columns must be set",
		},
		{
			name: "SelectOne",
			fn: func() {
				_, e.sqlErr = orm.Cmd().Table("t1").Columns("c1", "c2").Where(CondMap{"k1": 1}).GroupBy("g1").Having(CondMap{"h2": 2}).OrderBy("o1").SelectOne(&testModel{})
			},
			sqlText: "SELECT c1,c2 FROM t1 WHERE (k1=?) GROUP BY g1 HAVING (h2=?) ORDER BY o1 LIMIT 1",
			sqlArgs: []any{1, 2},
		},
		{
			name: "Select-All-PtrSlice",
			fn: func() {
				e.sqlErr = orm.Cmd().Where(CondMap{"k1": 1}).Limit(10).SelectAll(&[]*testModel{})
			},
			sqlText: "SELECT * FROM test_table WHERE (k1=?) LIMIT 10",
			sqlArgs: []any{1},
		},
		{
			name: "Select-All-StrictSlice",
			fn: func() {
				e.sqlErr = orm.Cmd().Where(CondMap{"k1": 1}).SelectAll(&[]testModel{})
			},
			sqlText: "SELECT * FROM test_table WHERE (k1=?)",
			sqlArgs: []any{1},
		},
	}

	callTestCase := func(fn func()) {
		defer func() {
			if err, ok := recover().(error); ok {
				e.sqlErr = err
			}
		}()
		fn()
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			*e = MockedDbExecutor{}
			callTestCase(tt.fn)
			if tt.sqlErrMsg == "" {
				assert.NoError(t, e.sqlErr)
				assert.EqualValues(t, tt.sqlText, e.sqlText)
				assert.EqualValues(t, tt.sqlArgs, e.sqlArgs)
			} else {
				if assert.Error(t, e.sqlErr) {
					assert.Contains(t, e.sqlErr.Error(), tt.sqlErrMsg)
				}
			}
		})
	}
}
