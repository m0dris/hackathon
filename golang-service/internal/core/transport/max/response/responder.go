package core_max_response

import (
	"context"
	"errors"

	core_errors "github.com/m0dris/hackathon/internal/core/errors"
	core_logger "github.com/m0dris/hackathon/internal/core/logger"
	core_max_bot "github.com/m0dris/hackathon/internal/core/transport/max/bot"
	maxbot "github.com/max-messenger/max-bot-api-client-go"
	"go.uber.org/zap"
)

type Responder struct {
	bot *core_max_bot.Client
}

func NewResponder(bot *core_max_bot.Client) *Responder {
	return &Responder{bot: bot}
}

func (r *Responder) Text(ctx context.Context, chatID int64, text string) {
	if err := r.bot.SendText(ctx, chatID, text); err != nil {
		core_logger.FromContext(ctx).Error("send text failed", zap.Error(err))
	}
}

func (r *Responder) Keyboard(ctx context.Context, chatID int64, text string, kb *maxbot.Keyboard) {
	if err := r.bot.SendKeyboard(ctx, chatID, text, kb); err != nil {
		core_logger.FromContext(ctx).Error("send keyboard failed", zap.Error(err))
	}
}

func (r *Responder) AnswerCallback(ctx context.Context, callbackID string, notification string) {
	if err := r.bot.AnswerCallback(ctx, callbackID, notification); err != nil {
		core_logger.FromContext(ctx).Error("answer callback failed", zap.Error(err))
	}
}

func (r *Responder) Error(ctx context.Context, chatID int64, err error) {
	core_logger.FromContext(ctx).Warn("handler returned error", zap.Error(err))

	switch {
	case errors.Is(err, core_errors.ErrNotFound):
		r.Text(ctx, chatID, "🔎 Не нашёл это в базе. Попробуйте ещё раз или вернитесь в меню командой /menu.")
	case errors.Is(err, core_errors.ErrForbidden):
		r.Text(ctx, chatID, "⛔ Это действие недоступно для вашей роли или этого объекта.")
	case errors.Is(err, core_errors.ErrConflict):
		r.Text(ctx, chatID, "⚠️ Не получилось: такая запись уже существует или в этом состоянии действие недоступно.")
	case errors.Is(err, core_errors.ErrInvalidArgument):
		r.Text(ctx, chatID, "❗ Проверьте данные и попробуйте ещё раз.")
	default:
		r.Text(ctx, chatID, "🚧 Что-то пошло не так на нашей стороне. Попробуйте позже.")
	}
}
