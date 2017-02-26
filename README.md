# erpc

go get -u -v github.com/euphie/erpc

```
package main

import (
	"fmt"

	erpc "github.com/euphie/erpc"
)

// S S
type S struct {
}

// M1 M1
func (s S) M1(a int64, b string) (resp erpc.Response) {
	resp = erpc.Response{}
	resp.Code = 10000
	resp.Message = fmt.Sprintf("M1方法调用成功, 参数的值:%d,%s", a, b)
	resp.Data = struct{}{}
	return
}

func main() {
	rpc := erpc.GetDefaultServer()
	rpc.Register(S{}, "AAA")
	rpc.Start()
}

```