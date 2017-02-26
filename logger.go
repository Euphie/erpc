package erpc

import (
	"github.com/euphie/erpc/logger"
)

// Logger 日志
type Logger interface {
	Info(format string, a ...interface{})
	Error(format string, a ...interface{})
	Debug(format string, a ...interface{})
	Warn(format string, a ...interface{})
}

const (
	// NONE 不记录日志
	NONE = 0
	// ERROR ERROR
	ERROR = 1
	// WARN WARN
	WARN = 2
	// INFO INFO
	INFO = 3
	// DEBUG DEBUG
	DEBUG = 4
)

var _logger Logger = &logger.SimpleLogger{}
var _level = INFO

// SetLogger 设置Logger
func SetLogger(l Logger) {
	_logger = l
}

// SetLogLevel 设置日志等级
func SetLogLevel(lv int) {
	_level = lv
}

// Error Error
func Error(format string, a ...interface{}) {
	if _level > NONE {
		_logger.Error(format, a...)
	}
}

// Warn Warn
func Warn(format string, a ...interface{}) {
	if _level > ERROR {
		_logger.Warn(format, a...)
	}
}

// Info Info
func Info(format string, a ...interface{}) {
	if _level > WARN {
		_logger.Info(format, a...)
	}
}

// Debug Debug
func Debug(format string, a ...interface{}) {
	if _level > INFO {
		_logger.Error(format, a...)
	}
}
