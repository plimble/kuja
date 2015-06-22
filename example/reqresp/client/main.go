package main

import (
	"fmt"
	"github.com/plimble/kuja/client"
	"github.com/plimble/kuja/example/reqresp"
	"time"
)

func main() {
	c, err := client.New("http://127.0.0.1:3000")
	if err != nil {
		panic(err)
	}
	defer c.Close()

	for i := 0; i < 1000; i++ {
		time.Sleep(time.Millisecond * 500)
		resp := &reqresp.AddResp{}
		status, err := c.Request("AddService", "Add", reqresp.AddReq{A: 5, B: 3}, resp, nil)
		fmt.Println(resp, status, err)
	}
	resp := &reqresp.AddResp{}
	status, err := c.Request("AddService", "Add", reqresp.AddReq{A: 5, B: 3}, resp, nil)

	fmt.Println(resp)   // C:8
	fmt.Println(status) // 200
	fmt.Println(err)    // nil

	// Async Requests
	resp1 := &reqresp.AddResp{}
	resp2 := &reqresp.AddResp{}
	resp3 := &reqresp.AddResp{}
	respinfos := c.AsyncRequests([]client.AsyncRequest{
		{Service: "AddService", Method: "Add", Request: &reqresp.AddReq{A: 11, B: 3}, Response: resp1, Metadata: nil},
		{Service: "AddService", Method: "Add", Request: &reqresp.AddReq{A: 10, B: 3}, Response: resp2, Metadata: nil},
		{Service: "AddService", Method: "Add", Request: &reqresp.AddReq{A: 15, B: 3}, Response: resp3, Metadata: nil},
	})

	fmt.Println(respinfos) // [{200,nil},{200,nil},{200,nil}]
	fmt.Println(resp1)     // C:14
	fmt.Println(resp2)     // C:13
	fmt.Println(resp3)     // C:18
}
