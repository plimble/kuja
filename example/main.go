package main

import (
	"github.com/Sirupsen/logrus"
	"github.com/plimble/kuja"
	"github.com/plimble/kuja/encoder/json"
	"github.com/plimble/kuja/registry/consul"
	// "github.com/plimble/kuja/registry/etcd"
)

type ServiceTest struct{}

func (s *ServiceTest) Add(ctx *kuja.Context, req *AddReq, resp *AddResp) error {
	resp.C = req.A + req.B + 10

	return nil
}

func main() {
	s := kuja.NewServer()
	s.Service(&ServiceTest{})
	s.Snappy(true)
	s.Encoder(json.NewEncoder())

	// reg := etcd.NewRegistry("/jack6", []string{"http://plimble.com:4001"})
	// s.Registry(reg)
	// reg.Watch()

	reg := consul.NewRegistry([]string{"http://plimble.com:8500"})
	s.Registry(reg)
	reg.Watch()

	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
		DisableColors: false,
	})

	s.Run(":3002", 0)
}
