package erpc

import "fmt"

// Log 记录日志
func Log(format string, a ...interface{}) {
	fmt.Printf(format+"\n", a...)
}
