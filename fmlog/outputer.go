package fmlog

import (
	"github.com/go-farmyard/farmyard/fmlog/termcolor"
	"github.com/go-farmyard/farmyard/fmutil"
	"io"
	"os"
	"strconv"
	"strings"
)

type Outputer interface {
	Output(r *Record)
}

type ConsoleOutputer struct {
	Writer     io.Writer
	TimeFormat string
	LevelMap   map[Level]string
}

var levelColorMap = map[Level][]byte{
	LevelTrace: termcolor.ColorBytes(termcolor.Bold, termcolor.FgCyan),
	LevelDebug: termcolor.ColorBytes(termcolor.Bold, termcolor.FgBlue),
	LevelInfo:  termcolor.ColorBytes(termcolor.Bold, termcolor.FgGreen),
	LevelWarn:  termcolor.ColorBytes(termcolor.Bold, termcolor.FgYellow),
	LevelError: termcolor.ColorBytes(termcolor.Bold, termcolor.FgRed),
	LevelFatal: termcolor.ColorBytes(termcolor.Bold, termcolor.FgRed),
	LevelPanic: termcolor.ColorBytes(termcolor.Bold, termcolor.BgRed),
}
var levelColorReset = termcolor.ColorBytes(termcolor.Reset)

func NewConsoleOutputer() *ConsoleOutputer {
	return &ConsoleOutputer{
		Writer:     os.Stdout,
		TimeFormat: "2006-01-02T15-04-05",
		LevelMap: map[Level]string{
			LevelTrace: "TRC",
			LevelDebug: "DBG",
			LevelInfo:  "INF",
			LevelWarn:  "WRN",
			LevelError: "ERR",
			LevelFatal: "FTL",
			LevelPanic: "PNC",
		},
	}
}

func readFieldPair(fields []any, pos int) (next int, k string, v any) {
	vk := fields[pos]
	if pos+1 < len(fields) {
		if s, ok := vk.(string); ok {
			return pos + 2, s, fields[pos+1]
		}
	}
	return pos + 1, "_" + strconv.Itoa(pos), vk
}

func (c *ConsoleOutputer) Output(r *Record) {
	levelStr, ok := c.LevelMap[r.level]
	if !ok {
		levelStr = "L" + strconv.Itoa(int(r.level))
	}
	buf := strings.Builder{}
	buf.WriteString(r.time.Format(c.TimeFormat))
	buf.WriteByte(' ')

	if color, ok := levelColorMap[r.level]; ok {
		buf.Write(color)
		buf.WriteString(levelStr)
		buf.Write(levelColorReset)
	} else {
		buf.WriteString(levelStr)
	}

	buf.WriteString(" {")
	buf.WriteString(r.logger.name)
	buf.WriteString("} ")
	buf.WriteString(r.msgFormatted())
	buf.WriteString(" ")
	for i := 0; i < len(r.fields); {
		var k string
		var v any
		i, k, v = readFieldPair(r.fields, i)
		buf.WriteString(" " + k + "=")
		buf.WriteString(fmutil.JsonEncodeString(v))
	}
	buf.WriteByte('\n')
	_, _ = io.WriteString(c.Writer, buf.String())
}
