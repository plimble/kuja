package registry

type Service struct {
}

type Registry interface {
	Register(*Service) error
	Deregister(*Service) error
	GetServiceEndpoint(string) string
	GetService(string) (*Service, error)
	ListServices() ([]*Service, error)
	// Watch() (Watcher, error)
}
