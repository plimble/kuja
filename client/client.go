package client

import (
	"bytes"
	"errors"
	"github.com/plimble/kuja/encoder"
	"github.com/plimble/kuja/registry"
	"io/ioutil"
	"net/http"
	"strings"
)

type Method interface {
	GetEndpoint(service, method string) string
}

type HeaderFunc func(header http.Header)
type QueryFunc func(header http.Header)

type Client struct {
	method     Method
	encoder    encoder.Encoder
	headerFunc HeaderFunc
}

func New(url string) *Client {
	if strings.HasPrefix(url, "/") {
		url = url[:len(url)-1]
	}
	return &Client{
		method: &Direct{url},
	}
}

func NewWithRegistry(r registry.Registry) {

}

func (c *Client) Encoder(enc encoder.Encoder) {
	c.encoder = enc
}

func (c *Client) Header(headerFunc HeaderFunc) {
	c.headerFunc = headerFunc
}

func (c *Client) Call(service, method string, reqv interface{}, respv interface{}, header http.Header) (int, error) {
	data, err := c.encoder.Marshal(reqv)
	if err != nil {
		return 0, err
	}
	buf := bytes.NewBuffer(data)

	req, err := http.NewRequest("POST", c.method.GetEndpoint(service, method), buf)
	if err != nil {
		return 0, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, err
	}

	if resp.StatusCode != 200 {
		errData, err := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return 0, err
		}

		return resp.StatusCode, errors.New(string(errData))
	}

	err = c.encoder.Decode(resp.Body, respv)
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
