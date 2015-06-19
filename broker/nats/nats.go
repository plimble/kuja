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

func (n *natsBroker) Publish(topic string, msg *broker.Message) error {
	data, err := msg.Marshal()
	if err != nil {
		return err
	}
	return n.conn.Publish(topic, data)
}

func (n *natsBroker) Subscribe(topic, appId string, h broker.Handler) {
	n.conn.QueueSubscribe(topic, appId, func(msg *nats.Msg) {
		brokerMsg := &broker.Message{}
		brokerMsg.Unmarshal(msg.Data)
		retryCount, reject := h(msg.Subject, brokerMsg)
		if reject {
			if retryCount == 0 {
				n.Publish(topic, brokerMsg)
			} else if brokerMsg.Retry < int32(retryCount) {
				brokerMsg.Retry++
				n.Publish(topic, brokerMsg)
			}
		}
	})
}
