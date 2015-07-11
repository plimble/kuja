package contract

import (
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

type Request struct {
	Metadata       map[string]string
	Method         string
	Body           interface{}
	ResponseObject interface{}
}

type Response struct {
	Metadata map[string]string
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
