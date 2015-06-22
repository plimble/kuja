package main

import (
	"fmt"
	"github.com/plimble/kuja/client"
	"github.com/plimble/kuja/encoder/proto"
	"github.com/plimble/kuja/example/protobuf"
)

func main() {
	c, err := client.New("http://127.0.0.1:3000", client.Encoder(proto.NewEncoder()))
	if err != nil {
		panic(err)
	}
	defer c.Close()

	// Request
	resp := &protobuf.AddResp{}
	status, err := c.Request("AddService", "Add", &protobuf.AddReq{A: 5, B: 3}, resp, nil)

	fmt.Println(resp)   // C:8
	fmt.Println(status) // 200
	fmt.Println(err)    // nil

	// Async Requests
	resp1 := &protobuf.AddResp{}
	resp2 := &protobuf.AddResp{}
	resp3 := &protobuf.AddResp{}
	respinfos := c.AsyncRequests([]client.AsyncRequest{
		{Service: "AddService", Method: "Add", Request: &protobuf.AddReq{A: 11, B: 3}, Response: resp1, Metadata: nil},
		{Service: "AddService", Method: "Add", Request: &protobuf.AddReq{A: 10, B: 3}, Response: resp2, Metadata: nil},
		{Service: "AddService", Method: "Add", Request: &protobuf.AddReq{A: 15, B: 3}, Response: resp3, Metadata: nil},
	})

	fmt.Println(respinfos) // [{200,nil},{200,nil},{200,nil}]
	fmt.Println(resp1)     // C:14
	fmt.Println(resp2)     // C:13
	fmt.Println(resp3)     // C:18
}
