package core_max_middleware

import (
	"context"
	"fmt"

	core_logger "github.com/m0dris/hackathon/internal/core/logger"
	"github.com/max-messenger/max-bot-api-client-go/schemes"
	"go.uber.org/zap"
)

func Logger(log *core_logger.Logger) Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(ctx context.Context, update schemes.UpdateInterface) {
			l := log.With(
				zap.String("update_type", string(update.GetUpdateType())),
				zap.Int64("chat_id", update.GetChatID()),
				zap.Int64("max_user_id", update.GetUserID()),
			)

			next(core_logger.ToContext(ctx, l), update)
		}
	}
}

func Recover() Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(ctx context.Context, update schemes.UpdateInterface) {
			defer func() {
				if p := recover(); p != nil {
					core_logger.FromContext(ctx).Error(
						"panic while handling MAX update",
						zap.Any("panic", p),
						zap.Error(fmt.Errorf("%v", p)),
					)
				}
			}()

			next(ctx, update)
		}
	}
}
