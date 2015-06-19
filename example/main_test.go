package main

import (
	"github.com/plimble/kuja/broker/nats"
	"github.com/plimble/kuja/client"
	"github.com/plimble/kuja/encoder/gogoproto"
	"github.com/streadway/amqp"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestRabbit(t *testing.T) {
	return
	conn, _ := amqp.Dial("amqp://guest:guest@plimble.com:5672/")
	defer conn.Close()

	ch, _ := conn.Channel()
	defer ch.Close()

	for i := 0; i < 2; i++ {
		ch.Publish(
			"uploader_received", // exchange
			"",                  // routing key
			false,               // mandatory
			false,               // immediate
			amqp.Publishing{
				ContentType: "text/plain",
				Body:        []byte("some-image.jpg"),
			},
		)
	}
}

func TestPublish(t *testing.T) {
	var err error
	c := client.New("http://127.0.0.1:3002")
	c.Encoder(gogoproto.NewEncoder())
	c.Broker(nats.NewBroker("nats://127.0.0.1:4222"))

	addreq := &AddReq{5, 7}
	addresp := &AddResp{}
	err = c.Publish("SubTest", "Add", &AddReq{1, 1}, map[string]string{"h1": "v1", "h2": "v2"})
	err = c.Publish("SubTest", "Add", &AddReq{2, 2}, map[string]string{"h1": "v1", "h2": "v2"})
	err = c.Publish("SubTest", "Add", &AddReq{3, 3}, map[string]string{"h1": "v1", "h2": "v2"})
	assert.NoError(t, err)

	status, err := c.Request("ServiceTest", "Add", addreq, addresp, nil)
	assert.Equal(t, 200, status)
	assert.NoError(t, err)
	assert.Equal(t, 22, addresp.C)

	addreq1 := &AddReq{5, 7}
	addresp1 := &AddResp{}
	err = c.AsyncRequest("ServiceTest", "Add", addreq1, addresp1, nil)
	assert.NoError(t, err)

	addreq2 := &AddReq{5, 7}
	addresp2 := &AddResp{}
	err = c.AsyncRequest("ServiceTest", "Add", addreq2, addresp2, nil)
	assert.NoError(t, err)
	time.Sleep(5 * time.Second)
}
