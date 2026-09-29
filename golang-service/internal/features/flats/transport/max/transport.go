package flats_max_transport

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	core_domain "github.com/m0dris/hackathon/internal/core/domain"
	core_errors "github.com/m0dris/hackathon/internal/core/errors"
	core_max_middleware "github.com/m0dris/hackathon/internal/core/transport/max/middleware"
	core_max_payload "github.com/m0dris/hackathon/internal/core/transport/max/payload"
	core_max_request "github.com/m0dris/hackathon/internal/core/transport/max/request"
	core_max_response "github.com/m0dris/hackathon/internal/core/transport/max/response"
	core_max_server "github.com/m0dris/hackathon/internal/core/transport/max/server"
	core_max_session "github.com/m0dris/hackathon/internal/core/transport/max/session"
	flats_service "github.com/m0dris/hackathon/internal/features/flats/service"
	houses_service "github.com/m0dris/hackathon/internal/features/houses/service"
	maxbot "github.com/max-messenger/max-bot-api-client-go"
	"github.com/max-messenger/max-bot-api-client-go/schemes"
)

const (
	stepAwaitCity        = "flat:add:city"
	stepAwaitStreet      = "flat:add:street"
	stepAwaitHouseNumber = "flat:add:house_number"
	stepAwaitFlatNumber  = "flat:add:flat_number"

	draftCity    = "city"
	draftStreet  = "street"
	draftHouseID = "house_id"
)

type Transport struct {
	service   *flats_service.Service
	houses    *houses_service.Service
	responder *core_max_response.Responder
	sessions  core_max_session.Store
}

func New(service *flats_service.Service, houses *houses_service.Service, responder *core_max_response.Responder, sessions core_max_session.Store) *Transport {
	return &Transport{service: service, houses: houses, responder: responder, sessions: sessions}
}

func (t *Transport) Routes() []core_max_server.Route {
	return []core_max_server.Route{
		{Key: core_max_payload.MenuFlats, Handler: t.handleMyFlats},
		{Key: core_max_payload.FlatAdd, Handler: t.handleAddStart},
		{Key: stepAwaitCity, Handler: t.handleAwaitCity},
		{Key: stepAwaitStreet, Handler: t.handleAwaitStreet},
		{Key: stepAwaitHouseNumber, Handler: t.handleAwaitHouseNumber},
		{Key: stepAwaitFlatNumber, Handler: t.handleAwaitFlatNumber},
	}
}

func (t *Transport) handleMyFlats(ctx context.Context, update schemes.UpdateInterface) {
	u, ok := core_max_request.AsMessageCallback(update)
	if !ok {
		return
	}

	t.responder.AnswerCallback(ctx, u.Callback.CallbackID, "")

	user, allowed := core_max_middleware.CheckRole(ctx, core_domain.RoleUser)
	if !allowed {
		t.responder.Text(ctx, u.GetChatID(), "⛔ Раздел доступен только жителям.")
		return
	}

	flats, err := t.service.My(ctx, user.Id)
	if err != nil {
		t.responder.Error(ctx, u.GetChatID(), err)
		return
	}

	text := "🏠 У вас пока нет зарегистрированных квартир."
	if len(flats) > 0 {
		text = "🏠 Ваши квартиры:\n\n"
		for _, f := range flats {
			text += "• " + f.Flat.Address() + "\n"
		}
	}

	kb := maxbot.InlineKeyboard(
		maxbot.Row(maxbot.Btn("➕ Добавить квартиру", core_max_payload.FlatAdd)),
		maxbot.Row(maxbot.Btn("🔙 В меню", core_max_payload.MenuMain)),
	)

	t.responder.Keyboard(ctx, u.GetChatID(), text, kb)
}

func (t *Transport) handleAddStart(ctx context.Context, update schemes.UpdateInterface) {
	u, ok := core_max_request.AsMessageCallback(update)
	if !ok {
		return
	}

	t.responder.AnswerCallback(ctx, u.Callback.CallbackID, "")

	if _, allowed := core_max_middleware.CheckRole(ctx, core_domain.RoleUser); !allowed {
		t.responder.Text(ctx, u.GetChatID(), "⛔ Раздел доступен только жителям.")
		return
	}

	t.sessions.Set(u.GetChatID(), core_max_session.Session{Step: stepAwaitCity, Draft: map[string]string{}})
	t.responder.Text(ctx, u.GetChatID(), "По какому адресу ваш дом? Введите город:")
}

func (t *Transport) handleAwaitCity(ctx context.Context, update schemes.UpdateInterface) {
	u, ok := core_max_request.AsMessageCreated(update)
	if !ok {
		return
	}

	chatID := u.GetChatID()

	city := strings.TrimSpace(u.GetText())
	if len([]rune(city)) < 2 {
		t.responder.Text(ctx, chatID, "❗ Слишком коротко. Введите город ещё раз:")
		return
	}

	sess, _ := t.sessions.Get(chatID)
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

	street := strings.TrimSpace(u.GetText())
	if len([]rune(street)) < 2 {
		t.responder.Text(ctx, chatID, "❗ Слишком коротко. Введите улицу ещё раз:")
		return
	}

	sess, _ := t.sessions.Get(chatID)
	sess.Draft[draftStreet] = street
	sess.Step = stepAwaitHouseNumber
	t.sessions.Set(chatID, sess)

	t.responder.Text(ctx, chatID, "Введите номер дома:")
}

func (t *Transport) handleAwaitHouseNumber(ctx context.Context, update schemes.UpdateInterface) {
	u, ok := core_max_request.AsMessageCreated(update)
	if !ok {
		return
	}

	chatID := u.GetChatID()
	sess, _ := t.sessions.Get(chatID)

	houseNumber := strings.TrimSpace(u.GetText())

	house, err := t.houses.SearchByAddress(ctx, sess.Draft[draftCity], sess.Draft[draftStreet], houseNumber)
	if err != nil {
		if errors.Is(err, core_errors.ErrNotFound) {
			t.sessions.Set(chatID, core_max_session.Session{Step: stepAwaitCity, Draft: map[string]string{}})
			t.responder.Text(
				ctx, chatID,
				"🔎 Такой дом не найден — проверьте адрес или уточните у своей УК, зарегистрирован ли дом в системе.\n\nВведите город ещё раз:",
			)
			return
		}

		t.responder.Error(ctx, chatID, err)
		return
	}

	sess.Draft[draftHouseID] = strconv.FormatInt(house.Id, 10)
	sess.Step = stepAwaitFlatNumber
	t.sessions.Set(chatID, sess)

	t.responder.Text(ctx, chatID, fmt.Sprintf("Нашёл дом: %s.\nВведите номер квартиры:", house.Address()))
}

func (t *Transport) handleAwaitFlatNumber(ctx context.Context, update schemes.UpdateInterface) {
	u, ok := core_max_request.AsMessageCreated(update)
	if !ok {
		return
	}

	chatID := u.GetChatID()
	sess, _ := t.sessions.Get(chatID)

	user, allowed := core_max_middleware.CheckRole(ctx, core_domain.RoleUser)
	if !allowed {
		t.sessions.Clear(chatID)
		t.responder.Text(ctx, chatID, "⛔ Раздел доступен только жителям.")
		return
	}

	flatNumber, err := strconv.Atoi(strings.TrimSpace(u.GetText()))
	if err != nil {
		t.responder.Text(ctx, chatID, "❗ Номер квартиры — это число. Попробуйте ещё раз:")
		return
	}

	houseID, _ := strconv.ParseInt(sess.Draft[draftHouseID], 10, 64)

	uf, err := t.service.AddMy(ctx, user.Id, houseID, flatNumber)
	if err != nil {
		t.responder.Error(ctx, chatID, err)
		return
	}

	t.sessions.Clear(chatID)
	t.responder.Keyboard(
		ctx,
		chatID,
		fmt.Sprintf("✅ Квартира добавлена: %s", uf.Flat.Address()),
		maxbot.InlineKeyboard(maxbot.Row(maxbot.Btn("🔙 К моим квартирам", core_max_payload.MenuFlats))),
	)
}
