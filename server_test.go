package kuja

import (
	"github.com/kr/pretty"
	"github.com/plimble/kuja/encoder/gogoproto"
	"github.com/plimble/kuja/encoder/json"
	"github.com/plimble/kuja/encoder/proto"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type ServiceTest struct{}

func (s *ServiceTest) Add(ctx *Ctx, req *AddReq, resp *AddResp) error {
	resp.C = req.A + req.B + 10

	return nil
}

func TestGetServiceMethod(t *testing.T) {
	service, method := getServerMethod("/ServiceTest/Add/")
	pretty.Println(service, method)
}

func BenchmarkGetServiceMethod(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		getServerMethod("/ServiceTest/Add/")
	}
}

func TestServer(t *testing.T) {
	encoder := proto.NewEncoder()
	s := NewServer()
	s.Use(func(ctx *Ctx, w http.ResponseWriter, r *http.Request) error {
		ctx.Next()

		return nil
	})
	s.Use(func(ctx *Ctx, w http.ResponseWriter, r *http.Request) error {
		ctx.Next()

		return nil
	})

	s.Service(&ServiceTest{})
	s.Encoder(encoder)

	reqData := &AddReq{1, 2}

	by, _ := encoder.Marshal(reqData)
	req, _ := http.NewRequest("POST", "http://plimble.com/ServiceTest/Add?id=1", strings.NewReader(string(by)))
	w := httptest.NewRecorder()

	s.ServeHTTP(w, req)

	resp := &AddResp{}

	encoder.Unmarshal(w.Body.Bytes(), resp)

	pretty.Println(w.Body.String())
	pretty.Println(resp)
}

func BenchmarkServerProto(b *testing.B) {
	encoder := proto.NewEncoder()
	s := NewServer()
	s.Use(func(ctx *Ctx, w http.ResponseWriter, r *http.Request) error {
		ctx.Next()

		return nil
	})
	s.Use(func(ctx *Ctx, w http.ResponseWriter, r *http.Request) error {
		ctx.Next()

		return nil
	})
	s.Service(&ServiceTest{})
	s.Encoder(encoder)

	reqData := &AddReq{1, 2}

	by, _ := encoder.Marshal(reqData)

	req, _ := http.NewRequest("POST", "http://plimble.com/ServiceTest/Add?id=1", strings.NewReader(string(by)))
	w := httptest.NewRecorder()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.ServeHTTP(w, req)
	}

}

func BenchmarkServerJson(b *testing.B) {
	encoder := json.NewEncoder()
	s := NewServer()
	s.Use(func(ctx *Ctx, w http.ResponseWriter, r *http.Request) error {
		ctx.Next()

		return nil
	})
	s.Use(func(ctx *Ctx, w http.ResponseWriter, r *http.Request) error {
		ctx.Next()

		return nil
	})
	s.Service(&ServiceTest{})
	s.Encoder(encoder)

	reqData := &AddReq{1, 2}

	by, _ := encoder.Marshal(reqData)

	req, _ := http.NewRequest("POST", "http://plimble.com/ServiceTest/Add?id=1", strings.NewReader(string(by)))
	w := httptest.NewRecorder()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.ServeHTTP(w, req)
	}

}

func BenchmarkServerGoGoProto(b *testing.B) {
	encoder := gogoproto.NewEncoder()
	s := NewServer()
	s.Use(func(ctx *Ctx, w http.ResponseWriter, r *http.Request) error {
		ctx.Next()

		return nil
	})
	s.Use(func(ctx *Ctx, w http.ResponseWriter, r *http.Request) error {
		ctx.Next()

		return nil
	})
	s.Service(&ServiceTest{})
	s.Encoder(encoder)

	reqData := &AddReq{1, 2}

	by, _ := encoder.Marshal(reqData)

	req, _ := http.NewRequest("POST", "http://plimble.com/ServiceTest/Add?id=1", strings.NewReader(string(by)))
	w := httptest.NewRecorder()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.ServeHTTP(w, req)
	}

}
