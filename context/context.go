package context

import (
	"golang.org/x/net/context"
)

type key int

const (
	mdKey     = key(0)
	respMdKey = key(1)
)

type Metadata map[string]string

func GetReqMetadata(ctx context.Context) Metadata {
	md := ctx.Value(0).(Metadata)
	return md
}

func WithReqMetadata(ctx context.Context, md Metadata) context.Context {
	return context.WithValue(ctx, 0, md)
}

func WithRespMetaData(ctx context.Context, md Metadata) context.Context {
	return context.WithValue(ctx, 1, md)
}

func GetRespMetadata(ctx context.Context) Metadata {
	md := ctx.Value(1).(Metadata)
	return md
}

func Value(ctx context.Context, key interface{}) interface{} {
	return ctx.Value(key)
}
