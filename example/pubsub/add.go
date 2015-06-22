package pubsub

import (
	"github.com/plimble/kuja"
	"github.com/plimble/kuja/client"
)

type AddReq struct {
	A int
	B int
}

type AddResp struct {
	C int
}

type AddService struct {
	Client client.Client
}

func (s *AddService) Add(ctx *kuja.Context, req *AddReq, resp *AddResp) error {
	resp.C = req.A + req.B
	s.Client.Publish("AddService.added", req, nil)
	return nil
}
