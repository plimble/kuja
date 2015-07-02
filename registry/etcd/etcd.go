package etcd

import (
	"errors"
	"github.com/coreos/go-etcd/etcd"
	"github.com/plimble/kuja/registry"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const (
	defaultProfix = "/_kuja-registry"
)

type EtcdRegistry struct {
	prefix string
	client *etcd.Client
	sync.RWMutex
	services map[string]*registry.Service
}

func (e *EtcdRegistry) servicePath(s string) string {
	return filepath.Join(e.prefix, s)
}

func (e *EtcdRegistry) nodePath(s, id string) string {
	return filepath.Join(e.prefix, s, id)
}

func (e *EtcdRegistry) Register(node *registry.Node) error {
	if node == nil {
		return errors.New("node should not be nil")
	}

	e.client.CreateDir(e.servicePath(node.Name), 0)

	_, err := e.client.Create(e.nodePath(node.Name, node.Id), node.URL, 120)

	go func() {
		for {
			select {
			case <-time.After(time.Minute):
				e.client.Set(e.nodePath(node.Name, node.Id), node.URL, 120)
			}
		}
	}()

	return err
}

func (e *EtcdRegistry) Deregister(node *registry.Node) error {
	if node == nil {
		return errors.New("node should not be nil")
	}

	_, err := e.client.Delete(e.nodePath(node.Name, node.Id), false)

	return err
}

func (e *EtcdRegistry) GetService(name string) (*registry.Service, error) {
	e.RLock()
	service, ok := e.services[name]
	e.RUnlock()

	if ok {
		return service, nil
	}

	rsp, err := e.client.Get(e.servicePath(name), false, true)
	if err != nil && !strings.HasPrefix(err.Error(), "100: Key not found") {
		return nil, err
	}

	s := &registry.Service{
		Name:  name,
		Nodes: []*registry.Node{},
	}

	if rsp == nil {
		return nil, errors.New("service " + name + " not found")
	}

	for _, n := range rsp.Node.Nodes {
		if n.Dir {
			continue
		}
		node := &registry.Node{}
		node.URL = n.Value
		s.Nodes = append(s.Nodes, node)
	}

	return s, nil
}

func (e *EtcdRegistry) ListServices() ([]*registry.Service, error) {
	var services []*registry.Service

	e.RLock()
	for _, service := range e.services {
		services = append(services, service)
	}

	if len(services) > 0 {
		return services, nil
	}
	e.RUnlock()

	rsp, err := e.client.Get(e.prefix, false, true)
	if err != nil && !strings.HasPrefix(err.Error(), "100: Key not found") {
		return nil, err
	}

	if rsp == nil {
		return services, nil
	}

	for _, node := range rsp.Node.Nodes {
		if node.Dir {
			service := &registry.Service{}
			service.Name = strings.Replace(node.Key, e.prefix+"/", "", 1)
			for _, subnode := range node.Nodes {
				n := &registry.Node{}
				n.URL = subnode.Value
				service.Nodes = append(service.Nodes, n)
			}
			services = append(services, service)
		}
	}

	return services, nil
}

func (e *EtcdRegistry) Watch() (registry.Watcher, error) {
	s, err := e.ListServices()
	if err != nil {
		return nil, err
	}

	for i := 0; i < len(s); i++ {
		e.services[s[i].Name] = s[i]
	}

	return newEtcdWatcher(e)
}

func NewRegistry(prefix string, addrs []string) registry.Registry {
	var cAddrs []string
	if prefix == "" {
		prefix = defaultProfix
	}

	for _, addr := range addrs {
		if len(addr) == 0 {
			continue
		}
		cAddrs = append(cAddrs, addr)
	}

	if len(cAddrs) == 0 {
		cAddrs = []string{"http://127.0.0.1:2379"}
	}

	e := &EtcdRegistry{
		prefix:   prefix,
		client:   etcd.NewClient(cAddrs),
		services: make(map[string]*registry.Service),
	}

	return e
}

func (e *EtcdRegistry) Close() {
	e.client.Close()
}
