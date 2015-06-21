package broker

import (
	"github.com/satori/go.uuid"
)

type Broker interface {
	Publish(topic string, msg *Message) error
	Subscribe(topic, queue, appId string, h Handler)
	Close()
	Connect() error
}

type Handler func(topic string, msg *Message) (int, error)

func NewMessage(meta map[string]string, body []byte) *Message {
	return &Message{
		Id:     uuid.NewV1().String(),
		Header: meta,
		Body:   body,
	}
}
