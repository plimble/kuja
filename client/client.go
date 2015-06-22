package client

import (
	"bytes"
	"errors"
	"github.com/afex/hystrix-go/hystrix"
	"github.com/golang/snappy/snappy"
	"github.com/plimble/kuja/broker"
	"github.com/plimble/kuja/registry"
	"net/http"
	"strings"
)

//go:generate mockery -name Client

type AsyncRequest struct {
	Service  string
	Method   string
	Request  interface{}
	Response interface{}
	Metadata map[string]string
}

type AsyncResponse struct {
	Status int
	Error  error
}

type CircuitBrakerConfig struct {
	Timeout                int
	MaxConcurrentRequests  int
	RequestVolumeThreshold int
	SleepWindow            int
	ErrorPercentThreshold  int
}

type Method interface {
	GetAddress(service, method string) (string, error)
}

type Client interface {
	Publish(topic string, v interface{}, meta map[string]string) error
	AsyncRequests(as []AsyncRequest) []AsyncResponse
	Request(service, method string, reqv interface{}, respv interface{}, header map[string]string) (int, error)
	Close()
}

type HeaderFunc func(header http.Header)

type DefaultClient struct {
	*option
	method        Method
	client        *http.Client
	hystrixConfig map[string]struct{}
}

func New(url string, opts ...Option) (Client, error) {
	if strings.HasPrefix(url, "/") {
		url = url[:len(url)-1]
	}

	opt := newOption()
	for i := 0; i < len(opts); i++ {
		opts[i](opt)
	}

	if opt.Broker != nil {
		if err := opt.Broker.Connect(); err != nil {
			return nil, err
		}
	}

	return &DefaultClient{
		option:        opt,
		method:        &Direct{url},
		hystrixConfig: make(map[string]struct{}),
		client: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: opt.TLSConfig,
			},
		},
	}, nil
}

func NewWithRegistry(r registry.Registry, watch bool, opts ...Option) (Client, error) {
	if watch {
		r.Watch()
	}

	opt := newOption()
	for i := 0; i < len(opts); i++ {
		opts[i](opt)
	}

	if opt.Broker != nil {
		if err := opt.Broker.Connect(); err != nil {
			return nil, err
		}
	}

	return &DefaultClient{
		option:        opt,
		method:        &Discovery{r},
		hystrixConfig: make(map[string]struct{}),
		client: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: opt.TLSConfig,
			},
		},
	}, nil
}

func (c *DefaultClient) Close() {
	if c.Broker != nil {
		c.Broker.Close()
	}
}

func (c *DefaultClient) Publish(topic string, v interface{}, meta map[string]string) error {
	data, err := c.Encoder.Marshal(v)
	if err != nil {
		return err
	}

	return c.Broker.Publish(topic, broker.NewMessage(meta, data))
}

func (c *DefaultClient) AsyncRequests(as []AsyncRequest) []AsyncResponse {
	done := make(chan AsyncResponse, len(as))
	aresps := make([]AsyncResponse, 0, len(as))

	for i := 0; i < len(as); i++ {
		go func(index int) {
			status, err := c.Request(as[index].Service, as[index].Method, as[index].Request, as[index].Response, as[index].Metadata)
			done <- AsyncResponse{status, err}
		}(i)
	}

	for aresp := range done {
		aresps = append(aresps, aresp)
		if len(aresps) == len(as) {
			break
		}
	}

	return aresps
}

func (c *DefaultClient) Request(service, method string, reqv interface{}, respv interface{}, header map[string]string) (int, error) {
	var err error

	addr, err := c.method.GetAddress(service, method)
	if err != nil {
		return 0, err
	}

	dataReq, err := c.Encoder.Marshal(reqv)
	if err != nil {
		return 0, err
	}
	buf := bytes.NewBuffer(dataReq)

	req, err := http.NewRequest("POST", addr, buf)
	if err != nil {
		return 0, err
	}

	for name, val := range c.Header {
		req.Header.Set(name, val)
	}

	for name, val := range header {
		req.Header.Set(name, val)
	}

	if _, ok := c.hystrixConfig[service]; !ok {
		hystrix.ConfigureCommand(service, hystrix.CommandConfig{
			Timeout:                c.Timeout,
			MaxConcurrentRequests:  c.MaxConcurrentRequests,
			RequestVolumeThreshold: c.RequestVolumeThreshold,
			SleepWindow:            c.SleepWindow,
			ErrorPercentThreshold:  c.ErrorPercentThreshold,
		})
		c.hystrixConfig[service] = struct{}{}
	}

	output := make(chan *http.Response, 1)
	errChan := hystrix.Go(service, func() error {
		resp, err := c.client.Do(req)
		if err != nil {
			return err
		}
		output <- resp
		return nil
	}, nil)

	select {
	case resp := <-output:
		// success
		buf.Reset()
		buf.ReadFrom(resp.Body)
		resp.Body.Close()

		if resp.StatusCode != 200 {
			return resp.StatusCode, errors.New(string(buf.Bytes()))
		}

		respData := buf.Bytes()
		if resp.Header.Get("Snappy") == "true" {
			respData, err = snappy.Decode(nil, buf.Bytes())
			if err != nil {

				return 0, err
			}
		}

		err = c.Encoder.Unmarshal(respData, respv)
		if err != nil {
			return 0, err
		}

		return resp.StatusCode, err
	case err := <-errChan:
		return 0, err
	}
}
