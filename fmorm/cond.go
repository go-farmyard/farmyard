package fmorm

import (
	"errors"
	"fmt"
	"strings"
)

type ClauseProvider interface {
	ToClause(orm *Orm, clause *strings.Builder, args *[]any) error
}

type CondMap map[string]any // NOTE: Go map has no order ....

type And []CondMap
type Or []CondMap
type NotAnd []CondMap
type NotOr []CondMap

const condKeySub = "__CondKeySub"
const condKeyModel = "__CondKeyModel"

func condMapToClause(orm *Orm, logicOp string, condMaps []CondMap, clause *strings.Builder, args *[]any) error {
	num := len(condMaps)
	if num == 0 {
		clause.WriteString("1=0")
		return nil
	}
	cnt := 0
	if num != 1 {
		clause.WriteByte('(')
	}
	for _, condMap := range condMaps {
		if err := condMap.ToClause(orm, clause, args); err != nil {
			return err
		}
		if cnt != num-1 {
			clause.WriteByte(' ')
			clause.WriteString(logicOp)
			clause.WriteByte(' ')
		}
		cnt++
	}
	if num != 1 {
		clause.WriteByte(')')
	}
	return nil
}

func (c And) ToClause(orm *Orm, clause *strings.Builder, args *[]any) error {
	return condMapToClause(orm, "AND", c, clause, args)
}

func (c Or) ToClause(orm *Orm, clause *strings.Builder, args *[]any) error {
	return condMapToClause(orm, "OR", c, clause, args)
}

func (c NotAnd) ToClause(orm *Orm, clause *strings.Builder, args *[]any) error {
	clause.WriteString("NOT ")
	return condMapToClause(orm, "AND", c, clause, args)
}

func (c NotOr) ToClause(orm *Orm, clause *strings.Builder, args *[]any) error {
	clause.WriteString("NOT ")
	return condMapToClause(orm, "OR", c, clause, args)
}

// longer first
var condOps = []string{">=", "<=", "<>", "!=", ">", "<", "=",
	" NOT BETWEEN", " BETWEEN", " NOT LIKE", " LIKE", " IS NOT", " NOT IN", " IS", " IN"}

var condOpSuffixBetween = " BETWEEN"
var condOpSuffixIn = " IN"

// ToClause NOTE: go map has no order ....
func (condMap CondMap) ToClause(orm *Orm, clause *strings.Builder, args *[]any) error {
	num := len(condMap)
	if num == 0 {
		clause.WriteString("1=0")
		return nil
	}

	if num == 1 {
		if condMap[condKeySub] != nil {
			switch sub := condMap[condKeySub].(type) {
			case ClauseProvider:
				if err := sub.ToClause(orm, clause, args); err != nil {
					return err
				}
			case string:
				clause.WriteByte('(')
				clause.WriteString(sub)
				clause.WriteByte(')')
			default:
				if m, ok := convertMapStringInterface(sub); ok {
					if err := CondMap(m).ToClause(orm, clause, args); err != nil {
						return err
					}
					return nil
				}
				return fmt.Errorf("unsupported cond (sub) type: %T", sub)
			}
			return nil
		}
		if sub, ok := condMap[condKeyModel]; ok {
			var fields ModelStructFields
			var err error
			if fields, err = modelStructFieldValues(orm.fieldMapper, sub, false); err == nil {
				if len(fields) == 0 {
					clause.WriteString("(1=0)")
				} else {
					clause.WriteByte('(')
					for i, field := range fields {
						clause.WriteString(orm.dialect.QuoteName(field.Name))
						if i != len(fields)-1 {
							clause.WriteString("=? AND ")
						} else {
							clause.WriteString("=?")
						}
						*args = append(*args, field.Value)
					}
					clause.WriteByte(')')
				}
				return nil
			}
			return fmt.Errorf("unsupported cond (non-zero) type: %T, err: %w", sub, err)
		}
	}

	cnt := 0
	if num != 1 {
		clause.WriteByte('(')
	}
	for k, v := range condMap {
		clause.WriteByte('(')
		if k == condKeySub || k == condKeyModel {
			return fmt.Errorf("unsupported clause key: %s", k)
		} else if v == nil {
			clause.WriteString(k)
		} else {
			condOp := ""
			for _, op := range condOps {
				if strings.HasSuffix(k, op) {
					condOp = op
					break
				}
			}

			if strings.HasSuffix(condOp, condOpSuffixBetween) {
				subArgs := interfaceSlice(v)
				if len(subArgs) != 2 {
					return errors.New("BETWEEN expects 2 args")
				}
				clause.WriteString(k)
				clause.WriteString(" ? AND ?")
				*args = append(*args, subArgs...)
			} else if strings.HasSuffix(condOp, condOpSuffixIn) {
				subArgs := interfaceSlice(v)
				if len(subArgs) == 0 {
					clause.WriteString("1=0")
				} else {
					clause.WriteString(k)
					clause.WriteString(" (?")
					for i := 0; i < len(subArgs)-1; i++ {
						clause.WriteString(",?")
					}
					clause.WriteByte(')')
					*args = append(*args, subArgs...)
				}
			} else {
				clause.WriteString(k)
				if condOp == "" {
					clause.WriteByte('=')
				}
				clause.WriteByte('?')
				*args = append(*args, v)
			}
		}
		clause.WriteByte(')')
		if cnt != num-1 {
			clause.WriteString(" AND ")
		}
		cnt++
	}
	if num != 1 {
		clause.WriteByte(')')
	}
	return nil
}

func Cond(sub any) CondMap {
	// model can only be converted to map in ToClause
	return CondMap{condKeySub: sub}
}

func CondModel(sub any) CondMap {
	// model can only be converted to map in ToClause
	return CondMap{condKeyModel: sub}
}

type condArgs struct {
	sql  string
	args []any
}

func (c condArgs) ToClause(orm *Orm, clause *strings.Builder, args *[]any) error {
	clause.WriteByte('(')
	clause.WriteString(c.sql)
	clause.WriteByte(')')
	*args = append(*args, c.args...)
	return nil
}

func CondArgs(sql string, args ...any) CondMap {
	// model can only be converted to map in ToClause
	return CondMap{condKeySub: &condArgs{
		sql:  sql,
		args: args,
	}}
}
