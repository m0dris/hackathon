package requests_max_transport

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	core_domain "github.com/m0dris/hackathon/internal/core/domain"
	core_max_middleware "github.com/m0dris/hackathon/internal/core/transport/max/middleware"
	core_max_payload "github.com/m0dris/hackathon/internal/core/transport/max/payload"
	core_max_request "github.com/m0dris/hackathon/internal/core/transport/max/request"
	core_max_response "github.com/m0dris/hackathon/internal/core/transport/max/response"
	core_max_server "github.com/m0dris/hackathon/internal/core/transport/max/server"
	core_max_session "github.com/m0dris/hackathon/internal/core/transport/max/session"
	flats_service "github.com/m0dris/hackathon/internal/features/flats/service"
	requests_service "github.com/m0dris/hackathon/internal/features/requests/service"
	maxbot "github.com/max-messenger/max-bot-api-client-go"
	"github.com/max-messenger/max-bot-api-client-go/schemes"
)

const (
	stepAwaitTitle       = "req:new:title"
	stepAwaitDescription = "req:new:description"

	draftUserFlatsID = "user_flats_id"
	draftType        = "type"
	draftTitle       = "title"
)

type Transport struct {
	service   *requests_service.Service
	flats     *flats_service.Service
	responder *core_max_response.Responder
	sessions  core_max_session.Store
}

func New(service *requests_service.Service, flats *flats_service.Service, responder *core_max_response.Responder, sessions core_max_session.Store) *Transport {
	return &Transport{service: service, flats: flats, responder: responder, sessions: sessions}
}

func (t *Transport) Routes() []core_max_server.Route {
	return []core_max_server.Route{
		{Key: core_max_payload.MenuRequests, Handler: t.handleMenu},
		{Key: core_max_payload.RequestNew, Handler: t.handleNewStart},
		{Key: core_max_payload.RequestNewFlat, Handler: t.handleFlatPicked},
		{Key: core_max_payload.RequestType, Handler: t.handleTypePicked},
		{Key: stepAwaitTitle, Handler: t.handleAwaitTitle},
		{Key: stepAwaitDescription, Handler: t.handleAwaitDescription},
		{Key: core_max_payload.RequestFilter, Handler: t.handleFilter},
		{Key: core_max_payload.RequestAdvance, Handler: t.handleAdvance},
	}
}

func (t *Transport) handleMenu(ctx context.Context, update schemes.UpdateInterface) {
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

	if user.IsCompany() {
		t.responder.Keyboard(ctx, u.GetChatID(), "📋 Заявки от жителей. Выберите фильтр:", filterKeyboard())
		return
	}

	t.showMyRequests(ctx, u.GetChatID(), user.Id)
}

func (t *Transport) showMyRequests(ctx context.Context, chatID, userID int64) {
	reqs, err := t.service.My(ctx, userID)
	if err != nil {
		t.responder.Error(ctx, chatID, err)
		return
	}

	text := "📝 У вас пока нет заявок."
	if len(reqs) > 0 {
		text = "📝 Ваши заявки:\n\n"
		for _, r := range reqs {
			text += fmt.Sprintf("• #%d %s — %s (%s)\n", r.Id, r.Type.Label(), r.Title, r.Status.Label())
		}
	}

	t.responder.Keyboard(ctx, chatID, text, maxbot.InlineKeyboard(
		maxbot.Row(maxbot.Btn("➕ Новая заявка", core_max_payload.RequestNew)),
		maxbot.Row(maxbot.Btn("🔙 В меню", core_max_payload.MenuMain)),
	))
}

func (t *Transport) handleFilter(ctx context.Context, update schemes.UpdateInterface) {
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

	_, arg := core_max_request.SplitPayload(u.Callback.Payload)

	var status *core_domain.RequestStatus
	if arg != "all" && arg != "" {
		s := core_domain.RequestStatus(arg)
		status = &s
	}

	reqs, err := t.service.Company(ctx, user.Id, status)
	if err != nil {
		t.responder.Error(ctx, u.GetChatID(), err)
		return
	}

	text := "📋 Заявок по этому фильтру нет."
	var rows [][]schemes.ButtonInterface
	if len(reqs) > 0 {
		text = "📋 Заявки от жителей:\n\n"
		for _, r := range reqs {
			text += fmt.Sprintf(
				"• #%d %s — %s (%s)\n  📝 %s\n  🏠 %s\n  👤 %s, тел. %s\n\n",
				r.Id, r.Type.Label(), r.Title, r.Status.Label(), r.Description, r.Address(), r.ApplicantName(), r.ApplicantPhone,
			)

			if next := r.Status.NextStatus(); next != "" {
				payload := fmt.Sprintf("%s:%d-%s", core_max_payload.RequestAdvance, r.Id, r.Status)
				rows = append(rows, maxbot.Row(maxbot.Btn(fmt.Sprintf("%s (#%d)", r.Status.NextActionLabel(), r.Id), payload)))
			}
		}
	}

	rows = append(rows, filterRows()...)
	rows = append(rows, maxbot.Row(maxbot.Btn("🔙 В меню", core_max_payload.MenuMain)))

	t.responder.Keyboard(ctx, u.GetChatID(), text, maxbot.InlineKeyboard(rows...))
}

func (t *Transport) handleAdvance(ctx context.Context, update schemes.UpdateInterface) {
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

	_, arg := core_max_request.SplitPayload(u.Callback.Payload)

	idStr, currentStatusStr, found := strings.Cut(arg, "-")
	if !found {
		t.responder.Text(ctx, u.GetChatID(), "❗ Некорректные данные кнопки.")
		return
	}

	requestID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		t.responder.Text(ctx, u.GetChatID(), "❗ Некорректный id заявки.")
		return
	}

	next := core_domain.RequestStatus(currentStatusStr).NextStatus()
	if next == "" {
		t.responder.Text(ctx, u.GetChatID(), "Эта заявка уже в терминальном статусе.")
		return
	}

	if _, err := t.service.UpdateStatus(ctx, user.Id, requestID, next); err != nil {
		t.responder.Error(ctx, u.GetChatID(), err)
		return
	}

	t.responder.Text(ctx, u.GetChatID(), fmt.Sprintf("✅ Заявка #%d переведена в статус «%s».", requestID, next.Label()))
	t.handleFilter(ctx, update)
}

func (t *Transport) handleNewStart(ctx context.Context, update schemes.UpdateInterface) {
	u, ok := core_max_request.AsMessageCallback(update)
	if !ok {
		return
	}

	t.responder.AnswerCallback(ctx, u.Callback.CallbackID, "")

	user, allowed := core_max_middleware.CheckRole(ctx, core_domain.RoleUser)
	if !allowed {
		t.responder.Text(ctx, u.GetChatID(), "⛔ Подавать заявки могут только жители.")
		return
	}

	myFlats, err := t.flats.My(ctx, user.Id)
	if err != nil {
		t.responder.Error(ctx, u.GetChatID(), err)
		return
	}

	if len(myFlats) == 0 {
		t.responder.Keyboard(
			ctx, u.GetChatID(),
			"Сначала добавьте квартиру — без неё некому подавать заявку.",
			maxbot.InlineKeyboard(maxbot.Row(maxbot.Btn("➕ Добавить квартиру", core_max_payload.FlatAdd))),
		)
		return
	}

	if len(myFlats) == 1 {
		t.sessions.Set(u.GetChatID(), core_max_session.Session{
			Draft: map[string]string{draftUserFlatsID: strconv.FormatInt(myFlats[0].UserFlatId, 10)},
		})
		t.responder.Keyboard(ctx, u.GetChatID(), "Что случилось?", typeKeyboard())

		return
	}

	var rows [][]schemes.ButtonInterface
	for _, f := range myFlats {
		payload := fmt.Sprintf("%s:%d", core_max_payload.RequestNewFlat, f.UserFlatId)
		rows = append(rows, maxbot.Row(maxbot.Btn(f.Flat.Address(), payload)))
	}

	t.responder.Keyboard(ctx, u.GetChatID(), "От лица какой квартиры подаём заявку?", maxbot.InlineKeyboard(rows...))
}

func (t *Transport) handleFlatPicked(ctx context.Context, update schemes.UpdateInterface) {
	u, ok := core_max_request.AsMessageCallback(update)
	if !ok {
		return
	}

	t.responder.AnswerCallback(ctx, u.Callback.CallbackID, "")

	if _, allowed := core_max_middleware.CheckRole(ctx, core_domain.RoleUser); !allowed {
		t.responder.Text(ctx, u.GetChatID(), "⛔ Подавать заявки могут только жители.")
		return
	}

	_, arg := core_max_request.SplitPayload(u.Callback.Payload)

	t.sessions.Set(u.GetChatID(), core_max_session.Session{Draft: map[string]string{draftUserFlatsID: arg}})
	t.responder.Keyboard(ctx, u.GetChatID(), "Что случилось?", typeKeyboard())
}

func (t *Transport) handleTypePicked(ctx context.Context, update schemes.UpdateInterface) {
	u, ok := core_max_request.AsMessageCallback(update)
	if !ok {
		return
	}

	t.responder.AnswerCallback(ctx, u.Callback.CallbackID, "")

	if _, allowed := core_max_middleware.CheckRole(ctx, core_domain.RoleUser); !allowed {
		t.responder.Text(ctx, u.GetChatID(), "⛔ Подавать заявки могут только жители.")
		return
	}

	chatID := u.GetChatID()

	sess, ok := t.sessions.Get(chatID)
	if !ok {
		t.responder.Text(ctx, chatID, "Сессия истекла, начните заново: /menu → «Заявки».")
		return
	}

	_, arg := core_max_request.SplitPayload(u.Callback.Payload)

	sess.Draft[draftType] = arg
	sess.Step = stepAwaitTitle
	t.sessions.Set(chatID, sess)

	t.responder.Text(ctx, chatID, "Коротко опишите проблему одной фразой (заголовок, до 255 символов):")
}

func (t *Transport) handleAwaitTitle(ctx context.Context, update schemes.UpdateInterface) {
	u, ok := core_max_request.AsMessageCreated(update)
	if !ok {
		return
	}

	chatID := u.GetChatID()
	sess, _ := t.sessions.Get(chatID)

	title := strings.TrimSpace(u.GetText())
	if l := len([]rune(title)); l < 3 || l > 255 {
		t.responder.Text(ctx, chatID, "❗ Заголовок должен быть от 3 до 255 символов. Попробуйте ещё раз:")
		return
	}

	sess.Draft[draftTitle] = title
	sess.Step = stepAwaitDescription
	t.sessions.Set(chatID, sess)

	t.responder.Text(ctx, chatID, "Опишите подробнее, что произошло:")
}

func (t *Transport) handleAwaitDescription(ctx context.Context, update schemes.UpdateInterface) {
	u, ok := core_max_request.AsMessageCreated(update)
	if !ok {
		return
	}

	chatID := u.GetChatID()
	sess, _ := t.sessions.Get(chatID)

	user, allowed := core_max_middleware.CheckRole(ctx, core_domain.RoleUser)
	if !allowed {
		t.sessions.Clear(chatID)
		t.responder.Text(ctx, chatID, "⛔ Подавать заявки могут только жители.")
		return
	}

	userFlatsID, _ := strconv.ParseInt(sess.Draft[draftUserFlatsID], 10, 64)

	req, err := t.service.Create(ctx, user.Id, core_domain.ServiceRequest{
		UserFlatsId: userFlatsID,
		Type:        core_domain.RequestType(sess.Draft[draftType]),
		Title:       sess.Draft[draftTitle],
		Description: strings.TrimSpace(u.GetText()),
	})
	if err != nil {
		t.responder.Error(ctx, chatID, err)
		return
	}

	t.sessions.Clear(chatID)
	t.responder.Keyboard(
		ctx, chatID,
		fmt.Sprintf("✅ Заявка #%d создана, статус: %s.", req.Id, req.Status.Label()),
		maxbot.InlineKeyboard(maxbot.Row(maxbot.Btn("🔙 К заявкам", core_max_payload.MenuRequests))),
	)
}

func typeKeyboard() *maxbot.Keyboard {
	var rows [][]schemes.ButtonInterface
	for _, l := range core_domain.RequestTypeLabels {
		payload := fmt.Sprintf("%s:%s", core_max_payload.RequestType, l.Type)
		rows = append(rows, maxbot.Row(maxbot.Btn(l.Label, payload)))
	}

	return maxbot.InlineKeyboard(rows...)
}

func filterRows() [][]schemes.ButtonInterface {
	return [][]schemes.ButtonInterface{
		maxbot.Row(
			maxbot.Btn("Все", core_max_payload.RequestFilter+":all"),
			maxbot.Btn("🆕 Новые", core_max_payload.RequestFilter+":"+string(core_domain.RequestStatusNew)),
		),
		maxbot.Row(
			maxbot.Btn("🔧 В работе", core_max_payload.RequestFilter+":"+string(core_domain.RequestStatusInProgress)),
			maxbot.Btn("✅ Выполненные", core_max_payload.RequestFilter+":"+string(core_domain.RequestStatusDone)),
		),
	}
}

func filterKeyboard() *maxbot.Keyboard {
	rows := filterRows()
	rows = append(rows, maxbot.Row(maxbot.Btn("🔙 В меню", core_max_payload.MenuMain)))

	return maxbot.InlineKeyboard(rows...)
}
