package erpc

import (
	"fmt"
	"io"
	"net"
	"reflect"
	"sync"
)

// ServiceRegisterFunc 服务注册后候执行的方法
type ServiceRegisterFunc func(serviceName string) error

// ServerOptions PRC服务器选项
type ServerOptions struct {
	Address             string
	Protocol            *Protocol
	ServiceRegisterFunc ServiceRegisterFunc
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
	options    *ServerOptions
	listener   *net.Listener
	serviceMap map[string]*Service
}

// NewServer 新建一个RPC服务器
func NewServer(options *ServerOptions) (server *Server) {
	server = new(Server)
	server.options = options
	server.serviceMap = make(map[string]*Service)
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
	if err := server.options.ServiceRegisterFunc(name); err != nil {
		Error("服务注册失败: %s", err.Error())
		return
	}
	server.serviceMap[name] = _service
	Info("服务 %s 注册成功", name)
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
			if err != io.EOF {
				Error(err.Error())
			}
			conn.Close()
			break
		}
	}
}

func (server *Server) execute(conn net.Conn) (err error) {
	req, err := server.options.Protocol.Codec.GetRequest(conn)
	if err != nil {
		if err != io.EOF {
			err = fmt.Errorf("获取请求失败: %s", err.Error())
		}
		return
	}
	defer func() {
		if p := recover(); p != nil {
			resp := new(Response)
			resp.Code = -10000
			resp.Message = fmt.Sprintf("方法调用失败: %v", p)
			resp.Seq = req.Seq
			server.response(conn, resp)
		}
	}()
	service, ok := server.serviceMap[req.ServiceName]
	if !ok {
		err = fmt.Errorf("服务不存在: %s", req.ServiceName)
	}
	method, ok := service.methodMap[req.MethodName]
	fmt.Printf("%v", service.methodMap)
	if !ok {
		err = fmt.Errorf("方法不存在: %s", req.MethodName)
	}

	params := make([]reflect.Value, len(req.Params))
	for i, p := range req.Params {
		if !ok {
			err = fmt.Errorf("参数类型不存在: %s", p.Type)
		}
		params[i] = reflect.ValueOf(p.GetValue())
	}
	resps := method.rvalue.Call(params)
	resp := resps[0].Interface().(Response)
	resp.Seq = req.Seq
	return server.response(conn, &resp)
}

func (server *Server) response(conn net.Conn, resp *Response) (err error) {
	err = server.options.Protocol.Codec.SendResponse(conn, *resp)
	Info("%v", *resp)
	return
}
