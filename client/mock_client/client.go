package mock_client

import "github.com/plimble/kuja/client"
import "github.com/stretchr/testify/mock"

type MockClient struct {
	mock.Mock
}

func NewMockClient() *MockClient {
	return &MockClient{}
}

func (m *MockClient) Publish(topic string, v interface{}, meta map[string]string) error {
	ret := m.Called(topic, v, meta)

	r0 := ret.Error(0)

	return r0
}
func (m *MockClient) AsyncRequests(as []client.AsyncRequest) []client.AsyncResponse {
	ret := m.Called(as)

	var r0 []client.AsyncResponse
	if ret.Get(0) != nil {
		r0 = ret.Get(0).([]client.AsyncResponse)
	}

	return r0
}
func (m *MockClient) Request(service string, method string, reqv interface{}, respv interface{}, header map[string]string) (int, error) {
	ret := m.Called(service, method, reqv, respv, header)

	r0 := ret.Get(0).(int)
	r1 := ret.Error(1)

	return r0, r1
}
func (m *MockClient) Close() {
	m.Called()
}
