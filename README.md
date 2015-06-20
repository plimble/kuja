Kuja [![godoc badge](http://godoc.org/github.com/plimble/ace?status.png)](http://godoc.org/github.com/plimble/ace)
========

Go microservice framework

##Features
- Client
  - Reuqest/Async Requests
  - Reties/Timeout
  - Service Discovery
  - Encoder
- Server
  - HTTP POST like RPC style
  - Pub/Sub Queue
  - Retries Queue
  - Graceful Shutdown
  - Custom Error
  - Registry Etcd/Consul
  - Snappy
  - Encoder
  - Middleware


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

