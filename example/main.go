package main

import (
	"errors"
	"github.com/Sirupsen/logrus"
	"github.com/plimble/kuja"
	"github.com/plimble/kuja/broker/nats"
	// "github.com/plimble/kuja/broker/nats"
	"github.com/plimble/kuja/encoder/gogoproto"
	// "github.com/plimble/kuja/registry/consul"
	// "github.com/plimble/kuja/registry/etcd"
	"github.com/streadway/amqp"
	"time"
)

type ServiceTest struct{}

func (s *ServiceTest) Add(ctx *kuja.Context, req *AddReq, resp *AddResp) error {
	resp.C = req.A + req.B + 10
	logrus.Info("request", req, "resp", resp)

	return nil
}

type SubTest struct{}

func (s *SubTest) Add(ctx *kuja.SubscribeContext, data *AddReq) error {
	logrus.Info("get", data)
	time.Sleep(time.Second * 2)
	logrus.Info(ctx)
	return errors.New("test error")
}

func main() {
	s := kuja.NewServer()
	s.Service(&ServiceTest{})
	s.Snappy(true)
	s.Encoder(gogoproto.NewEncoder())
	s.Broker(nats.NewBroker("nats://127.0.0.1:4222"))
	sub := &SubTest{}
	s.Subscribe("SubTest", "Add", sub.Add)
	s.SubscribeSize(10)

	conn, _ := amqp.Dial("amqp://guest:guest@plimble.com:5672/")
	defer conn.Close()
	ch, _ := conn.Channel()
	defer ch.Close()

	ch.ExchangeDeclare(
		"uploader_received", // name
		"fanout",            // type
		true,                // durable
		false,               // auto-deleted
		false,               // internal
		false,               // no-wait
		nil,                 // arguments
	)

	q, _ := ch.QueueDeclare(
		"q1",  // name
		true,  // durable
		false, // delete when usused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)

	q2, _ := ch.QueueDeclare(
		"q2",  // name
		true,  // durable
		false, // delete when usused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)

	ch.QueueBind(
		q.Name,              // queue name
		"",                  // routing key
		"uploader_received", // exchange
		false,
		nil,
	)

	ch.QueueBind(
		q2.Name,             // queue name
		"",                  // routing key
		"uploader_received", // exchange
		false,
		nil,
	)

	msgs1, _ := ch.Consume(
		q.Name, // queue
		"",     // consumer
		false,  // no-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)

	msgs2, _ := ch.Consume(
		q.Name, // queue
		"",     // consumer
		false,  // no-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)

	msgs3, _ := ch.Consume(
		q2.Name, // queue
		"",      // consumer
		false,   // no-ack
		false,   // exclusive
		false,   // no-local
		false,   // no-wait
		nil,     // args
	)

	go func() {
		for d := range msgs1 {
			logrus.Println("11111 Triggered event uploader_received")
			time.Sleep(time.Second * 3)
			logrus.Printf("11111 Resizing %s", d.Body)
			d.Ack(false)
		}
	}()

	go func() {
		for d := range msgs2 {
			logrus.Println("22222 Triggered event uploader_received")
			time.Sleep(time.Second * 3)
			logrus.Printf("22222 Resizing %s", d.Body)
			d.Ack(false)
		}
	}()

	go func() {
		for d := range msgs3 {
			logrus.Println("33333 Triggered event uploader_received")
			time.Sleep(time.Second * 3)
			logrus.Printf("33333 Send Email %s", d.Body)
			d.Ack(false)
		}
	}()

	// reg := etcd.NewRegistry("/jack6", []string{"http://plimble.com:4001"})
	// s.Registry(reg)
	// reg.Watch()

	// reg := consul.NewRegistry([]string{"http://plimble.com:8500"})
	// s.Registry(reg)
	// reg.Watch()

	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
		DisableColors: false,
	})

	s.Run(":3002", 0)
}
