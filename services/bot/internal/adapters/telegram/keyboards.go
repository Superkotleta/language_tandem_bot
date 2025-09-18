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

func (h *TelegramHandler) createSaveEditsKeyboard(interfaceLang string) tgbotapi.InlineKeyboardMarkup {
	save := tgbotapi.NewInlineKeyboardButtonData(
		h.service.Localizer.Get(interfaceLang, "save_button"),
		"save_edits",
	)
	cancel := tgbotapi.NewInlineKeyboardButtonData(
		h.service.Localizer.Get(interfaceLang, "cancel_button"),
		"cancel_edits",
	)
	buttons := [][]tgbotapi.InlineKeyboardButton{
		{save, cancel},
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

func (h *TelegramHandler) createEditInterestsKeyboard(interests map[int]string, selectedInterests []int, interfaceLang string) tgbotapi.InlineKeyboardMarkup {
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

		button := tgbotapi.NewInlineKeyboardButtonData(label, fmt.Sprintf("edit_interest_%d", id))
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

	// Добавляем кнопки сохранить/отменить вместо продолжить/назад
	saveCancelRow := h.createSaveEditsKeyboard(interfaceLang).InlineKeyboard[0]
	buttonRows = append(buttonRows, []tgbotapi.InlineKeyboardButton{saveCancelRow[0], saveCancelRow[1]})

	return tgbotapi.NewInlineKeyboardMarkup(buttonRows...)
}

// Создание клавиатуры подтверждения выбора языков.
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

// Создание клавиатуры выбора уровня владения языком.
func (h *TelegramHandler) createLanguageLevelKeyboard(interfaceLang, languageCode string) tgbotapi.InlineKeyboardMarkup {
	return h.createLanguageLevelKeyboardWithPrefix(interfaceLang, languageCode, "level_", true)
}

// createLanguageLevelKeyboardWithPrefix создает клавиатуру выбора уровня с произвольным префиксом колбэков.
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

func (h *TelegramHandler) createEditLanguagesKeyboard(interfaceLang, currentNative, currentTarget, currentLevel string) tgbotapi.InlineKeyboardMarkup {
	nativeName := h.service.Localizer.GetLanguageName(currentNative, interfaceLang)
	targetName := h.service.Localizer.GetLanguageName(currentTarget, interfaceLang)
	levelName := h.service.Localizer.Get(interfaceLang, "choose_level_"+currentLevel)

	// Определяем флаг для текущего родного языка
	var nativeFlag string
	switch currentNative {
	case "ru":
		nativeFlag = "🇷🇺"
	case "en":
		nativeFlag = "🇺🇸"
	case "es":
		nativeFlag = "🇪🇸"
	case "zh":
		nativeFlag = "🇨🇳"
	default:
		nativeFlag = "🌍"
	}

	editNative := tgbotapi.NewInlineKeyboardButtonData(
		fmt.Sprintf("%s %s: %s", nativeFlag, h.service.Localizer.Get(interfaceLang, "languages_selected_native"), nativeName),
		"edit_native_lang",
	)

	// Добавляем кнопки save/cancel
	saveCancelRow := h.createSaveEditsKeyboard(interfaceLang).InlineKeyboard[0]

	var buttons [][]tgbotapi.InlineKeyboardButton

	// Всегда добавляем кнопку редактирования родного языка
	buttons = append(buttons, []tgbotapi.InlineKeyboardButton{editNative})

	// Добавляем кнопку редактирования изучаемого языка только если родной - русский
	if currentNative == "ru" {
		editTarget := tgbotapi.NewInlineKeyboardButtonData(
			fmt.Sprintf("📚 %s: %s", h.service.Localizer.Get(interfaceLang, "languages_selected_target"), targetName),
			"edit_target_lang",
		)
		buttons = append(buttons, []tgbotapi.InlineKeyboardButton{editTarget})
	}

	// Всегда добавляем кнопку редактирования уровня владения языком
	editLevel := tgbotapi.NewInlineKeyboardButtonData(
		fmt.Sprintf("🎯 %s: %s", "Уровень языка", levelName),
		"edit_level",
	)
	buttons = append(buttons, []tgbotapi.InlineKeyboardButton{editLevel})

	// Добавляем save/cancel внизу
	buttons = append(buttons, []tgbotapi.InlineKeyboardButton{saveCancelRow[0], saveCancelRow[1]})

	return tgbotapi.NewInlineKeyboardMarkup(buttons...)
}
