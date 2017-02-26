package erpc

// Logger 日志
type Logger interface {
	Info(format string, a ...interface{})
	Error(format string, a ...interface{})
	Debug(format string, a ...interface{})
	Warn(format string, a ...interface{})
}
