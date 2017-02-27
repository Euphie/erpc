package erpc

import (
	"encoding/json"
	"errors"
	"io"
	"net"
	"reflect"
	"strconv"
)

// ParamTypes 参数类型映射
var ParamTypes = make(map[string]reflect.Type)

func init() {
	ParamTypes["int"] = reflect.TypeOf(reflect.Int32)
	ParamTypes["long"] = reflect.TypeOf(reflect.Int64)
	ParamTypes["float"] = reflect.TypeOf(reflect.Float32)
	ParamTypes["double"] = reflect.TypeOf(reflect.Float64)
	ParamTypes["string"] = reflect.TypeOf(reflect.String)
}

func checkIn(intype reflect.Type) bool {
	switch intype {
	case reflect.TypeOf(reflect.Int32):
		return true
	case reflect.TypeOf(reflect.Int64):
		return true
	case reflect.TypeOf(reflect.Float32):
		return true
	case reflect.TypeOf(reflect.Float64):
		return true
	case reflect.TypeOf(reflect.String):
		return true

	}
	return true
}

func convert(param RequestParam) (value interface{}) {
	switch param.Type {
	case "int":
		value, _ = strconv.ParseInt(param.Value, 10, 32)
	case "float":
		value, _ = strconv.ParseFloat(param.Value, 32)
	case "long":
		value, _ = strconv.ParseInt(param.Value, 10, 64)
	case "double":
		value, _ = strconv.ParseFloat(param.Value, 64)
	case "string":
		value = param.Value
	}
	return
}

// Protocol 协议
type Protocol struct {
	Codec Codec
}

// Codec 编码器
type Codec interface {
	getRequest(r net.Conn) (req Request, err error)
	getResponse(r net.Conn) (resp Response, err error)
	sendRequest(r net.Conn) bool
	sendResponse(r net.Conn, resp Response) (err error)
}

// Request 请求
type Request struct {
	ServiceName string         `json:"ServiceName"`
	MethodName  string         `json:"MethodName"`
	Params      []RequestParam `json:"Params"`
}

// Response 请求
type Response struct {
	Code    int
	Message string
	Data    interface{}
}

// RequestParam 方法参数
type RequestParam struct {
	Type  string
	Value string
}

//=============实现一个JSON的编码器=================

// JSONCodec JSON
type JSONCodec struct {
}

func (jc *JSONCodec) getRequest(r net.Conn) (req Request, err error) {
	// 方便telnet测试，取前8个字节的字符，转成int
	buf := make([]byte, 8)
	n, _ := io.ReadFull(r, buf)
	req = Request{}
	if n < 4 {
		err = errors.New("读取报文长度错误")
		return
	}
	len, _ := strconv.Atoi(string(buf))
	buf = make([]byte, len)
	n, _ = io.ReadFull(r, buf)
	if n < len || n == 0 {
		err = errors.New("报文读取错误")
		return
	}
	err = json.Unmarshal(buf, &req)
	if err != nil {
		err = errors.New("报文解析错误")
		return
	}
	return
}

func (jc *JSONCodec) getResponse(r net.Conn) (resp Response, err error) {
	return Response{}, nil
}
func (jc *JSONCodec) sendRequest(r net.Conn) bool {
	return false
}
func (jc *JSONCodec) sendResponse(r net.Conn, resp Response) (err error) {
	bytes, err := json.Marshal(resp)
	if err != nil {
		return err
	}
	ilen := len(bytes)
	slen := strconv.Itoa(ilen)
	bit := len(slen)
	if bit > 8 {
		return errors.New("写入报文太大")
	}
	if bit < 8 {
		for i := 0; i < 8-bit; i++ {
			slen = "0" + slen
		}
	}
	buf := []byte(slen)
	for _, v := range bytes {
		buf = append(buf, v)
	}
	n, err := r.Write(buf)
	if n == 0 || n != ilen {
		return errors.New("数据报文失败")
	}
	return err
}

//=============实现一个测试的JSON的编码器=================

// TestJSONCodec JSON
type TestJSONCodec struct {
}

func (jc *TestJSONCodec) getRequest(r net.Conn) (req Request, err error) {
	str := "{\"ServiceName\":\"AAA\",\"MethodName\":\"M1\",\"Params\":[{\"Value\":\"111\",\"Type\":\"int\"},{\"Value\":\"ssbb\",\"Type\":\"string\"}]}"
	req = Request{}
	json.Unmarshal([]byte(str), &req)
	return req, nil
}

func (jc *TestJSONCodec) getResponse(r net.Conn) (resp Response, err error) {
	return Response{}, nil
}
func (jc *TestJSONCodec) sendRequest(r net.Conn) bool {
	return false
}
func (jc *TestJSONCodec) sendResponse(r net.Conn, resp Response) (err error) {
	bytes, err := json.Marshal(resp)
	if err != nil {
		return err
	}
	ilen := len(bytes)
	slen := strconv.Itoa(ilen)
	bit := len(slen)
	if bit > 8 {
		return errors.New("写入报文太大")
	}
	if bit < 8 {
		for i := 0; i < 8-bit; i++ {
			slen = "0" + slen
		}
	}
	buf := []byte(slen)
	for _, v := range bytes {
		buf = append(buf, v)
	}
	n, err := r.Write(buf)
	if n == 0 || n != ilen {
		return errors.New("数据报文失败")
	}
	return err
}
