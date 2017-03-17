package config

import (
	"time"

	"github.com/euphie/erpc"
	"github.com/euphie/erpc/protocol"
)

// ServerConfig ServerConfig
type ServerConfig struct {
	Address  string `erpc:"server:address"`
	Protocol string `erpc:"server:protocol"`
}

// ClientConfig ClientConfig
type ClientConfig struct {
	Address  string `erpc:"client:address"`
	Protocol string `erpc:"client:protocol"`
	Timeout  int    `erpc:"client:timeout"`
}

// GetServerOptions GetServerOptions
func GetServerOptions(file string) (so *erpc.ServerOptions) {
	conf := NewConfig()
	if err := conf.Parse(file); err != nil {
		panic(err)
	}

	sc := new(ServerConfig)
	if err := conf.Unmarshal(sc); err != nil {
		panic(err)
	}

	so = new(erpc.ServerOptions)
	so.Address = sc.Address
	switch sc.Protocol {
	default:
		so.Protocol = new(erpc.Protocol)
		so.Protocol.Codec = new(protocol.JSONCodec)
		so.Protocol.Name = "json"
		so.Protocol.Version = "1"
	}
	return
}

// GetClientOptions GetClientOptions
func GetClientOptions(file string) (co *erpc.ClientOptions) {
	conf := NewConfig()
	if err := conf.Parse(file); err != nil {
		panic(err)
	}

	cc := new(ClientConfig)
	if err := conf.Unmarshal(cc); err != nil {
		panic(err)
	}

	co = new(erpc.ClientOptions)
	co.Address = cc.Address
	co.Timeout = time.Duration(cc.Timeout)
	switch cc.Protocol {
	default:
		co.Protocol = new(erpc.Protocol)
		co.Protocol.Codec = new(protocol.JSONCodec)
		co.Protocol.Name = "json"
		co.Protocol.Version = "1"
	}
	return
}
