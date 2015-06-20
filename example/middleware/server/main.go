package main

import (
	"github.com/plimble/kuja"
	"github.com/plimble/kuja/example/middleware"
	"log"
	"net/http"
)

func GlobalMiddleware(ctx *kuja.Context, w http.ResponseWriter, r *http.Request) error {
	log.Println("Before Global")

	if err := ctx.Next(); err != nil {
		return err
	}

	log.Println("After Global")

	return nil
}

func AddMiddleware(ctx *kuja.Context, w http.ResponseWriter, r *http.Request) error {
	log.Println("Before Add")

	if err := ctx.Next(); err != nil {
		return err
	}

	log.Println("After Add")

	return nil
}

func main() {
	server := kuja.NewServer()
	server.Use(GlobalMiddleware)
	server.Service(&middleware.AddService{}, AddMiddleware)
	server.Snappy(true)

	server.Run(":3000", 0)
}
