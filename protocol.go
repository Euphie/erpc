package erpc

import (
	"errors"
	"fmt"
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
	GetRequest(conn net.Conn) (req Request, err error)
	GetResponse(conn net.Conn) (resp Response, err error)
	SendRequest(conn net.Conn, req Request) (err error)
	SendResponse(conn net.Conn, resp Response) (err error)
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
