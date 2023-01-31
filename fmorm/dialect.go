package fmorm

type Dialect interface {
	QuoteName(string) string
	NameQuoter() (string, string)
}

type DialectNop struct {
}

func (d *DialectNop) NameQuoter() (string, string) {
	return "", ""
}

func (d *DialectNop) QuoteName(s string) string {
	return s
}

type DialectMySQL struct {
}

func (d *DialectMySQL) NameQuoter() (string, string) {
	return "`", "`"
}

func (d *DialectMySQL) QuoteName(s string) string {
	return "`" + s + "`"
}
