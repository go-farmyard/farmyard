package fmlog

var DefaultLogger *Logger

func init() {
	InitDefault("app")
}

func Tracef(msg string, args ...any) bool {
	return DefaultLogger.Tracef(msg, args...)
}
func Debugf(msg string, args ...any) bool {
	return DefaultLogger.Debugf(msg, args...)
}
func Infof(msg string, args ...any) bool {
	return DefaultLogger.Infof(msg, args...)
}
func Warnf(msg string, args ...any) bool {
	return DefaultLogger.Warnf(msg, args...)
}
func Errorf(msg string, args ...any) bool {
	return DefaultLogger.Errorf(msg, args...)
}
func Fatalf(msg string, args ...any) bool {
	return DefaultLogger.Fatalf(msg, args...)
}
func Panicf(msg string, args ...any) bool {
	return DefaultLogger.Panicf(msg, args...)
}

func IsTrace() bool {
	return DefaultLogger.IsTrace()
}

func IsDebug() bool {
	return DefaultLogger.IsDebug()
}

func InitDefault(name string) {
	DefaultLogger = newLogger(nil, name)
	DefaultLogger.AddOutputer(NewConsoleOutputer())
}
