package erpc

import "io"

// Protocol 协议
type Protocol struct {
	codec Codec
}

// Codec 解码器
type Codec interface {
	getRequest(r io.Reader) (req Request)
}

// Request 请求
type Request struct {
	ServiceName string `json:"ServiceName"`
	MethodName  string `json:"MethodName"`
}
