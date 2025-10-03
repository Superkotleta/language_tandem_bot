package telegram

import (
	"fmt"
	"sort"
	"strconv"

	"language-exchange-bot/internal/core"
	"language-exchange-bot/internal/models"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Константы для символов.
const (
	SymbolUnchecked = "☐ "
)

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
			"back_to_main_menu",
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

		prefix1 := SymbolUnchecked
		if selectedMap[interest1.ID] {
			prefix1 = "✅ "
		}

		button1 := tgbotapi.NewInlineKeyboardButtonData(
			prefix1+interestName1,
			"interest_select_"+strconv.Itoa(interest1.ID),
		)
		row = append(row, button1)

		// Вторая кнопка в ряду (если есть)
		if i+1 < len(interests) {
			interest2 := interests[i+1]
			interestName2 := kb.service.Localizer.Get(interfaceLang, "interest_"+interest2.KeyName)

			prefix2 := SymbolUnchecked
			if selectedMap[interest2.ID] {
				prefix2 = "✅ "
			}

			button2 := tgbotapi.NewInlineKeyboardButtonData(
				prefix2+interestName2,
				"interest_select_"+strconv.Itoa(interest2.ID),
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
			"back_to_categories",
		),
	}
	buttonRows = append(buttonRows, controlRow)

	return tgbotapi.NewInlineKeyboardMarkup(buttonRows...)
}

// CreatePrimaryInterestsKeyboard создает клавиатуру для выбора основных интересов.
func (kb *KeyboardBuilder) CreatePrimaryInterestsKeyboard(
	selections interface{},
	interfaceLang string,
) tgbotapi.InlineKeyboardMarkup {
	var buttonRows [][]tgbotapi.InlineKeyboardButton

	// Приводим к правильному типу
	var tempSelections []models.InterestSelection
	if modelsSelections, ok := selections.([]models.InterestSelection); ok {
		tempSelections = modelsSelections
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
		// Получаем название интереса (упрощенно, в реальности нужно загружать из БД)
		interestName1 := fmt.Sprintf("Интерес %d", selection1.InterestID)

		prefix1 := SymbolUnchecked
		if selection1.IsPrimary {
			prefix1 = "⭐ "
		}

		button1 := tgbotapi.NewInlineKeyboardButtonData(
			prefix1+interestName1,
			"primary_interest_"+strconv.Itoa(selection1.InterestID),
		)
		row = append(row, button1)

		// Вторая кнопка в ряду (если есть)
		if i+1 < len(tempSelections) {
			selection2 := tempSelections[i+1]
			interestName2 := fmt.Sprintf("Интерес %d", selection2.InterestID)

			prefix2 := SymbolUnchecked
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

// CreateProfileCompletedKeyboard создает клавиатуру после завершения настройки профиля.
func (kb *KeyboardBuilder) CreateProfileCompletedKeyboard(interfaceLang string) tgbotapi.InlineKeyboardMarkup {
	mainMenu := tgbotapi.NewInlineKeyboardButtonData(
		kb.service.Localizer.Get(interfaceLang, "main_menu_title"),
		"back_to_main_menu",
	)
	viewProfile := tgbotapi.NewInlineKeyboardButtonData(
		kb.service.Localizer.Get(interfaceLang, "profile_show"),
		"profile_show",
	)
	buttons := [][]tgbotapi.InlineKeyboardButton{
		{mainMenu, viewProfile},
	}

	return tgbotapi.NewInlineKeyboardMarkup(buttons...)
}
