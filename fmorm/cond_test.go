package fmorm

import (
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestCond_ToClause(t *testing.T) {

	orm := NewOrm(nil)

	tests := []struct {
		name   string
		cond   ClauseProvider
		clause string
		args   []any
	}{
		{
			name:   "CondMap-empty",
			cond:   CondMap{},
			clause: "1=0",
			args:   nil,
		},
		{
			name: "CondMap-k",
			cond: CondMap{
				"k": 1,
			},
			clause: "(k=?)",
			args:   []any{1},
		},
		{
			name: "CondMap-no-order",
			cond: CondMap{
				"sub": nil,
			},
			clause: "(sub)",
			args:   nil,
		},
		{
			name:   "And",
			cond:   And{{"k": 1}, Cond("sub")},
			clause: "((k=?) AND (sub))",
			args:   []any{1},
		},
		{
			name:   "Or",
			cond:   Or{{"k": 1}, {"sub": nil}},
			clause: "((k=?) OR (sub))",
			args:   []any{1},
		},
		{
			name: "In",
			cond: And{
				{"x IN": []int{1, 2, 3}},
				Cond(Or{
					{"k<>": 100},
					{"sub": nil},
				}),
			},
			clause: "((x IN (?,?,?)) AND ((k<>?) OR (sub)))",
			args:   []any{1, 2, 3, 100},
		},
		{
			name:   "Between",
			cond:   CondMap{"x NOT BETWEEN": []int{1, 2}},
			clause: "(x NOT BETWEEN ? AND ?)",
			args:   []any{1, 2},
		},
		{
			name: "Not Or",
			cond: NotOr{
				{"a=b": nil},
				{"x=y": nil},
			},
			clause: "NOT ((a=b) OR (x=y))",
			args:   nil,
		},
		{
			name:   "Cond Struct",
			cond:   CondModel(&testModel{ID: 1, K1: 2}),
			clause: "(id=? AND k1=?)",
			args:   []any{1, 2},
		},
		{
			name:   "Cond Struct Model",
			cond:   CondModel(&testModel{K1: 2}),
			clause: "(k1=?)",
			args:   []any{2},
		},
	}

	toClause := func(orm *Orm, c ClauseProvider) (string, []any, error) {
		clause := &strings.Builder{}
		var args []any
		err := c.ToClause(orm, clause, &args)
		return clause.String(), args, err
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clause, args, _ := toClause(orm, tt.cond)
			assert.EqualValues(t, tt.clause, clause)
			assert.EqualValues(t, tt.args, args)
		})
	}
}
