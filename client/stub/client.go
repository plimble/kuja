package stub

import (
	"errors"
	"github.com/plimble/kuja/contract"
	"github.com/plimble/kuja/encoder"
	"reflect"

	"github.com/plimble/kuja/client"
)

type StubClient struct {
	contracts contract.Contracts
	encoder   encoder.Encoder
}

func NewStubClient(contracts contract.Contracts, encoder encoder.Encoder) *StubClient {
	return &StubClient{contracts, encoder}
}

func (s *StubClient) Publish(topic string, v interface{}, meta map[string]string) error {
	for _, contract := range s.contracts {
		for _, inter := range contract.Interactions {
			if inter.Publish != nil {
				if inter.Publish.Topic == topic {
					if reflect.TypeOf(v).Name() != reflect.TypeOf(inter.Publish.Message).Name() {
						return errors.New("invalid message type")
					}

					return nil
				}
			}
		}
	}

	return errors.New("no topic in contact")
}

func (s *StubClient) AsyncRequests(as []client.AsyncRequest) []client.AsyncResponse {
	done := make(chan client.AsyncResponse, len(as))
	aresps := make([]client.AsyncResponse, 0, len(as))

	for i := 0; i < len(as); i++ {
		go func(index int) {
			status, err := s.Request(as[index].Service, as[index].Method, as[index].Request, as[index].Response, as[index].Metadata)
			done <- client.AsyncResponse{status, err}
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

func (s *StubClient) Request(service string, method string, reqv interface{}, respv interface{}, header map[string]string) (int, error) {
	for _, contract := range s.contracts {
		if contract.Provider != service {
			continue
		}

		for _, inter := range contract.Interactions {
			if inter.Request != nil {
				if inter.Request.Method != method {
					return 500, errors.New("no method in contact")
				}

				if reflect.TypeOf(reqv).Name() != reflect.TypeOf(inter.Request.Body).Name() {
					return 500, errors.New("invalid request type")
				}

				if reflect.TypeOf(respv).Name() != reflect.TypeOf(inter.Request.ResponseObject).Name() {
					return 500, errors.New("invalid response type")
				}

				if inter.Response == nil {
					return 500, errors.New("no response")
				}

				if inter.Response.Error != nil {
					return 500, inter.Response.Error
				}

				data, err := s.encoder.Marshal(inter.Response.Body)
				if err != nil {
					return 500, err
				}

				if err := s.encoder.Unmarshal(data, respv); err != nil {
					return 500, err
				}

				return 200, nil
			}
		}
	}

	return 500, errors.New("no service in contact")
}

func (s *StubClient) Close() {}
