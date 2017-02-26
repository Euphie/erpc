package logger

import "fmt"

// SimpleLogger 简单日志
type SimpleLogger struct {
}

// Info Info
func (sl *SimpleLogger) Info(format string, a ...interface{}) {
	fmt.Printf(format+"\n", a...)
}

// Error Error
func (sl *SimpleLogger) Error(format string, a ...interface{}) {
	fmt.Printf(format+"\n", a...)
}

// Debug Debug
func (sl *SimpleLogger) Debug(format string, a ...interface{}) {
	fmt.Printf(format+"\n", a...)
}

// Warn Warn
func (sl *SimpleLogger) Warn(format string, a ...interface{}) {
	fmt.Printf(format+"\n", a...)
}
