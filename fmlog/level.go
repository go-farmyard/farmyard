package fmlog

import "github.com/go-farmyard/farmyard/fmutil"

type Level int

const (
	LevelTrace Level = 10
	LevelDebug Level = 20
	LevelInfo  Level = 30
	LevelWarn  Level = 40
	LevelError Level = 50
	LevelNone  Level = 100
	LevelFatal Level = 110
	LevelPanic Level = 120
)

var nameLevelMap = map[string]Level{
	"trace":   LevelTrace,
	"debug":   LevelDebug,
	"info":    LevelInfo,
	"warn":    LevelWarn,
	"warning": LevelWarn,
	"error":   LevelError,
	"none":    LevelNone,
}

var levelNameMap = map[Level]string{
	LevelTrace: "trace",
	LevelDebug: "debug",
	LevelInfo:  "info",
	LevelWarn:  "warn",
	LevelError: "error",
	LevelNone:  "none",
}

func (l Level) IsSevererOrEqual(t Level) bool {
	return l >= t
}

func (l Level) String() string {
	name, ok := levelNameMap[l]
	fmutil.MustTrue(ok, "unknown log level: %d", l)
	return name
}
