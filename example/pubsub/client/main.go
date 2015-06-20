package main

import (
	"fmt"
	"github.com/plimble/kuja/broker/rabbitmq"
	"github.com/plimble/kuja/client"
	"github.com/plimble/kuja/example/pubsub"
)

func main() {
	c := client.New("http://127.0.0.1:3000", nil)
	c.Broker(rabbitmq.NewBroker("amqp://guest:guest@plimble.com:5672/"))

	resp := &pubsub.AddResp{}
	status, err := c.Request("AddService", "Add", pubsub.AddReq{A: 5, B: 3}, resp, nil)

	fmt.Println(resp)   // C:8
	fmt.Println(status) // 200
	fmt.Println(err)    // nil

	c.Publish("AddService", "added", &pubsub.AddReq{A: 6, B: 6}, nil)
}
