package contract

import (
	"errors"
	"github.com/plimble/kuja/encoder/gogoproto"
	"github.com/plimble/kuja/encoder/json"
	"github.com/plimble/kuja/encoder/proto"
	"gopkg.in/bluesuncorp/validator.v5"
)

const (
	JSON = iota
	PROTOBUF
	GOGOPROTOBUF
)

var (
	jsonEnc  = json.NewEncoder()
	protoEnc = proto.NewEncoder()
	gogoEnc  = gogoproto.NewEncoder()
)

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
	Subscribes    []*Subscribe
}

type Request struct {
	Metadata map[string]string
	Method   string
	Body     interface{}
}

func (r *Request) Validate() error {
	v := validator.New("validate", validator.BakedInValidators)
	errs := v.Struct(r.Body)

	return errors.New(errs.Error())
}

func (r *Request) Marshal(codec int) ([]byte, error) {
	switch codec {
	case JSON:
		return jsonEnc.Marshal(r.Body)
	case PROTOBUF:
		return protoEnc.Marshal(r.Body)
	case GOGOPROTOBUF:
		return gogoEnc.Marshal(r.Body)
	}

	return nil, errors.New("no codec")
}

type Response struct {
	Metadata map[string]string
	Body     interface{}
	Error    error
}

// func (r *Request) Unmarshal(codec int) error {
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

type Publish struct {
	Metadata map[string]string
	Topic    string
	Message  interface{}
}

type Subscribe struct {
	Queue string
	Error error
}

type ContractManager struct {
	contracts []*Contract
}

func New() *ContractManager {
	return &ContractManager{}
}

type Option func(c *Contract)

func (c *ContractManager) Add(contract *Contract) {
	c.contracts = append(c.contracts, contract)
}

func (c *ContractManager) Get() []*Contract {
	return c.contracts
}
