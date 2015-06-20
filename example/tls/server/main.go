package main

import (
	"github.com/plimble/kuja"
	"github.com/plimble/kuja/example/tls"
)

func main() {
	server := kuja.NewServer()
	server.Service(&tls.AddService{})
	server.Snappy(true)

	server.RunTLS(":3000", 0, "../server.pem", "../server-key.pem")
}
