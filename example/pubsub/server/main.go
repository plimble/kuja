package main

import (
	"github.com/plimble/kuja"
	"github.com/plimble/kuja/broker/rabbitmq"
	"github.com/plimble/kuja/example/pubsub"
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
	server.Subscribe("AddService", "added", "add", sub.Add)
	server.Subscribe("AddService", "added", "sub", sub.Sub)
	server.Subscribe("AddService", "added", "multiply", sub.Multiply)
	server.Subscribe("AddService", "added", "devide", sub.Divide)

	// 10 workers for each subscription
	server.SubscribeSize(10)

	server.Snappy(true)

	server.Run(":3000", 0)
}
