package client

import (
	"github.com/plimble/kuja/registry"
	"math/rand"
)

type Discovery struct {
	registry registry.Registry
}

func (d *Discovery) GetAddress(service, method string) (string, error) {
	s, err := d.registry.GetService(service)
	if err != nil {
		return "", err
	}

	n := rand.Int() % len(s.Nodes)
	node := s.Nodes[n]

	return node.URL + "/" + service + "/" + method, nil
}
