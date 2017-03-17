# erpc

go get -u -v github.com/euphie/erpc


* server

```
package main

import (
	"fmt"

	"github.com/euphie/erpc"
	"github.com/euphie/erpc/config"
)

// S S
type S struct {
}

// M1 M1
func (s S) M1(a int, b string) (resp erpc.Response) {
	resp = erpc.Response{}
	resp.Code = 10000
	resp.Message = fmt.Sprintf("M1方法调用成功, 参数的值:%d,%s", a, b)
	resp.Data = struct{}{}
	return
}

func (s S) M2(a float32, b bool) (resp erpc.Response) {
	resp = erpc.Response{}
	resp.Code = 10000
	resp.Message = fmt.Sprintf("M2方法调用成功, 参数的值:%v,%v", a, b)
	resp.Data = struct{}{}
	return
}

func main() {
	conf := config.GetServerOptions("./erpc.conf")
	rpc := erpc.NewServer(conf)
	rpc.Register(S{}, "AAA")
	rpc.Start()
}

```

* client

```
package main

import "fmt"
import "github.com/euphie/erpc"
import "github.com/euphie/erpc/config"

var pf = fmt.Printf
var pl = fmt.Println
var spf = fmt.Sprintf
var pt = func(i interface{}) {
	pf("%T", i)
}

func main() {
	conf := config.GetClientOptions("./erpc.conf")
	c, err := erpc.NewClient(conf)
	if err != nil {
		pf("错误:%v", err.Error())
		return
	}

	var a float64 = 3.1
	b := false
	resp, err := c.Call("AAA", "M2", a, b)
	pl(resp)
	resp, err = c.Call("AAA", "M2", a, b)
	pl(resp)
}


```