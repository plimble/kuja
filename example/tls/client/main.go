package main

import (
	"crypto/tls"
	"fmt"
	"github.com/plimble/kuja/client"
	tlsexample "github.com/plimble/kuja/example/tls"
)

func main() {
	tlsConfig := &tls.Config{}
	tlsConfig.Certificates = make([]tls.Certificate, 1)
	tlsConfig.Certificates[0], _ = tls.LoadX509KeyPair("../cert.pem", "../key.pem")
	tlsConfig.InsecureSkipVerify = true

	c := client.New("https://127.0.0.1:3000", tlsConfig)

	resp := &tlsexample.AddResp{}
	status, err := c.Request("AddService", "Add", tlsexample.AddReq{A: 5, B: 3}, resp, nil)

	fmt.Println(resp)   // C:8
	fmt.Println(status) // 200
	fmt.Println(err)    // nil

	// Async Requests
	resp1 := &tlsexample.AddResp{}
	resp2 := &tlsexample.AddResp{}
	resp3 := &tlsexample.AddResp{}
	respinfos := c.AsyncRequests([]client.AsyncRequest{
		{Service: "AddService", Method: "Add", Request: &tlsexample.AddReq{A: 11, B: 3}, Response: resp1, Metadata: nil},
		{Service: "AddService", Method: "Add", Request: &tlsexample.AddReq{A: 10, B: 3}, Response: resp2, Metadata: nil},
		{Service: "AddService", Method: "Add", Request: &tlsexample.AddReq{A: 15, B: 3}, Response: resp3, Metadata: nil},
	})

	fmt.Println(respinfos) // [{200,nil},{200,nil},{200,nil}]
	fmt.Println(resp1)     // C:14
	fmt.Println(resp2)     // C:13
	fmt.Println(resp3)     // C:18
}
