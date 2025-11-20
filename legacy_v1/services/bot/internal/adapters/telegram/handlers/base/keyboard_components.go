package base

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// =============================================================================
// STANDARD BUTTON HELPERS
// =============================================================================
// Эти методы создают стандартные кнопки, которые используются по всему приложению.
// Преимущества:
// 1. Единая точка изменений - если нужно поменять текст/стиль кнопки
// 2. Автоматическая локализация - не нужно помнить ключи локализации
// 3. Меньше дублирования кода
// 4. Защита от опечаток в callback данных

// CreateBackButton создает кнопку "Назад" с указанным callback.
// Используется для навигации назад в многошаговых формах.
//
// Параметры:
//   - lang: код языка интерфейса (ru, en, es, zh)
//   - callbackData: callback данные для кнопки (например: "availability_back_to_days")
//
// Пример:
//
//	backBtn := kb.CreateBackButton("ru", "availability_back_to_daytype")
func (kb *KeyboardBuilder) CreateBackButton(lang, callbackData string) tgbotapi.InlineKeyboardButton {
	return tgbotapi.NewInlineKeyboardButtonData(
		kb.service.Localizer.Get(lang, "back_button"),
		callbackData,
	)
}

// CreateContinueButton создает кнопку "Продолжить" с указанным callback.
// Используется для перехода к следующему шагу в многошаговых формах.
//
// Параметры:
//   - lang: код языка интерфейса
//   - callbackData: callback данные для кнопки (например: "availability_proceed_to_time")
//
// Пример:
//
//	continueBtn := kb.CreateContinueButton("en", "availability_proceed_to_communication")
func (kb *KeyboardBuilder) CreateContinueButton(lang, callbackData string) tgbotapi.InlineKeyboardButton {
	return tgbotapi.NewInlineKeyboardButtonData(
		kb.service.Localizer.Get(lang, "continue_button"),
		callbackData,
	)
}

// CreateBackToMainButton создает кнопку "В главное меню".
// Всегда ведет в главное меню (callback: "back_to_main_menu").
//
// Параметры:
//   - lang: код языка интерфейса
//
// Пример:
//
//	mainMenuBtn := kb.CreateBackToMainButton("ru")
func (kb *KeyboardBuilder) CreateBackToMainButton(lang string) tgbotapi.InlineKeyboardButton {
	return tgbotapi.NewInlineKeyboardButtonData(
		kb.service.Localizer.Get(lang, "back_to_main"),
		"back_to_main_menu",
	)
}

// CreateViewProfileButton создает кнопку "Посмотреть профиль".
// Всегда ведет к просмотру профиля (callback: "view_profile").
//
// Параметры:
//   - lang: код языка интерфейса
//
// Пример:
//
//	profileBtn := kb.CreateViewProfileButton("en")
func (kb *KeyboardBuilder) CreateViewProfileButton(lang string) tgbotapi.InlineKeyboardButton {
	return tgbotapi.NewInlineKeyboardButtonData(
		kb.service.Localizer.Get(lang, "profile_show"),
		"view_profile",
	)
}

// CreateSaveButton создает кнопку "Сохранить изменения".
// Используется для подтверждения сохранения изменений.
//
// Параметры:
//   - lang: код языка интерфейса
//
// Пример:
//
//	saveBtn := kb.CreateSaveButton("ru")
func (kb *KeyboardBuilder) CreateSaveButton(lang string) tgbotapi.InlineKeyboardButton {
	return tgbotapi.NewInlineKeyboardButtonData(
		kb.service.Localizer.Get(lang, "save_changes"),
		"save_changes",
	)
}

// CreateCancelButton создает кнопку "Отменить".
// Используется для отмены текущей операции.
//
// Параметры:
//   - lang: код языка интерфейса
//
// Пример:
//
//	cancelBtn := kb.CreateCancelButton("en")
func (kb *KeyboardBuilder) CreateCancelButton(lang string) tgbotapi.InlineKeyboardButton {
	return tgbotapi.NewInlineKeyboardButtonData(
		kb.service.Localizer.Get(lang, "cancel_edit"),
		"cancel_edit",
	)
}

// CreateUndoButton создает кнопку "Отменить последнее изменение".
// Используется для отката последнего действия.
//
// Параметры:
//   - lang: код языка интерфейса
//
// Пример:
//
//	undoBtn := kb.CreateUndoButton("es")
func (kb *KeyboardBuilder) CreateUndoButton(lang string) tgbotapi.InlineKeyboardButton {
	return tgbotapi.NewInlineKeyboardButtonData(
		kb.service.Localizer.Get(lang, "undo_last_change"),
		"undo_last_change",
	)
}

// CreateYesButton создает кнопку "Да".
// Используется в диалогах подтверждения.
//
// Параметры:
//   - lang: код языка интерфейса
//   - callbackData: callback данные для кнопки
//
// Пример:
//
//	yesBtn := kb.CreateYesButton("ru", "confirm_delete")
func (kb *KeyboardBuilder) CreateYesButton(lang, callbackData string) tgbotapi.InlineKeyboardButton {
	return tgbotapi.NewInlineKeyboardButtonData(
		kb.service.Localizer.Get(lang, "yes_button"),
		callbackData,
	)
}

// CreateNoButton создает кнопку "Нет".
// Используется в диалогах подтверждения.
//
// Параметры:
//   - lang: код языка интерфейса
//   - callbackData: callback данные для кнопки
//
// Пример:
//
//	noBtn := kb.CreateNoButton("en", "cancel_delete")
func (kb *KeyboardBuilder) CreateNoButton(lang, callbackData string) tgbotapi.InlineKeyboardButton {
	return tgbotapi.NewInlineKeyboardButtonData(
		kb.service.Localizer.Get(lang, "no_button"),
		callbackData,
	)
}

// CreateNavigationRow создает строку с кнопками "Назад" и "Продолжить".
// Это самый частый паттерн в многошаговых формах.
//
// Параметры:
//   - lang: код языка интерфейса
//   - backCallback: callback для кнопки "Назад"
//   - continueCallback: callback для кнопки "Продолжить"
//
// Пример:
//
//	navRow := kb.CreateNavigationRow("ru", "back_to_days", "proceed_to_time")
//	keyboard := tgbotapi.NewInlineKeyboardMarkup(navRow)
func (kb *KeyboardBuilder) CreateNavigationRow(lang, backCallback, continueCallback string) []tgbotapi.InlineKeyboardButton {
	return []tgbotapi.InlineKeyboardButton{
		kb.CreateBackButton(lang, backCallback),
		kb.CreateContinueButton(lang, continueCallback),
	}
}

// CreateProfileActionsRow создает строку с кнопками "Посмотреть профиль" и "В главное меню".
// Используется после завершения настройки профиля или других операций.
//
// Параметры:
//   - lang: код языка интерфейса
//
// Пример:
//
//	actionsRow := kb.CreateProfileActionsRow("en")
//	keyboard := tgbotapi.NewInlineKeyboardMarkup(actionsRow)
func (kb *KeyboardBuilder) CreateProfileActionsRow(lang string) []tgbotapi.InlineKeyboardButton {
	return []tgbotapi.InlineKeyboardButton{
		kb.CreateViewProfileButton(lang),
		kb.CreateBackToMainButton(lang),
	}
}

// =============================================================================
// COMMON KEYBOARD PATTERNS
// =============================================================================
// Эти методы создают часто используемые комбинации клавиатур.
// Преимущества:
// 1. Стандартизация интерфейса - одинаковый вид во всем приложении
// 2. Быстрое создание - одна строка вместо нескольких
// 3. Автоматическая локализация всех кнопок

// CreateSaveCancelKeyboard создает клавиатуру "Сохранить изменения" + "Отменить".
// Используется в формах редактирования для подтверждения или отмены изменений.
//
// Параметры:
//   - lang: код языка интерфейса
//
// Пример:
//
//	keyboard := kb.CreateSaveCancelKeyboard("ru")
//	editMsg := tgbotapi.NewEditMessageTextAndMarkup(chatID, messageID, text, keyboard)
func (kb *KeyboardBuilder) CreateSaveCancelKeyboard(lang string) tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		[]tgbotapi.InlineKeyboardButton{
			kb.CreateSaveButton(lang),
			kb.CreateCancelButton(lang),
		},
	)
}

// CreateSaveUndoKeyboard создает клавиатуру "Сохранить изменения" + "Отменить последнее".
// Используется в пошаговых редакторах для сохранения или отката последнего изменения.
//
// Параметры:
//   - lang: код языка интерфейса
//
// Пример:
//
//	keyboard := kb.CreateSaveUndoKeyboard("en")
func (kb *KeyboardBuilder) CreateSaveUndoKeyboard(lang string) tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		[]tgbotapi.InlineKeyboardButton{
			kb.CreateSaveButton(lang),
			kb.CreateUndoButton(lang),
		},
	)
}

// CreateConfirmationKeyboard создает клавиатуру подтверждения с кнопками "Да" и "Нет".
// Используется для вопросов требующих подтверждения пользователя.
//
// Параметры:
//   - lang: код языка интерфейса
//   - yesCallback: callback data для кнопки "Да"
//   - noCallback: callback data для кнопки "Нет"
//
// Пример:
//
//	keyboard := kb.CreateConfirmationKeyboard("ru", "confirm_delete", "cancel_delete")
func (kb *KeyboardBuilder) CreateConfirmationKeyboard(lang, yesCallback, noCallback string) tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		[]tgbotapi.InlineKeyboardButton{
			kb.CreateYesButton(lang, yesCallback),
			kb.CreateNoButton(lang, noCallback),
		},
	)
}

// CreateBackOnlyKeyboard создает клавиатуру с одной кнопкой "Назад".
// Используется когда нужно предоставить только возможность вернуться.
//
// Параметры:
//   - lang: код языка интерфейса
//   - backCallback: callback data для кнопки "Назад"
//
// Пример:
//
//	keyboard := kb.CreateBackOnlyKeyboard("es", "back_to_menu")
func (kb *KeyboardBuilder) CreateBackOnlyKeyboard(lang, backCallback string) tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		[]tgbotapi.InlineKeyboardButton{
			kb.CreateBackButton(lang, backCallback),
		},
	)
}

// CreateMenuNavigationKeyboard создает клавиатуру с кнопками "В главное меню" и "Назад".
// Стандартная навигация для большинства экранов.
//
// Параметры:
//   - lang: код языка интерфейса
//   - backCallback: callback data для кнопки "Назад" (может быть пустой строкой)
//
// Пример:
//
//	keyboard := kb.CreateMenuNavigationKeyboard("ru", "back_to_profile")
func (kb *KeyboardBuilder) CreateMenuNavigationKeyboard(lang, backCallback string) tgbotapi.InlineKeyboardMarkup {
	var buttons []tgbotapi.InlineKeyboardButton

	if backCallback != "" {
		buttons = append(buttons, kb.CreateBackButton(lang, backCallback))
	}
	buttons = append(buttons, kb.CreateBackToMainButton(lang))

	return tgbotapi.NewInlineKeyboardMarkup(buttons)
}
