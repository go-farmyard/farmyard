package fmlog

import "time"

type Facade struct {
	logger *Logger
	fields []any
}

func (l *Facade) newRecord(level Level) *Record {
	r := &Record{
		logger: l.logger,
		level:  level,
		time:   time.Now(),
		fields: l.fields,
	}
	return r
}

func (l *Facade) Tracef(msg string, args ...any) bool {
	return l.newRecord(LevelTrace).Msgf(msg, args...)
}

func (l *Facade) Debugf(msg string, args ...any) bool {
	return l.newRecord(LevelDebug).Msgf(msg, args...)
}

func (l *Facade) Infof(msg string, args ...any) bool {
	return l.newRecord(LevelInfo).Msgf(msg, args...)
}

func (l *Facade) Warnf(msg string, args ...any) bool {
	return l.newRecord(LevelWarn).Msgf(msg, args...)
}

func (l *Facade) Errorf(msg string, args ...any) bool {
	return l.newRecord(LevelError).Msgf(msg, args...)
}

func (l *Facade) Fatalf(msg string, args ...any) bool {
	return l.newRecord(LevelFatal).Msgf(msg, args...)
}

func (l *Facade) Panicf(msg string, args ...any) bool {
	return l.newRecord(LevelPanic).Msgf(msg, args...)
}

func (l *Facade) WithFields(fields ...any) *Facade {
	return &Facade{
		logger: l.logger,
		fields: append(l.fields, fields...),
	}
}

func (l *Facade) TraceFields(msg string, fields ...any) bool {
	return l.newRecord(LevelTrace).MsgFields(msg, fields...)
}

func (l *Facade) DebugFields(msg string, fields ...any) bool {
	return l.newRecord(LevelDebug).MsgFields(msg, fields...)
}

func (l *Facade) InfoFields(msg string, fields ...any) bool {
	return l.newRecord(LevelInfo).MsgFields(msg, fields...)
}

func (l *Facade) WarnFields(msg string, fields ...any) bool {
	return l.newRecord(LevelWarn).MsgFields(msg, fields...)
}

func (l *Facade) ErrorFields(msg string, fields ...any) bool {
	return l.newRecord(LevelError).MsgFields(msg, fields...)
}

func (l *Facade) FatalFields(msg string, fields ...any) bool {
	return l.newRecord(LevelFatal).MsgFields(msg, fields...)
}

func (l *Facade) PanicFields(msg string, fields ...any) bool {
	return l.newRecord(LevelPanic).MsgFields(msg, fields...)
}

func (l *Facade) Level(level Level) *Record {
	return l.newRecord(level)
}
