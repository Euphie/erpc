package erpc

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"reflect"
	"strconv"
)

// ParamTypes 参数类型映射
var ParamTypes = make(map[string]string)

func init() {
	ParamTypes[reflect.Int.String()] = "int"
	ParamTypes[reflect.Int32.String()] = "int32"
	ParamTypes[reflect.Int64.String()] = "int64"
	ParamTypes[reflect.Float32.String()] = "float32"
	ParamTypes[reflect.Float64.String()] = "float64"
	ParamTypes[reflect.String.String()] = "string"
	ParamTypes[reflect.Bool.String()] = "bool"
}

// GetValue 获取参数值
func (param RequestParam) GetValue() (value interface{}) {
	switch param.Type {
	case "int":
		value, _ = strconv.Atoi(param.Value)
	case "int32":
		temp, _ := strconv.ParseInt(param.Value, 10, 32)
		value = int32(temp)
	case "int64":
		temp, _ := strconv.ParseInt(param.Value, 10, 32)
		value = int64(temp)
	case "float32":
		temp, _ := strconv.ParseFloat(param.Value, 32)
		value = float32(temp)
	case "float64":
		value, _ = strconv.ParseFloat(param.Value, 64)
	case "string":
		value = param.Value
	case "bool":
		value, _ = strconv.ParseBool(param.Value)
	default:
		value = nil
	}
	return
}

// GetRequestParam 参数转换成RequestParam
func GetRequestParam(value interface{}) (RequestParam, error) {
	rp := new(RequestParam)
	switch value.(type) {
	case int:
		rp.Type = "int"
		rp.Value = fmt.Sprint(value.(int))
	case int32:
		rp.Type = "int32"
		rp.Value = fmt.Sprint(value.(int32))
	case int64:
		rp.Type = "int64"
		rp.Value = fmt.Sprint(value.(int64))
	case float32:
		rp.Type = "float32"
		rp.Value = strconv.FormatFloat(float64(value.(float32)), 'f', -1, 32)
	case float64:
		rp.Type = "float64"
		rp.Value = strconv.FormatFloat(value.(float64), 'f', -1, 64)
	case string:
		rp.Type = "string"
		rp.Value = value.(string)
	case bool:
		rp.Type = "bool"
		rp.Value = fmt.Sprint(value.(bool))
	default:
		return *rp, errors.New("不支持的参数类型")
	}
	return *rp, nil
}

func checkIn(intype reflect.Type) bool {
	_, ok := ParamTypes[intype.String()]
	return ok
}

func convert(param *RequestParam) (value interface{}) {
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
	//版本号
	Version string
	//协议名称
	Name string
	//编码器，用于实现数据收发的过程
	Codec Codec
}

// Codec 编码器接口
type Codec interface {
	getRequest(r net.Conn) (req Request, err error)
	getResponse(r net.Conn) (resp Response, err error)
	sendRequest(r net.Conn, req Request) (err error)
	sendResponse(r net.Conn, resp Response) (err error)
}

// Request 请求
type Request struct {
	// 请求序列，唯一
	Seq uint64
	// 服务名称
	ServiceName string `json:"ServiceName"`
	// 方法名称
	MethodName string `json:"MethodName"`
	// 请求参数
	Params []RequestParam `json:"Params"`
}

// Response 响应，注册的方法返回值必须是Response类型
type Response struct {
	// 响应码
	Code int
	// 响应消息
	Message string
	// 结果数据
	Data interface{}
	// 客户端过来的请求序列，原样返回
	Seq uint64
}

// RequestParam 方法参数
type RequestParam struct {
	Type  string
	Value string
}

//=============实现一个简单的JSON编码器=================

// JSONCodec JSON
type JSONCodec struct {
}

func (jc *JSONCodec) getRequest(r net.Conn) (req Request, err error) {
	// 方便telnet测试，取前8个字节的字符，转成int
	buf := make([]byte, 8)
	n, _ := io.ReadFull(r, buf)
	req = Request{}
	if err == io.EOF || n == 0 {
		err = io.EOF
		return
	}
	if n < 8 {
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
	// 方便telnet测试，取前8个字节的字符，转成int
	buf := make([]byte, 8)
	n, err := io.ReadFull(r, buf)
	if err == io.EOF || n == 0 {
		err = errors.New("请求结束")
		return
	}
	resp = Response{}
	if n < 8 {
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
	err = json.Unmarshal(buf, &resp)
	if err != nil {
		err = errors.New("报文解析错误")
		return
	}
	return
}
func (jc *JSONCodec) sendRequest(r net.Conn, req Request) (err error) {
	buff, err := json.Marshal(req)
	if err != nil {
		err = errors.New("报文生成错误")
		return
	}

	pedding := ""
	slen := strconv.Itoa(len(buff))
	for len := len(slen); len < 8; len++ {
		pedding += "0"
	}
	body := pedding + slen + string(buff)
	n, err := r.Write([]byte(body))
	if err != nil || n != len(buff)+8 {
		err = errors.New("报文写入错误")
		return
	}

	return nil
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
	if n == 0 || n != ilen+8 {
		return errors.New("数据报文写入失败")
	}
	return err
}
