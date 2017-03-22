package scheduler

import (
	"errors"
	"strconv"

	"github.com/euphie/erpc"
	consulapi "github.com/hashicorp/consul/api"
)

// ConsulScheduler ConsulScheduler
type ConsulScheduler struct {
	Address  string
	Protocol *erpc.Protocol
}

// GetConsulClient GetConsulClient
func (cs *ConsulScheduler) GetConsulClient() *consulapi.Client {
	config := consulapi.DefaultConfig()
	//config.Address = "10.211.55.5:8500"
	config.Address = cs.Address
	client, err := consulapi.NewClient(config)
	if err == nil {
		return client
	}

	return nil
}

// GetServiceRegisterFunc GetServiceRegisterFunc
func (cs *ConsulScheduler) GetServiceRegisterFunc() *erpc.ServiceRegisterFunc {
	var f erpc.ServiceRegisterFunc = func(serviceName string) (err error) {
		client := cs.GetConsulClient()
		//创建一个新服务。
		registration := new(consulapi.AgentServiceRegistration)
		registration.ID = serviceName
		registration.Name = serviceName
		registration.Port = 9001
		registration.Tags = []string{serviceName}
		registration.Address = "10.211.55.2"

		//增加check。
		check := new(consulapi.AgentServiceCheck)
		check.TCP = "10.211.55.2:9001"
		//设置超时 5s。
		check.Timeout = "5s"
		//设置间隔 5s。
		check.Interval = "5s"
		//注册check服务。
		registration.Check = check
		return client.Agent().ServiceRegister(registration)
	}
	return &f
}

// GetClient GetClient
func (cs *ConsulScheduler) GetClient(serviceName string) (c *erpc.Client, err error) {
	client := cs.GetConsulClient()

	q := new(consulapi.QueryOptions)
	r, _, _ := client.Health().Service(serviceName, "", true, q)
	if len(r) != 1 {
		return nil, errors.New("no service found")
	}
	co := new(erpc.ClientOptions)
	co.Address = r[0].Service.Address + ":" + strconv.Itoa(r[0].Service.Port)
	co.Timeout = 3
	co.Protocol = cs.Protocol
	return erpc.NewClient(co)
}

// Call Call
func (cs *ConsulScheduler) Call(serviceName string, methodName string, params ...interface{}) (resp erpc.Response, err error) {
	client, err := cs.GetClient(serviceName)
	if err != nil {
		return
	}
	return client.Call(serviceName, methodName, params)
}
