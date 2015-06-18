package kuja

import (
	"errors"
	log "github.com/Sirupsen/logrus"
	"github.com/plimble/kuja/broker"
	"github.com/satori/go.uuid"
	"reflect"
)

type SubscriberErrorFunc func(subscriberID, subscriber, topic string, err error)

type subscriberContext struct {
	info *SubscriberInfo
	msg  *broker.Message
}

type SubscriberInfo struct {
	Id         string
	Subscriber string
	Topic      string
}

type subscriber struct {
	id     string
	name   string
	rcvr   reflect.Value             // receiver of methods for the service
	typ    reflect.Type              // type of the receiver
	method map[string]*methodSubType // registered methods
}

type methodSubType struct {
	method   reflect.Method
	MetaType reflect.Type
	DataType reflect.Type
}

func defaulSubscriberErr(subscriberID, subscriber, topic string, err error) {
	log.Infof("Subscriber Error %s %s %s %s", subscriberID, subscriber, topic, err)
}

func (server *Server) SubscriberError(fn SubscriberErrorFunc) {
	server.subscriberError = fn
}

func (server *Server) Broker(b broker.Broker) {
	server.broker = b
}

func (server *Server) Subscriber(subscriber interface{}) {
	if server.broker == nil {
		panic(errors.New("no broker registered"))
	}

	if err := server.registerSub(subscriber, "", false); err != nil {
		panic(err)
	}
}

func (server *Server) registerSub(rcvr interface{}, name string, useName bool) error {
	server.mu.Lock()
	defer server.mu.Unlock()
	if server.subscriberMap == nil {
		server.subscriberMap = make(map[string]*subscriber)
	}
	s := new(subscriber)
	s.typ = reflect.TypeOf(rcvr)
	s.rcvr = reflect.ValueOf(rcvr)
	sname := reflect.Indirect(s.rcvr).Type().Name()
	if useName {
		sname = name
	}
	if sname == "" {
		log.Fatal("rpc: no service name for type", s.typ.String())
	}
	if !isExported(sname) && !useName {
		s := "rpc Register: type " + sname + " is not exported"
		log.Print(s)
		return errors.New(s)
	}
	if _, present := server.serviceMap[sname]; present {
		return errors.New("rpc: service already defined: " + sname)
	}
	s.name = sname
	s.id = s.name + "-" + uuid.NewV1().String()
	s.method = make(map[string]*methodSubType)

	// Install the methods
	for m := 0; m < s.typ.NumMethod(); m++ {
		method := s.typ.Method(m)
		if mt := prepareSubMethod(method); mt != nil {
			s.method[method.Name] = mt
		}

		topic := s.name + "." + method.Name
		server.broker.Subscribe(topic, func(topic string, header map[string]string, data []byte) {
			info := &SubscriberInfo{}
			msg := &broker.Message{}
			info.Id = s.id
			info.Subscriber = s.name
			info.Topic = method.Name
			datav := reflect.New(s.method[method.Name].DataType.Elem())

			if err := msg.Unmarshal(data); err != nil {
				server.subscriberError(info.Id, info.Subscriber, info.Topic, err)
				return
			}
			if err := server.encoder.Unmarshal(msg.Body, datav.Interface()); err != nil {
				server.subscriberError(info.Id, info.Subscriber, info.Topic, err)
				return
			}

			returnValues := s.method[method.Name].method.Func.Call([]reflect.Value{s.rcvr, reflect.ValueOf(info), reflect.ValueOf(msg.Header), datav})
			errInter := returnValues[0].Interface()
			if errInter != nil {
				server.subscriberError(info.Id, info.Subscriber, info.Topic, errInter.(error))
			}
		})
	}

	if len(s.method) == 0 {
		s := "rpc Register: type " + sname + " has no exported methods of suitable type"
		log.Print(s)
		return errors.New(s)
	}
	server.subscriberMap[s.name] = s

	return nil
}

func prepareSubMethod(method reflect.Method) *methodSubType {
	mtype := method.Type
	mname := method.Name
	var subinfoType, dataType, MetaType reflect.Type

	// Method must be exported.
	if method.PkgPath != "" {
		return nil
	}

	switch mtype.NumIn() {
	case 4:
		// normal method
		subinfoType = mtype.In(1)
		MetaType = mtype.In(2)
		dataType = mtype.In(3)
	default:
		log.Println("method", mname, "of", mtype, "has wrong number of ins:", mtype.NumIn())
		return nil
	}

	// First arg need not be a pointer.
	if !isExportedOrBuiltinType(MetaType) {
		log.Println(mname, "argument type not exported:", MetaType)
		return nil
	}

	// the second argument will tell us if it's a streaming call
	// or a regular call
	if dataType.Kind() == reflect.Func {
		// this is a streaming call
		if dataType.NumIn() != 1 {
			log.Println("method", mname, "sendReply has wrong number of ins:", dataType.NumIn())
			return nil
		}
		if dataType.In(0).Kind() != reflect.Interface {
			log.Println("method", mname, "sendReply parameter type not an interface:", dataType.In(0))
			return nil
		}
		if dataType.NumOut() != 1 {
			log.Println("method", mname, "sendReply has wrong number of outs:", dataType.NumOut())
			return nil
		}
		if returnType := dataType.Out(0); returnType != typeOfError {
			log.Println("method", mname, "sendReply returns", returnType.String(), "not error")
			return nil
		}

	} else if dataType.Kind() != reflect.Ptr {
		log.Println("method", mname, "reply type not a pointer:", dataType)
		return nil
	}

	// Reply type must be exported.
	if !isExportedOrBuiltinType(dataType) {
		log.Println("method", mname, "reply type not exported:", dataType)
		return nil
	}
	// Method needs one out.
	if mtype.NumOut() != 1 {
		log.Println("method", mname, "has wrong number of outs:", mtype.NumOut())
		return nil
	}

	if subinfoType.String() != "*kuja.SubscriberInfo" {
		log.Println("method", mname, subinfoType.String(), "is not SubscriberInfo")
		return nil
	}

	// The return type of the method must be error.
	if returnType := mtype.Out(0); returnType != typeOfError {
		log.Println("method", mname, "returns", returnType.String(), "not error")
		return nil
	}
	return &methodSubType{method: method, MetaType: MetaType, DataType: dataType}
}
