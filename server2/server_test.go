package server

import (
	"github.com/kr/pretty"
	"github.com/plimble/kuja/encoder/json"
	"golang.org/x/net/context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type AddReq struct {
	A int
	B int
}

type AddResp struct {
	C int
}

type ServiceTest struct{}

func (s *ServiceTest) Add(ctx context.Context, req *AddReq, resp *AddResp) error {
	resp.C = req.A + req.B

	return nil
}

func TestServer(t *testing.T) {
	s := NewServer()
	s.Register(&ServiceTest{})
	s.Encoder(json.NewJsonEncoder())

	server := httptest.NewServer(s)
	defer server.Close()

	reqBody := `{"method":"ServiceTest.Add","params":{"A":1,"B":2}}`
	req, _ := http.NewRequest("POST", server.URL, strings.NewReader(reqBody))

	resp, _ := http.DefaultClient.Do(req)
	resp.Body.Close()

	pretty.Println(resp.StatusCode)
}

func BenchmarkServer(b *testing.B) {
	s := NewServer()
	s.Register(&ServiceTest{})
	s.Encoder(json.NewJsonEncoder())

	server := httptest.NewServer(s)
	defer server.Close()

	reqBody := `{"method":"ServiceTest.Add","params":{"A":1,"B":2}}`
	body := strings.NewReader(reqBody)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("POST", server.URL, body)

		resp, _ := http.DefaultClient.Do(req)
		if resp != nil {
			resp.Body.Close()
		}
	}

}
