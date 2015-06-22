package client

import (
	"crypto/tls"
	"github.com/afex/hystrix-go/hystrix"
	"github.com/plimble/kuja/broker"
	"github.com/plimble/kuja/encoder"
	"github.com/plimble/kuja/encoder/json"
)

var (
	DefaultEncoder                = json.NewEncoder()
	DefaultTimeout                = 1000
	DefaultMaxConcurrentRequests  = 0
	DefaultRequestVolumeThreshold = 5
	DefaultSleepWindow            = 1000
	DefaultErrorPercentThreshold  = 50
)

type Option func(o *option)

type option struct {
	Broker                 broker.Broker
	Encoder                encoder.Encoder
	Header                 map[string]string
	Timeout                int
	MaxConcurrentRequests  int
	RequestVolumeThreshold int
	SleepWindow            int
	ErrorPercentThreshold  int
	ErrMaxConcurrency      string
	ErrCircuitOpen         string
	ErrTimeout             string
	TLSConfig              *tls.Config
}

func newOption() *option {
	return &option{
		Encoder:                DefaultEncoder,
		Timeout:                DefaultTimeout,
		MaxConcurrentRequests:  DefaultMaxConcurrentRequests,
		RequestVolumeThreshold: DefaultRequestVolumeThreshold,
		SleepWindow:            DefaultSleepWindow,
		ErrorPercentThreshold:  DefaultErrorPercentThreshold,
	}
}

func Broker(b broker.Broker) Option {
	return func(o *option) {
		o.Broker = b
	}
}

func Encoder(enc encoder.Encoder) Option {
	return func(o *option) {
		o.Encoder = enc
	}
}

func Header(hdr map[string]string) Option {
	return func(o *option) {
		o.Header = hdr
	}
}

func Timeout(n int) Option {
	return func(o *option) {
		o.Timeout = n
	}
}

func MaxConcurrentRequests(n int) Option {
	return func(o *option) {
		o.MaxConcurrentRequests = n
	}
}

func RequestVolumeThreshold(n int) Option {
	return func(o *option) {
		o.RequestVolumeThreshold = n
	}
}

func SleepWindow(n int) Option {
	return func(o *option) {
		o.SleepWindow = n
	}
}

func ErrorPercentThreshold(n int) Option {
	return func(o *option) {
		o.ErrorPercentThreshold = n
	}
}

func ErrMaxConcurrency(s string) Option {
	return func(o *option) {
		hystrix.ErrMaxConcurrency = hystrix.CircuitError{Message: s}
	}
}

func ErrCircuitOpen(s string) Option {
	return func(o *option) {
		hystrix.ErrCircuitOpen = hystrix.CircuitError{Message: s}
	}
}

func ErrTimeout(s string) Option {
	return func(o *option) {
		hystrix.ErrTimeout = hystrix.CircuitError{Message: s}
	}
}

func TLSConfig(config *tls.Config) Option {
	return func(o *option) {
		o.TLSConfig = config
	}
}
