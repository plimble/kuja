package main

import (
	c "github.com/plimble/kuja/context"
	"github.com/plimble/kuja/encoder/json"
	"github.com/plimble/kuja/server"
	"golang.org/x/net/context"
	"net/http"
)

type ServiceTest struct{}

func (s *ServiceTest) Add(ctx context.Context, req *AddReq, resp *AddResp) error {
	resp.C = req.A + req.B + 10

	return nil
}

func main() {

	s := server.NewServer()
	s.Register(&ServiceTest{})
	s.Encoder(json.NewEncoder())

	http.ListenAndServe(":3000", s)

}
