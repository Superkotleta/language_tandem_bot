package telegram

import (
	"fmt"
	"sort"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (h *TelegramHandler) createLanguageKeyboard(interfaceLang, keyboardType string) tgbotapi.InlineKeyboardMarkup {
	type langOption struct{ code, flag string }
	languages := []langOption{
		{"en", "🇺🇸"}, {"ru", "🇷🇺"}, {"es", "🇪🇸"}, {"zh", "🇨🇳"},
	}

	var buttons [][]tgbotapi.InlineKeyboardButton
	for _, lang := range languages {
		name := h.service.Localizer.GetLanguageName(lang.code, interfaceLang)
		label := fmt.Sprintf("%s %s", lang.flag, name)
		button := tgbotapi.NewInlineKeyboardButtonData(label, fmt.Sprintf("lang_%s_%s", keyboardType, lang.code))
		buttons = append(buttons, []tgbotapi.InlineKeyboardButton{button})
	}
	return tgbotapi.NewInlineKeyboardMarkup(buttons...)
}

func (h *TelegramHandler) createInterestsKeyboard(interests map[int]string, selectedInterests []int) tgbotapi.InlineKeyboardMarkup {
	var buttons [][]tgbotapi.InlineKeyboardButton

	// ✅ Создаём мапу для быстрой проверки выбранных
	selectedMap := make(map[int]bool)
	for _, id := range selectedInterests {
		selectedMap[id] = true
	}

	// ✅ КЛЮЧЕВОЙ ФИК: Получаем отсортированные ID для стабильного порядка
	var sortedIDs []int
	for id := range interests {
		sortedIDs = append(sortedIDs, id)
	}
	// Сортируем по возрастанию ID
	sort.Ints(sortedIDs)

	// ✅ Создаём кнопки в стабильном порядке
	for _, id := range sortedIDs {
		name := interests[id]
		var label string
		if selectedMap[id] {
			label = "✅ " + name // Галочка для выбранных
		} else {
			label = name
		}

		button := tgbotapi.NewInlineKeyboardButtonData(label, fmt.Sprintf("interest_%d", id))
		buttons = append(buttons, []tgbotapi.InlineKeyboardButton{button})
	}

	// Кнопка "Продолжить" если есть выбранные интересы
	if len(selectedInterests) > 0 {
		continueButton := tgbotapi.NewInlineKeyboardButtonData(
			"▶️ Продолжить", "interests_continue",
		)
		buttons = append(buttons, []tgbotapi.InlineKeyboardButton{continueButton})
	}

	return tgbotapi.NewInlineKeyboardMarkup(buttons...)
}
