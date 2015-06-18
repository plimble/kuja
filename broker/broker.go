package broker

type Broker interface {
	Address() string
	Connect() error
	Disconnect() error
	Init() error
	Publish(string, *Message) error
	Subscribe(string, Handler) (Subscriber, error)
}

type Message struct {
	Header map[string]string
	Body   []byte
}

type Handler func()

type Subscriber struct{}
