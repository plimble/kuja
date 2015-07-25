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

func (s *AddService) Add(ctx *kuja.Context, req *AddReq) (*AddResp, error) {
	resp := &AddResp{}
	resp.C = req.A + req.B
	return resp, nil
}
