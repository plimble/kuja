package mocks

import "github.com/stretchr/testify/mock"

import "github.com/plimble/kuja/encoder"

import "net/http"

type Client struct {
	mock.Mock
}

func (m *Client) Call(service string, method string, reqv interface{}, respv interface{}, header http.Header) (int, error) {
	ret := m.Called(service, method, reqv, respv, header)

	r0 := ret.Get(0).(int)
	r1 := ret.Error(1)

	return r0, r1
}
func (m *Client) Encoder(enc encoder.Encoder) {
	m.Called(enc)
}
