package core_max_middleware

import (
	"context"

	"github.com/max-messenger/max-bot-api-client-go/schemes"
)

type HandlerFunc func(ctx context.Context, update schemes.UpdateInterface)

type Middleware func(HandlerFunc) HandlerFunc

func Chain(h HandlerFunc, mw ...Middleware) HandlerFunc {
	for i := len(mw) - 1; i >= 0; i-- {
		h = mw[i](h)
	}

	return h
}
