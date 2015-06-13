package server

import (
	"github.com/kr/pretty"
	"github.com/youtube/vitess/go/rpcplus"
	"github.com/youtube/vitess/go/rpcplus/pbrpc"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestPlus(t *testing.T) {
	s := rpcplus.NewServer()
	rpcplus.PB = func(conn io.ReadWriteCloser) rpcplus.ServerCodec {
		return pbrpc.NewServerCodec(conn)
	}
	s.Register(&ServiceTest{})

	server := httptest.NewServer(s)
	defer server.Close()

	reqBody := `{"method":"ServiceTest.Add","params":{"A":1,"B":2}}`
	req, _ := http.NewRequest("CONNECT", server.URL+"/_goRPC_", strings.NewReader(reqBody))

	resp, _ := http.DefaultClient.Do(req)
	resp.Body.Close()

	pretty.Println(resp.StatusCode)
}

func BenchmarkPlus(b *testing.B) {
	s := rpcplus.NewServer()
	rpcplus.PB = func(conn io.ReadWriteCloser) rpcplus.ServerCodec {
		return pbrpc.NewServerCodec(conn)
	}
	s.Register(&ServiceTest{})

	server := httptest.NewServer(s)
	defer server.Close()

	reqBody := `{"method":"ServiceTest.Add","params":{"A":1,"B":2}}`
	// req, _ := http.NewRequest("CONNECT", server.URL+"/_goRPC_", strings.NewReader(reqBody))

	url := server.URL + "/_goRPC_"
	body := strings.NewReader(reqBody)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("CONNECT", url, body)
		resp, _ := http.DefaultClient.Do(req)
		if resp != nil {
			resp.Body.Close()
		}
	}
}
