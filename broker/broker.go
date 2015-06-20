package broker

type Broker interface {
	Publish(topic string, msg *Message) error
	Subscribe(topic, queue, appId string, h Handler)
	Close()
	Connect() error
}

type Handler func(topic string, msg *Message) (int, bool)
