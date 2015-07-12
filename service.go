package kuja

import (
	log "github.com/Sirupsen/logrus"
	"github.com/plimble/errors"
	"github.com/plimble/kuja/registry"
	"github.com/satori/go.uuid"
	"reflect"
	"unicode"
	"unicode/utf8"
)

type ServiceErrorFunc func(serviceID, service, method string, status int, err error)

// Precompute the reflect type for error. Can't use error directly
// because Typeof takes an empty interface value. This is annoying.
var typeOfError = reflect.TypeOf((*error)(nil)).Elem()

type service struct {
	id       string
	name     string                 // name of service
	rcvr     reflect.Value          // receiver of methods for the service
	typ      reflect.Type           // type of the receiver
	method   map[string]*methodType // registered methods
	handlers []Handler
	node     *registry.Node
}

type methodType struct {
	method      reflect.Method
	ArgType     reflect.Type
	ReplyType   reflect.Type
	ContextType reflect.Type
	stream      bool
	numCalls    uint
}

func (server *Server) ServiceError(fn ServiceErrorFunc) {
	server.serviceError = fn
}

func defaulServiceErr(serviceID, service, method string, status int, err error) {
	if err2, ok := err.(errors.Error); ok {
		switch err2.Code() {
		case 400, 404:
			log.Infof("Service Error %s %s %s %d %s", serviceID, service, method, status, err)
		case 403, 401:
			log.Warnf("Service Error %s %s %s %d %s", serviceID, service, method, status, err)
		case 500:
			log.Errorf("Service Error %s %s %s %d %s", serviceID, service, method, status, err)
		}
	} else {
		log.Errorf("Service Error %s %s %s %d %s", serviceID, service, method, status, err)
	}
}

func (server *Server) Service(service interface{}, h ...Handler) {
	if err := server.register(service, "", false, h); err != nil {
		panic(err)
	}
}

func (m *methodType) prepareContext(ctx *Context) reflect.Value {
	if contextv := reflect.ValueOf(ctx); contextv.IsValid() {
		return contextv
	}
	return reflect.Zero(m.ContextType)
}

func (server *Server) register(rcvr interface{}, name string, useName bool, h []Handler) error {
	server.mu.Lock()
	defer server.mu.Unlock()
	if server.serviceMap == nil {
		server.serviceMap = make(map[string]*service)
	}
	s := new(service)
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
		log.Errorln(s)
		return errors.New(s)
	}
	if _, present := server.serviceMap[sname]; present {
		return errors.New("rpc: service already defined: " + sname)
	}
	s.name = sname
	s.method = make(map[string]*methodType)

	// Install the methods
	for m := 0; m < s.typ.NumMethod(); m++ {
		method := s.typ.Method(m)
		if mt := prepareMethod(method); mt != nil {
			s.method[method.Name] = mt
		}
	}

	if len(s.method) == 0 {
		s := "rpc Register: type " + sname + " has no exported methods of suitable type"
		log.Errorln(s)
		return errors.New(s)
	}
	server.serviceMap[s.name] = s

	s.id = s.name + "-" + uuid.NewV1().String()
	s.handlers = append(s.handlers, server.middleware...)
	s.handlers = append(s.handlers, h...)

	return nil
}

// prepareMethod returns a methodType for the provided method or nil
// in case if the method was unsuitable.
func prepareMethod(method reflect.Method) *methodType {
	mtype := method.Type
	mname := method.Name
	var replyType, argType, contextType reflect.Type

	stream := false
	// Method must be exported.
	if method.PkgPath != "" {
		return nil
	}

	switch mtype.NumIn() {
	case 4:
		// method that takes a context
		argType = mtype.In(2)
		replyType = mtype.In(3)
		contextType = mtype.In(1)
	default:
		log.Errorln("method", mname, "of", mtype, "has wrong number of ins:", mtype.NumIn())
		return nil
	}

	// First arg need not be a pointer.
	if !isExportedOrBuiltinType(argType) {
		log.Errorln(mname, "argument type not exported:", argType)
		return nil
	}

	// the second argument will tell us if it's a streaming call
	// or a regular call
	if replyType.Kind() != reflect.Ptr {
		log.Errorln("method", mname, "reply type not a pointer:", replyType)
		return nil
	}

	// Reply type must be exported.
	if !isExportedOrBuiltinType(replyType) {
		log.Errorln("method", mname, "reply type not exported:", replyType)
		return nil
	}
	// Method needs one out.
	if mtype.NumOut() != 1 {
		log.Errorln("method", mname, "has wrong number of outs:", mtype.NumOut())
		return nil
	}
	// The return type of the method must be error.
	if returnType := mtype.Out(0); returnType != typeOfError {
		log.Errorln("method", mname, "returns", returnType.String(), "not error")
		return nil
	}
	return &methodType{method: method, ArgType: argType, ReplyType: replyType, ContextType: contextType, stream: stream}
}

// Is this an exported - upper case - name?
func isExported(name string) bool {
	rune, _ := utf8.DecodeRuneInString(name)
	return unicode.IsUpper(rune)
}

// Is this type exported or a builtin?
func isExportedOrBuiltinType(t reflect.Type) bool {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	// PkgPath will be non-empty even for an exported type,
	// so we need to check the type name as well.
	return isExported(t.Name()) || t.PkgPath() == ""
}
