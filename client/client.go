package client

import (
	"bytes"
	"errors"
	"github.com/golang/snappy/snappy"
	"github.com/plimble/kuja/encoder"
	"github.com/plimble/kuja/encoder/json"
	"github.com/plimble/kuja/registry"
	"io/ioutil"
	"net/http"
	"strings"
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
	method        Method
	encoder       encoder.Encoder
	DefaultHeader http.Header
}

func New(url string) *DefaultClient {
	if strings.HasPrefix(url, "/") {
		url = url[:len(url)-1]
	}
	return &DefaultClient{
		method:  &Direct{url},
		encoder: json.NewEncoder(),
	}
}

func NewWithRegistry(r registry.Registry) *DefaultClient {
	return &DefaultClient{
		method:  &Discovery{r},
		encoder: json.NewEncoder(),
	}
}

func (c *DefaultClient) Encoder(enc encoder.Encoder) {
	c.encoder = enc
}

func (c *DefaultClient) Call(service, method string, reqv interface{}, respv interface{}, header http.Header) (int, error) {
	var err error

	addr, err := c.method.GetAddress(service, method)
	if err != nil {
		return 0, err
	}

	data, err := c.encoder.Marshal(reqv)
	if err != nil {
		return 0, err
	}
	buf := bytes.NewBuffer(data)

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

	respData, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return 0, err
	}

	if resp.StatusCode != 200 {
		return resp.StatusCode, errors.New(string(respData))
	}

	if resp.Header.Get("Snappy") == "true" {
		respData, err = snappy.Decode(nil, respData)
		if err != nil {
			return 0, err
		}
	}

	err = c.encoder.Unmarshal(respData, respv)
	resp.Body.Close()
	if err != nil {
		return 0, err
	}

	return resp.StatusCode, err
}

func concat(s ...string) string {
	size := 0
	for i := 0; i < len(s); i++ {
		size += len(s[i])
	}

	buf := make([]byte, 0, size)

	for i := 0; i < len(s); i++ {
		buf = append(buf, []byte(s[i])...)
	}

	return string(buf)
}
