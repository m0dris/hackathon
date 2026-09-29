package core_max_server

import core_max_middleware "github.com/m0dris/hackathon/internal/core/transport/max/middleware"

type Route struct {
	Key     string
	Handler core_max_middleware.HandlerFunc
}
