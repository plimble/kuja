package stub_registry

import (
	"github.com/plimble/kuja/registry"
)

type stubWatcher struct{}

func (s *stubWatcher) Stop() {}

type StubRegistry struct{}

func (s *StubRegistry) Register(node *registry.Node) error {
	return nil
}

func (s *StubRegistry) Deregister(node *registry.Node) error {
	return nil
}

func (s *StubRegistry) GetService(name string) (*registry.Service, error) {
	return nil, nil
}

func (s *StubRegistry) ListServices() ([]*registry.Service, error) {
	return nil, nil
}

func (s *StubRegistry) Watch() (registry.Watcher, error) {
	return &stubWatcher{}, nil
}

func (s *StubRegistry) Close() {}

func NewRegistry() registry.Registry {
	return &StubRegistry{}
}
