package fmutil

import "time"

func SqlDatetime(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format("2006-01-02 15-04-05")
}
