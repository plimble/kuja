package kuja

import (
	"github.com/golang/snappy/snappy"
	"github.com/plimble/kuja/encoder"
	"github.com/plimble/kuja/encoder/json"
	"io"
	"log"
	"net/http"
	"reflect"
	"strings"
	"sync"
)

type Handler func(ctx *Ctx, w http.ResponseWriter, r *http.Request) error
type LogErrorFunc func(serviceID, service, method string, status int, err error)

type Server struct {
	pool       sync.Pool
	middleware []Handler
	mu         sync.Mutex // protects the serviceMap
	serviceMap map[string]*service
	encoder    encoder.Encoder
	snappy     bool
	logError   LogErrorFunc
}

func defaulLogErr(serviceID, service, method string, status int, err error) {
	log.Printf("Error %s %s %s %d %s", serviceID, service, method, status, err)
}

func NewServer() *Server {
	server := &Server{
		serviceMap: make(map[string]*service),
		encoder:    json.NewEncoder(),
		logError:   defaulLogErr,
	}

	server.pool.New = func() interface{} {
		return &Ctx{
			ReqMetadata:  make(Metadata),
			RespMetadata: make(Metadata),
			returnValues: make([]reflect.Value, 1),
			logError:     server.logError,
			isResp:       false,
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
	ctx.ServiceID = s.id
	ctx.ServiceName = serviceName
	ctx.MethodName = methodName

	if len(s.handlers) > 0 {
		if err := s.handlers[0](ctx, w, req); err != nil && !ctx.isResp {
			respError(err, ctx)
		}
	} else {
		if err := serve(ctx); err != nil && !ctx.isResp {
			respError(err, ctx)
		}
	}
	server.pool.Put(ctx)
}

func respError(err error, ctx *Ctx) {
	if errs, ok := err.(Errors); ok {
		ctx.isResp = true
		go ctx.logError(ctx.ServiceID, ctx.ServiceName, ctx.MethodName, errs.Status(), err)
		ctx.w.WriteHeader(errs.Status())
		ctx.w.Write([]byte(errs.Error()))
	} else {
		ctx.isResp = true
		go ctx.logError(ctx.ServiceID, ctx.ServiceName, ctx.MethodName, 500, err)
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

	for name, val := range ctx.RespMetadata {
		ctx.w.Header().Set(name, val)
	}

	if ctx.snappy {
		data, err := ctx.encoder.Marshal(replyv.Interface())
		if err != nil {
			return err
		}
		data, err = snappy.Encode(nil, data)
		if err != nil {
			return err
		}
		ctx.isResp = true
		ctx.w.Header().Set("Snappy", "true")
		ctx.w.WriteHeader(200)
		ctx.w.Write(data)
	} else {
		ctx.isResp = true
		ctx.w.WriteHeader(200)
		ctx.encoder.Encode(ctx.w, replyv.Interface())
	}

	return nil
}
