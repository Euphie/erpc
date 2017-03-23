package consul

import (
	"errors"
	"strconv"

	"github.com/euphie/erpc"
	"github.com/euphie/erpc/protocol"
	consulapi "github.com/hashicorp/consul/api"
)

func NewScheduler(file string) (sc *Scheduler) {
	conf := erpc.NewConfig()
	if err := conf.Parse(file); err != nil {
		panic(err)
	}

	sc = new(Scheduler)
	if err := conf.Unmarshal(sc); err != nil {
		panic(err)
	}

	return
}

// Scheduler Scheduler
type Scheduler struct {
	Protocol      string `erpc:"server:protocol"`
	ServerAddress string `erpc:"server:address"`
	CheckTimeout  string `erpc:"check:timeout"`
	CheckInterval string `erpc:"check:interval"`
	ConsulAddress string `erpc:"consul:address"`
}

func (sc *Scheduler) GetServerOptions() *erpc.ServerOptions {
	so := new(erpc.ServerOptions)
	so.Address = sc.ServerAddress
	so.ServiceRegisterFunc = sc.GetServiceRegisterFunc()
	so.Protocol = new(erpc.Protocol)
	so.Protocol.Codec = new(protocol.JSONCodec)
	so.Protocol.Name = "json"
	so.Protocol.Version = "1"
	return so
}

func (c *Scheduler) GetServiceRegisterFunc() erpc.ServiceRegisterFunc {
	return func(serviceName string) (err error) {
		client := c.getConsulClient()
		//创建一个新服务。
		registration := new(consulapi.AgentServiceRegistration)
		registration.ID = serviceName
		registration.Name = serviceName
		registration.Port = 9001
		registration.Tags = []string{serviceName}
		registration.Address = "10.211.55.2"

		check := new(consulapi.AgentServiceCheck)
		check.TCP = c.ServerAddress
		check.Timeout = c.CheckTimeout
		check.Interval = c.CheckInterval
		registration.Check = check
		return client.Agent().ServiceRegister(registration)
	}
}

func (scheduler *Scheduler) getConsulClient() *consulapi.Client {
	config := consulapi.DefaultConfig()
	config.Address = scheduler.ConsulAddress
	client, err := consulapi.NewClient(config)
	if err == nil {
		return client
	}

	return nil
}

func (scheduler *Scheduler) GetClient(serviceName string) (c *erpc.Client, err error) {
	client := scheduler.getConsulClient()
	q := new(consulapi.QueryOptions)
	r, _, _ := client.Health().Service(serviceName, "", true, q)
	if len(r) != 1 {
		return nil, errors.New("no service found")
	}
	co := new(erpc.ClientOptions)
	co.Address = r[0].Service.Address + ":" + strconv.Itoa(r[0].Service.Port)
	co.Timeout = 3
	co.Protocol = new(erpc.Protocol)
	co.Protocol.Codec = new(protocol.JSONCodec)
	co.Protocol.Name = "json"
	return erpc.NewClient(co)
}

func (scheduler *Scheduler) Call(serviceName string, methodName string, params ...interface{}) (resp erpc.Response, err error) {
	client, err := scheduler.GetClient(serviceName)
	if err != nil {
		return
	}
	return client.Call(serviceName, methodName, params)
}
