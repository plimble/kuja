package etcd

import (
	"encoding/json"
	"errors"
	"github.com/coreos/go-etcd/etcd"
	"github.com/plimble/kuja/registry"
	"path/filepath"
	"strings"
	"sync"
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

func encode(s *registry.Node) string {
	b, _ := json.Marshal(s)
	return string(b)
}

func decode(ds string) *registry.Node {
	var s *registry.Node
	json.Unmarshal([]byte(ds), &s)
	return s
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

	_, err := e.client.Create(e.nodePath(node.Name, node.Id), encode(node), 0)

	return err
}

func (e *EtcdRegistry) Deregister(name, id string) error {
	if name == "" || id == "" {
		return errors.New("node should not be nil")
	}

	_, err := e.client.Delete(e.nodePath(name, id), false)

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

	//prefix/servicename/ssss=val
	for _, n := range rsp.Node.Nodes {
		if n.Dir {
			continue
		}
		n := decode(n.Value)
		s.Nodes = append(s.Nodes, n)
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

	for _, node := range rsp.Node.Nodes {
		if node.Dir {
			service := &registry.Service{}
			service.Name = strings.Replace(node.Key, e.prefix+"/", "", 1)
			for _, subnode := range node.Nodes {
				n := decode(subnode.Value)
				service.Nodes = append(service.Nodes, n)
			}
			services = append(services, service)
		}
	}

	return services, nil
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
