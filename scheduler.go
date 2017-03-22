package erpc

// Scheduler 调度器
type Scheduler interface {
	GetServiceRegisterFunc() *ServiceRegisterFunc
	GetClient(serviceName string) (c *Client, err error)
	Call(serviceName string, methodName string, params ...interface{}) (resp Response, err error)
}
