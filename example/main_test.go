package main

import (
	"github.com/plimble/kuja/broker/nats"
	"github.com/plimble/kuja/client"
	"github.com/plimble/kuja/encoder/json"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPublish(t *testing.T) {
	c := client.New("http://127.0.0.1:3000")
	c.Encoder(json.NewEncoder())
	broker, err := nats.NewBroker("http://127.0.0.1:4222")
	if err != nil {
		panic(err)
	}
	c.Broker(broker)

	err = c.Publish("SubTest.Add", &AddReq{10, 20}, map[string]string{"h1": "v1", "h2": "v2"})
	assert.NoError(t, err)
}
