package main

import (
	"github.com/plimble/kuja"
	"github.com/plimble/kuja/encoder/json"
	"github.com/plimble/kuja/registry/etcd"
)

type ServiceTest struct{}

func (s *ServiceTest) Add(ctx *kuja.Ctx, req *AddReq, resp *AddResp) error {
	resp.C = req.A + req.B + 10

	return nil
}

func main() {

	s := kuja.NewServer()
	s.Service(&ServiceTest{})
	s.Snappy(true)
	s.Encoder(json.NewEncoder())
	s.Registry(etcd.NewRegistry("/jack", []string{"http://plimble.com:4001"}))

	s.Run(":3000", 0)
}
