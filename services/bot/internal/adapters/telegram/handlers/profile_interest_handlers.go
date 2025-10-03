package handlers

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"language-exchange-bot/internal/core"
	"language-exchange-bot/internal/errors"
	"language-exchange-bot/internal/models"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Константы для работы с callback data.
const (
	MinPartsForInterestCallback = 4 // Минимальное количество частей в callback data для интересов
)

// ProfileInterestHandler обрабатывает редактирование интересов из профиля.
type ProfileInterestHandler struct {
	service         *core.BotService
	interestService *core.InterestService
	bot             *tgbotapi.BotAPI
	keyboardBuilder *KeyboardBuilder
	errorHandler    *errors.ErrorHandler
}

// NewProfileInterestHandler создает новый обработчик интересов для профиля.
func NewProfileInterestHandler(
	service *core.BotService,
	interestService *core.InterestService,
	bot *tgbotapi.BotAPI,
	keyboardBuilder *KeyboardBuilder,
	errorHandler *errors.ErrorHandler,
) *ProfileInterestHandler {
	return &ProfileInterestHandler{
		service:         service,
		interestService: interestService,
		bot:             bot,
		keyboardBuilder: keyboardBuilder,
		errorHandler:    errorHandler,
	}
}

// HandleEditInterestsFromProfile обрабатывает редактирование интересов из профиля.
func (pih *ProfileInterestHandler) HandleEditInterestsFromProfile(callback *tgbotapi.CallbackQuery, user *models.User) error {
	// Получаем категории интересов через кэш
	categories, err := pih.interestService.GetInterestCategories()
	if err != nil {
		return pih.errorHandler.HandleTelegramError(err, callback.Message.Chat.ID, int64(user.ID), "GetInterestCategories")
	}

	// Логируем получение категорий
	log.Printf("ProfileInterestHandler: Retrieved %d interest categories for user %d", len(categories), user.ID)

	// Создаем клавиатуру с категориями
	keyboard := pih.keyboardBuilder.CreateInterestCategoriesKeyboard(user.InterfaceLanguageCode)

	// Создаем текст с инструкциями
	text := pih.service.Localizer.Get(user.InterfaceLanguageCode, "edit_interests_from_profile") + "\n\n" +
		pih.service.Localizer.Get(user.InterfaceLanguageCode, "choose_interest_category")

	// Обновляем сообщение
	editMsg := tgbotapi.NewEditMessageTextAndMarkup(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		text,
		keyboard,
	)

	_, err = pih.bot.Request(editMsg)
	if err != nil {
		return pih.errorHandler.HandleTelegramError(err, callback.Message.Chat.ID, int64(user.ID), "EditMessage")
	}

	return nil
}

// HandleEditInterestCategoryFromProfile обрабатывает выбор категории для редактирования.
func (pih *ProfileInterestHandler) HandleEditInterestCategoryFromProfile(callback *tgbotapi.CallbackQuery, user *models.User, categoryKey string) error {
	log.Printf("ProfileInterestHandler: User %d selected category '%s' for editing", user.ID, categoryKey)

	// Получаем категории
	categories, err := pih.interestService.GetInterestCategories()
	if err != nil {
		return pih.errorHandler.HandleTelegramError(err, callback.Message.Chat.ID, int64(user.ID), "GetInterestCategories")
	}

	// Находим выбранную категорию
	var selectedCategory *models.InterestCategory

	for _, category := range categories {
		if category.KeyName == categoryKey {
			selectedCategory = &category

			break
		}
	}

	if selectedCategory == nil {
		return pih.errorHandler.HandleTelegramError(err, callback.Message.Chat.ID, int64(user.ID), "CategoryNotFound")
	}

	// Получаем интересы в категории
	interests, err := pih.interestService.GetInterestsByCategory(selectedCategory.ID)
	if err != nil {
		return pih.errorHandler.HandleTelegramError(err, callback.Message.Chat.ID, int64(user.ID), "GetInterestsByCategory")
	}

	// Получаем текущие выборы пользователя
	userSelections, err := pih.interestService.GetUserInterestSelections(user.ID)
	if err != nil {
		return pih.errorHandler.HandleTelegramError(
			err,
			callback.Message.Chat.ID,
			int64(user.ID),
			"GetUserInterestSelections",
		)
	}

	// Создаем карту выбранных интересов
	selectedMap := make(map[int]bool)
	for _, selection := range userSelections {
		selectedMap[selection.InterestID] = true
	}

	// Создаем клавиатуру с интересами
	keyboard := pih.keyboardBuilder.CreateCategoryInterestsKeyboard(interests, selectedMap, categoryKey, user.InterfaceLanguageCode)

	// Создаем текст
	categoryName := pih.service.Localizer.Get(user.InterfaceLanguageCode, "category_"+categoryKey)
	text := pih.service.Localizer.Get(user.InterfaceLanguageCode, "edit_interests_in_category") + " " + categoryName

	// Обновляем сообщение
	editMsg := tgbotapi.NewEditMessageTextAndMarkup(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		text,
		keyboard,
	)

	_, err = pih.bot.Request(editMsg)
	if err != nil {
		return pih.errorHandler.HandleTelegramError(err, callback.Message.Chat.ID, int64(user.ID), "EditMessage")
	}

	return nil
}

// HandleEditInterestSelectionFromProfile обрабатывает выбор/отмену интереса при редактировании из профиля.
func (pih *ProfileInterestHandler) HandleEditInterestSelectionFromProfile(callback *tgbotapi.CallbackQuery, user *models.User, interestIDStr string) error {
	interestID, err := strconv.Atoi(interestIDStr)
	if err != nil {
		return pih.errorHandler.HandleTelegramError(err, callback.Message.Chat.ID, int64(user.ID), "ParseInterestID")
	}

	log.Printf("ProfileInterestHandler: User %d toggling interest %d", user.ID, interestID)

	// Получаем текущие выборы пользователя
	userSelections, err := pih.interestService.GetUserInterestSelections(user.ID)
	if err != nil {
		return pih.errorHandler.HandleTelegramError(err, callback.Message.Chat.ID, int64(user.ID), "GetUserInterestSelections")
	}

	// Проверяем, выбран ли уже этот интерес
	isSelected := false

	for _, selection := range userSelections {
		if selection.InterestID == interestID {
			isSelected = true

			break
		}
	}

	// Переключаем выбор
	if isSelected {
		// Удаляем выбор
		err = pih.interestService.RemoveUserInterestSelection(user.ID, interestID)
	} else {
		// Добавляем выбор
		err = pih.interestService.AddUserInterestSelection(user.ID, interestID, false)
	}

	if err != nil {
		return pih.errorHandler.HandleTelegramError(err, callback.Message.Chat.ID, int64(user.ID), "ToggleInterestSelection")
	}

	// Обновляем клавиатуру
	return pih.updateCategoryInterestsKeyboardFromProfile(callback, user, interestIDStr)
}

// HandleEditPrimaryInterestsFromProfile обрабатывает редактирование основных интересов из профиля.
func (pih *ProfileInterestHandler) HandleEditPrimaryInterestsFromProfile(callback *tgbotapi.CallbackQuery, user *models.User) error {
	// Получаем текущие выборы пользователя
	userSelections, err := pih.interestService.GetUserInterestSelections(user.ID)
	if err != nil {
		return pih.errorHandler.HandleTelegramError(err, callback.Message.Chat.ID, int64(user.ID), "GetUserInterestSelections")
	}

	// Создаем клавиатуру для выбора основных интересов
	keyboard := pih.keyboardBuilder.CreatePrimaryInterestsKeyboard(userSelections, user.InterfaceLanguageCode)

	// Создаем текст
	text := pih.service.Localizer.Get(user.InterfaceLanguageCode, "edit_primary_interests") + "\n\n" +
		pih.service.Localizer.Get(user.InterfaceLanguageCode, "choose_primary_interests")

	// Обновляем сообщение
	editMsg := tgbotapi.NewEditMessageTextAndMarkup(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		text,
		keyboard,
	)

	_, err = pih.bot.Request(editMsg)
	if err != nil {
		return pih.errorHandler.HandleTelegramError(err, callback.Message.Chat.ID, int64(user.ID), "EditMessage")
	}

	return nil
}

// HandleEditPrimaryInterestSelectionFromProfile обрабатывает выбор/отмену основного интереса.
func (pih *ProfileInterestHandler) HandleEditPrimaryInterestSelectionFromProfile(callback *tgbotapi.CallbackQuery, user *models.User, interestIDStr string) error {
	interestID, err := strconv.Atoi(interestIDStr)
	if err != nil {
		return pih.errorHandler.HandleTelegramError(
			err,
			callback.Message.Chat.ID,
			int64(user.ID),
			"ParseInterestID",
		)
	}

	// Получаем текущие выборы пользователя
	userSelections, err := pih.interestService.GetUserInterestSelections(user.ID)
	if err != nil {
		return pih.errorHandler.HandleTelegramError(err, callback.Message.Chat.ID, int64(user.ID), "GetUserInterestSelections")
	}

	// Находим текущий выбор
	var currentSelection *models.InterestSelection

	for _, selection := range userSelections {
		if selection.InterestID == interestID {
			currentSelection = &selection

			break
		}
	}

	if currentSelection == nil {
		// Интерес не выбран, ничего не делаем
		return nil
	}

	// Переключаем статус основного
	err = pih.interestService.SetPrimaryInterest(
		user.ID,
		interestID,
		!currentSelection.IsPrimary,
	)
	if err != nil {
		return pih.errorHandler.HandleTelegramError(err, callback.Message.Chat.ID, int64(user.ID), "TogglePrimaryInterest")
	}

	// Обновляем клавиатуру
	return pih.updatePrimaryInterestsKeyboardFromProfile(callback, user)
}

// HandleSaveInterestEditsFromProfile сохраняет изменения интересов и возвращается к профилю.
func (pih *ProfileInterestHandler) HandleSaveInterestEditsFromProfile(callback *tgbotapi.CallbackQuery, user *models.User) error {
	log.Printf("ProfileInterestHandler: User %d saving interest edits", user.ID)

	// Получаем сводку интересов пользователя
	summary, err := pih.interestService.GetUserInterestSummary(user.ID)
	if err != nil {
		return pih.errorHandler.HandleTelegramError(err, callback.Message.Chat.ID, int64(user.ID), "GetUserInterestSummary")
	}

	// Создаем текст с обновленными интересами
	text := pih.service.Localizer.Get(user.InterfaceLanguageCode, "interests_updated_successfully") + "\n\n" +
		fmt.Sprintf("%s: %d\n%s: %d\n%s: %d",
			pih.service.Localizer.Get(user.InterfaceLanguageCode, "total_interests"),
			summary.TotalInterests,
			pih.service.Localizer.Get(user.InterfaceLanguageCode, "primary_interests_label"),
			len(summary.PrimaryInterests),
			pih.service.Localizer.Get(user.InterfaceLanguageCode, "additional_interests_label"),
			len(summary.AdditionalInterests))

	// Создаем клавиатуру для возврата к профилю
	keyboard := pih.keyboardBuilder.CreateProfileMenuKeyboard(user.InterfaceLanguageCode)

	// Обновляем сообщение
	editMsg := tgbotapi.NewEditMessageTextAndMarkup(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		text,
		keyboard,
	)

	_, err = pih.bot.Request(editMsg)
	if err != nil {
		return pih.errorHandler.HandleTelegramError(err, callback.Message.Chat.ID, int64(user.ID), "EditMessage")
	}

	return nil
}

// updateCategoryInterestsKeyboardFromProfile обновляет клавиатуру интересов в категории.
func (pih *ProfileInterestHandler) updateCategoryInterestsKeyboardFromProfile(callback *tgbotapi.CallbackQuery, user *models.User, interestIDStr string) error {
	// Извлекаем categoryKey из callback data (предполагаем формат "edit_interest_category_<key>_<id>")
	parts := strings.Split(callback.Data, "_")
	if len(parts) < MinPartsForInterestCallback {
		log.Printf("Invalid callback data format: %s", callback.Data)

		return nil
	}

	categoryKey := parts[3] // "edit_interest_category_<key>_<id>"

	// Получаем категории
	categories, err := pih.interestService.GetInterestCategories()
	if err != nil {
		return pih.errorHandler.HandleTelegramError(err, callback.Message.Chat.ID, int64(user.ID), "GetInterestCategories")
	}

	// Находим выбранную категорию
	var selectedCategory *models.InterestCategory

	for _, category := range categories {
		if category.KeyName == categoryKey {
			selectedCategory = &category

			break
		}
	}

	if selectedCategory == nil {
		return pih.errorHandler.HandleTelegramError(err, callback.Message.Chat.ID, int64(user.ID), "CategoryNotFound")
	}

	// Получаем интересы в категории
	interests, err := pih.interestService.GetInterestsByCategory(selectedCategory.ID)
	if err != nil {
		return pih.errorHandler.HandleTelegramError(err, callback.Message.Chat.ID, int64(user.ID), "GetInterestsByCategory")
	}

	// Получаем обновленные выборы пользователя
	userSelections, err := pih.interestService.GetUserInterestSelections(user.ID)
	if err != nil {
		return pih.errorHandler.HandleTelegramError(err, callback.Message.Chat.ID, int64(user.ID), "GetUserInterestSelections")
	}

	// Создаем карту выбранных интересов
	selectedMap := make(map[int]bool)
	for _, selection := range userSelections {
		selectedMap[selection.InterestID] = true
	}

	// Создаем обновленную клавиатуру
	keyboard := pih.keyboardBuilder.CreateCategoryInterestsKeyboard(
		interests,
		selectedMap,
		categoryKey,
		user.InterfaceLanguageCode,
	)

	// Обновляем только клавиатуру
	editMsg := tgbotapi.NewEditMessageReplyMarkup(callback.Message.Chat.ID, callback.Message.MessageID, keyboard)
	_, err = pih.bot.Request(editMsg)

	return err
}

// updatePrimaryInterestsKeyboardFromProfile обновляет клавиатуру основных интересов.
func (pih *ProfileInterestHandler) updatePrimaryInterestsKeyboardFromProfile(callback *tgbotapi.CallbackQuery, user *models.User) error {
	// Получаем обновленные выборы пользователя
	userSelections, err := pih.interestService.GetUserInterestSelections(user.ID)
	if err != nil {
		return pih.errorHandler.HandleTelegramError(err, callback.Message.Chat.ID, int64(user.ID), "GetUserInterestSelections")
	}

	// Создаем обновленную клавиатуру
	keyboard := pih.keyboardBuilder.CreatePrimaryInterestsKeyboard(userSelections, user.InterfaceLanguageCode)

	// Обновляем только клавиатуру
	editMsg := tgbotapi.NewEditMessageReplyMarkup(callback.Message.Chat.ID, callback.Message.MessageID, keyboard)
	_, err = pih.bot.Request(editMsg)

	return err
}
