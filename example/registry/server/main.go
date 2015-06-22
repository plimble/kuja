package main

import (
	"github.com/plimble/kuja"
	"github.com/plimble/kuja/example/registry"
	"github.com/plimble/kuja/registry/etcd"
)

func main() {
	server := kuja.NewServer()
	server.Service(&registry.AddService{})
	server.Registry(etcd.NewRegistry("/kuja_example/services", []string{"http://127.0.0.1.com:4001"}))
	server.Snappy(true)

	server.Run(":3000", 0)
}
