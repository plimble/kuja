package main

import (
	"errors"
	"github.com/Sirupsen/logrus"
	"github.com/plimble/kuja"
	"github.com/plimble/kuja/broker/nats"
	"github.com/plimble/kuja/encoder/gogoproto"
	"github.com/plimble/kuja/registry/consul"
	"time"
	// "github.com/plimble/kuja/registry/etcd"
)

type ServiceTest struct{}

func (s *ServiceTest) Add(ctx *kuja.Context, req *AddReq, resp *AddResp) error {
	resp.C = req.A + req.B + 10
	logrus.Info("request", req, "resp", resp)

	return nil
}

type SubTest struct{}

func (s *SubTest) Add(info *kuja.SubscriberInfo, meta kuja.Metadata, data *AddReq) error {
	logrus.Info("get", data)
	time.Sleep(time.Second * 2)
	logrus.Info(meta)
	logrus.Info(info)
	return errors.New("test error")
}

func main() {
	s := kuja.NewServer()
	s.Service(&ServiceTest{})
	s.Snappy(true)
	s.Encoder(gogoproto.NewEncoder())
	broker, err := nats.NewBroker("nats://127.0.0.1:4222")
	if err != nil {
		panic(err)
	}
	s.Broker(broker)
	s.Subscriber(&SubTest{})

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
