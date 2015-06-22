package main

import (
	"fmt"
	"github.com/plimble/kuja/client"
	"github.com/plimble/kuja/example/middleware"
)

func main() {
	c, err := client.New("http://127.0.0.1:3000")
	if err != nil {
		panic(err)
	}
	defer c.Close()

	resp := &middleware.AddResp{}
	status, err := c.Request("AddService", "Add", middleware.AddReq{A: 5, B: 3}, resp, nil)

	fmt.Println(resp)   // C:8
	fmt.Println(status) // 200
	fmt.Println(err)    // nil

	// Async Requests
	resp1 := &middleware.AddResp{}
	resp2 := &middleware.AddResp{}
	resp3 := &middleware.AddResp{}
	respinfos := c.AsyncRequests([]client.AsyncRequest{
		{Service: "AddService", Method: "Add", Request: &middleware.AddReq{A: 11, B: 3}, Response: resp1, Metadata: nil},
		{Service: "AddService", Method: "Add", Request: &middleware.AddReq{A: 10, B: 3}, Response: resp2, Metadata: nil},
		{Service: "AddService", Method: "Add", Request: &middleware.AddReq{A: 15, B: 3}, Response: resp3, Metadata: nil},
	})

	fmt.Println(respinfos) // [{200,nil},{200,nil},{200,nil}]
	fmt.Println(resp1)     // C:14
	fmt.Println(resp2)     // C:13
	fmt.Println(resp3)     // C:18
}
