package kuja

import (
	"github.com/plimble/kuja/encoder"
	"net/http"
	"reflect"
)

type Ctx struct {
	ReqMetadata  Metadata
	RespMetadata Metadata
	index        int
	handlers     []Handler
	mt           *methodType
	req          *http.Request
	w            http.ResponseWriter
	rcvr         reflect.Value
	encoder      encoder.Encoder
	returnValues []reflect.Value
	snappy       bool
}

func (ctx *Ctx) Next() error {
	if ctx.index+1 == len(ctx.handlers) {
		return serve(ctx)
	} else if ctx.index+1 < len(ctx.handlers) {
		ctx.index++
		return ctx.handlers[ctx.index](ctx, ctx.w, ctx.req)
	}

	return nil
}

type key int

const (
	mdKey           = key(0)
	respMdKey       = key(1)
	serveContextKey = key(3)
)

type Metadata map[string]string
