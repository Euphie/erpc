package erpc

import (
	"net"
	"sync"
	"time"
)

// ClientOptions RPC客户端选项
type ClientOptions struct {
	Address  string
	Protocol *Protocol
	Timeout  time.Duration
}

// Client RPC客户端
type Client struct {
	options *ClientOptions
	mutex   sync.Mutex
	conn    net.Conn
	pool    map[uint64]*Call
	seq     uint64
}

// Call RPC调用
type Call struct {
	Req   *Request
	Resp  *Response
	Done  chan *Call
	Error error
}

func (client *Client) dispatch() {
	for {
		resp, err := client.options.Protocol.Codec.getResponse(client.conn)
		if err != nil {
			Error("获取响应失败: %s", err.Error())
			continue
		}
		call, ok := client.pool[resp.Seq]
		if !ok {
			//可能发送就失败了，或者服务端错误，先忽略
			continue
		} else {
			call.Resp = &resp
		}
		call.done()
	}
}

func (c *Call) done() {
	select {
	case c.Done <- c:
	}
}

func (client *Client) request(req *Request) *Call {
	call := new(Call)
	client.mutex.Lock()
	client.seq++
	client.pool[client.seq] = call
	client.mutex.Unlock()
	req.Seq = client.seq
	call.Req = req
	call.Done = make(chan *Call)
	err := client.options.Protocol.Codec.sendRequest(client.conn, *req)
	if err != nil {
		call.Error = err
		call.done()
	}

	return call
}

func (client *Client) call(req *Request) (resp Response, err error) {
	call := new(Call)
	done := false
	for done == false {
		select {
		case call = <-client.request(req).Done:
			done = true
		case <-time.After(time.Second * client.options.Timeout):
			done = true
		}
	}

	return *call.Resp, call.Error
}

// Call 调用RPC方法
func (client *Client) Call(serviceName string, methodName string, params ...interface{}) (resp Response, err error) {
	req := new(Request)
	req.ServiceName = serviceName
	req.MethodName = methodName
	req.Params = make([]RequestParam, len(params))
	for i, p := range params {
		rp, err := GetRequestParam(p)
		if err == nil {
			req.Params[i] = rp
		} else {
			return resp, err
		}
	}
	return client.call(req)
}

// NewClient 实例化一个RPC客户端
func NewClient(options *ClientOptions) (client *Client, err error) {
	client = new(Client)
	client.options = options
	client.pool = make(map[uint64]*Call)
	client.conn, err = net.Dial("tcp", options.Address)
	if err != nil {
		return nil, err
	}
	go client.dispatch()
	return
}
