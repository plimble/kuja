package contract

import (
	"github.com/plimble/kuja/client"
	"github.com/stretchr/testify/assert"
	"testing"

	// "errors"
	// "github.com/plimble/kuja/encoder/gogoproto"
	// "github.com/plimble/kuja/encoder/json"
	// "github.com/plimble/kuja/encoder/proto"
)

// const (
// 	JSON = iota
// 	PROTOBUF
// 	GOGOPROTOBUF
// )

// var (
// 	jsonEnc  = json.NewEncoder()
// 	protoEnc = proto.NewEncoder()
// 	gogoEnc  = gogoproto.NewEncoder()
// )

type Contracts []*Contract

type Contract struct {
	Codec        int
	Snappy       bool
	Provider     string
	Consumer     string
	Interactions []*Interaction
	Metadata     map[string]string
}

type Interaction struct {
	Description   string
	ProviderState string
	Request       *Request
	Response      *Response
	Publish       *Publish
}

func (c *Contract) ContractTest(t *testing.T, kujaClient client.Client) {
	for _, inter := range c.Interactions {
		switch {
		case inter.Publish != nil:
			err := kujaClient.Publish(inter.Publish.Topic, inter.Publish.Message, inter.Publish.Metadata)
			assert.NoError(t, err)
		case inter.Request != nil:
			status, err := kujaClient.Request(
				c.Provider,
				inter.Request.Method,
				inter.Request.Body,
				inter.Request.ResponseObject,
				inter.Request.Metadata,
			)

			assert.NoError(t, err)
			assert.Equal(t, inter.Response.Status, status)
		}
	}
}

type Request struct {
	Metadata       map[string]string
	Method         string
	Body           interface{}
	ResponseObject interface{}
}

type Response struct {
	Metadata map[string]string
	Status   int
	Body     interface{}
	Error    error
}

type Publish struct {
	Metadata map[string]string
	Topic    string
	Message  interface{}
}

// func Marshal(codec int) ([]byte, error) {
// 	switch codec {
// 	case JSON:
// 		return jsonEnc.Marshal(r.Body)
// 	case PROTOBUF:
// 		return protoEnc.Marshal(r.Body)
// 	case GOGOPROTOBUF:
// 		return gogoEnc.Marshal(r.Body)
// 	}

// 	return nil, errors.New("no codec")
// }

// func Unmarshal(codec int) error {
// 	switch codec {
// 	case JSON:
// 		return jsonEnc.Unmarshal(data, v)
// 	case PROTOBUF:
// 		return protoEnc.Unmarshal(r.Body)
// 	case GOGOPROTOBUF:
// 		return gogoEnc.Unmarshal(r.Body)
// 	}

// 	return nil, errors.New("no codec")
// }
