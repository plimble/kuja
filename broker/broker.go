package broker

import (
	"golang.org/x/net/context"
)

type Endpoint func(ctx context.Context, v interface{}) (v interface{}, error)
