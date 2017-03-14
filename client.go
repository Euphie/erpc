package erpc

import (
	"net"
	"sync"
	"time"
)

// ClientOptions RPC客户端选项
type ClientOptions struct {
	Retries  int
	Timeout  time.Duration
	Address  string
	Port     int
	Protocol *Protocol
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

func (c *Client) dispatch() {
	for {
		resp, err := c.options.Protocol.Codec.getResponse(c.conn)
		if err != nil {
			Error("获取响应失败: %s", err.Error())
			continue
		}
		call, ok := c.pool[resp.Seq]
		call.Resp = &resp
		if !ok {
			//可能发送就失败了，或者服务端错误，先忽略
			continue
		}
		call.done()
	}
}

func (c *Call) done() {
	select {
	case c.Done <- c:
	}
}

func (c *Client) request(req *Request) *Call {
	call := new(Call)
	c.mutex.Lock()
	c.seq++
	c.pool[c.seq] = call
	c.mutex.Unlock()
	req.Seq = c.seq
	call.Req = req
	call.Done = make(chan *Call)
	err := c.options.Protocol.Codec.sendRequest(c.conn, *req)
	if err != nil {
		call.Error = err
		call.done()
	}

	return call
}

func (c *Client) call(req *Request) (resp Response, err error) {
	call := new(Call)
	done := false
	for done == false {
		select {
		case call = <-c.request(req).Done:
			done = true
		case <-time.After(time.Second * c.options.Timeout):
			done = true
		}
	}

	return *call.Resp, call.Error
}

// Call 调用RPC方法
func (c *Client) Call(serviceName string, methodName string, params ...interface{}) (resp Response, err error) {
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
	return c.call(req)
}

// NewClient 实例化一个RPC客户端
func NewClient(options *ClientOptions) (c *Client, err error) {
	c = new(Client)
	options.Address = "127.0.0.1:9999"
	options.Timeout = 2
	options.Protocol = &Protocol{Codec: &JSONCodec{}}
	c.options = options
	c.pool = make(map[uint64]*Call)
	c.conn, err = net.Dial("tcp", options.Address)
	if err != nil {
		return nil, err
	}
	go c.dispatch()
	return
}
