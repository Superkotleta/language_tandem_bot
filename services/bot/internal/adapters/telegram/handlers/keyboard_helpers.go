package handlers

import (
	"fmt"
	"sort"

	"language-exchange-bot/internal/core"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// KeyboardBuilder создает различные типы клавиатур для Telegram
type KeyboardBuilder struct {
	service *core.BotService
}

// NewKeyboardBuilder создает новый экземпляр KeyboardBuilder
func NewKeyboardBuilder(service *core.BotService) *KeyboardBuilder {
	return &KeyboardBuilder{
		service: service,
	}
}

// CreateLanguageKeyboard создает клавиатуру выбора языка
func (kb *KeyboardBuilder) CreateLanguageKeyboard(interfaceLang, keyboardType string, excludeLang string, showBackButton bool) tgbotapi.InlineKeyboardMarkup {
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

		name := kb.service.Localizer.GetLanguageName(lang.code, interfaceLang)
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
			kb.service.Localizer.Get(interfaceLang, "back_button"),
			backCallback,
		)
		buttons = append(buttons, []tgbotapi.InlineKeyboardButton{backButton})
	}

	return tgbotapi.NewInlineKeyboardMarkup(buttons...)
}

// CreateInterestsKeyboard создает клавиатуру для выбора интересов
func (kb *KeyboardBuilder) CreateInterestsKeyboard(interests map[int]string, selectedInterests []int, interfaceLang string) tgbotapi.InlineKeyboardMarkup {
	// Создаем карту для быстрого поиска выбранных интересов
	selectedMap := make(map[int]bool)
	for _, id := range selectedInterests {
		selectedMap[id] = true
	}

	// Сортируем интересы по ID для стабильного порядка
	type interestPair struct {
		id   int
		name string
	}
	var sortedInterests []interestPair
	for id, name := range interests {
		sortedInterests = append(sortedInterests, interestPair{id, name})
	}
	sort.Slice(sortedInterests, func(i, j int) bool {
		return sortedInterests[i].id < sortedInterests[j].id
	})

	var buttons [][]tgbotapi.InlineKeyboardButton
	for _, interest := range sortedInterests {
		label := interest.name
		if selectedMap[interest.id] {
			label = "✅ " + label
		}

		button := tgbotapi.NewInlineKeyboardButtonData(
			label,
			fmt.Sprintf("interest_%d", interest.id),
		)
		buttons = append(buttons, []tgbotapi.InlineKeyboardButton{button})
	}

	// Добавляем кнопку "Продолжить"
	continueButton := tgbotapi.NewInlineKeyboardButtonData(
		kb.service.Localizer.Get(interfaceLang, "interests_continue"),
		"interests_continue",
	)
	buttons = append(buttons, []tgbotapi.InlineKeyboardButton{continueButton})

	return tgbotapi.NewInlineKeyboardMarkup(buttons...)
}

// CreateMainMenuKeyboard создает главное меню
func (kb *KeyboardBuilder) CreateMainMenuKeyboard(interfaceLang string) tgbotapi.InlineKeyboardMarkup {
	viewProfile := tgbotapi.NewInlineKeyboardButtonData(
		kb.service.Localizer.Get(interfaceLang, "main_menu_view_profile"),
		"main_view_profile",
	)
	editProfile := tgbotapi.NewInlineKeyboardButtonData(
		kb.service.Localizer.Get(interfaceLang, "main_menu_edit_profile"),
		"main_edit_profile",
	)
	changeLang := tgbotapi.NewInlineKeyboardButtonData(
		kb.service.Localizer.Get(interfaceLang, "main_menu_change_lang"),
		"main_change_language",
	)
	feedback := tgbotapi.NewInlineKeyboardButtonData(
		kb.service.Localizer.Get(interfaceLang, "main_menu_feedback"),
		"main_feedback",
	)

	// Компонуем меню по 2 кнопки в ряд для лучшей организации
	buttons := [][]tgbotapi.InlineKeyboardButton{
		{viewProfile, editProfile},
		{changeLang, feedback},
	}
	return tgbotapi.NewInlineKeyboardMarkup(buttons...)
}

// CreateProfileMenuKeyboard создает меню профиля
func (kb *KeyboardBuilder) CreateProfileMenuKeyboard(interfaceLang string) tgbotapi.InlineKeyboardMarkup {
	// Кнопки для управления профилем
	editInterests := tgbotapi.NewInlineKeyboardButtonData(
		kb.service.Localizer.Get(interfaceLang, "profile_edit_interests"),
		"edit_interests",
	)
	editLanguages := tgbotapi.NewInlineKeyboardButtonData(
		kb.service.Localizer.Get(interfaceLang, "profile_edit_languages"),
		"edit_languages",
	)
	changeInterfaceLang := tgbotapi.NewInlineKeyboardButtonData(
		kb.service.Localizer.Get(interfaceLang, "main_menu_change_lang"),
		"main_change_language",
	)
	reconfig := tgbotapi.NewInlineKeyboardButtonData(
		kb.service.Localizer.Get(interfaceLang, "profile_reconfigure"),
		"profile_reset_ask",
	)
	backToMain := tgbotapi.NewInlineKeyboardButtonData(
		kb.service.Localizer.Get(interfaceLang, "back_to_main"),
		"back_to_main_menu",
	)

	// Пять рядов: интересы, языки, язык интерфейса, сброс, главное меню
	buttons := [][]tgbotapi.InlineKeyboardButton{
		{editInterests},
		{editLanguages},
		{changeInterfaceLang},
		{reconfig},
		{backToMain},
	}
	return tgbotapi.NewInlineKeyboardMarkup(buttons...)
}

// CreateResetConfirmKeyboard создает клавиатуру подтверждения сброса
func (kb *KeyboardBuilder) CreateResetConfirmKeyboard(interfaceLang string) tgbotapi.InlineKeyboardMarkup {
	yes := tgbotapi.NewInlineKeyboardButtonData(
		kb.service.Localizer.Get(interfaceLang, "profile_reset_yes"),
		"profile_reset_yes",
	)
	no := tgbotapi.NewInlineKeyboardButtonData(
		kb.service.Localizer.Get(interfaceLang, "profile_reset_no"),
		"profile_reset_no",
	)
	return tgbotapi.NewInlineKeyboardMarkup([][]tgbotapi.InlineKeyboardButton{{yes}, {no}}...)
}

// CreateLanguageLevelKeyboard создает клавиатуру для выбора уровня языка
func (kb *KeyboardBuilder) CreateLanguageLevelKeyboard(interfaceLang, targetLanguage string) tgbotapi.InlineKeyboardMarkup {
	return kb.CreateLanguageLevelKeyboardWithPrefix(interfaceLang, targetLanguage, "level_", true)
}

// CreateLanguageLevelKeyboardWithPrefix создает клавиатуру уровня языка с кастомным префиксом
func (kb *KeyboardBuilder) CreateLanguageLevelKeyboardWithPrefix(interfaceLang, targetLanguage, prefix string, showBackButton bool) tgbotapi.InlineKeyboardMarkup {
	levels := []string{"beginner", "elementary", "intermediate", "upper_intermediate"}

	var buttons [][]tgbotapi.InlineKeyboardButton
	for _, level := range levels {
		text := kb.service.Localizer.Get(interfaceLang, "choose_level_"+level)
		callback := prefix + level
		button := tgbotapi.NewInlineKeyboardButtonData(text, callback)
		buttons = append(buttons, []tgbotapi.InlineKeyboardButton{button})
	}

	if showBackButton {
		backButton := tgbotapi.NewInlineKeyboardButtonData(
			kb.service.Localizer.Get(interfaceLang, "back_button"),
			"back_to_previous_step",
		)
		buttons = append(buttons, []tgbotapi.InlineKeyboardButton{backButton})
	}

	return tgbotapi.NewInlineKeyboardMarkup(buttons...)
}

// CreateProfileCompletedKeyboard создает клавиатуру для завершенного профиля
func (kb *KeyboardBuilder) CreateProfileCompletedKeyboard(interfaceLang string) tgbotapi.InlineKeyboardMarkup {
	mainMenuButton := tgbotapi.NewInlineKeyboardButtonData(
		kb.service.Localizer.Get(interfaceLang, "main_menu_button"),
		"back_to_main_menu",
	)
	return tgbotapi.NewInlineKeyboardMarkup([]tgbotapi.InlineKeyboardButton{mainMenuButton})
}

// CreateEditInterestsKeyboard создает клавиатуру для редактирования интересов
func (kb *KeyboardBuilder) CreateEditInterestsKeyboard(interests map[int]string, selectedInterests []int, interfaceLang string) tgbotapi.InlineKeyboardMarkup {
	// Создаем карту для быстрого поиска выбранных интересов
	selectedMap := make(map[int]bool)
	for _, id := range selectedInterests {
		selectedMap[id] = true
	}

	// Сортируем интересы по ID
	type interestPair struct {
		id   int
		name string
	}
	var sortedInterests []interestPair
	for id, name := range interests {
		sortedInterests = append(sortedInterests, interestPair{id, name})
	}
	sort.Slice(sortedInterests, func(i, j int) bool {
		return sortedInterests[i].id < sortedInterests[j].id
	})

	var buttons [][]tgbotapi.InlineKeyboardButton
	for _, interest := range sortedInterests {
		label := interest.name
		if selectedMap[interest.id] {
			label = "✅ " + label
		}

		button := tgbotapi.NewInlineKeyboardButtonData(
			label,
			fmt.Sprintf("edit_interest_%d", interest.id),
		)
		buttons = append(buttons, []tgbotapi.InlineKeyboardButton{button})
	}

	// Добавляем кнопки сохранить/отменить
	saveButton := tgbotapi.NewInlineKeyboardButtonData(
		kb.service.Localizer.Get(interfaceLang, "save_button"),
		"save_edits",
	)
	cancelButton := tgbotapi.NewInlineKeyboardButtonData(
		kb.service.Localizer.Get(interfaceLang, "cancel_button"),
		"cancel_edits",
	)

	buttons = append(buttons, []tgbotapi.InlineKeyboardButton{saveButton, cancelButton})
	return tgbotapi.NewInlineKeyboardMarkup(buttons...)
}

// CreateEditLanguagesKeyboard создает клавиатуру для редактирования языков
func (kb *KeyboardBuilder) CreateEditLanguagesKeyboard(interfaceLang, nativeLang, targetLang, level string) tgbotapi.InlineKeyboardMarkup {
	var buttons [][]tgbotapi.InlineKeyboardButton

	// Родной язык
	nativeName := kb.service.Localizer.GetLanguageName(nativeLang, interfaceLang)
	nativeButton := tgbotapi.NewInlineKeyboardButtonData(
		fmt.Sprintf("🏠 %s: %s", kb.service.Localizer.Get(interfaceLang, "profile_field_native"), nativeName),
		"edit_native_lang",
	)
	buttons = append(buttons, []tgbotapi.InlineKeyboardButton{nativeButton})

	// Изучаемый язык (только если родной язык - русский)
	if nativeLang == "ru" {
		targetName := kb.service.Localizer.GetLanguageName(targetLang, interfaceLang)
		targetButton := tgbotapi.NewInlineKeyboardButtonData(
			fmt.Sprintf("📚 %s: %s", kb.service.Localizer.Get(interfaceLang, "profile_field_target"), targetName),
			"edit_target_lang",
		)
		buttons = append(buttons, []tgbotapi.InlineKeyboardButton{targetButton})
	}

	// Уровень владения языком
	if level != "" {
		levelName := kb.service.Localizer.Get(interfaceLang, "choose_level_"+level)
		levelButton := tgbotapi.NewInlineKeyboardButtonData(
			fmt.Sprintf("🎯 %s: %s", kb.service.Localizer.Get(interfaceLang, "level_label"), levelName),
			"edit_level",
		)
		buttons = append(buttons, []tgbotapi.InlineKeyboardButton{levelButton})
	}

	// Кнопки сохранить/отменить
	saveButton := tgbotapi.NewInlineKeyboardButtonData(
		kb.service.Localizer.Get(interfaceLang, "save_button"),
		"save_edits",
	)
	cancelButton := tgbotapi.NewInlineKeyboardButtonData(
		kb.service.Localizer.Get(interfaceLang, "cancel_button"),
		"cancel_edits",
	)

	buttons = append(buttons, []tgbotapi.InlineKeyboardButton{saveButton, cancelButton})
	return tgbotapi.NewInlineKeyboardMarkup(buttons...)
}

// CreateSaveEditsKeyboard создает клавиатуру для сохранения/отмены изменений
func (kb *KeyboardBuilder) CreateSaveEditsKeyboard(interfaceLang string) tgbotapi.InlineKeyboardMarkup {
	saveButton := tgbotapi.NewInlineKeyboardButtonData(
		kb.service.Localizer.Get(interfaceLang, "save_button"),
		"save_edits",
	)
	cancelButton := tgbotapi.NewInlineKeyboardButtonData(
		kb.service.Localizer.Get(interfaceLang, "cancel_button"),
		"cancel_edits",
	)

	return tgbotapi.NewInlineKeyboardMarkup([]tgbotapi.InlineKeyboardButton{saveButton, cancelButton})
}

// CreateFeedbackAdminKeyboard создает клавиатуру для управления отзывами (временная заглушка)
func (kb *KeyboardBuilder) CreateFeedbackAdminKeyboard(interfaceLang string) tgbotapi.InlineKeyboardMarkup {
	// Временная реализация, которая будет обновлена при переносе статистики
	buttons := [][]tgbotapi.InlineKeyboardButton{
		{tgbotapi.NewInlineKeyboardButtonData("🆕 Активные", "show_active_feedbacks")},
		{tgbotapi.NewInlineKeyboardButtonData("📚 Архив", "show_archive_feedbacks")},
		{tgbotapi.NewInlineKeyboardButtonData("📋 Все", "show_all_feedbacks")},
	}
	return tgbotapi.NewInlineKeyboardMarkup(buttons...)
}
