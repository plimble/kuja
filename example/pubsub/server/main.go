package main

import (
	"github.com/plimble/kuja"
	"github.com/plimble/kuja/broker/rabbitmq"
	"github.com/plimble/kuja/example/pubsub"
	"time"
)

func main() {
	server := kuja.NewServer()
	server.Service(&pubsub.AddService{})
	server.Broker(rabbitmq.NewBroker("amqp://guest:guest@plimble.com:5672/"))
	sub := &pubsub.Subscibers{}

	// All subscriptions with the same queue name will form a queue group.
	// Each message will be delivered to only one subscriber per queue group,
	// using queuing semantics. You can have as many queue groups as you wish.
	// Normal subscribers will continue to work as expected.
	server.Subscribe("AddService.added", "add", 0, sub.Add)
	server.Subscribe("AddService.added", "sub", 0, sub.Sub)
	server.Subscribe("AddService.added", "multiply", 0, sub.Multiply)
	server.Subscribe("AddService.added", "devide", 0, sub.Divide)
	server.Subscribe("AddService.added", "long", time.Second*2, sub.Longrunning)

	// 10 workers for each subscription
	server.SubscribeSize(10)

	server.Snappy(true)

	server.Run(":3000", 0)
}
