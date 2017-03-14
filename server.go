package erpc

import (
	"fmt"
	"io"
	"net"
	"reflect"
	"sync"
)

// Options PRC服务器选项
type Options struct {
	Address  string
	Protocol *Protocol
}

// Service 服务
type Service struct {
	name      string
	rtype     reflect.Type
	rvalue    reflect.Value
	methodMap map[string]*SerivceMethod
}

// SerivceMethod 服务方法
type SerivceMethod struct {
	method reflect.Method
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
	if options.Address == "" {
		options.Address = "0.0.0.0:9999"
	}
	if options.Protocol == nil {
		options.Protocol = &Protocol{Codec: &JSONCodec{}}
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
		Error("服务注册失败")
		return
	}
	Debug("注册的服务类型: %s", name)
	if alias != "" {
		Info("设置服务别名: %s", alias)
		name = alias
	}
	if _, ok := server.serviceMap[name]; ok {
		Warn("服务已经注册过了: %s", name)
		return
	}
	_service.name = name
	methodNum := _service.rtype.NumMethod()
	if methodNum == 0 {
		Error("没有找到方法, 服务: %s 注册失败", name)
		return
	}
	if _service.methodMap == nil {
		_service.methodMap = make(map[string]*SerivceMethod)
	}
	for i := 0; i < methodNum; i++ {
		value := _service.rvalue.Method(i)
		method := _service.rtype.Method(i)
		incheck := true
		for j := 1; j < method.Type.NumIn(); j++ {
			intype := method.Type.In(j)
			if !checkIn(intype) {
				incheck = false
				continue
			}
		}
		if !incheck {
			continue
		}
		if method.Type.NumOut() != 1 {
			continue
		}
		if method.Type.Out(0) != reflect.TypeOf(Response{}) {
			continue
		}
		Info("发现方法: %s", method.Name)
		_service.methodMap[method.Name] = &SerivceMethod{
			rvalue: value,
			method: method,
		}
	}
	server.serviceMap[name] = _service
}

// Start 启动RPC服务器
func (server *Server) Start() {
	listener, err := net.Listen("tcp", server.options.Address)
	if err != nil {
		Error("监听失败: %s", err.Error())
	} else {
		Info("监听地址: %s", server.options.Address)
	}

	server.listener = &listener
	for {
		conn, err := listener.Accept()
		// conn.SetReadDeadline(time.Now().Add(time.Duration(60 * time.Second)))
		if err != nil {
			Error("连接失败: %s", err.Error())
		}
		go server.handleConn(conn)
	}
}

func (server *Server) handleConn(conn net.Conn) {
	for {
		err := server.execute(conn)
		if err != nil {
			break
		}
	}
}

func (server *Server) execute(conn net.Conn) error {
	req, err := server.options.Protocol.Codec.getRequest(conn)
	if err != nil {
		if err != io.EOF {
			Error("获取请求失败: %s", err.Error())
		}
		return err
	}
	service, ok := server.serviceMap[req.ServiceName]
	if !ok {
		Error("服务不存在: %s", req.ServiceName)
		return err
	}
	method, ok := service.methodMap[req.MethodName]
	fmt.Printf("%v", service.methodMap)
	if !ok {
		Error("方法不存在: %s", req.MethodName)
		return err
	}

	params := make([]reflect.Value, len(req.Params))
	for i, p := range req.Params {
		if !ok {
			Error("参数类型不存在: %s", p.Type)
			return err
		}
		params[i] = reflect.ValueOf(p.GetValue())
	}
	resps := method.rvalue.Call(params)
	resp := resps[0].Interface().(Response)
	resp.Seq = req.Seq
	server.options.Protocol.Codec.sendResponse(conn, resp)
	Info("%v", resp)
	return nil
}
