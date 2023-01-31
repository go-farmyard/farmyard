package fmorm

import (
	"errors"
	"strings"
)

type SqlCmd struct {
	Query strings.Builder
	Args  []any
}

func (sc *SqlCmd) Append(strs ...string) *SqlCmd {
	for _, s := range strs {
		sc.Query.WriteString(s)
	}
	return sc
}

func (sc *SqlCmd) AppendStrings(strs []string, optSep ...string) *SqlCmd {
	sep := ""
	if len(optSep) == 1 {
		sep = optSep[0]
	}
	for i, s := range strs {
		sc.Query.WriteString(s)
		if sep != "" && i != len(strs)-1 {
			sc.Query.WriteString(sep)
		}
	}
	return sc
}

func (sc *SqlCmd) AppendClause(orm *Orm, c ClauseProvider) error {
	if c == nil {
		return errors.New("clause is nil")
	}
	return c.ToClause(orm, &sc.Query, &sc.Args)
}
