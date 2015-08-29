package msgp

import (
	"github.com/plimble/errors"
	"github.com/tinylib/msgp/msgp"
	"io"
)

type MsgpEncoder struct{}

func NewEncoder() *MsgpEncoder {
	return &MsgpEncoder{}
}

func (e *MsgpEncoder) Encode(w io.Writer, v interface{}) error {
	vMsgp, ok := v.(msgp.Encodable)
	if !ok {
		return errors.InternalError("object is not implement msgp Encodable interface")
	}

	return msgp.Encode(w, vMsgp)
}

func (e *MsgpEncoder) Decode(r io.Reader, v interface{}) error {
	vMsgp, ok := v.(msgp.Decodable)
	if !ok {
		return errors.InternalError("object is not implement msgp Decodable interface")
	}

	return msgp.Decode(r, vMsgp)
}

func (e *MsgpEncoder) Marshal(v interface{}) ([]byte, error) {
	vMsgp, ok := v.(msgp.Marshaler)
	if !ok {
		return nil, errors.InternalError("object is not implement msgp Marshaler interface")
	}

	return vMsgp.MarshalMsg(nil)
}

func (e *MsgpEncoder) Unmarshal(data []byte, v interface{}) error {
	vMsgp, ok := v.(msgp.Unmarshaler)
	if !ok {
		return errors.InternalError("object is not implement msgp Unmarshaler interface")
	}

	_, err := vMsgp.UnmarshalMsg(data)
	return err
}
