package server

import (
	"encoding/json"
	"errors"
	c "github.com/plimble/kuja/context"
	"github.com/plimble/kuja/encoder"
	"golang.org/x/net/context"
	"io"
	"log"
	"net/http"
	"reflect"
	"strings"
	"sync"
	"unicode"
	"unicode/utf8"
)

// Precompute the reflect type for error. Can't use error directly
// because Typeof takes an empty interface value. This is annoying.
var typeOfError = reflect.TypeOf((*error)(nil)).Elem()

type ErrorHandler func(err error) interface{}

type serverRequest struct {
	Method string           `json:"method"`
	Params *json.RawMessage `json:"params"`
	ID     string           `json:"id"`
}

type serverResponse struct {
	ID     string         `json:"id"`
	Result interface{}    `json:"result,omitempty"`
	Error  *ResponseError `json:"error,omitempty"`
}

type ResponseError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e *ResponseError) Error() string {
	return e.Message
}

func Error(code int, msg string) error {
	return &ResponseError{code, msg}
}

type service struct {
	name   string                 // name of service
	rcvr   reflect.Value          // receiver of methods for the service
	typ    reflect.Type           // type of the receiver
	method map[string]*methodType // registered methods
}

type methodType struct {
	sync.Mutex  // protects counters
	method      reflect.Method
	ArgType     reflect.Type
	ReplyType   reflect.Type
	ContextType reflect.Type
	stream      bool
	numCalls    uint
}

func (m *methodType) prepareContext(ctx context.Context) reflect.Value {
	if contextv := reflect.ValueOf(ctx); contextv.IsValid() {
		return contextv
	}
	return reflect.Zero(m.ContextType)
}

type Server struct {
	mu         sync.Mutex // protects the serviceMap
	serviceMap map[string]*service
	encoder    encoder.Encoder
}

func NewServer() *Server {
	return &Server{serviceMap: make(map[string]*service)}
}

func (server *Server) Register(service interface{}) {
	if err := server.register(service, "", false); err != nil {
		panic(err)
	}
}

func (server *Server) Encoder(enc encoder.Encoder) {
	server.encoder = enc
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

func (server *Server) register(rcvr interface{}, name string, useName bool) error {
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
		log.Print(s)
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
		log.Print(s)
		return errors.New(s)
	}
	server.serviceMap[s.name] = s
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
	case 3:
		// normal method
		argType = mtype.In(1)
		replyType = mtype.In(2)
		contextType = nil
	case 4:
		// method that takes a context
		argType = mtype.In(2)
		replyType = mtype.In(3)
		contextType = mtype.In(1)
	default:
		log.Println("method", mname, "of", mtype, "has wrong number of ins:", mtype.NumIn())
		return nil
	}

	// First arg need not be a pointer.
	if !isExportedOrBuiltinType(argType) {
		log.Println(mname, "argument type not exported:", argType)
		return nil
	}

	// the second argument will tell us if it's a streaming call
	// or a regular call
	if replyType.Kind() == reflect.Func {
		// this is a streaming call
		stream = true
		if replyType.NumIn() != 1 {
			log.Println("method", mname, "sendReply has wrong number of ins:", replyType.NumIn())
			return nil
		}
		if replyType.In(0).Kind() != reflect.Interface {
			log.Println("method", mname, "sendReply parameter type not an interface:", replyType.In(0))
			return nil
		}
		if replyType.NumOut() != 1 {
			log.Println("method", mname, "sendReply has wrong number of outs:", replyType.NumOut())
			return nil
		}
		if returnType := replyType.Out(0); returnType != typeOfError {
			log.Println("method", mname, "sendReply returns", returnType.String(), "not error")
			return nil
		}

	} else if replyType.Kind() != reflect.Ptr {
		log.Println("method", mname, "reply type not a pointer:", replyType)
		return nil
	}

	// Reply type must be exported.
	if !isExportedOrBuiltinType(replyType) {
		log.Println("method", mname, "reply type not exported:", replyType)
		return nil
	}
	// Method needs one out.
	if mtype.NumOut() != 1 {
		log.Println("method", mname, "has wrong number of outs:", mtype.NumOut())
		return nil
	}
	// The return type of the method must be error.
	if returnType := mtype.Out(0); returnType != typeOfError {
		log.Println("method", mname, "returns", returnType.String(), "not error")
		return nil
	}
	return &methodType{method: method, ArgType: argType, ReplyType: replyType, ContextType: contextType, stream: stream}
}

func (server *Server) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusMethodNotAllowed)
		io.WriteString(w, "405 must POST\n")
		return
	}

	server.Serve(w, req)
}

func (server *Server) Serve(w http.ResponseWriter, req *http.Request) {
	sreq := &serverRequest{}

	server.encoder.Decode(req.Body, sreq)
	defer req.Body.Close()

	header := make(c.Metadata)
	for name, vals := range req.Header {
		header[name] = vals[0]
	}

	v, err := server.serveServiceMethod(header, sreq)

	resp := &serverResponse{}
	if err != nil {
		respErr, ok := err.(*ResponseError)
		if ok {
			resp.Error = respErr
		} else {
			resp.Error = &ResponseError{
				Code:    -32600,
				Message: err.Error(),
			}
		}
		w.WriteHeader(500)
		server.encoder.Encode(w, resp)
		return
	}

	w.WriteHeader(200)
	resp.Result = v
	resp.ID = sreq.ID
	server.encoder.Encode(w, resp)

}

func (server *Server) serveServiceMethod(header c.Metadata, sreq *serverRequest) (interface{}, error) {
	var err error
	serviceMethod := strings.Split(sreq.Method, ".")
	if len(serviceMethod) != 2 {
		err = errors.New("rpc: service/method request ill-formed: " + sreq.Method)
		return nil, err
	}
	// Look up the request.
	server.mu.Lock()
	s := server.serviceMap[serviceMethod[0]]
	server.mu.Unlock()
	if s == nil {
		err = errors.New("rpc: can't find service " + sreq.Method)
		return nil, err
	}
	mt := s.method[serviceMethod[1]]
	if mt == nil {
		err = errors.New("rpc: can't find method " + sreq.Method)
		return nil, err
	}

	ctx := c.WithMetadata(context.Background(), header)

	argv := reflect.New(mt.ArgType.Elem())
	replyv := reflect.New(mt.ReplyType.Elem())

	json.Unmarshal(*sreq.Params, argv.Interface())

	return server.call(ctx, mt, s, argv, replyv)
}

func (server *Server) call(ctx context.Context, mt *methodType, s *service, argv, replyv reflect.Value) (interface{}, error) {
	mt.Lock()
	mt.numCalls++
	mt.Unlock()

	function := mt.method.Func
	var returnValues []reflect.Value

	returnValues = function.Call([]reflect.Value{s.rcvr, mt.prepareContext(ctx), argv, replyv})

	var err error
	errInter := returnValues[0].Interface()
	if errInter != nil {
		err = errInter.(error)
	}

	return replyv.Interface(), err
}
