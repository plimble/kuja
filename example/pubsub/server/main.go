package main

import (
	"github.com/plimble/kuja"
	"github.com/plimble/kuja/broker/rabbitmq"
	"github.com/plimble/kuja/client"
	"github.com/plimble/kuja/example/pubsub"
	"time"
)

func main() {
	c, err := client.New("http://127.0.0.1:3000", client.Broker(rabbitmq.NewBroker("amqp://guest:guest@plimble.com:5672/")))
	if err != nil {
		panic(err)
	}
	defer c.Close()

	server := kuja.NewServer()
	server.Service(&pubsub.AddService{c})
	server.Broker(rabbitmq.NewBroker("amqp://guest:guest@plimble.com:5672/"))
	sub := &pubsub.Subscibers{}

	// All subscriptions with the same queue name will form a queue group.
	// Each message will be delivered to only one subscriber per queue group,
	// using queuing semantics. You can have as many queue groups as you wish.
	// Normal subscribers will continue to work as expected.
	server.Subscribe("AddService.added", "add", 0, 1, sub.Add)
	server.Subscribe("AddService.added", "sub", 0, 1, sub.Sub)
	server.Subscribe("AddService.added", "multiply", 0, 1, sub.Multiply)
	// 10 workers for devide
	server.Subscribe("AddService.added", "devide", 0, 10, sub.Divide)
	// timeout 2 seconds, 10 workers
	server.Subscribe("AddService.added", "long", time.Second*2, 10, sub.Longrunning)

	server.Snappy(true)

	server.Run(":3000", 0)
}
