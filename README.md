Kuja [![godoc badge](http://godoc.org/github.com/plimble/ace?status.png)](http://godoc.org/github.com/plimble/ace)
========

Go microservice framework

##Features
- Client
    - Reuqest/Async Requests
    - Publish
    - Reties/Timeout
    - Service Discovery/Watch
    - Encoder
    - Circuit breaker
- Server
    - HTTP POST like RPC style
    - Publish/Subscribe Queue
    - Retries Queue
    - Graceful Shutdown
    - Custom Error
    - Register Etcd/Consul
    - Snappy
    - Encoder
    - Middleware
    - RateLimit

##TODO
- RateLimit
- Circuit breaker
- Trace
- Doc
- Test

##Usage

###Server

```go
type AddService struct{}

func (s *AddService) Add(ctx *Ctx, req *AddReq, resp *AddResp) error {
    resp.C = req.A + req.B
    return nil
}

func main() {
    server :=
}
```

