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
}

type Registry interface {
	Register(node *Node) error
	Deregister(name, id string) error
	GetService(name string) (*Service, error)
	ListServices() ([]*Service, error)
	// Watch() (Watcher, error)
}
