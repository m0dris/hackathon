package core_max_middleware

import (
	"context"

	core_domain "github.com/m0dris/hackathon/internal/core/domain"
	core_logger "github.com/m0dris/hackathon/internal/core/logger"
	"github.com/max-messenger/max-bot-api-client-go/schemes"
	"go.uber.org/zap"
)

type ctxKey struct{ name string }

var userKey = ctxKey{"max_user"}

type UserResolver interface {
	Me(ctx context.Context, maxUserID int64) (core_domain.User, bool, error)
}

func UserFromContext(ctx context.Context) (core_domain.User, bool) {
	u, ok := ctx.Value(userKey).(core_domain.User)

	return u, ok
}

func CheckRole(ctx context.Context, want core_domain.Role) (core_domain.User, bool) {
	user, found := UserFromContext(ctx)

	return user, found && user.Role == want
}

func ResolveUser(resolver UserResolver) Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(ctx context.Context, update schemes.UpdateInterface) {
			user, found, err := resolver.Me(ctx, update.GetUserID())
			if err != nil {
				core_logger.FromContext(ctx).Warn("resolve MAX user failed", zap.Error(err))
				next(ctx, update)

				return
			}

			if found {
				ctx = context.WithValue(ctx, userKey, user)
			}

			next(ctx, update)
		}
	}
}
