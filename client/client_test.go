package client

import (
	"github.com/plimble/kuja"
	"github.com/plimble/kuja/encoder/json"
	"github.com/stretchr/testify/assert"
	"net/http/httptest"
	"testing"
)

type AddReq struct {
	A int32 `json:"A,omitempty"`
	B int32 `json:"B,omitempty"`
}

type AddResp struct {
	C int32 `json:"C,omitempty"`
}

type ServiceTest struct{}

func (s *ServiceTest) Add(ctx *kuja.Ctx, req *AddReq, resp *AddResp) error {
	resp.C = req.A + req.B + 10

	return nil
}

func TestDirectClient(t *testing.T) {
	k := kuja.NewServer()
	k.Encoder(json.NewEncoder())
	k.Service(&ServiceTest{})

	server := httptest.NewServer(k)
	defer server.Close()

	client := New(server.URL)
	client.Encoder(json.NewEncoder())

	addreq := &AddReq{5, 7}
	addresp := &AddResp{}
	status, err := client.Call("ServiceTest", "Add", addreq, addresp, nil)

	assert.NoError(t, err)
	assert.Equal(t, 200, status)
	assert.Equal(t, 22, addresp.C)
}
