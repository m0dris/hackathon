package users_max_transport

import (
	"context"
	"strings"

	core_domain "github.com/m0dris/hackathon/internal/core/domain"
	core_max_menu "github.com/m0dris/hackathon/internal/core/transport/max/menu"
	core_max_middleware "github.com/m0dris/hackathon/internal/core/transport/max/middleware"
	core_max_payload "github.com/m0dris/hackathon/internal/core/transport/max/payload"
	core_max_request "github.com/m0dris/hackathon/internal/core/transport/max/request"
	core_max_response "github.com/m0dris/hackathon/internal/core/transport/max/response"
	core_max_server "github.com/m0dris/hackathon/internal/core/transport/max/server"
	core_max_session "github.com/m0dris/hackathon/internal/core/transport/max/session"
	users_service "github.com/m0dris/hackathon/internal/features/users/service"
	maxbot "github.com/max-messenger/max-bot-api-client-go"
	"github.com/max-messenger/max-bot-api-client-go/schemes"
)

const (
	stepAwaitName  = "reg:await_name"
	stepAwaitPhone = "reg:await_phone"

	draftRole      = "role"
	draftFirstName = "first_name"
	draftLastName  = "last_name"
)

type Transport struct {
	service   *users_service.Service
	responder *core_max_response.Responder
	sessions  core_max_session.Store
}

func New(service *users_service.Service, responder *core_max_response.Responder, sessions core_max_session.Store) *Transport {
	return &Transport{service: service, responder: responder, sessions: sessions}
}

func (t *Transport) Routes() []core_max_server.Route {
	return []core_max_server.Route{
		{Key: "menu", Handler: t.handleMenuCommand},
		{Key: core_max_payload.RoleUser, Handler: t.handleRolePicked(core_domain.RoleUser)},
		{Key: core_max_payload.RoleCompany, Handler: t.handleRolePicked(core_domain.RoleCompany)},
		{Key: stepAwaitName, Handler: t.handleAwaitName},
		{Key: stepAwaitPhone, Handler: t.handleAwaitPhone},
		{Key: core_max_payload.MenuProfile, Handler: t.handleProfile},
		{Key: core_max_payload.MenuMain, Handler: t.handleShowMenuCallback},
	}
}

func (t *Transport) BotStarted(ctx context.Context, update schemes.UpdateInterface) {
	u, ok := core_max_request.AsBotStarted(update)
	if !ok {
		return
	}

	t.startOrShowMenu(ctx, u.ChatId)
}

func (t *Transport) DefaultText(ctx context.Context, update schemes.UpdateInterface) {
	u, ok := core_max_request.AsMessageCreated(update)
	if !ok {
		return
	}

	if user, found := core_max_middleware.UserFromContext(ctx); found {
		t.responder.Keyboard(ctx, u.GetChatID(), "Не понял сообщение. "+core_max_menu.Text, core_max_menu.Keyboard(user.Role))
		return
	}

	t.startOrShowMenu(ctx, u.GetChatID())
}

func (t *Transport) startOrShowMenu(ctx context.Context, chatID int64) {
	if user, found := core_max_middleware.UserFromContext(ctx); found {
		t.responder.Keyboard(ctx, chatID, "С возвращением! "+core_max_menu.Text, core_max_menu.Keyboard(user.Role))
		return
	}

	t.sessions.Clear(chatID)
	t.responder.Keyboard(
		ctx,
		chatID,
		"👋 Привет! Это бот вашей управляющей компании.\n\nВы житель дома или представитель УК?",
		roleKeyboard(),
	)
}

func (t *Transport) handleMenuCommand(ctx context.Context, update schemes.UpdateInterface) {
	u, ok := core_max_request.AsMessageCreated(update)
	if !ok {
		return
	}

	t.startOrShowMenu(ctx, u.GetChatID())
}

func (t *Transport) handleShowMenuCallback(ctx context.Context, update schemes.UpdateInterface) {
	u, ok := core_max_request.AsMessageCallback(update)
	if !ok {
		return
	}

	t.responder.AnswerCallback(ctx, u.Callback.CallbackID, "")
	t.startOrShowMenu(ctx, u.GetChatID())
}

func (t *Transport) handleRolePicked(role core_domain.Role) core_max_middleware.HandlerFunc {
	return func(ctx context.Context, update schemes.UpdateInterface) {
		u, ok := core_max_request.AsMessageCallback(update)
		if !ok {
			return
		}

		t.responder.AnswerCallback(ctx, u.Callback.CallbackID, "")

		if _, found := core_max_middleware.UserFromContext(ctx); found {
			t.responder.Text(ctx, u.GetChatID(), "Вы уже зарегистрированы. Наберите /menu.")
			return
		}

		t.sessions.Set(u.GetChatID(), core_max_session.Session{
			Step:  stepAwaitName,
			Draft: map[string]string{draftRole: string(role)},
		})

		t.responder.Text(ctx, u.GetChatID(), "Введите имя и фамилию одним сообщением, например:\nИван Иванов")
	}
}

func (t *Transport) handleAwaitName(ctx context.Context, update schemes.UpdateInterface) {
	u, ok := core_max_request.AsMessageCreated(update)
	if !ok {
		return
	}

	chatID := u.GetChatID()
	sess, _ := t.sessions.Get(chatID)

	first, last, ok := splitFullName(u.GetText())
	if !ok {
		t.responder.Text(ctx, chatID, "❗ Нужно и имя, и фамилия одним сообщением, например: Иван Иванов")
		return
	}

	sess.Draft[draftFirstName] = first
	sess.Draft[draftLastName] = last
	sess.Step = stepAwaitPhone
	t.sessions.Set(chatID, sess)

	t.responder.Text(ctx, chatID, "Теперь укажите номер телефона, например: +79991234567")
}

func (t *Transport) handleAwaitPhone(ctx context.Context, update schemes.UpdateInterface) {
	u, ok := core_max_request.AsMessageCreated(update)
	if !ok {
		return
	}

	chatID := u.GetChatID()
	sess, _ := t.sessions.Get(chatID)

	phone := strings.TrimSpace(u.GetText())

	user, err := t.service.Register(ctx, core_domain.User{
		Id:          u.GetUserID(),
		ChatId:      chatID,
		FirstName:   sess.Draft[draftFirstName],
		LastName:    sess.Draft[draftLastName],
		PhoneNumber: phone,
		Role:        core_domain.Role(sess.Draft[draftRole]),
	})
	if err != nil {
		t.responder.Error(ctx, chatID, err)
		return
	}

	t.sessions.Clear(chatID)
	t.responder.Keyboard(
		ctx,
		chatID,
		"✅ Готово, "+user.FirstName+"! Вы зарегистрированы как "+roleLabel(user.Role)+".\n\n"+core_max_menu.Text,
		core_max_menu.Keyboard(user.Role),
	)
}

func (t *Transport) handleProfile(ctx context.Context, update schemes.UpdateInterface) {
	u, ok := core_max_request.AsMessageCallback(update)
	if !ok {
		return
	}

	t.responder.AnswerCallback(ctx, u.Callback.CallbackID, "")

	user, found := core_max_middleware.UserFromContext(ctx)
	if !found {
		t.responder.Text(ctx, u.GetChatID(), "Сначала нужно зарегистрироваться: наберите /start")
		return
	}

	text := "👤 Профиль\n\n" +
		"Имя: " + user.FirstName + " " + user.LastName + "\n" +
		"Телефон: " + user.PhoneNumber + "\n" +
		"Роль: " + roleLabel(user.Role)

	t.responder.Keyboard(ctx, u.GetChatID(), text, core_max_menu.Keyboard(user.Role))
}

func splitFullName(text string) (first string, last string, ok bool) {
	fields := strings.Fields(text)
	if len(fields) < 2 {
		return "", "", false
	}

	return fields[0], strings.Join(fields[1:], " "), true
}

func roleLabel(r core_domain.Role) string {
	if r == core_domain.RoleCompany {
		return "управляющая компания"
	}

	return "житель"
}

func roleKeyboard() *maxbot.Keyboard {
	return maxbot.InlineKeyboard(
		maxbot.Row(maxbot.Btn("🏠 Я житель", core_max_payload.RoleUser)),
		maxbot.Row(maxbot.Btn("🏢 Я управляющая компания", core_max_payload.RoleCompany)),
	)
}
