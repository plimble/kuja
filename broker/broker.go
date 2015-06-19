package broker

type Broker interface {
	Publish(topic string, data []byte) error
	Subscribe(topic, appId string, h Handler)
	Queue(workers int, topic string, h Handler)
	Close()
	Connect() error
}

type Handler func(topic string, header map[string]string, data []byte)
