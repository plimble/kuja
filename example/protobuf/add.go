package protobuf

import (
	"github.com/plimble/kuja"
)

type AddService struct{}

func (s *AddService) Add(ctx *kuja.Context, req *AddReq, resp *AddResp) error {
	resp.C = req.A + req.B
	return nil
}