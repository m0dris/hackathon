package houses_max_transport

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	core_max_middleware "github.com/m0dris/hackathon/internal/core/transport/max/middleware"
	core_max_payload "github.com/m0dris/hackathon/internal/core/transport/max/payload"
	core_max_request "github.com/m0dris/hackathon/internal/core/transport/max/request"
	core_max_response "github.com/m0dris/hackathon/internal/core/transport/max/response"
	core_max_server "github.com/m0dris/hackathon/internal/core/transport/max/server"
	core_max_session "github.com/m0dris/hackathon/internal/core/transport/max/session"

	core_domain "github.com/m0dris/hackathon/internal/core/domain"
	houses_service "github.com/m0dris/hackathon/internal/features/houses/service"
	maxbot "github.com/max-messenger/max-bot-api-client-go"
	"github.com/max-messenger/max-bot-api-client-go/schemes"
)

const (
	stepAwaitCity   = "house:add:city"
	stepAwaitStreet = "house:add:street"
	stepAwaitNumber = "house:add:number"

	draftCity   = "city"
	draftStreet = "street"
)

type Transport struct {
	service   *houses_service.Service
	responder *core_max_response.Responder
	sessions  core_max_session.Store
}

func New(service *houses_service.Service, responder *core_max_response.Responder, sessions core_max_session.Store) *Transport {
	return &Transport{service: service, responder: responder, sessions: sessions}
}

func (t *Transport) Routes() []core_max_server.Route {
	return []core_max_server.Route{
		{Key: core_max_payload.MenuHouses, Handler: t.handleMyHouses},
		{Key: core_max_payload.HouseAdd, Handler: t.handleAddStart},
		{Key: stepAwaitCity, Handler: t.handleAwaitCity},
		{Key: stepAwaitStreet, Handler: t.handleAwaitStreet},
		{Key: stepAwaitNumber, Handler: t.handleAwaitNumber},
		{Key: core_max_payload.HouseOpen, Handler: t.handleOpen},
	}
}

func (t *Transport) handleMyHouses(ctx context.Context, update schemes.UpdateInterface) {
	u, ok := core_max_request.AsMessageCallback(update)
	if !ok {
		return
	}

	t.responder.AnswerCallback(ctx, u.Callback.CallbackID, "")

	user, allowed := core_max_middleware.CheckRole(ctx, core_domain.RoleCompany)
	if !allowed {
		t.responder.Text(ctx, u.GetChatID(), "⛔ Раздел доступен только представителям УК.")
		return
	}

	houses, err := t.service.My(ctx, user.Id)
	if err != nil {
		t.responder.Error(ctx, u.GetChatID(), err)
		return
	}

	var rows [][]schemes.ButtonInterface
	for _, h := range houses {
		rows = append(rows, maxbot.Row(maxbot.Btn(h.Address(), fmt.Sprintf("%s:%d", core_max_payload.HouseOpen, h.Id))))
	}
	rows = append(rows, maxbot.Row(maxbot.Btn("➕ Зарегистрировать дом", core_max_payload.HouseAdd)))
	rows = append(rows, maxbot.Row(maxbot.Btn("🔙 В меню", core_max_payload.MenuMain)))

	text := "🏢 Ваши дома:"
	if len(houses) == 0 {
		text = "🏢 У вас пока нет зарегистрированных домов."
	}

	t.responder.Keyboard(ctx, u.GetChatID(), text, maxbot.InlineKeyboard(rows...))
}

func (t *Transport) handleAddStart(ctx context.Context, update schemes.UpdateInterface) {
	u, ok := core_max_request.AsMessageCallback(update)
	if !ok {
		return
	}

	t.responder.AnswerCallback(ctx, u.Callback.CallbackID, "")

	if _, allowed := core_max_middleware.CheckRole(ctx, core_domain.RoleCompany); !allowed {
		t.responder.Text(ctx, u.GetChatID(), "⛔ Раздел доступен только представителям УК.")
		return
	}

	t.sessions.Set(u.GetChatID(), core_max_session.Session{Step: stepAwaitCity, Draft: map[string]string{}})
	t.responder.Text(ctx, u.GetChatID(), "Введите город:")
}

func (t *Transport) handleAwaitCity(ctx context.Context, update schemes.UpdateInterface) {
	u, ok := core_max_request.AsMessageCreated(update)
	if !ok {
		return
	}

	chatID := u.GetChatID()
	sess, _ := t.sessions.Get(chatID)

	city := strings.TrimSpace(u.GetText())
	if len([]rune(city)) < 2 {
		t.responder.Text(ctx, chatID, "❗ Слишком коротко. Введите город ещё раз:")
		return
	}

	sess.Draft[draftCity] = city
	sess.Step = stepAwaitStreet
	t.sessions.Set(chatID, sess)

	t.responder.Text(ctx, chatID, "Введите улицу:")
}

func (t *Transport) handleAwaitStreet(ctx context.Context, update schemes.UpdateInterface) {
	u, ok := core_max_request.AsMessageCreated(update)
	if !ok {
		return
	}

	chatID := u.GetChatID()
	sess, _ := t.sessions.Get(chatID)

	street := strings.TrimSpace(u.GetText())
	if len([]rune(street)) < 2 {
		t.responder.Text(ctx, chatID, "❗ Слишком коротко. Введите улицу ещё раз:")
		return
	}

	sess.Draft[draftStreet] = street
	sess.Step = stepAwaitNumber
	t.sessions.Set(chatID, sess)

	t.responder.Text(ctx, chatID, "Введите номер дома:")
}

func (t *Transport) handleAwaitNumber(ctx context.Context, update schemes.UpdateInterface) {
	u, ok := core_max_request.AsMessageCreated(update)
	if !ok {
		return
	}

	chatID := u.GetChatID()
	sess, _ := t.sessions.Get(chatID)

	user, allowed := core_max_middleware.CheckRole(ctx, core_domain.RoleCompany)
	if !allowed {
		t.sessions.Clear(chatID)
		t.responder.Text(ctx, chatID, "⛔ Раздел доступен только представителям УК.")
		return
	}

	house, err := t.service.Create(ctx, user.Id, core_domain.House{
		City:        sess.Draft[draftCity],
		Street:      sess.Draft[draftStreet],
		HouseNumber: strings.TrimSpace(u.GetText()),
	})
	if err != nil {
		t.responder.Error(ctx, chatID, err)
		return
	}

	t.sessions.Clear(chatID)
	t.responder.Keyboard(
		ctx,
		chatID,
		fmt.Sprintf(
			"✅ Дом зарегистрирован: %s (id %d)\n\nТеперь жители смогут найти его сами — по точному адресу, при добавлении своей квартиры.",
			house.Address(), house.Id,
		),
		maxbot.InlineKeyboard(maxbot.Row(maxbot.Btn("🔙 К моим домам", core_max_payload.MenuHouses))),
	)
}

func (t *Transport) handleOpen(ctx context.Context, update schemes.UpdateInterface) {
	u, ok := core_max_request.AsMessageCallback(update)
	if !ok {
		return
	}

	t.responder.AnswerCallback(ctx, u.Callback.CallbackID, "")

	if _, allowed := core_max_middleware.CheckRole(ctx, core_domain.RoleCompany); !allowed {
		t.responder.Text(ctx, u.GetChatID(), "⛔ Раздел доступен только представителям УК.")
		return
	}

	_, arg := core_max_request.SplitPayload(u.Callback.Payload)

	houseID, err := strconv.ParseInt(arg, 10, 64)
	if err != nil {
		t.responder.Text(ctx, u.GetChatID(), "❗ Некорректный id дома.")
		return
	}

	flats, err := t.service.FlatsByHouse(ctx, houseID)
	if err != nil {
		t.responder.Error(ctx, u.GetChatID(), err)
		return
	}

	var text string
	if len(flats) == 0 {
		text = fmt.Sprintf("📋 В доме #%d пока нет ни одной зарегистрированной квартиры.", houseID)
	} else {
		text = fmt.Sprintf("📋 Квартиры дома (%s):\n\n", flats[0].Address())
		for _, f := range flats {
			text += fmt.Sprintf("• кв. %d\n", f.FlatNumber)
		}
	}

	t.responder.Keyboard(ctx, u.GetChatID(), text, maxbot.InlineKeyboard(maxbot.Row(maxbot.Btn("🔙 К моим домам", core_max_payload.MenuHouses))))
}
