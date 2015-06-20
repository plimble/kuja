package pubsub

import (
	"github.com/plimble/kuja"
)

type AddReq struct {
	A int
	B int
}

type AddResp struct {
	C int
}

type AddService struct{}

func (s *AddService) Add(ctx *kuja.Context, req *AddReq, resp *AddResp) error {
	resp.C = req.A + req.B
	ctx.Publish("AddService.added", req, nil)
	return nil
}
