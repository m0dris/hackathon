package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	core_grpc_client "github.com/m0dris/hackathon/internal/core/grpc/client"
	core_logger "github.com/m0dris/hackathon/internal/core/logger"
	core_max_bot "github.com/m0dris/hackathon/internal/core/transport/max/bot"
	core_max_middleware "github.com/m0dris/hackathon/internal/core/transport/max/middleware"
	core_max_response "github.com/m0dris/hackathon/internal/core/transport/max/response"
	core_max_server "github.com/m0dris/hackathon/internal/core/transport/max/server"
	core_max_session "github.com/m0dris/hackathon/internal/core/transport/max/session"

	core_http_middleware "github.com/m0dris/hackathon/internal/core/transport/http/middleware"
	core_http_server "github.com/m0dris/hackathon/internal/core/transport/http/server"

	pb "github.com/m0dris/hackathon/internal/grpc/pb"

	flats_service "github.com/m0dris/hackathon/internal/features/flats/service"
	flats_max_transport "github.com/m0dris/hackathon/internal/features/flats/transport/max"
	houses_service "github.com/m0dris/hackathon/internal/features/houses/service"
	houses_max_transport "github.com/m0dris/hackathon/internal/features/houses/transport/max"
	requests_service "github.com/m0dris/hackathon/internal/features/requests/service"
	requests_max_transport "github.com/m0dris/hackathon/internal/features/requests/transport/max"
	users_service "github.com/m0dris/hackathon/internal/features/users/service"
	users_max_transport "github.com/m0dris/hackathon/internal/features/users/transport/max"

	"github.com/max-messenger/max-bot-api-client-go/schemes"
	"go.uber.org/zap"
)

func main() {
	ctx, cancel := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT, syscall.SIGTERM,
	)
	defer cancel()

	logger, err := core_logger.NewLogger(core_logger.NewConfigMust())
	if err != nil {
		fmt.Println("falied to init aplication logger:", err)
		os.Exit(1)
	}
	defer logger.Close()

	grpcConn, err := core_grpc_client.NewConn(core_grpc_client.NewConfigMust())
	if err != nil {
		logger.Error("init grpc connection to db microservice failed", zap.Error(err))
		os.Exit(1)
	}
	defer grpcConn.Close()

	db := pb.NewUkServiceClient(grpcConn)

	usersSvc := users_service.New(db)
	housesSvc := houses_service.New(db)
	flatsSvc := flats_service.New(db)
	requestsSvc := requests_service.New(db)

	maxBotConfig := core_max_bot.NewConfigMust()

	maxClient, err := core_max_bot.NewClient(maxBotConfig)
	if err != nil {
		logger.Error("init MAX bot client failed", zap.Error(err))
		os.Exit(1)
	}

	sessions := core_max_session.NewMemoryStore()
	responder := core_max_response.NewResponder(maxClient)

	usersTransport := users_max_transport.New(usersSvc, responder, sessions)
	housesTransport := houses_max_transport.New(housesSvc, responder, sessions)
	flatsTransport := flats_max_transport.New(flatsSvc, housesSvc, responder, sessions)
	requestsTransport := requests_max_transport.New(requestsSvc, flatsSvc, responder, sessions)

	router := core_max_server.NewRouter(
		sessions,
		core_max_middleware.Recover(),
		core_max_middleware.Logger(logger),
		core_max_middleware.ResolveUser(usersSvc),
	)
	router.RegisterBotStarted(usersTransport.BotStarted)
	router.RegisterDefaultText(usersTransport.DefaultText)
	router.Register(usersTransport.Routes()...)
	router.Register(housesTransport.Routes()...)
	router.Register(flatsTransport.Routes()...)
	router.Register(requestsTransport.Routes()...)

	if maxBotConfig.AutoSubscribe {
		if err := maxClient.Subscribe(ctx); err != nil {
			logger.Warn("subscribe MAX webhook failed", zap.Error(err))
		}
	}

	logger.Debug("initializing HTTP server")

	httpServer := core_http_server.NewHTTPServer(
		core_http_server.NewConfigMust(),
		logger,
		core_http_middleware.RequestId(),
		core_http_middleware.Logger(logger),
		core_http_middleware.Trace(),
		core_http_middleware.Panic(),
	)

	apiVersionRouter := core_http_server.NewAPIVersionRouter(core_http_server.ApiVersion1)
	apiVersionRouter.RegisterRoutes(core_http_server.Route{
		Method: "POST",
		Path:   "/webhook",
		Handler: maxClient.WebhookHandler(func(update schemes.UpdateInterface) {
			router.Dispatch(context.Background(), update)
		}),
	})
	httpServer.RegisterAPIRoutes(apiVersionRouter)

	if err := httpServer.Run(ctx); err != nil {
		logger.Error("HTTP server run error", zap.Error(err))
	}
}
