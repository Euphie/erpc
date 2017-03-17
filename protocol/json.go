package protocol

import (
	"encoding/json"
	"errors"
	"io"
	"net"
	"strconv"

	"github.com/euphie/erpc"
)

//=============实现一个简单的JSON编码器=================

// JSONCodec JSON
type JSONCodec struct {
}

// GetRequest GetRequest
func (jc *JSONCodec) GetRequest(conn net.Conn) (req erpc.Request, err error) {
	// 方便telnet测试，取前8个字节的字符，转成int
	buf := make([]byte, 8)
	n, _ := io.ReadFull(conn, buf)
	req = erpc.Request{}
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
	n, _ = io.ReadFull(conn, buf)
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

// GetResponse GetResponse
func (jc *JSONCodec) GetResponse(conn net.Conn) (resp erpc.Response, err error) {
	// 方便telnet测试，取前8个字节的字符，转成int
	buf := make([]byte, 8)
	n, err := io.ReadFull(conn, buf)
	if err == io.EOF || n == 0 {
		err = errors.New("请求结束")
		return
	}
	resp = erpc.Response{}
	if n < 8 {
		err = errors.New("读取报文长度错误")
		return
	}
	len, _ := strconv.Atoi(string(buf))
	buf = make([]byte, len)
	n, _ = io.ReadFull(conn, buf)
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

// SendRequest SendRequest
func (jc *JSONCodec) SendRequest(conn net.Conn, req erpc.Request) (err error) {
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
	n, err := conn.Write([]byte(body))
	if err != nil || n != len(buff)+8 {
		err = errors.New("报文写入错误")
		return
	}

	return nil
}

// SendResponse SendResponse
func (jc *JSONCodec) SendResponse(conn net.Conn, resp erpc.Response) (err error) {
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
	n, err := conn.Write(buf)
	if n == 0 || n != ilen+8 {
		return errors.New("数据报文写入失败")
	}
	return err
}
