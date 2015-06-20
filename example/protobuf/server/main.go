package main

import (
	"github.com/plimble/kuja"
	"github.com/plimble/kuja/encoder/proto"
	"github.com/plimble/kuja/example/protobuf"
)

func main() {
	server := kuja.NewServer()
	server.Encoder(proto.NewEncoder())
	server.Snappy(true)
	server.Service(&protobuf.AddService{})

	server.Run(":3000", 0)
}
