package core_max_menu

import (
	core_domain "github.com/m0dris/hackathon/internal/core/domain"
	core_max_payload "github.com/m0dris/hackathon/internal/core/transport/max/payload"
	maxbot "github.com/max-messenger/max-bot-api-client-go"
	"github.com/max-messenger/max-bot-api-client-go/schemes"
)

const Text = "Главное меню:"

func Keyboard(role core_domain.Role) *maxbot.Keyboard {
	var rows [][]schemes.ButtonInterface

	if role == core_domain.RoleCompany {
		rows = append(rows, maxbot.Row(maxbot.Btn("🏢 Мои дома", core_max_payload.MenuHouses)))
	} else {
		rows = append(rows, maxbot.Row(maxbot.Btn("🏠 Мои квартиры", core_max_payload.MenuFlats)))
	}

	rows = append(rows,
		maxbot.Row(maxbot.Btn("📝 Заявки", core_max_payload.MenuRequests)),
		maxbot.Row(maxbot.Btn("👤 Профиль", core_max_payload.MenuProfile)),
	)

	return maxbot.InlineKeyboard(rows...)
}
