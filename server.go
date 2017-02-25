package erpc

import (
	"net"
	"reflect"
	"strconv"
	"sync"
)

// Options PRC服务器选项
type Options struct {
	address  string
	port     int
	protocol *Protocol
}

// Service 服务
type Service struct {
	name   string
	rtype  reflect.Type
	rvalue reflect.Value
}

// Server PRC服务器
type Server struct {
	mutex      sync.RWMutex
	options    *Options
	listener   *net.Listener
	serviceMap map[string]*Service
}

// GetDefaultServer 创建一个默认的RPC服务器
func GetDefaultServer() (server *Server) {
	return NewServer(Options{})
}

// NewServer 新建一个RPC服务器
func NewServer(options Options) (server *Server) {
	// 默认选项
	if options.address == "" {
		options.address = "0.0.0.0"
	}

	if options.port == 0 {
		options.port = 9999
	}

	if options.protocol == nil {
		options.protocol = &Protocol{}
	}

	server = &Server{
		options:    &options,
		serviceMap: make(map[string]*Service),
	}

	return
}

// Register 注册服务
func (server *Server) Register(service interface{}, alias string) {
	server.mutex.Lock()
	defer server.mutex.Unlock()
	_service := new(Service)
	_service.rtype = reflect.TypeOf(service)
	_service.rvalue = reflect.ValueOf(service)
	name := reflect.Indirect(_service.rvalue).Type().Name()
	if name == "" {
		Log("服务注册失败")
		return
	}

	Log("注册的服务类型: %s", name)
	if alias != "" {
		Log("设置服务别名: %s", alias)
		name = alias
	}

	if _, ok := server.serviceMap[name]; ok {
		Log("服务已经注册过了: %s", name)
		return
	}

	_service.name = name

	methodNum := _service.rtype.NumMethod()
	if methodNum == 0 {
		Log("没有找到方法, 服务: %s 注册失败", name)
		return
	}
	for i := 0; i < methodNum; i++ {
		method := _service.rtype.Method(i)
		methodName := method.Name
		Log("发现方法: %s", methodName)
		methodType := method.Type
		Log("方法类型: %v", methodType)
		//Log("%v", methodType.In(0))
	}
}

// Start 启动RPC服务器
func (server *Server) Start() {
	address := server.options.address + ":" + strconv.Itoa(server.options.port)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		Log("监听失败: %s", err.Error())
	} else {
		Log("监听地址: %s", address)
	}

	server.listener = &listener
	for {
		conn, err := listener.Accept()
		if err != nil {
			Log("连接失败: %s", err.Error())
		}
		go server.handleConn(conn)
	}
}

func (server *Server) handleConn(conn net.Conn) {

}

func (server *Server) exec() {

}
