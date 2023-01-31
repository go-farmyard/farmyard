package fmlog

type GeneralLogger interface {
	Tracef(msg string, args ...any) bool
	Debugf(msg string, args ...any) bool
	Infof(msg string, args ...any) bool
	Warnf(msg string, args ...any) bool
	Errorf(msg string, args ...any) bool
	Fatalf(msg string, args ...any) bool
	Panicf(msg string, args ...any) bool
}
