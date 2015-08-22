package kuja

import (
	"crypto/tls"
	"io"
	"net"
	"net/http"
	"reflect"
	"strings"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/golang/snappy"
	"github.com/plimble/errors"
	"github.com/plimble/kuja/broker"
	"github.com/plimble/kuja/encoder"
	"github.com/plimble/kuja/encoder/json"
	"github.com/plimble/kuja/registry"
	"github.com/satori/go.uuid"
	"golang.org/x/net/netutil"
	"gopkg.in/tylerb/graceful.v1"
)

const (
	DEVELOPMENT = "development"
	STAGING     = "staging"
	PRODUCTION  = "production"
)

type Handler func(ctx *Context, w http.ResponseWriter, r *http.Request) error

type Server struct {
	stage           string
	id              string
	pool            sync.Pool
	middleware      []Handler
	mu              sync.Mutex // protects the serviceMap
	serviceMap      map[string]*service
	subscriberMap   map[string]*subscriber
	broker          broker.Broker
	encoder         encoder.Encoder
	snappy          bool
	serviceError    ServiceErrorFunc
	subscriberError SubscriberErrorFunc
	registry        registry.Registry
	srv             *graceful.Server
}

func NewServer() *Server {
	server := &Server{
		stage:           DEVELOPMENT,
		id:              uuid.NewV1().String(),
		serviceMap:      make(map[string]*service),
		encoder:         json.NewEncoder(),
		serviceError:    defaulServiceErr,
		subscriberError: defaulSubscriberErr,
		subscriberMap:   make(map[string]*subscriber),
	}

	server.pool.New = func() interface{} {
		return &Context{
			ReqMetadata:  make(Metadata),
			RespMetadata: make(Metadata),
			returnValues: make([]reflect.Value, 1),
			serviceError: server.serviceError,
			isResp:       false,
		}
	}

	return server
}

func (server *Server) Id() string {
	return server.id
}

func (server *Server) Use(h ...Handler) {
	server.middleware = append(server.middleware, h...)
}

func (server *Server) Stage(stage string) {
	server.stage = stage
}

func (server *Server) Snappy(enable bool) {
	server.snappy = enable
}

func (server *Server) Registry(r registry.Registry) {
	server.registry = r
}

func (server *Server) Encoder(enc encoder.Encoder) {
	server.encoder = enc
}

func (server *Server) Broker(b broker.Broker) {
	server.broker = b
}

func (server *Server) Run(addr string, timeout time.Duration) {
	server.srv = &graceful.Server{
		Timeout: timeout,
		Server: &http.Server{
			Addr:    addr,
			Handler: server,
		},
	}

	l, err := net.Listen("tcp", addr)
	if err != nil {
		log.Error(err)
		return
	}

	if server.srv.ListenLimit != 0 {
		l = netutil.LimitListener(l, server.srv.ListenLimit)
	}

	if err := server.startRegistry("http://", addr); err != nil {
		log.Error(err)
		return
	}

	if err := server.startBroker(); err != nil {
		log.Error(err)
		return
	}

	if err := server.startSubscribe(); err != nil {
		log.Error(err)
		return
	}

	log.Infof("Start server id %s on %s stage %s", server.id, addr, server.stage)
	server.srv.Serve(l)
	log.Info("Stop server")
	server.stopRegistry()
	server.stopBroker()
}

func (server *Server) Close() {
	server.srv.Stop(0)
}

func (server *Server) RunTLS(addr string, timeout time.Duration, certFile, keyFile string) {
	server.srv = &graceful.Server{
		Timeout: timeout,
		Server: &http.Server{
			Addr:    addr,
			Handler: server,
		},
	}

	config := &tls.Config{}
	if server.srv.TLSConfig != nil {
		*config = *server.srv.TLSConfig
	}

	if config.NextProtos == nil {
		config.NextProtos = []string{"http/1.1"}
	}

	var err error
	config.Certificates = make([]tls.Certificate, 1)
	config.Certificates[0], err = tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		log.Error(err)
		return
	}

	conn, err := net.Listen("tcp", addr)
	if err != nil {
		log.Error(err)
		return
	}

	tlsListener := tls.NewListener(conn, config)

	if err := server.startRegistry("https://", addr); err != nil {
		log.Error(err)
		return
	}

	if err := server.startBroker(); err != nil {
		log.Error(err)
		return
	}

	if err := server.startSubscribe(); err != nil {
		log.Error(err)
		return
	}

	log.Infof("Start server id %s on %s stage %s", server.id, addr, server.stage)
	server.srv.Serve(tlsListener)
	log.Info("Stop server")
	server.stopRegistry()
	server.stopBroker()
}

func (server *Server) startRegistry(scheme, addr string) error {
	if server.registry == nil {
		return nil
	}

	var host string
	var port string
	parts := strings.Split(addr, ":")
	if len(parts) > 1 {
		host = strings.Join(parts[:len(parts)-1], ":")
		port = parts[len(parts)-1]
	} else {
		host = parts[0]
	}

	for _, service := range server.serviceMap {
		service.node = &registry.Node{
			Id:      service.id,
			Name:    service.name,
			Host:    host,
			Port:    port,
			Address: addr,
			URL:     scheme + addr,
		}
		err := server.registry.Register(service.node)
		log.Infof("Registerd %s %s %s", service.name, service.id, addr)
		if err != nil {
			return err
		}
	}

	return nil
}

func (server *Server) stopRegistry() {
	if server.registry != nil {
		for _, service := range server.serviceMap {
			if service.node == nil {
				continue
			}
			err := server.registry.Deregister(service.node)
			if err != nil {
				log.WithFields(log.Fields{
					"error":      err,
					"service":    service.name,
					"service_id": service.id,
				}).Error("Unable to deregister")
				continue
			}
			log.Infof("Deregisterd %s %s", service.name, service.id)
		}
		server.registry.Close()
	}
}

func (server *Server) startBroker() error {
	if server.broker == nil {
		return nil
	}

	if err := server.broker.Connect(); err != nil {
		return err
	}

	log.Infof("Connected to broker")

	return nil
}

func (server *Server) stopBroker() {
	if server.broker == nil {
		return
	}

	server.broker.Close()
	log.Infof("Close broker connection")
}

func (server *Server) startSubscribe() error {
	if server.broker == nil {
		return nil
	}

	for _, s := range server.subscriberMap {
		server.broker.Subscribe(s.topic, s.queue, server.id, s.size, s.handler)
		log.Infof("Subscribe topic: %s queue: %s size: %d", s.topic, s.queue, s.size)
	}

	return nil
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
	if req.Method == "GET" && req.URL.Path == "/health" {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, "ok\n")
		return
	}

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

	ctx := server.pool.Get().(*Context)

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

func respError(err error, ctx *Context) {
	if errs, ok := err.(errors.Error); ok {
		ctx.isResp = true
		go ctx.serviceError(ctx.ServiceID, ctx.ServiceName, ctx.MethodName, errs.Code(), err)
		ctx.w.WriteHeader(errs.Code())
		ctx.w.Write([]byte(errs.Error()))
	} else {
		ctx.isResp = true
		go ctx.serviceError(ctx.ServiceID, ctx.ServiceName, ctx.MethodName, 500, err)
		ctx.w.WriteHeader(500)
		ctx.w.Write([]byte(err.Error()))
	}
}

func serve(ctx *Context) error {
	argv := reflect.New(ctx.mt.ArgType.Elem())
	var replyv interface{}

	argvInter := argv.Interface()
	err := ctx.encoder.Decode(ctx.req.Body, argvInter)
	ctx.req.Body.Close()
	if err != nil {
		return errors.InternalError("unable to encode response")
	}

	function := ctx.mt.method.Func
	if ctx.mt.ContextType == nil {
		ctx.returnValues = function.Call([]reflect.Value{ctx.rcvr, argv})
	} else {
		ctx.returnValues = function.Call([]reflect.Value{ctx.rcvr, ctx.mt.prepareContext(ctx), argv})
	}

	if len(ctx.returnValues) == 1 {
		if ctx.returnValues[0].Interface() != nil {
			return ctx.returnValues[0].Interface().(error)
		}
	} else if len(ctx.returnValues) == 2 {
		if ctx.returnValues[1].Interface() != nil {
			return ctx.returnValues[1].Interface().(error)
		}
		replyv = ctx.returnValues[0].Interface()
	}

	for name, val := range ctx.RespMetadata {
		ctx.w.Header().Set(name, val)
	}

	if replyv != nil {
		if ctx.snappy {
			data, err := ctx.encoder.Marshal(replyv)
			if err != nil {
				return err
			}
			data = snappy.Encode(nil, data)
			ctx.isResp = true
			ctx.w.Header().Set("Snappy", "true")
			ctx.w.WriteHeader(200)
			ctx.w.Write(data)
		} else {
			ctx.isResp = true
			ctx.w.WriteHeader(200)
			ctx.encoder.Encode(ctx.w, replyv)
		}
	} else {
		ctx.isResp = true
		ctx.w.WriteHeader(200)
		ctx.w.Write(nil)
	}

	return nil
}
