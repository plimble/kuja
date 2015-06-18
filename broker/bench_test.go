package broker

import (
	"testing"
)

func BenchmarkGogoProtoEncode(b *testing.B) {
	msg := &Message{
		Header: map[string]string{
			"h1": "v1",
			"h2": "v2",
		},
		Body: []byte("MyMonster"),
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		msg.Marshal()
	}
}

func BenchmarkGogoProtoDecode(b *testing.B) {
	msg := &Message{
		Header: map[string]string{
			"h1": "v1",
			"h2": "v2",
		},
		Body: []byte("MyMonster"),
	}

	data, _ := msg.Marshal()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		result := &Message{}
		result.Unmarshal(data)
	}
}
