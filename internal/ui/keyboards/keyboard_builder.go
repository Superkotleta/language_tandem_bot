package keyboards

import (
	"language-exchange-bot/internal/pkg/i18n"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// KeyboardBuilder builds Telegram keyboards
type KeyboardBuilder struct {
	localizer *i18n.Localizer
}

// NewKeyboardBuilder creates a new keyboard builder
func NewKeyboardBuilder(localizer *i18n.Localizer) *KeyboardBuilder {
	return &KeyboardBuilder{
		localizer: localizer,
	}
}

// MainMenu creates the main menu keyboard
// Shows "Заполнить Профиль" if hasProfile is false
// Shows "Мой Профиль" if hasProfile is true
func (kb *KeyboardBuilder) MainMenu(lang string, hasProfile bool) tgbotapi.ReplyKeyboardMarkup {
	var buttons [][]tgbotapi.KeyboardButton

	if hasProfile {
		// User has completed profile
		buttons = [][]tgbotapi.KeyboardButton{
			{
				tgbotapi.NewKeyboardButton(kb.localizer.Get(lang, "btn_profile")),
			},
		}
	} else {
		// User needs to fill profile
		buttons = [][]tgbotapi.KeyboardButton{
			{
				tgbotapi.NewKeyboardButton(kb.localizer.Get(lang, "btn_fill_profile")),
			},
		}
	}

	return tgbotapi.NewReplyKeyboard(buttons...)
}

// ProfileActions creates inline keyboard for profile actions
func (kb *KeyboardBuilder) ProfileActions(lang string) tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				kb.localizer.Get(lang, "btn_edit_profile"),
				"edit_profile",
			),
		),
	)
}
