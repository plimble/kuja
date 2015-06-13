package proto

import (
	"errors"
	"github.com/golang/protobuf/proto"
	"github.com/plimble/utils/pool"
	"io"
)

type ProtoEncoder struct {
	pool *pool.BufferPool
}

func NewEncoder() *ProtoEncoder {
	return &ProtoEncoder{pool.NewBufferPool(512)}
}

func (e *ProtoEncoder) Encode(w io.Writer, v interface{}) error {
	pb, ok := v.(proto.Message)
	if !ok {
		errors.New("does not proto message interface")
	}

	data, err := proto.Marshal(pb)
	if err != nil {
		return err
	}

	_, err = w.Write(data)

	return err
}

func (e *ProtoEncoder) Decode(r io.Reader, v interface{}) error {
	buf := e.pool.Get()
	buf.ReadFrom(r)

	pb, ok := v.(proto.Message)
	if !ok {
		errors.New("does not proto message interface")
	}

	err := proto.Unmarshal(buf.Bytes(), pb)
	e.pool.Put(buf)

	return err
}

func (e *ProtoEncoder) Marshal(v interface{}) ([]byte, error) {
	pb, ok := v.(proto.Message)
	if !ok {
		errors.New("does not proto message interface")
	}

	return proto.Marshal(pb)
}

func (e *ProtoEncoder) Unmarshal(data []byte, v interface{}) error {
	pb, ok := v.(proto.Message)
	if !ok {
		errors.New("does not proto message interface")
	}

	return proto.Unmarshal(data, pb)
}
