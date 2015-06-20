package pubsub

import (
	"errors"
	"github.com/plimble/kuja"
	"log"
)

type Subscibers struct{}

func (s *Subscibers) Add(ctx *kuja.SubscribeContext, req *AddReq) error {
	log.Println("Subscribe Add triggered")
	return ctx.Ack()
}

func (s *Subscibers) Sub(ctx *kuja.SubscribeContext, req *AddReq) error {
	log.Println("Subscribe Sub triggered")
	// retry 3 times
	return ctx.Reject(3, errors.New("error in Sub"))
}

func (s *Subscibers) Multiply(ctx *kuja.SubscribeContext, req *AddReq) error {
	log.Println("Subscribe Multiply triggered")
	// no retry
	return ctx.Reject(0, errors.New("error in Multiply"))
}

func (s *Subscibers) Divide(ctx *kuja.SubscribeContext, req *AddReq) error {
	log.Println("Subscribe Divide triggered")
	return errors.New("error in divide")
}
