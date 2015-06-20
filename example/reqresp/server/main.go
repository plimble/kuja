package main

import (
	"github.com/plimble/kuja"
	"github.com/plimble/kuja/example/reqresp"
)

func main() {
	server := kuja.NewServer()
	server.Service(&reqresp.AddService{})
	server.Snappy(true)

	server.Run(":3000", 0)
}
