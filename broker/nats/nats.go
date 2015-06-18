package nats

import (
	"github.com/apcera/nats"
	"github.com/plimble/kuja/broker"
)

type natsBroker struct {
	conn *nats.Conn
}

func NewBroker(url string) (*natsBroker, error) {
	conn, err := nats.Connect(url)
	if err != nil {
		return nil, err
	}

	return &natsBroker{
		conn: conn,
	}, nil
}

func (n *natsBroker) Publish(topic string, data []byte) error {
	return n.conn.Publish(topic, data)
}

func (n *natsBroker) Subscribe(topic string, h broker.Handler) {
	n.conn.Subscribe(topic, func(msg *nats.Msg) {
		h(msg.Subject, nil, msg.Data)
	})
}

func (n *natsBroker) Queue(workers int, topic string, h broker.Handler) {
	n.conn.QueueSubscribe(topic, "queue", func(msg *nats.Msg) {
		h(msg.Subject, nil, msg.Data)
	})
}
