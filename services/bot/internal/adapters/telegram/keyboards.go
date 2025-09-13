package telegram

import (
	"fmt"
	"sort"

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
		backButton := tgbotapi.NewInlineKeyboardButtonData(
			h.service.Localizer.Get(interfaceLang, "back_button"),
			"back_to_previous_step",
		)
		buttons = append(buttons, []tgbotapi.InlineKeyboardButton{backButton})
	}

	return tgbotapi.NewInlineKeyboardMarkup(buttons...)
}

func (h *TelegramHandler) createInterestsKeyboard(interests map[int]string, selectedInterests []int, interfaceLang string) tgbotapi.InlineKeyboardMarkup {
	var buttonRows [][]tgbotapi.InlineKeyboardButton

	// Создаём мапу для быстрой проверки выбранных
	selectedMap := make(map[int]bool)
	for _, id := range selectedInterests {
		selectedMap[id] = true
	}

	// Получаем отсортированные ID для стабильного порядка
	var sortedIDs []int
	for id := range interests {
		sortedIDs = append(sortedIDs, id)
	}
	sort.Ints(sortedIDs)

	// Создаём кнопки по 2 в ряд для более компактного вида
	var currentRow []tgbotapi.InlineKeyboardButton

	for _, id := range sortedIDs {
		name := interests[id]
		var label string
		if selectedMap[id] {
			label = "✅ " + name
		} else {
			label = name
		}

		button := tgbotapi.NewInlineKeyboardButtonData(label, fmt.Sprintf("interest_%d", id))
		currentRow = append(currentRow, button)

		// Когда в ряду 2 кнопки, добавляем ряд полностью
		if len(currentRow) == 2 {
			buttonRows = append(buttonRows, currentRow)
			currentRow = []tgbotapi.InlineKeyboardButton{}
		}
	}

	// Добавляем остаток последнего ряда
	if len(currentRow) > 0 {
		buttonRows = append(buttonRows, currentRow)
	}

	// Нижний блок с кнопками управления
	var controlRow []tgbotapi.InlineKeyboardButton

	// Кнопка "Продолжить" всегда видна
	continueText := h.service.Localizer.Get(interfaceLang, "interests_continue")
	continueButton := tgbotapi.NewInlineKeyboardButtonData(
		continueText,
		"interests_continue",
	)
	controlRow = append(controlRow, continueButton)

	// Добавляем кнопку "Назад"
	backButton := tgbotapi.NewInlineKeyboardButtonData(
		h.service.Localizer.Get(interfaceLang, "back_button"),
		"back_to_previous_step",
	)
	controlRow = append(controlRow, backButton)

	buttonRows = append(buttonRows, controlRow)

	return tgbotapi.NewInlineKeyboardMarkup(buttonRows...)
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
		button := tgbotapi.NewInlineKeyboardButtonData(label, fmt.Sprintf("level_%s", level.code))
		buttons = append(buttons, []tgbotapi.InlineKeyboardButton{button})
	}

	// Добавляем кнопку "Назад"
	backButton := tgbotapi.NewInlineKeyboardButtonData(
		h.service.Localizer.Get(interfaceLang, "back_button"),
		"back_to_previous_step",
	)
	buttons = append(buttons, []tgbotapi.InlineKeyboardButton{backButton})

	return tgbotapi.NewInlineKeyboardMarkup(buttons...)
}
