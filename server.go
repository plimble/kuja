package kuja

import (
	"github.com/golang/snappy/snappy"
	"github.com/plimble/kuja/encoder"
	"io"
	"net/http"
	"reflect"
	"strings"
	"sync"
)

type Handler func(ctx *Ctx, w http.ResponseWriter, r *http.Request) error

type service struct {
	name     string                 // name of service
	rcvr     reflect.Value          // receiver of methods for the service
	typ      reflect.Type           // type of the receiver
	method   map[string]*methodType // registered methods
	handlers []Handler
}

type Server struct {
	pool       sync.Pool
	middleware []Handler
	mu         sync.Mutex // protects the serviceMap
	serviceMap map[string]*service
	encoder    encoder.Encoder
	snappy     bool
}

func NewServer() *Server {
	server := &Server{
		serviceMap: make(map[string]*service),
	}

	server.pool.New = func() interface{} {
		return &Ctx{
			ReqMetadata:  make(Metadata),
			RespMetadata: make(Metadata),
			returnValues: make([]reflect.Value, 1),
		}
	}

	return server
}

func (server *Server) Use(h ...Handler) {
	server.middleware = append(server.middleware, h...)
}

func (server *Server) Snappy(enable bool) {
	server.snappy = enable
}

func (server *Server) Service(service interface{}, h ...Handler) {
	if err := server.register(service, "", false, h); err != nil {
		panic(err)
	}
}

func (server *Server) Encoder(enc encoder.Encoder) {
	server.encoder = enc
}

func (server *Server) Run(addr string) {
	http.ListenAndServe(addr, server)
}

func getServiceMethod(s string) (string, string) {
	if strings.HasPrefix(s, "/") {
		s = s[1:]
	}

	if strings.HasSuffix(s, "/") {
		s = s[:len(s)-1]
	}

	for i := 0; i < len(s); i++ {
		if s[i] == '/' {
			return s[:i], s[i+1:]
		}
	}

	return "", ""
}

func (server *Server) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusMethodNotAllowed)
		io.WriteString(w, "405 must POST\n")
		return
	}

	serviceName, methodName := getServiceMethod(req.URL.Path)

	if serviceName == "" || methodName == "" {
		w.WriteHeader(404)
		w.Write([]byte("rpc: can't find service or method"))
		return
	}

	ctx := server.pool.Get().(*Ctx)
	// defer server.pool.Put(ctx)

	for name, vals := range req.Header {
		ctx.ReqMetadata[name] = vals[0]
	}

	// server.mu.Lock()
	s := server.serviceMap[serviceName]
	// server.mu.Unlock()
	if s == nil {
		w.WriteHeader(404)
		w.Write([]byte("rpc: can't find service " + serviceName))
		return
	}
	mt := s.method[methodName]
	if mt == nil {
		w.WriteHeader(404)
		w.Write([]byte("rpc: can't find method " + methodName))
		return
	}

	ctx.handlers = s.handlers
	ctx.mt = mt
	ctx.req = req
	ctx.w = w
	ctx.rcvr = s.rcvr
	ctx.encoder = server.encoder
	ctx.snappy = server.snappy

	if len(s.handlers) > 0 {
		if err := s.handlers[0](ctx, w, req); err != nil {
			server.respError(err, ctx)
		}
	} else {
		if err := serve(ctx); err != nil {
			server.respError(err, ctx)
		}
	}
	server.pool.Put(ctx)
}

func (server *Server) respError(err error, ctx *Ctx) {
	if errs, ok := err.(Errors); ok {
		ctx.w.WriteHeader(errs.Status())
		ctx.w.Write([]byte(errs.Error()))
	} else {
		ctx.w.WriteHeader(500)
		ctx.w.Write([]byte(err.Error()))
	}
}

func serve(ctx *Ctx) error {
	argv := reflect.New(ctx.mt.ArgType.Elem())
	replyv := reflect.New(ctx.mt.ReplyType.Elem())

	argvInter := argv.Interface()
	err := ctx.encoder.Decode(ctx.req.Body, argvInter)
	ctx.req.Body.Close()
	if err != nil {
		return Error(500, "unable to encode response")
	}

	function := ctx.mt.method.Func
	ctx.returnValues = function.Call([]reflect.Value{ctx.rcvr, ctx.mt.prepareContext(ctx), argv, replyv})

	if ctx.returnValues[0].Interface() != nil {
		return ctx.returnValues[0].Interface().(error)
	}

	if ctx.snappy {
		ctx.w.Header().Set("Snappy", "true")
		ctx.w.WriteHeader(200)
		data, _ := ctx.encoder.Marshal(replyv.Interface())
		data, _ = snappy.Encode(nil, data)
		ctx.w.Write(data)
	} else {
		ctx.w.WriteHeader(200)
		ctx.encoder.Encode(ctx.w, replyv.Interface())
	}

	return nil
}
