package telegram

import (
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (h *TelegramHandler) createLanguageKeyboard(interfaceLang, keyboardType string) tgbotapi.InlineKeyboardMarkup {
	type langOpt struct{ code, flag string }
	langs := []langOpt{
		{"en", "🇺🇸"}, {"ru", "🇷🇺"}, {"es", "🇪🇸"}, {"zh", "🇨🇳"},
	}
	var rows [][]tgbotapi.InlineKeyboardButton
	for _, l := range langs {
		name := h.service.Localizer.GetLanguageName(l.code, interfaceLang)
		label := fmt.Sprintf("%s %s", l.flag, name)
		btn := tgbotapi.NewInlineKeyboardButtonData(label, fmt.Sprintf("lang_%s_%s", keyboardType, l.code))
		rows = append(rows, []tgbotapi.InlineKeyboardButton{btn})
	}
	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

func (h *TelegramHandler) createInterestsKeyboard(interests map[int]string) tgbotapi.InlineKeyboardMarkup {
	buttons := make([][]tgbotapi.InlineKeyboardButton, 0, len(interests))
	for id, name := range interests {
		button := tgbotapi.NewInlineKeyboardButtonData(
			name,
			fmt.Sprintf("interest_%d", id),
		)
		buttons = append(buttons, []tgbotapi.InlineKeyboardButton{button})
	}
	return tgbotapi.NewInlineKeyboardMarkup(buttons...)
}
