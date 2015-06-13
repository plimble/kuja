package json

import (
	"encoding/json"
	"io"
)

type JsonEncoder struct{}

func NewEncoder() *JsonEncoder {
	return &JsonEncoder{}
}

func (e *JsonEncoder) Encode(w io.Writer, v interface{}) error {
	return json.NewEncoder(w).Encode(v)
}

func (e *JsonEncoder) Decode(r io.Reader, v interface{}) error {
	return json.NewDecoder(r).Decode(v)
}

func (e *JsonEncoder) Marshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

func (e *JsonEncoder) Unmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}
