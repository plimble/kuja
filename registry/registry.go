package registry

type Service struct {
	Name  string
	Nodes []*Node
}

type Node struct {
	Id      string
	Name    string
	Host    string
	Port    string
	Address string
	URL     string
}

type Watcher interface {
	Stop()
}

type Registry interface {
	Register(node *Node) error
	Deregister(node *Node) error
	GetService(name string) (*Service, error)
	ListServices() ([]*Service, error)
	Watch() (Watcher, error)
	Close()
}
