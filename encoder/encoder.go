package encoder

import (
	"io"
)

type Encoder interface {
	ContentType() string
	Encode(w io.Writer, v interface{}) error
	Decode(r io.Reader, v interface{}) error
	Marshal(v interface{}) ([]byte, error)
	Unmarshal(data []byte, v interface{}) error
}
