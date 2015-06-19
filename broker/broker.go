package broker

import (
	"errors"
)

var (
	ErrRejectWithRequeue = errors.New("reject but requeue")
	ErrReject            = errors.New("reject")
)

type Broker interface {
	Publish(topic string, msg *Message) error
	Subscribe(topic, appId string, h Handler)
	Close()
	Connect() error
}

type Handler func(topic string, msg *Message) (int, bool)
