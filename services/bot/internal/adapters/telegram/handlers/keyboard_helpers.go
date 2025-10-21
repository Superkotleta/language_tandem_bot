package handlers

import (
	"fmt"
	"sort"
	"strconv"

	"language-exchange-bot/internal/core"
	"language-exchange-bot/internal/localization"
	"language-exchange-bot/internal/models"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Keyboard symbols are now defined in localization/constants.go

// Константы для callback команд.
const (
	CallbackBackToMainMenu     = "back_to_main_menu"
	CallbackBackToPreviousStep = "back_to_previous_step"
)

// TemporaryInterestSelection представляет временный выбор интереса пользователем.
type TemporaryInterestSelection struct {
	InterestID     int
	IsPrimary      bool
	SelectionOrder int
}

// KeyboardBuilder создает различные типы клавиатур для Telegram.
type KeyboardBuilder struct {
	service *core.BotService
}

// NewKeyboardBuilder создает новый экземпляр KeyboardBuilder.
func NewKeyboardBuilder(service *core.BotService) *KeyboardBuilder {
	return &KeyboardBuilder{
		service: service,
	}
}

// CreateLanguageKeyboard создает клавиатуру выбора языка.
func (kb *KeyboardBuilder) CreateLanguageKeyboard(
	interfaceLang,
	keyboardType string,
	excludeLang string,
	showBackButton bool,
) tgbotapi.InlineKeyboardMarkup {
	// Получаем языки из кэша или БД
	languages, err := kb.service.GetCachedLanguages(interfaceLang)
	if err != nil {
		// Fallback на хардкод если кэш не работает
		languages = []*models.Language{
			{ID: localization.LanguageIDEnglish, Code: "en", NameNative: "English", NameEn: "English"},
			{ID: localization.LanguageIDRussian, Code: "ru", NameNative: "Русский", NameEn: "Russian"},
			{ID: localization.LanguageIDSpanish, Code: "es", NameNative: "Español", NameEn: "Spanish"},
			{ID: localization.LanguageIDChinese, Code: "zh", NameNative: "中文", NameEn: "Chinese"},
		}
	}

	// Используем Map для автоматического удаления дубликатов
	uniqueButtons := make(map[string]tgbotapi.InlineKeyboardButton)

	for _, lang := range languages {
		// Исключаем указанный язык из списка
		if lang.Code == excludeLang {
			continue
		}

		// Получаем флаг для языка
		flag := getLanguageFlag(lang.Code)
		name := kb.service.Localizer.GetLanguageName(lang.Code, interfaceLang)
		label := fmt.Sprintf("%s %s", flag, name)
		callbackData := fmt.Sprintf("lang_%s_%s", keyboardType, lang.Code)

		// Избегаем дубликатов по callback data (на случай если название языка совпадает)
		if _, exists := uniqueButtons[callbackData]; !exists {
			button := tgbotapi.NewInlineKeyboardButtonData(label, callbackData)
			uniqueButtons[callbackData] = button
		}
	}

	// Преобразуем map в массив кнопок
	buttons := make([][]tgbotapi.InlineKeyboardButton, 0, len(uniqueButtons))
	for _, button := range uniqueButtons {
		buttons = append(buttons, []tgbotapi.InlineKeyboardButton{button})
	}

	// Добавляем кнопку "Назад", если это необходимо
	if showBackButton {
		var backCallback string
		if keyboardType == "interface" || keyboardType == "native" {
			backCallback = CallbackBackToMainMenu
		} else {
			backCallback = CallbackBackToPreviousStep
		}

		backButton := tgbotapi.NewInlineKeyboardButtonData(
			kb.service.Localizer.Get(interfaceLang, "back_button"),
			backCallback,
		)
		buttons = append(buttons, []tgbotapi.InlineKeyboardButton{backButton})
	}

	return tgbotapi.NewInlineKeyboardMarkup(buttons...)
}

// CreateInterestsKeyboard создает клавиатуру для выбора интересов.
func (kb *KeyboardBuilder) CreateInterestsKeyboard(
	interests map[int]string,
	selectedInterests []int,
	interfaceLang string,
) tgbotapi.InlineKeyboardMarkup {
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

	sortedInterests := make([]interestPair, 0, len(interests))
	for id, name := range interests {
		sortedInterests = append(sortedInterests, interestPair{id, name})
	}

	sort.Slice(sortedInterests, func(i, j int) bool {
		return sortedInterests[i].id < sortedInterests[j].id
	})

	buttons := make([][]tgbotapi.InlineKeyboardButton, 0, len(sortedInterests))

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

// CreateMainMenuKeyboard создает главное меню.
func (kb *KeyboardBuilder) CreateMainMenuKeyboard(interfaceLang string, hasProfile bool) tgbotapi.InlineKeyboardMarkup {
	changeLang := tgbotapi.NewInlineKeyboardButtonData(
		kb.service.Localizer.Get(interfaceLang, "main_menu_change_lang"),
		"main_change_language",
	)
	feedback := tgbotapi.NewInlineKeyboardButtonData(
		kb.service.Localizer.Get(interfaceLang, "main_menu_feedback"),
		"main_feedback",
	)

	var buttons [][]tgbotapi.InlineKeyboardButton

	if hasProfile {
		// У пользователя есть профиль - показываем обычные кнопки
		viewProfile := tgbotapi.NewInlineKeyboardButtonData(
			kb.service.Localizer.Get(interfaceLang, "main_menu_view_profile"),
			"main_view_profile",
		)
		editProfile := tgbotapi.NewInlineKeyboardButtonData(
			kb.service.Localizer.Get(interfaceLang, "main_menu_edit_profile"),
			"main_edit_profile",
		)
		buttons = [][]tgbotapi.InlineKeyboardButton{
			{viewProfile, editProfile},
			{changeLang, feedback},
		}
	} else {
		// У пользователя нет профиля - показываем большую кнопку "Создать профиль"
		createProfile := tgbotapi.NewInlineKeyboardButtonData(
			kb.service.Localizer.Get(interfaceLang, "main_menu_create_profile"),
			"start_profile_setup",
		)
		buttons = [][]tgbotapi.InlineKeyboardButton{
			{createProfile},
			{changeLang, feedback},
		}
	}

	return tgbotapi.NewInlineKeyboardMarkup(buttons...)
}

// CreateProfileMenuKeyboard создает меню профиля.
func (kb *KeyboardBuilder) CreateProfileMenuKeyboard(interfaceLang string) tgbotapi.InlineKeyboardMarkup {
	// Кнопки для управления профилем
	editInterestsIsolated := tgbotapi.NewInlineKeyboardButtonData(
		kb.service.Localizer.Get(interfaceLang, "profile_edit_interests_isolated"),
		"isolated_edit_start",
	)
	editLanguages := tgbotapi.NewInlineKeyboardButtonData(
		kb.service.Localizer.Get(interfaceLang, "profile_edit_languages"),
		"edit_languages",
	)
	changeInterfaceLang := tgbotapi.NewInlineKeyboardButtonData(
		kb.service.Localizer.Get(interfaceLang, "main_menu_change_lang"),
		"main_change_language",
	)
	editAvailability := tgbotapi.NewInlineKeyboardButtonData(
		"⏰ "+kb.service.Localizer.Get(interfaceLang, "edit_availability"),
		"edit_availability",
	)
	reconfig := tgbotapi.NewInlineKeyboardButtonData(
		kb.service.Localizer.Get(interfaceLang, "profile_reconfigure"),
		"profile_reset_ask",
	)
	backToMain := tgbotapi.NewInlineKeyboardButtonData(
		kb.service.Localizer.Get(interfaceLang, "back_to_main"),
		"back_to_main_menu",
	)

	// Группировка кнопок для лучшего UX:
	// Ряд 1: Интересы и доступность
	// Ряд 2: Языки и интерфейс
	// Ряд 3: Сброс профиля
	// Ряд 4: Главное меню
	buttons := [][]tgbotapi.InlineKeyboardButton{
		{editInterestsIsolated, editAvailability},
		{editLanguages, changeInterfaceLang},
		{reconfig},
		{backToMain},
	}

	return tgbotapi.NewInlineKeyboardMarkup(buttons...)
}

// CreateResetConfirmKeyboard создает клавиатуру подтверждения сброса.
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

// CreateLanguageLevelKeyboardWithPrefix создает клавиатуру уровня языка с кастомным префиксом.
func (kb *KeyboardBuilder) CreateLanguageLevelKeyboardWithPrefix(interfaceLang, targetLanguage, prefix string, showBackButton bool) tgbotapi.InlineKeyboardMarkup {
	levels := []string{"beginner", "elementary", "intermediate", "upper_intermediate"}

	buttons := make([][]tgbotapi.InlineKeyboardButton, 0, len(levels))

	for _, level := range levels {
		text := kb.service.Localizer.Get(interfaceLang, "choose_level_"+level)
		callback := prefix + level
		button := tgbotapi.NewInlineKeyboardButtonData(text, callback)
		buttons = append(buttons, []tgbotapi.InlineKeyboardButton{button})
	}

	if showBackButton {
		backButton := tgbotapi.NewInlineKeyboardButtonData(
			kb.service.Localizer.Get(interfaceLang, "back_button"),
			CallbackBackToPreviousStep,
		)
		buttons = append(buttons, []tgbotapi.InlineKeyboardButton{backButton})
	}

	return tgbotapi.NewInlineKeyboardMarkup(buttons...)
}

// CreateProfileCompletedKeyboard создает клавиатуру для завершенного профиля.
func (kb *KeyboardBuilder) CreateProfileCompletedKeyboard(interfaceLang string) tgbotapi.InlineKeyboardMarkup {
	viewProfileButton := tgbotapi.NewInlineKeyboardButtonData(
		kb.service.Localizer.Get(interfaceLang, "profile_completed_view"),
		"profile_show",
	)
	mainMenuButton := tgbotapi.NewInlineKeyboardButtonData(
		kb.service.Localizer.Get(interfaceLang, "profile_completed_main"),
		"back_to_main_menu",
	)

	buttons := [][]tgbotapi.InlineKeyboardButton{
		{viewProfileButton, mainMenuButton},
	}

	return tgbotapi.NewInlineKeyboardMarkup(buttons...)
}

// CreateAvailabilitySetupKeyboard создает клавиатуру для перехода к настройке доступности.
func (kb *KeyboardBuilder) CreateAvailabilitySetupKeyboard(interfaceLang string) tgbotapi.InlineKeyboardMarkup {
	setupAvailabilityButton := tgbotapi.NewInlineKeyboardButtonData(
		"⏰ "+kb.service.Localizer.Get(interfaceLang, "setup_availability_button"),
		"setup_availability",
	)

	viewProfileButton := tgbotapi.NewInlineKeyboardButtonData(
		kb.service.Localizer.Get(interfaceLang, "profile_show"),
		"profile_show",
	)

	mainMenuButton := tgbotapi.NewInlineKeyboardButtonData(
		kb.service.Localizer.Get(interfaceLang, "profile_completed_main"),
		"back_to_main_menu",
	)

	buttons := [][]tgbotapi.InlineKeyboardButton{
		{setupAvailabilityButton},
		{viewProfileButton, mainMenuButton},
	}

	return tgbotapi.NewInlineKeyboardMarkup(buttons...)
}

// CreateEditLanguagesKeyboard создает клавиатуру для редактирования языков.
// DEPRECATED: Use IsolatedLanguageEditor instead (isolated_language_editor.go).
// This method is kept for backward compatibility only and will be removed in a future version.
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

// CreateSaveEditsKeyboard создает клавиатуру для сохранения/отмены изменений.
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

// CreateFeedbackAdminKeyboard создает клавиатуру для управления отзывами (временная заглушка).
func (kb *KeyboardBuilder) CreateFeedbackAdminKeyboard(interfaceLang string) tgbotapi.InlineKeyboardMarkup {
	// Временная реализация, которая будет обновлена при переносе статистики
	buttons := [][]tgbotapi.InlineKeyboardButton{
		{tgbotapi.NewInlineKeyboardButtonData("🆕 Активные", "show_active_feedbacks")},
		{tgbotapi.NewInlineKeyboardButtonData("📚 Архив", "show_archive_feedbacks")},
		{tgbotapi.NewInlineKeyboardButtonData("📋 Все", "show_all_feedbacks")},
	}

	return tgbotapi.NewInlineKeyboardMarkup(buttons...)
}

// getLanguageFlag возвращает флаг для языка.
func getLanguageFlag(langCode string) string {
	switch langCode {
	case "ru":
		return "🇷🇺"
	case "en":
		return "🇺🇸"
	case "es":
		return "🇪🇸"
	case "zh":
		return "🇨🇳"
	default:
		return "🌍"
	}
}

// CreateInterestCategoriesKeyboard создает клавиатуру для выбора категорий интересов.
func (kb *KeyboardBuilder) CreateInterestCategoriesKeyboard(interfaceLang string) tgbotapi.InlineKeyboardMarkup {
	categories := []struct {
		key  string
		icon string
	}{
		{"entertainment", "🎬"},
		{"education", "📚"},
		{"active", "⚽"},
		{"creative", "🎨"},
		{"social", "👥"},
	}

	var buttonRows [][]tgbotapi.InlineKeyboardButton

	// Создаем кнопки категорий (по 2 в ряд)
	for i := 0; i < len(categories); i += 2 {
		var row []tgbotapi.InlineKeyboardButton

		// Первая кнопка в ряду
		categoryName := kb.service.Localizer.Get(interfaceLang, "category_"+categories[i].key)
		button1 := tgbotapi.NewInlineKeyboardButtonData(
			categoryName,
			"interest_category_"+categories[i].key,
		)
		row = append(row, button1)

		// Вторая кнопка в ряду (если есть)
		if i+1 < len(categories) {
			categoryName2 := kb.service.Localizer.Get(interfaceLang, "category_"+categories[i+1].key)
			button2 := tgbotapi.NewInlineKeyboardButtonData(
				categoryName2,
				"interest_category_"+categories[i+1].key,
			)
			row = append(row, button2)
		}

		buttonRows = append(buttonRows, row)
	}

	// Добавляем кнопки управления
	controlRow := []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData(
			kb.service.Localizer.Get(interfaceLang, "continue_button"),
			"interests_continue",
		),
		tgbotapi.NewInlineKeyboardButtonData(
			kb.service.Localizer.Get(interfaceLang, "back_button"),
			"back_to_language_level",
		),
	}
	buttonRows = append(buttonRows, controlRow)

	return tgbotapi.NewInlineKeyboardMarkup(buttonRows...)
}

// CreateCategoryInterestsKeyboard создает клавиатуру для выбора интересов в категории.
func (kb *KeyboardBuilder) CreateCategoryInterestsKeyboard(interests []models.Interest, selectedMap map[int]bool, categoryKey, interfaceLang string) tgbotapi.InlineKeyboardMarkup {
	var buttonRows [][]tgbotapi.InlineKeyboardButton

	// Сортируем интересы по display_order
	sort.Slice(interests, func(i, j int) bool {
		return interests[i].DisplayOrder < interests[j].DisplayOrder
	})

	// Создаем кнопки интересов (по 2 в ряд)
	for i := 0; i < len(interests); i += 2 {
		var row []tgbotapi.InlineKeyboardButton

		// Первая кнопка в ряду
		interest1 := interests[i]
		interestName1 := kb.service.Localizer.Get(interfaceLang, "interest_"+interest1.KeyName)

		prefix1 := localization.SymbolUnchecked
		if selectedMap[interest1.ID] {
			prefix1 = localization.SymbolChecked
		}

		button1 := tgbotapi.NewInlineKeyboardButtonData(
			prefix1+interestName1,
			"edit_interest_select_"+strconv.Itoa(interest1.ID),
		)
		row = append(row, button1)

		// Вторая кнопка в ряду (если есть)
		if i+1 < len(interests) {
			interest2 := interests[i+1]
			interestName2 := kb.service.Localizer.Get(interfaceLang, "interest_"+interest2.KeyName)

			prefix2 := localization.SymbolUnchecked
			if selectedMap[interest2.ID] {
				prefix2 = "✅ "
			}

			button2 := tgbotapi.NewInlineKeyboardButtonData(
				prefix2+interestName2,
				"edit_interest_select_"+strconv.Itoa(interest2.ID),
			)
			row = append(row, button2)
		}

		buttonRows = append(buttonRows, row)
	}

	// Добавляем кнопку "К категориям" для возврата к редактированию интересов
	controlRow := []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData(
			kb.service.Localizer.Get(interfaceLang, "to_categories_button"),
			"back_to_categories",
		),
	}
	buttonRows = append(buttonRows, controlRow)

	return tgbotapi.NewInlineKeyboardMarkup(buttonRows...)
}

// CreatePrimaryInterestsKeyboard создает клавиатуру для выбора основных интересов.
//
//nolint:funlen
func (kb *KeyboardBuilder) CreatePrimaryInterestsKeyboard(selections interface{}, interfaceLang string) tgbotapi.InlineKeyboardMarkup {
	var buttonRows [][]tgbotapi.InlineKeyboardButton

	// Получаем локализованные названия интересов
	localizedInterests, err := kb.service.GetCachedInterests(interfaceLang)
	if err != nil {
		// Fallback - создаем пустую карту
		localizedInterests = make(map[int]string)
	}

	// Приводим к правильному типу
	var tempSelections []TemporaryInterestSelection
	if tempSelectionsInterface, ok := selections.([]TemporaryInterestSelection); ok {
		tempSelections = tempSelectionsInterface
	} else if modelsSelections, ok := selections.([]models.InterestSelection); ok {
		// Конвертируем из models.InterestSelection в TemporaryInterestSelection
		for _, sel := range modelsSelections {
			tempSelections = append(tempSelections, TemporaryInterestSelection{
				InterestID:     sel.InterestID,
				IsPrimary:      sel.IsPrimary,
				SelectionOrder: sel.SelectionOrder,
			})
		}
	}

	// Сортируем выборы по порядку выбора
	sort.Slice(tempSelections, func(i, j int) bool {
		return tempSelections[i].SelectionOrder < tempSelections[j].SelectionOrder
	})

	// Создаем кнопки для каждого выбранного интереса
	for i := 0; i < len(tempSelections); i += 2 {
		var row []tgbotapi.InlineKeyboardButton

		// Первая кнопка в ряду
		selection1 := tempSelections[i]
		// Получаем локализованное название интереса
		interestName1, exists := localizedInterests[selection1.InterestID]
		if !exists {
			// Fallback: пытаемся получить через getInterestName
			var err error

			interestName1, err = kb.getInterestName(selection1.InterestID, interfaceLang)
			if err != nil {
				interestName1 = fmt.Sprintf("Интерес %d", selection1.InterestID)
			}
		} else {
			// Проверяем, что это не просто key_name (английское название)
			// Если название совпадает с key_name, пытаемся найти перевод
			interest, err := kb.service.DB.GetInterestByID(selection1.InterestID)
			if err == nil {
				// Всегда пытаемся найти перевод в JSON файлах
				translatedName := kb.service.Localizer.Get(interfaceLang, "interest_"+interest.KeyName)
				if translatedName != "interest_"+interest.KeyName {
					interestName1 = translatedName
				}
			}
		}

		prefix1 := localization.SymbolUnchecked
		if selection1.IsPrimary {
			prefix1 = localization.SymbolStar
		}

		button1 := tgbotapi.NewInlineKeyboardButtonData(
			prefix1+interestName1,
			"primary_interest_"+strconv.Itoa(selection1.InterestID),
		)
		row = append(row, button1)

		// Вторая кнопка в ряду (если есть)
		if i+1 < len(tempSelections) {
			selection2 := tempSelections[i+1]

			// Получаем локализованное название интереса
			interestName2, exists := localizedInterests[selection2.InterestID]
			if !exists {
				// Fallback: пытаемся получить через getInterestName
				var err error

				interestName2, err = kb.getInterestName(selection2.InterestID, interfaceLang)
				if err != nil {
					interestName2 = fmt.Sprintf("Интерес %d", selection2.InterestID)
				}
			} else {
				// Проверяем, что это не просто key_name (английское название)
				// Если название совпадает с key_name, пытаемся найти перевод
				interest, err := kb.service.DB.GetInterestByID(selection2.InterestID)
				if err == nil {
					// Всегда пытаемся найти перевод в JSON файлах
					translatedName := kb.service.Localizer.Get(interfaceLang, "interest_"+interest.KeyName)
					if translatedName != "interest_"+interest.KeyName {
						interestName2 = translatedName
					}
				}
			}

			prefix2 := localization.SymbolUnchecked
			if selection2.IsPrimary {
				prefix2 = "⭐ "
			}

			button2 := tgbotapi.NewInlineKeyboardButtonData(
				prefix2+interestName2,
				"primary_interest_"+strconv.Itoa(selection2.InterestID),
			)
			row = append(row, button2)
		}

		buttonRows = append(buttonRows, row)
	}

	// Добавляем кнопки управления
	controlRow := []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData(
			kb.service.Localizer.Get(interfaceLang, "continue_button"),
			"primary_interests_continue",
		),
		tgbotapi.NewInlineKeyboardButtonData(
			kb.service.Localizer.Get(interfaceLang, "back_button"),
			"back_to_interests",
		),
	}
	buttonRows = append(buttonRows, controlRow)

	return tgbotapi.NewInlineKeyboardMarkup(buttonRows...)
}

// CreateEditInterestCategoriesKeyboard создает клавиатуру для выбора категорий в режиме редактирования.
func (kb *KeyboardBuilder) CreateEditInterestCategoriesKeyboard(interfaceLang string) tgbotapi.InlineKeyboardMarkup {
	categories := []struct {
		key  string
		icon string
	}{
		{"entertainment", "🎬"},
		{"education", "📚"},
		{"active", "⚽"},
		{"creative", "🎨"},
		{"social", "👥"},
	}

	var buttonRows [][]tgbotapi.InlineKeyboardButton

	// Создаем кнопки категорий (по 2 в ряд)
	for i := 0; i < len(categories); i += 2 {
		var row []tgbotapi.InlineKeyboardButton

		// Первая кнопка в ряду
		categoryName := kb.service.Localizer.Get(interfaceLang, "category_"+categories[i].key)
		button1 := tgbotapi.NewInlineKeyboardButtonData(
			categoryName,
			"edit_interest_category_"+categories[i].key,
		)
		row = append(row, button1)

		// Вторая кнопка в ряду (если есть)
		if i+1 < len(categories) {
			categoryName2 := kb.service.Localizer.Get(interfaceLang, "category_"+categories[i+1].key)
			button2 := tgbotapi.NewInlineKeyboardButtonData(
				categoryName2,
				"edit_interest_category_"+categories[i+1].key,
			)
			row = append(row, button2)
		}

		buttonRows = append(buttonRows, row)
	}

	// Добавляем кнопки управления для режима редактирования
	controlRow := []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData(
			kb.service.Localizer.Get(interfaceLang, "continue_button"),
			"interests_continue",
		),
		tgbotapi.NewInlineKeyboardButtonData(
			kb.service.Localizer.Get(interfaceLang, "cancel_button"),
			"back_to_profile",
		),
	}
	buttonRows = append(buttonRows, controlRow)

	return tgbotapi.NewInlineKeyboardMarkup(buttonRows...)
}

// CreateEditPrimaryInterestsKeyboard создает клавиатуру для выбора основных интересов в режиме редактирования.
func (kb *KeyboardBuilder) CreateEditPrimaryInterestsKeyboard(selections interface{}, interfaceLang string) tgbotapi.InlineKeyboardMarkup {
	var buttonRows [][]tgbotapi.InlineKeyboardButton

	// Получаем локализованные названия интересов
	localizedInterests, err := kb.service.GetCachedInterests(interfaceLang)
	if err != nil {
		// Fallback - создаем пустую карту
		localizedInterests = make(map[int]string)
	}

	// Конвертируем selections в нужный формат
	var tempSelections []struct {
		InterestID int
		IsPrimary  bool
	}

	switch s := selections.(type) {
	case []models.InterestSelection:
		for _, selection := range s {
			tempSelections = append(tempSelections, struct {
				InterestID int
				IsPrimary  bool
			}{
				InterestID: selection.InterestID,
				IsPrimary:  selection.IsPrimary,
			})
		}
	}

	// Сортируем: сначала основные, потом дополнительные
	sort.Slice(tempSelections, func(i, j int) bool {
		if tempSelections[i].IsPrimary != tempSelections[j].IsPrimary {
			return tempSelections[i].IsPrimary
		}

		return tempSelections[i].InterestID < tempSelections[j].InterestID
	})

	// Создаем кнопки для каждого выбранного интереса
	for i := 0; i < len(tempSelections); i += 2 {
		var row []tgbotapi.InlineKeyboardButton

		// Первая кнопка в ряду
		selection1 := tempSelections[i]
		// Получаем локализованное название интереса
		interestName1, exists := localizedInterests[selection1.InterestID]
		if !exists {
			// Fallback: пытаемся получить через getInterestName
			var err error

			interestName1, err = kb.getInterestName(selection1.InterestID, interfaceLang)
			if err != nil {
				interestName1 = fmt.Sprintf("Интерес %d", selection1.InterestID)
			}
		} else {
			// Проверяем, что это не просто key_name (английское название)
			// Если название совпадает с key_name, пытаемся найти перевод
			interest, err := kb.service.DB.GetInterestByID(selection1.InterestID)
			if err == nil {
				// Всегда пытаемся найти перевод в JSON файлах
				translatedName := kb.service.Localizer.Get(interfaceLang, "interest_"+interest.KeyName)
				if translatedName != "interest_"+interest.KeyName {
					interestName1 = translatedName
				}
			}
		}

		prefix1 := localization.SymbolUnchecked
		if selection1.IsPrimary {
			prefix1 = localization.SymbolStar
		}

		button1 := tgbotapi.NewInlineKeyboardButtonData(
			prefix1+interestName1,
			"edit_primary_interest_"+strconv.Itoa(selection1.InterestID),
		)
		row = append(row, button1)

		// Вторая кнопка в ряду (если есть)
		if i+1 < len(tempSelections) {
			selection2 := tempSelections[i+1]

			// Получаем локализованное название интереса
			interestName2, exists := localizedInterests[selection2.InterestID]
			if !exists {
				// Fallback: пытаемся получить через getInterestName
				var err error

				interestName2, err = kb.getInterestName(selection2.InterestID, interfaceLang)
				if err != nil {
					interestName2 = fmt.Sprintf("Интерес %d", selection2.InterestID)
				}
			} else {
				// Проверяем, что это не просто key_name (английское название)
				// Если название совпадает с key_name, пытаемся найти перевод
				interest, err := kb.service.DB.GetInterestByID(selection2.InterestID)
				if err == nil {
					// Всегда пытаемся найти перевод в JSON файлах
					translatedName := kb.service.Localizer.Get(interfaceLang, "interest_"+interest.KeyName)
					if translatedName != "interest_"+interest.KeyName {
						interestName2 = translatedName
					}
				}
			}

			prefix2 := localization.SymbolUnchecked
			if selection2.IsPrimary {
				prefix2 = "⭐ "
			}

			button2 := tgbotapi.NewInlineKeyboardButtonData(
				prefix2+interestName2,
				"edit_primary_interest_"+strconv.Itoa(selection2.InterestID),
			)
			row = append(row, button2)
		}

		buttonRows = append(buttonRows, row)
	}

	// Добавляем кнопки управления для режима редактирования
	controlRow := []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData(
			kb.service.Localizer.Get(interfaceLang, "to_categories_button"),
			"back_to_categories",
		),
		tgbotapi.NewInlineKeyboardButtonData(
			kb.service.Localizer.Get(interfaceLang, "cancel_button"),
			"back_to_profile",
		),
		tgbotapi.NewInlineKeyboardButtonData(
			kb.service.Localizer.Get(interfaceLang, "save_button"),
			"save_interest_edits",
		),
	}
	buttonRows = append(buttonRows, controlRow)

	return tgbotapi.NewInlineKeyboardMarkup(buttonRows...)
}

// CreateEditCategoryInterestsKeyboard создает клавиатуру для редактирования интересов в категории.
func (kb *KeyboardBuilder) CreateEditCategoryInterestsKeyboard(interests []models.Interest, selectedMap map[int]bool, categoryKey, interfaceLang string) tgbotapi.InlineKeyboardMarkup {
	var buttonRows [][]tgbotapi.InlineKeyboardButton

	// Сортируем интересы по display_order
	sort.Slice(interests, func(i, j int) bool {
		return interests[i].DisplayOrder < interests[j].DisplayOrder
	})

	// Создаем кнопки интересов (по 2 в ряд)
	for i := 0; i < len(interests); i += 2 {
		var row []tgbotapi.InlineKeyboardButton

		// Первая кнопка в ряду
		interest1 := interests[i]
		interestName1 := kb.service.Localizer.Get(interfaceLang, "interest_"+interest1.KeyName)

		prefix1 := localization.SymbolUnchecked
		if selectedMap[interest1.ID] {
			prefix1 = localization.SymbolChecked
		}

		button1 := tgbotapi.NewInlineKeyboardButtonData(
			prefix1+interestName1,
			"edit_interest_select_"+strconv.Itoa(interest1.ID),
		)
		row = append(row, button1)

		// Вторая кнопка в ряду (если есть)
		if i+1 < len(interests) {
			interest2 := interests[i+1]
			interestName2 := kb.service.Localizer.Get(interfaceLang, "interest_"+interest2.KeyName)

			prefix2 := localization.SymbolUnchecked
			if selectedMap[interest2.ID] {
				prefix2 = "✅ "
			}

			button2 := tgbotapi.NewInlineKeyboardButtonData(
				prefix2+interestName2,
				"edit_interest_select_"+strconv.Itoa(interest2.ID),
			)
			row = append(row, button2)
		}

		buttonRows = append(buttonRows, row)
	}

	// Добавляем только кнопку "Назад" для возврата к редактированию интересов
	controlRow := []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData(
			kb.service.Localizer.Get(interfaceLang, "back_button"),
			"back_to_edit_categories",
		),
	}
	buttonRows = append(buttonRows, controlRow)

	return tgbotapi.NewInlineKeyboardMarkup(buttonRows...)
}

// getInterestName получает название интереса из базы данных.
func (kb *KeyboardBuilder) getInterestName(interestID int, interfaceLang string) (string, error) {
	// Получаем все интересы из кэша
	interests, err := kb.service.GetCachedInterests(interfaceLang)
	if err != nil {
		return "", err
	}

	// Ищем интерес по ID
	if name, exists := interests[interestID]; exists {
		return name, nil
	}

	// Fallback: пытаемся получить из базы данных напрямую
	interest, err := kb.service.DB.GetInterestByID(interestID)
	if err != nil {
		return "", err
	}

	// Пытаемся получить локализованное название через ключ
	interestName := kb.service.Localizer.Get(interfaceLang, "interest_"+interest.KeyName)
	if interestName != "interest_"+interest.KeyName {
		return interestName, nil
	}

	// Последний fallback - используем key_name из базы данных
	return interest.KeyName, nil
}

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
