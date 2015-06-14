package registry

type Service struct {
}

type Endpoint struct {
	URL  string
	Host string
	Port string
}

type Registry interface {
	Register(*Service) error
	Deregister(*Service) error
	GetServiceEndpoint(string)
	GetService(string) (*Service, error)
	ListServices() ([]*Service, error)
	// Watch() (Watcher, error)
}
