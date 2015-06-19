package client

import (
	"bytes"
	"errors"
	"github.com/golang/snappy/snappy"
	"github.com/plimble/kuja/broker"
	"github.com/plimble/kuja/encoder"
	"github.com/plimble/kuja/encoder/json"
	"github.com/plimble/kuja/registry"
	"net/http"
	"strings"
	"sync"
)

//go:generate mockery -name Client

type Method interface {
	GetAddress(service, method string) (string, error)
}

type Client interface {
	Call(service, method string, reqv interface{}, respv interface{}, header http.Header) (int, error)
	Encoder(enc encoder.Encoder)
}

type HeaderFunc func(header http.Header)

type DefaultClient struct {
	pool          sync.Pool
	method        Method
	encoder       encoder.Encoder
	broker        broker.Broker
	DefaultHeader http.Header
}

func New(url string) *DefaultClient {
	if strings.HasPrefix(url, "/") {
		url = url[:len(url)-1]
	}

	d := &DefaultClient{
		method:  &Direct{url},
		encoder: json.NewEncoder(),
	}

	d.pool.New = func() interface{} {
		return bytes.NewBuffer(make([]byte, 0, bytes.MinRead))
	}

	return d
}

func NewWithRegistry(r registry.Registry, watch bool) *DefaultClient {
	if watch {
		r.Watch()
	}

	return &DefaultClient{
		method: &Discovery{
			registry: r,
		},
		encoder: json.NewEncoder(),
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

	msg := broker.Message{
		Header: meta,
		Body:   data,
	}

	msgData, err := msg.Marshal()
	if err != nil {
		return err
	}

	return c.broker.Publish(service+"."+topic, msgData)
}

func (c *DefaultClient) Encoder(enc encoder.Encoder) {
	c.encoder = enc
}

func (c *DefaultClient) AsyncRequest(service, method string, reqv interface{}, respv interface{}, header http.Header) error {
	var err error

	addr, err := c.method.GetAddress(service, method)
	if err != nil {
		return err
	}

	data, err := c.encoder.Marshal(reqv)
	if err != nil {
		return err
	}
	buf := bytes.NewBuffer(data)

	req, err := http.NewRequest("POST", addr, buf)
	if err != nil {

		return err
	}

	for name, val := range c.DefaultHeader {
		req.Header.Set(name, val[0])
	}

	for name, val := range header {
		req.Header.Set(name, val[0])
	}

	go http.DefaultClient.Do(req)

	return nil
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

	resp, err := http.DefaultClient.Do(req)
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
