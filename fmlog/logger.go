package fmlog

import (
	"github.com/go-farmyard/farmyard/fmutil"
	"os"
	"strings"
	"sync"
)

type Logger struct {
	*Facade
	parent *Logger

	name     string
	fullName string
	level    Level

	mu         sync.RWMutex
	outputers  []Outputer
	subLoggers map[string]*Logger
}

var _ GeneralLogger = (*Logger)(nil)

func newLogger(parent *Logger, name string) *Logger {
	l := &Logger{
		parent: parent,
		name:   name,
		level:  LevelInfo,
	}
	l.Facade = &Facade{
		logger: l,
	}
	if parent != nil {
		l.level = parent.level
		l.fullName = parent.fullName + fmutil.Iif(parent.fullName == "", "", ".") + name
	} else {
		l.fullName = name
	}
	return l
}

func (l *Logger) processRecord(r *Record) bool {
	l1 := l
	for l1 != nil {
		if !r.level.IsSevererOrEqual(l1.level) {
			l1 = nil
			break
		}
		if l1.outputers != nil {
			break
		}
		l1 = l1.parent
	}
	if l1 != nil {
		l1.mu.RLock()
		for _, o := range l1.outputers {
			o.Output(r)
		}
		l1.mu.RUnlock()
	}

	if r.level == LevelFatal {
		os.Exit(1)
	} else if r.level == LevelPanic {
		panic(r.msgFormatted())
	}

	return true
}

func (l *Logger) SubLogger(name string) *Logger {
	if pos := strings.Index(name, "."); pos >= 0 {
		return l.SubLogger(name[:pos]).SubLogger(name[pos+1:])
	} else {
		var sub *Logger
		ok := false

		l.mu.RLock()
		if l.subLoggers != nil {
			sub, ok = l.subLoggers[name]
		}
		l.mu.RUnlock()
		if !ok {
			l.mu.Lock()
			if l.subLoggers == nil {
				l.subLoggers = map[string]*Logger{}
			}
			if sub, ok = l.subLoggers[name]; !ok {
				sub = newLogger(l, name)
				l.subLoggers[name] = sub
			}
			l.mu.Unlock()
		}
		return sub
	}
}

func (l *Logger) AddOutputer(o Outputer) {
	l.mu.Lock()
	l.outputers = append(l.outputers, o)
	l.mu.Unlock()
}

func (l *Logger) SetLevel(level Level) {
	l.level = level
}

func (l *Logger) SetLevelName(levelName string) {
	level, ok := nameLevelMap[strings.ToLower(levelName)]
	fmutil.MustTrue(ok, "invalid log level name: %s", levelName)
	l.SetLevel(level)
}

func (l *Logger) GetLevel() Level {
	return l.level
}

func (l *Logger) IsTrace() bool {
	return LevelTrace.IsSevererOrEqual(l.level)
}

func (l *Logger) IsDebug() bool {
	return LevelDebug.IsSevererOrEqual(l.level)
}
