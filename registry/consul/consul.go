package consul

import (
	"errors"
	consul "github.com/hashicorp/consul/api"
	"github.com/plimble/kuja/registry"
	"strconv"
	"strings"
	"sync"
)

type ConsulRegistry struct {
	addr   string
	client *consul.Client
	sync.RWMutex
	services map[string]*registry.Service
}

func (c *ConsulRegistry) Register(node *registry.Node) error {
	if node == nil {
		return errors.New("node should not be nil")
	}

	portInt, _ := strconv.Atoi(node.Port)

	_, err := c.client.Catalog().Register(&consul.CatalogRegistration{
		Node:    node.Id,
		Address: node.Address,
		Service: &consul.AgentService{
			ID:      node.Id,
			Service: node.Name,
			Port:    portInt,
		},
	}, nil)

	return err
}

func (c *ConsulRegistry) Deregister(node *registry.Node) error {
	if node == nil {
		return errors.New("node should not be nil")
	}

	_, err := c.client.Catalog().Deregister(&consul.CatalogDeregistration{
		Node: node.Id,
	}, nil)

	return err
}

func (c *ConsulRegistry) GetService(name string) (*registry.Service, error) {
	c.RLock()
	service, ok := c.services[name]
	c.RUnlock()

	if ok {
		return service, nil
	}

	rsp, _, err := c.client.Catalog().Service(name, "", nil)
	if err != nil {
		return nil, err
	}

	s := &registry.Service{
		Name:  name,
		Nodes: []*registry.Node{},
	}

	for _, consulService := range rsp {
		if consulService.ServiceName != name {
			continue
		}

		s.Nodes = append(s.Nodes, &registry.Node{
			Id:      consulService.ServiceID,
			Address: consulService.Address,
			Port:    strconv.Itoa(consulService.ServicePort),
		})
	}

	return s, nil
}

func (c *ConsulRegistry) ListServices() ([]*registry.Service, error) {
	var services []*registry.Service

	c.RLock()
	for _, service := range c.services {
		services = append(services, service)
	}

	if len(services) > 0 {
		return services, nil
	}
	c.RUnlock()

	rsp, _, err := c.client.Catalog().Services(&consul.QueryOptions{})
	if err != nil {
		return nil, err
	}

	for service, _ := range rsp {
		services = append(services, &registry.Service{Name: service})
	}

	return services, nil
}

func (e *ConsulRegistry) Watch() (registry.Watcher, error) {
	services, err := e.ListServices()
	if err != nil {
		return nil, err
	}

	for i := 0; i < len(services); i++ {
		s, err := e.GetService(services[i].Name)
		if err != nil {
			return nil, err
		}

		e.services[s.Name] = s
	}

	return newConsulWatcher(e)
}

func NewRegistry(addrs []string) registry.Registry {
	replacer := strings.NewReplacer("http://", "", "https://", "")
	for i := 0; i < len(addrs); i++ {
		addrs[i] = replacer.Replace(addrs[i])
	}
	config := consul.DefaultConfig()
	if len(addrs) > 0 {
		config.Address = addrs[0]
	}
	client, _ := consul.NewClient(config)

	c := &ConsulRegistry{
		addr:     config.Address,
		client:   client,
		services: make(map[string]*registry.Service),
	}

	return c
}
