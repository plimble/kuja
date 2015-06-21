package client

import (
	"bytes"
	"crypto/tls"
	"errors"
	"github.com/facebookgo/httpcontrol"
	"github.com/golang/snappy/snappy"
	"github.com/plimble/kuja/broker"
	"github.com/plimble/kuja/encoder"
	"github.com/plimble/kuja/encoder/json"
	"github.com/plimble/kuja/registry"
	"net/http"
	"strings"
	"time"
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

type Method interface {
	GetAddress(service, method string) (string, error)
}

type Client interface {
	Broker(b broker.Broker)
	Publish(topic string, v interface{}, meta map[string]string) error
	Encoder(enc encoder.Encoder)
	AsyncRequests(as []AsyncRequest) []AsyncResponse
	DefaultHeader(hdr map[string]string)
	Request(service, method string, reqv interface{}, respv interface{}, header map[string]string) (int, error)
	DefaultTimeout(d time.Duration)
	DefaultMaxTries(n uint)
}

type HeaderFunc func(header http.Header)

type DefaultClient struct {
	method        Method
	encoder       encoder.Encoder
	broker        broker.Broker
	defaultHeader map[string]string
	retry         int
	timeout       time.Duration
	client        *http.Client
	httpControl   *httpcontrol.Transport
}

func New(url string, tlsConfig *tls.Config) Client {
	if strings.HasPrefix(url, "/") {
		url = url[:len(url)-1]
	}

	d := &DefaultClient{
		method:  &Direct{url},
		encoder: json.NewEncoder(),
		client:  &http.Client{},
		httpControl: &httpcontrol.Transport{
			RequestTimeout:  0,
			MaxTries:        0,
			TLSClientConfig: tlsConfig,
		},
	}

	d.client.Transport = d.httpControl
	return d
}

func NewWithRegistry(r registry.Registry, watch bool, tlsConfig *tls.Config) Client {
	if watch {
		r.Watch()
	}

	d := &DefaultClient{
		method: &Discovery{
			registry: r,
		},
		encoder: json.NewEncoder(),
		client:  &http.Client{},
		httpControl: &httpcontrol.Transport{
			RequestTimeout:  0,
			MaxTries:        0,
			TLSClientConfig: tlsConfig,
		},
	}

	d.client.Transport = d.httpControl
	return d
}

func (c *DefaultClient) DefaultTimeout(d time.Duration) {
	c.httpControl.RequestTimeout = d
}

func (c *DefaultClient) DefaultMaxTries(n uint) {
	c.httpControl.MaxTries = n
}

func (c *DefaultClient) Broker(b broker.Broker) {
	c.broker = b
	if err := c.broker.Connect(); err != nil {
		panic(err)
	}
}

func (c *DefaultClient) DefaultHeader(hdr map[string]string) {
	c.defaultHeader = hdr
}

func (c *DefaultClient) Publish(topic string, v interface{}, meta map[string]string) error {
	data, err := c.encoder.Marshal(v)
	if err != nil {
		return err
	}

	return c.broker.Publish(topic, broker.NewMessage(meta, data))
}

func (c *DefaultClient) Encoder(enc encoder.Encoder) {
	c.encoder = enc
}

func (c *DefaultClient) AsyncRequests(as []AsyncRequest) []AsyncResponse {
	done := make(chan AsyncResponse, len(as))
	// arespChan := make(chan *AsyncResponse, len(as))
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

	dataReq, err := c.encoder.Marshal(reqv)
	if err != nil {
		return 0, err
	}
	buf := bytes.NewBuffer(dataReq)

	req, err := http.NewRequest("POST", addr, buf)
	if err != nil {
		return 0, err
	}

	for name, val := range c.defaultHeader {
		req.Header.Set(name, val)
	}

	for name, val := range header {
		req.Header.Set(name, val)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return 0, err
	}
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

	err = c.encoder.Unmarshal(respData, respv)
	if err != nil {
		return 0, err
	}

	return resp.StatusCode, err
}
