package telegram

import (
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (h *TelegramHandler) createLanguageKeyboard(interfaceLang, keyboardType string, excludeLang string, showBackButton bool) tgbotapi.InlineKeyboardMarkup {
	type langOption struct{ code, flag string }
	languages := []langOption{
		{"en", "🇺🇸"}, {"ru", "🇷🇺"}, {"es", "🇪🇸"}, {"zh", "🇨🇳"},
	}

	// Используем Map для автоматического удаления дубликатов
	uniqueButtons := make(map[string]tgbotapi.InlineKeyboardButton)

	for _, lang := range languages {
		// Исключаем указанный язык из списка
		if lang.code == excludeLang {
			continue
		}

		name := h.service.Localizer.GetLanguageName(lang.code, interfaceLang)
		label := fmt.Sprintf("%s %s", lang.flag, name)
		callbackData := fmt.Sprintf("lang_%s_%s", keyboardType, lang.code)

		// Избегаем дубликатов по callback data (на случай если название языка совпадает)
		if _, exists := uniqueButtons[callbackData]; !exists {
			button := tgbotapi.NewInlineKeyboardButtonData(label, callbackData)
			uniqueButtons[callbackData] = button
		}
	}

	// Преобразуем map в массив кнопок
	var buttons [][]tgbotapi.InlineKeyboardButton
	for _, button := range uniqueButtons {
		buttons = append(buttons, []tgbotapi.InlineKeyboardButton{button})
	}

	// Добавляем кнопку "Назад", если это необходимо
	if showBackButton {
		var backCallback string
		if keyboardType == "interface" || keyboardType == "native" {
			backCallback = "back_to_main_menu"
		} else {
			backCallback = "back_to_previous_step"
		}

		backButton := tgbotapi.NewInlineKeyboardButtonData(
			h.service.Localizer.Get(interfaceLang, "back_button"),
			backCallback,
		)
		buttons = append(buttons, []tgbotapi.InlineKeyboardButton{backButton})
	}

	return tgbotapi.NewInlineKeyboardMarkup(buttons...)
}

func (h *TelegramHandler) createProfileCompletedKeyboard(interfaceLang string) tgbotapi.InlineKeyboardMarkup {
	mainMenu := tgbotapi.NewInlineKeyboardButtonData(
		h.service.Localizer.Get(interfaceLang, "main_menu_title"),
		"back_to_main_menu",
	)
	viewProfile := tgbotapi.NewInlineKeyboardButtonData(
		h.service.Localizer.Get(interfaceLang, "profile_show"),
		"profile_show",
	)
	buttons := [][]tgbotapi.InlineKeyboardButton{
		{mainMenu, viewProfile},
	}
	return tgbotapi.NewInlineKeyboardMarkup(buttons...)
}

// Создание клавиатуры подтверждения выбора языков
func (h *TelegramHandler) createLanguageConfirmationKeyboard(interfaceLang string) tgbotapi.InlineKeyboardMarkup {
	continueButton := tgbotapi.NewInlineKeyboardButtonData(
		h.service.Localizer.Get(interfaceLang, "languages_continue_filling"),
		"languages_continue_filling",
	)
	reselectButton := tgbotapi.NewInlineKeyboardButtonData(
		h.service.Localizer.Get(interfaceLang, "languages_reselect"),
		"languages_reselect",
	)

	// Только две кнопки без кнопки "Назад"
	buttons := [][]tgbotapi.InlineKeyboardButton{
		{continueButton},
		{reselectButton},
	}

	return tgbotapi.NewInlineKeyboardMarkup(buttons...)
}

// Создание клавиатуры выбора уровня владения языком
func (h *TelegramHandler) createLanguageLevelKeyboard(interfaceLang, languageCode string) tgbotapi.InlineKeyboardMarkup {
	return h.createLanguageLevelKeyboardWithPrefix(interfaceLang, languageCode, "level_", true)
}

// createLanguageLevelKeyboardWithPrefix создает клавиатуру выбора уровня с произвольным префиксом колбэков
func (h *TelegramHandler) createLanguageLevelKeyboardWithPrefix(interfaceLang, languageCode, callbackPrefix string, showBackButton bool) tgbotapi.InlineKeyboardMarkup {
	levels := []struct {
		code, key string
	}{
		{"beginner", "choose_level_beginner"},
		{"elementary", "choose_level_elementary"},
		{"intermediate", "choose_level_intermediate"},
		{"upper_intermediate", "choose_level_upper_intermediate"},
		{"advanced", "choose_level_advanced"},
	}

	var buttons [][]tgbotapi.InlineKeyboardButton
	for _, level := range levels {
		label := h.service.Localizer.Get(interfaceLang, level.key)
		button := tgbotapi.NewInlineKeyboardButtonData(label, fmt.Sprintf("%s%s", callbackPrefix, level.code))
		buttons = append(buttons, []tgbotapi.InlineKeyboardButton{button})
	}

	// Добавляем кнопку "Назад" только если необходимо
	if showBackButton {
		backButton := tgbotapi.NewInlineKeyboardButtonData(
			h.service.Localizer.Get(interfaceLang, "back_button"),
			"back_to_previous_step",
		)
		buttons = append(buttons, []tgbotapi.InlineKeyboardButton{backButton})
	}

	return tgbotapi.NewInlineKeyboardMarkup(buttons...)
}
