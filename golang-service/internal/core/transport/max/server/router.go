package core_max_server

import (
	"context"

	core_logger "github.com/m0dris/hackathon/internal/core/logger"
	core_max_middleware "github.com/m0dris/hackathon/internal/core/transport/max/middleware"
	core_max_request "github.com/m0dris/hackathon/internal/core/transport/max/request"
	core_max_session "github.com/m0dris/hackathon/internal/core/transport/max/session"
	"github.com/max-messenger/max-bot-api-client-go/schemes"
)

type Router struct {
	routes      map[string]core_max_middleware.HandlerFunc
	botStarted  core_max_middleware.HandlerFunc
	defaultText core_max_middleware.HandlerFunc
	middlewares []core_max_middleware.Middleware
	sessions    core_max_session.Store
}

func NewRouter(sessions core_max_session.Store, middlewares ...core_max_middleware.Middleware) *Router {
	return &Router{
		routes:      make(map[string]core_max_middleware.HandlerFunc),
		middlewares: middlewares,
		sessions:    sessions,
	}
}

func (r *Router) Register(routes ...Route) {
	for _, route := range routes {
		if _, exists := r.routes[route.Key]; exists {
			panic("MAX route key already registered: " + route.Key)
		}

		r.routes[route.Key] = route.Handler
	}
}

func (r *Router) RegisterBotStarted(h core_max_middleware.HandlerFunc) {
	r.botStarted = h
}

func (r *Router) RegisterDefaultText(h core_max_middleware.HandlerFunc) {
	r.defaultText = h
}

func (r *Router) Dispatch(ctx context.Context, update schemes.UpdateInterface) {
	core_max_middleware.Chain(r.resolve(update), r.middlewares...)(ctx, update)
}

func (r *Router) resolve(update schemes.UpdateInterface) core_max_middleware.HandlerFunc {
	switch u := update.(type) {
	case *schemes.BotStartedUpdate:
		if r.botStarted != nil {
			return r.botStarted
		}

	case *schemes.MessageCallbackUpdate:
		key, _ := core_max_request.SplitPayload(u.Callback.Payload)
		if h, ok := r.routes[key]; ok {
			return h
		}

	case *schemes.MessageCreatedUpdate:
		if cmd := u.GetCommand(); cmd != "" {
			if h, ok := r.routes[cmd]; ok {
				return h
			}
		} else if sess, ok := r.sessions.Get(u.GetChatID()); ok && sess.Step != "" {
			if h, ok := r.routes[sess.Step]; ok {
				return h
			}
		}

		if r.defaultText != nil {
			return r.defaultText
		}
	}

	return unhandled
}

func unhandled(ctx context.Context, update schemes.UpdateInterface) {
	core_logger.FromContext(ctx).Debug("no handler for MAX update")
}
