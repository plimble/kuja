Kuja [![godoc badge](http://godoc.org/github.com/plimble/kuja?status.png)](http://godoc.org/github.com/plimble/kuja)
========

Go webservice framework

##Features
- Client
    - Reuqest/Async Requests
    - Publish
    - Service Discovery/Watch
    - Encoder
    - Circuit breaker
- Server
    - HTTP POST like RPC style
    - Publish/Subscribe Queue
    - Retries/Timeout Queue
    - Graceful Shutdown
    - Custom Error
    - Register Etcd/Consul and Hartbeat
    - Snappy
    - Encoder
    - Middleware
    - Health Endpoint (/health)
    - Built-in Client
    - Metric

##TODO
- Config func
- Circuit breaker Publish
- Circuit breaker option and custom error
- Circuit breaker metric prometheus
- Close method for client
- Built-in Client
- Metric
- prometheus metric
- Doc
- [Phosphor Distributed Tracing system](https://github.com/mattheath/phosphor)
- Test

##Usage
See in [example](https://github.com/plimble/kuja/tree/master/example)

##Note
This framwork is designed for internal services, no rate limit, connection limit and securities.
Only have tls for protocal security. Please add api gateway or create securities by yourself.


