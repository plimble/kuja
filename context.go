package kuja

import (
	"github.com/plimble/kuja/encoder"
	"net/http"
	"reflect"
)

type Context struct {
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
	serviceError ServiceErrorFunc
	ServiceID    string
	ServiceName  string
	MethodName   string
	isResp       bool
}

func (ctx *Context) Next() error {
	if ctx.index+1 == len(ctx.handlers) {
		if err := serve(ctx); err != nil && !ctx.isResp {
			respError(err, ctx)
		}
	} else if ctx.index+1 < len(ctx.handlers) {
		ctx.index++
		if err := ctx.handlers[ctx.index](ctx, ctx.w, ctx.req); err != nil && !ctx.isResp {
			respError(err, ctx)
		}
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
