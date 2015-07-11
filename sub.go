package kuja

import (
	"errors"
	log "github.com/Sirupsen/logrus"
	"github.com/plimble/kuja/broker"
	"reflect"
	"strings"
	"time"
)

type SubscriberErrorFunc func(appId, topic, queue string, err error)

type SubscribeContext struct {
	Id       string
	Topic    string
	Metadata Metadata
	status   string
	retry    int
	Queue    string
}

func (ctx *SubscribeContext) Reject(retry int, err error) error {
	if retry > 0 {
		ctx.retry = retry
	}

	if err == nil {
		err = errors.New("rejected")
	}

	return err
}

func (ctx *SubscribeContext) Ack() error {
	return nil
}

type subscriber struct {
	rcvr     reflect.Value // receiver of methods for the service
	typ      reflect.Type  // type of the receiver
	topic    string
	queue    string
	dataType reflect.Type
	handler  broker.Handler
	timeout  time.Duration
	size     int
}

func defaulSubscriberErr(msgId, topic, queue string, err error) {
	log.Errorf("Subscriber Error msg id: %s, topic: %s, queue: %s, err: %s", msgId, topic, queue, err)
}

func (server *Server) SubscriberError(fn SubscriberErrorFunc) {
	server.subscriberError = fn
}

func (server *Server) Subscribe(topic, queue string, timeout time.Duration, size int, method interface{}) {
	if topic == "" || queue == "" {
		panic(errors.New("topic, queue should not be empty"))
	}

	s := &subscriber{
		topic:   topic,
		queue:   queue,
		timeout: timeout,
		size:    size,
	}

	if err := server.registerSub(method, s); err != nil {
		panic(err)
	}
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

	if s.timeout == 0 {
		s.handler = server.subscribe(s)
	} else {
		s.handler = server.subscribeTimeout(s)
	}
	server.subscriberMap[s.topic+"."+s.queue] = s

	return nil
}

func (server *Server) ServePubSub(topic string, msg []byte) error {
	var err error

	brokerMsg := &broker.Message{}
	if err = brokerMsg.Unmarshal(msg); err != nil {
		return err
	}

	err = nil
	for key, handler := range server.subscriberMap {
		if strings.HasPrefix(key, topic) {
			_, err := handler.handler(topic, brokerMsg)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

type subResponse struct {
	retry int
	err   error
}

func (server *Server) subscribeTimeout(s *subscriber) broker.Handler {
	return func(topic string, msg *broker.Message) (int, error) {
		var err error
		ch := make(chan *subResponse)

		ctx := &SubscribeContext{
			Id:       msg.Id,
			Topic:    s.topic,
			Queue:    s.queue,
			Metadata: msg.Header,
			retry:    0,
		}

		if msg.Retry > 0 {
			log.Infof("Subscriber Error msg id: %s, topic: %s, queue: %s, retry: %d", ctx.Id, ctx.Topic, ctx.Queue, msg.Retry)
		}

		datav := reflect.New(s.dataType.Elem())
		if err = server.encoder.Unmarshal(msg.Body, datav.Interface()); err != nil {
			server.subscriberError(ctx.Id, ctx.Topic, ctx.Queue, err)
			return 0, err
		}

		go func() {
			returnValues := s.rcvr.Call([]reflect.Value{reflect.ValueOf(ctx), datav})
			errInter := returnValues[0].Interface()
			if errInter != nil {
				ch <- &subResponse{ctx.retry, errInter.(error)}
			} else {
				ch <- &subResponse{0, nil}
			}
		}()

		select {
		case resp := <-ch:
			if resp.err != nil {
				server.subscriberError(ctx.Id, ctx.Topic, ctx.Queue, resp.err)
			}
			return resp.retry, resp.err
		case <-time.After(s.timeout):
			server.subscriberError(ctx.Id, ctx.Topic, ctx.Queue, errors.New("time out"))
			return 0, errors.New("timeout")
		}

		return 0, nil
	}
}

func (server *Server) subscribe(s *subscriber) broker.Handler {
	return func(topic string, msg *broker.Message) (int, error) {
		var err error
		ctx := &SubscribeContext{
			Id:       msg.Id,
			Topic:    s.topic,
			Queue:    s.queue,
			Metadata: msg.Header,
			retry:    0,
		}

		if msg.Retry > 0 {
			log.Infof("Subscriber Error msg id: %s, topic: %s, queue: %s, retry: %d", ctx.Id, ctx.Topic, ctx.Queue, msg.Retry)
		}

		datav := reflect.New(s.dataType.Elem())
		if err = server.encoder.Unmarshal(msg.Body, datav.Interface()); err != nil {
			server.subscriberError(ctx.Id, ctx.Topic, ctx.Queue, err)
			return 0, err
		}

		returnValues := s.rcvr.Call([]reflect.Value{reflect.ValueOf(ctx), datav})
		errInter := returnValues[0].Interface()
		if errInter != nil {
			err = errInter.(error)
			server.subscriberError(ctx.Id, ctx.Topic, ctx.Queue, err)
			return ctx.retry, err
		}

		return 0, nil
	}
}
