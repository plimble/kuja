package gogoproto

import (
	"errors"
	"github.com/gogo/protobuf/proto"
	"github.com/plimble/utils/pool"
	"io"
)

type GoGoProtoEncoder struct {
	pool *pool.BufferPool
}

func NewEncoder() *GoGoProtoEncoder {
	return &GoGoProtoEncoder{pool.NewBufferPool(512)}
}

func (e *GoGoProtoEncoder) Encode(w io.Writer, v interface{}) error {
	pb, ok := v.(proto.Marshaler)
	if !ok {
		errors.New("does not proto message interface")
	}

	data, err := pb.Marshal()
	if err != nil {
		return err
	}

	_, err = w.Write(data)

	return err
}

func (e *GoGoProtoEncoder) Decode(r io.Reader, v interface{}) error {
	buf := e.pool.Get()
	buf.ReadFrom(r)

	pb, ok := v.(proto.Unmarshaler)
	if !ok {
		errors.New("does not proto message interface")
	}

	err := pb.Unmarshal(buf.Bytes())
	e.pool.Put(buf)

	return err
}

func (e *GoGoProtoEncoder) Marshal(v interface{}) ([]byte, error) {
	pb, ok := v.(proto.Marshaler)
	if !ok {
		errors.New("does not proto message interface")
	}

	return pb.Marshal()
}

func (e *GoGoProtoEncoder) Unmarshal(data []byte, v interface{}) error {
	pb, ok := v.(proto.Unmarshaler)
	if !ok {
		errors.New("does not proto message interface")
	}

	return pb.Unmarshal(data)
}
