package nats

import (
	"github.com/apcera/nats"
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

func (n *natsBroker) Publish(topic string) {
	n.conn.Publish(subj, data)
}

func (n *natsBroker) Subscribe() {
    
}

func (n *natsBroker) Queue() {

}
