package fmlog

import (
	"fmt"
	"time"
)

type Record struct {
	logger   *Logger
	time     time.Time
	level    Level
	fields   []any
	msg      string
	msgArgs  []any
	msgCache string
}

func (r *Record) msgFormatted() string {
	if r.msgCache == "" {
		if len(r.msgArgs) == 0 {
			r.msgCache = r.msg
		} else {
			r.msgCache = fmt.Sprintf(r.msg, r.msgArgs...)
		}
	}
	return r.msgCache
}

func (r *Record) MsgFields(msg string, fields ...any) bool {
	r.msgCache = ""
	r.msg = msg
	r.msgArgs = nil
	r.fields = append(r.fields, fields...)
	return r.Msgf(msg)
}

func (r *Record) Msgf(msg string, args ...any) bool {
	r.msgCache = ""
	r.msg = msg
	r.msgArgs = args
	return r.logger.processRecord(r)
}
