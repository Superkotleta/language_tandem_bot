package handlers

import (
	"fmt"
	"sort"
	"strconv"

	"language-exchange-bot/internal/localization"
	"language-exchange-bot/internal/models"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// createEditMainMenuKeyboard создает клавиатуру главного меню редактирования.
func (e *IsolatedInterestEditor) createEditMainMenuKeyboard(interfaceLang string, stats EditStats) tgbotapi.InlineKeyboardMarkup {
	var buttonRows [][]tgbotapi.InlineKeyboardButton

	// Основные действия
	mainRow := []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData(
			"🎯 "+e.service.Localizer.Get(interfaceLang, "edit_interests_by_category"),
			"isolated_edit_categories",
		),
		tgbotapi.NewInlineKeyboardButtonData(
			localization.SymbolStar+e.service.Localizer.Get(interfaceLang, "edit_primary_interests"),
			"isolated_edit_primary",
		),
	}
	buttonRows = append(buttonRows, mainRow)

	// Управление
	controlRow := []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData(
			"💾 "+e.service.Localizer.Get(interfaceLang, "save_changes"),
			"isolated_save_changes",
		),
		tgbotapi.NewInlineKeyboardButtonData(
			"❌ "+e.service.Localizer.Get(interfaceLang, "cancel_edit"),
			"isolated_cancel_edit",
		),
	}
	buttonRows = append(buttonRows, controlRow)

	return tgbotapi.NewInlineKeyboardMarkup(buttonRows...)
}

// createEditCategoriesKeyboard создает клавиатуру категорий для редактирования.
func (e *IsolatedInterestEditor) createEditCategoriesKeyboard(categories []models.InterestCategory, session *EditSession, interfaceLang string) tgbotapi.InlineKeyboardMarkup {
	var buttonRows [][]tgbotapi.InlineKeyboardButton

	// Сортируем категории по display_order
	sort.Slice(categories, func(i, j int) bool {
		return categories[i].DisplayOrder < categories[j].DisplayOrder
	})

	// Создаем кнопки категорий (по 2 в ряд)
	for i := 0; i < len(categories); i += 2 {
		var row []tgbotapi.InlineKeyboardButton

		// Первая кнопка в ряду
		category1 := categories[i]
		categoryName1 := e.service.Localizer.Get(interfaceLang, "category_"+category1.KeyName)

		// Добавляем индикатор прогресса
		progress1 := e.getCategoryProgress(session, category1.KeyName)
		buttonText1 := fmt.Sprintf("%s %s", categoryName1, progress1)

		button1 := tgbotapi.NewInlineKeyboardButtonData(
			buttonText1,
			"isolated_edit_category_"+category1.KeyName,
		)
		row = append(row, button1)

		// Вторая кнопка в ряду (если есть)
		if i+1 < len(categories) {
			category2 := categories[i+1]
			categoryName2 := e.service.Localizer.Get(interfaceLang, "category_"+category2.KeyName)

			progress2 := e.getCategoryProgress(session, category2.KeyName)
			buttonText2 := fmt.Sprintf("%s %s", categoryName2, progress2)

			button2 := tgbotapi.NewInlineKeyboardButtonData(
				buttonText2,
				"isolated_edit_category_"+category2.KeyName,
			)
			row = append(row, button2)
		}

		buttonRows = append(buttonRows, row)
	}

	// Навигация
	navRow := []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData(
			"🏠 "+e.service.Localizer.Get(interfaceLang, "back_to_edit_menu"),
			"isolated_main_menu",
		),
		tgbotapi.NewInlineKeyboardButtonData(
			localization.SymbolStar+e.service.Localizer.Get(interfaceLang, "edit_primary_interests"),
			"isolated_edit_primary",
		),
	}
	buttonRows = append(buttonRows, navRow)

	return tgbotapi.NewInlineKeyboardMarkup(buttonRows...)
}

// createEditCategoryInterestsKeyboard создает клавиатуру интересов в категории для редактирования.
func (e *IsolatedInterestEditor) createEditCategoryInterestsKeyboard(interests []models.Interest, selectedMap map[int]bool, categoryKey, interfaceLang string) tgbotapi.InlineKeyboardMarkup {
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
		interestName1 := e.service.Localizer.Get(interfaceLang, "interest_"+interest1.KeyName)

		prefix1 := localization.SymbolUnchecked
		if selectedMap[interest1.ID] {
			prefix1 = localization.SymbolChecked
		}

		button1 := tgbotapi.NewInlineKeyboardButtonData(
			prefix1+interestName1,
			"isolated_toggle_interest_"+strconv.Itoa(interest1.ID),
		)
		row = append(row, button1)

		// Вторая кнопка в ряду (если есть)
		if i+1 < len(interests) {
			interest2 := interests[i+1]
			interestName2 := e.service.Localizer.Get(interfaceLang, "interest_"+interest2.KeyName)

			prefix2 := localization.SymbolUnchecked
			if selectedMap[interest2.ID] {
				prefix2 = localization.SymbolChecked
			}

			button2 := tgbotapi.NewInlineKeyboardButtonData(
				prefix2+interestName2,
				"isolated_toggle_interest_"+strconv.Itoa(interest2.ID),
			)
			row = append(row, button2)
		}

		buttonRows = append(buttonRows, row)
	}

	// Массовые операции
	if len(interests) > 0 {
		massRow := []tgbotapi.InlineKeyboardButton{
			tgbotapi.NewInlineKeyboardButtonData(
				localization.SymbolChecked+e.service.Localizer.Get(interfaceLang, "select_all_in_category"),
				"isolated_select_all_"+categoryKey,
			),
			tgbotapi.NewInlineKeyboardButtonData(
				"❌ "+e.service.Localizer.Get(interfaceLang, "clear_all_in_category"),
				"isolated_clear_all_"+categoryKey,
			),
		}
		buttonRows = append(buttonRows, massRow)
	}

	// Навигация
	navRow := []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData(
			"⬅️ "+e.service.Localizer.Get(interfaceLang, "back_to_categories"),
			"isolated_edit_categories",
		),
		tgbotapi.NewInlineKeyboardButtonData(
			"🏠 "+e.service.Localizer.Get(interfaceLang, "back_to_edit_menu"),
			"isolated_main_menu",
		),
	}
	buttonRows = append(buttonRows, navRow)

	return tgbotapi.NewInlineKeyboardMarkup(buttonRows...)
}

// createChangesPreviewKeyboard создает клавиатуру для предварительного просмотра изменений.
func (e *IsolatedInterestEditor) createChangesPreviewKeyboard(interfaceLang string) tgbotapi.InlineKeyboardMarkup {
	var buttonRows [][]tgbotapi.InlineKeyboardButton

	// Основные действия
	mainRow := []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData(
			"💾 "+e.service.Localizer.Get(interfaceLang, "save_changes"),
			"isolated_save_changes",
		),
		tgbotapi.NewInlineKeyboardButtonData(
			"↩️ "+e.service.Localizer.Get(interfaceLang, "undo_last_change"),
			"isolated_undo_last",
		),
	}
	buttonRows = append(buttonRows, mainRow)

	// Навигация
	navRow := []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData(
			"🏠 "+e.service.Localizer.Get(interfaceLang, "back_to_edit_menu"),
			"isolated_main_menu",
		),
		tgbotapi.NewInlineKeyboardButtonData(
			"❌ "+e.service.Localizer.Get(interfaceLang, "cancel_edit"),
			"isolated_cancel_edit",
		),
	}
	buttonRows = append(buttonRows, navRow)

	return tgbotapi.NewInlineKeyboardMarkup(buttonRows...)
}

// createEditPrimaryInterestsKeyboard создает клавиатуру для редактирования основных интересов.
func (e *IsolatedInterestEditor) createEditPrimaryInterestsKeyboard(selections []models.InterestSelection, interfaceLang string) tgbotapi.InlineKeyboardMarkup {
	var buttonRows [][]tgbotapi.InlineKeyboardButton

	// Сортируем только по ID для стабильного порядка
	sort.Slice(selections, func(i, j int) bool {
		return selections[i].InterestID < selections[j].InterestID
	})

	// Создаем кнопки для каждого выбранного интереса
	for i := 0; i < len(selections); i += 2 {
		var row []tgbotapi.InlineKeyboardButton

		// Первая кнопка в ряду
		selection1 := selections[i]

		interest1, err := e.interestService.GetInterestByID(selection1.InterestID)
		if err != nil {
			continue
		}

		interestName1 := e.service.Localizer.Get(interfaceLang, "interest_"+interest1.KeyName)

		prefix1 := localization.SymbolUnchecked
		if selection1.IsPrimary {
			prefix1 = localization.SymbolStar
		}

		button1 := tgbotapi.NewInlineKeyboardButtonData(
			prefix1+interestName1,
			"isolated_toggle_primary_"+strconv.Itoa(selection1.InterestID),
		)
		row = append(row, button1)

		// Вторая кнопка в ряду (если есть)
		if i+1 < len(selections) {
			selection2 := selections[i+1]

			interest2, err := e.interestService.GetInterestByID(selection2.InterestID)
			if err != nil {
				continue
			}

			interestName2 := e.service.Localizer.Get(interfaceLang, "interest_"+interest2.KeyName)

			prefix2 := localization.SymbolUnchecked
			if selection2.IsPrimary {
				prefix2 = localization.SymbolStar
			}

			button2 := tgbotapi.NewInlineKeyboardButtonData(
				prefix2+interestName2,
				"isolated_toggle_primary_"+strconv.Itoa(selection2.InterestID),
			)
			row = append(row, button2)
		}

		buttonRows = append(buttonRows, row)
	}

	// Навигация
	navRow := []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData(
			"🏠 "+e.service.Localizer.Get(interfaceLang, "back_to_edit_menu"),
			"isolated_main_menu",
		),
		tgbotapi.NewInlineKeyboardButtonData(
			"🎯 "+e.service.Localizer.Get(interfaceLang, "edit_interests_by_category"),
			"isolated_edit_categories",
		),
	}
	buttonRows = append(buttonRows, navRow)

	return tgbotapi.NewInlineKeyboardMarkup(buttonRows...)
}

// Вспомогательные методы

func (e *IsolatedInterestEditor) getCategoryProgress(session *EditSession, categoryKey string) string {
	// Подсчитываем количество выбранных интересов в категории
	count := 0
	for range session.CurrentSelections {
		// TODO: Добавить проверку категории интереса
		count++
	}

	if count == 0 {
		return localization.SymbolEmpty
	} else if count < localization.ProgressDisplayThreshold {
		return "◐"
	} else {
		return "◉"
	}
}
