package nats

import (
	"github.com/apcera/nats"
	"github.com/plimble/kuja/broker"
)

type natsBroker struct {
	url  string
	conn *nats.Conn
}

func NewBroker(url string) *natsBroker {
	return &natsBroker{
		url: url,
	}
}

func (n *natsBroker) Connect() error {
	var err error
	n.conn, err = nats.Connect(n.url)

	return err
}

func (n *natsBroker) Close() {
	n.conn.Close()
}

func (n *natsBroker) Publish(topic string, data []byte) error {
	return n.conn.Publish(topic, data)
}

func (n *natsBroker) Subscribe(topic, appId string, h broker.Handler) {
	n.conn.QueueSubscribe(topic, appId, func(msg *nats.Msg) {
		h(msg.Subject, nil, msg.Data)
	})
}

func (n *natsBroker) Queue(workers int, topic string, h broker.Handler) {
	n.conn.QueueSubscribe(topic, "queue", func(msg *nats.Msg) {
		h(msg.Subject, nil, msg.Data)
	})
}
