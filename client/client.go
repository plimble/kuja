package client

import (
	"bytes"
	"errors"
	"github.com/facebookgo/httpcontrol"
	"github.com/golang/snappy/snappy"
	"github.com/plimble/kuja/broker"
	"github.com/plimble/kuja/encoder"
	"github.com/plimble/kuja/encoder/json"
	"github.com/plimble/kuja/registry"
	"net/http"
	"strings"
	"sync"
	"time"
)

//go:generate mockery -name Client

type AsyncRequest struct {
	Service  string
	Method   string
	Request  interface{}
	Response interface{}
	Metadata http.Header
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
	Publish(service, topic string, v interface{}, meta map[string]string) error
	Encoder(enc encoder.Encoder)
	AsyncRequests(as []AsyncRequest) []AsyncResponse
	Request(service, method string, reqv interface{}, respv interface{}, header http.Header) (int, error)
}

type HeaderFunc func(header http.Header)

type DefaultClient struct {
	pool          sync.Pool
	method        Method
	encoder       encoder.Encoder
	broker        broker.Broker
	DefaultHeader http.Header
	retry         int
	timeout       time.Duration
	client        *http.Client
}

func New(url string) Client {
	if strings.HasPrefix(url, "/") {
		url = url[:len(url)-1]
	}

	d := &DefaultClient{
		method:  &Direct{url},
		encoder: json.NewEncoder(),
		client: &http.Client{
			Transport: &httpcontrol.Transport{
				RequestTimeout: 0,
				MaxTries:       0,
			},
		},
	}

	d.pool.New = func() interface{} {
		return bytes.NewBuffer(make([]byte, 0, bytes.MinRead))
	}

	return d
}

func NewWithRegistry(r registry.Registry, watch bool) Client {
	if watch {
		r.Watch()
	}

	return &DefaultClient{
		method: &Discovery{
			registry: r,
		},
		encoder: json.NewEncoder(),
		client: &http.Client{
			Transport: &httpcontrol.Transport{
				RequestTimeout: 0,
				MaxTries:       0,
			},
		},
	}
}

func (c *DefaultClient) Broker(b broker.Broker) {
	c.broker = b
	if err := c.broker.Connect(); err != nil {
		panic(err)
	}
}

func (c *DefaultClient) Publish(service, topic string, v interface{}, meta map[string]string) error {
	data, err := c.encoder.Marshal(v)
	if err != nil {
		return err
	}

	msg := &broker.Message{
		Header: meta,
		Body:   data,
	}

	return c.broker.Publish(service+"."+topic, msg)
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

func (c *DefaultClient) Request(service, method string, reqv interface{}, respv interface{}, header http.Header) (int, error) {
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

	for name, val := range c.DefaultHeader {
		req.Header.Set(name, val[0])
	}

	for name, val := range header {
		req.Header.Set(name, val[0])
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
