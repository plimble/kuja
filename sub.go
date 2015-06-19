package kuja

import (
	"errors"
	log "github.com/Sirupsen/logrus"
	"github.com/plimble/kuja/broker"
	"reflect"
)

type SubscriberErrorFunc func(serverId, service, topic string, err error)

type SubscribeContext struct {
	Service  string
	Topic    string
	Metadata Metadata
	status   string
	retry    int
	reject   bool
}

func (ctx *SubscribeContext) Reject(retry int) error {
	ctx.reject = true
	if retry > -1 {
		ctx.retry = retry
		return nil
	}

	ctx.retry = -1

	return nil
}

func (ctx *SubscribeContext) Ack() error {
	return nil
}

type subscriber struct {
	name     string
	rcvr     reflect.Value // receiver of methods for the service
	typ      reflect.Type  // type of the receiver
	service  string
	topic    string
	dataType reflect.Type
}

func defaulSubscriberErr(subscriberID, subscriber, topic string, err error) {
	log.Errorf("Subscriber Error %s %s %s %s", subscriberID, subscriber, topic, err)
}

func (server *Server) SubscriberError(fn SubscriberErrorFunc) {
	server.subscriberError = fn
}

func (server *Server) Subscribe(service, topic string, method interface{}) {
	if server.broker == nil {
		panic(errors.New("no broker registered"))
	}

	s := &subscriber{
		service: service,
		topic:   topic,
		name:    service + "." + topic,
	}

	if err := server.registerSub(method, s); err != nil {
		panic(err)
	}
}

func (server *Server) SubscribeSize(n int) {
	server.subscribeSize = n
}

func (server *Server) registerSub(method interface{}, s *subscriber) error {
	server.mu.Lock()
	defer server.mu.Unlock()

	s.rcvr = reflect.ValueOf(method)
	s.typ = reflect.TypeOf(method)
	methodName := s.typ.String()

	if s.typ.NumOut() != 1 {
		log.Errorln("method", methodName, "has wrong number of outs:", s.typ.NumOut())
		return nil
	}

	if s.typ.NumIn() != 2 {
		log.Errorln("method", methodName, "has wrong number of ins:", s.typ.NumIn())
		return nil
	}

	returnType := s.typ.Out(0)
	ctxType := s.typ.In(0)
	s.dataType = s.typ.In(1)

	if s.dataType.Kind() != reflect.Ptr {
		log.Errorln("method", methodName, "data type not a pointer:", s.dataType)
		return nil
	}

	if !isExportedOrBuiltinType(s.dataType) {
		log.Errorln("method", methodName, "data type not exported:", s.dataType)
		return nil
	}

	if ctxType.String() != "*kuja.SubscribeContext" {
		log.Error("method", methodName, "argument 2", returnType.String(), "not SubscribeContext")
		return nil
	}

	if returnType != typeOfError {
		log.Error("method", methodName, "returns", returnType.String(), "not error")
		return nil
	}

	server.subscriberMap[s.name] = s

	return nil
}

func (server *Server) subscribe(s *subscriber) broker.Handler {
	return func(topic string, msg *broker.Message) (int, bool) {
		sub, ok := server.subscriberMap[topic]
		if !ok {
			server.subscriberError("", "", "", errors.New("no topic"+topic))
			return -1, false
		}

		ctx := &SubscribeContext{
			Service:  sub.service,
			Topic:    sub.topic,
			Metadata: msg.Header,
			retry:    -1,
			reject:   false,
		}

		datav := reflect.New(sub.dataType.Elem())
		if err := server.encoder.Unmarshal(msg.Body, datav.Interface()); err != nil {
			server.subscriberError(server.id, ctx.Service, ctx.Topic, err)
			return -1, false
		}

		returnValues := sub.rcvr.Call([]reflect.Value{reflect.ValueOf(ctx), datav})
		errInter := returnValues[0].Interface()
		if errInter != nil {
			server.subscriberError(server.id, ctx.Service, ctx.Topic, errInter.(error))
		}

		if ctx.reject {
			switch ctx.retry {
			case 0:
				log.Errorf("Subscriber Rejected %s %s %s %s", server.id, ctx.Service, ctx.Topic, "reject, retry: always")
				server.subscriberError(server.id, ctx.Service, ctx.Topic, errors.New("reject, retry: always"))
			case -1:
				log.Errorf("Subscriber Rejected %s %s %s %s", server.id, ctx.Service, ctx.Topic, "reject, retry: no")
			default:
				log.Errorf("Subscriber Rejected %s %s %s %s %d", server.id, ctx.Service, ctx.Topic, "reject, retry:", ctx.retry)
			}
		}

		return ctx.retry, ctx.reject
	}
}
