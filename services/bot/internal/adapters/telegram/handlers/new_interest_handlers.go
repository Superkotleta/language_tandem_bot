package handlers

import (
	"log"
	"strconv"
	"strings"

	"language-exchange-bot/internal/core"
	"language-exchange-bot/internal/errors"
	"language-exchange-bot/internal/models"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// NewInterestHandler интерфейс для новой системы интересов
type NewInterestHandler interface {
	HandleInterestCategorySelection(callback *tgbotapi.CallbackQuery, user *models.User, categoryKey string) error
	HandleInterestSelection(callback *tgbotapi.CallbackQuery, user *models.User, interestIDStr string) error
	HandlePrimaryInterestSelection(callback *tgbotapi.CallbackQuery, user *models.User, interestIDStr string) error
	HandleInterestsContinue(callback *tgbotapi.CallbackQuery, user *models.User) error
	HandlePrimaryInterestsContinue(callback *tgbotapi.CallbackQuery, user *models.User) error
	HandleBackToCategories(callback *tgbotapi.CallbackQuery, user *models.User) error
	HandleBackToInterests(callback *tgbotapi.CallbackQuery, user *models.User) error
}

// NewInterestHandlerImpl реализация нового обработчика интересов
type NewInterestHandlerImpl struct {
	service         *core.BotService
	interestService *core.InterestService
	bot             *tgbotapi.BotAPI
	keyboardBuilder *KeyboardBuilder
	errorHandler    *errors.ErrorHandler
}

// NewNewInterestHandler создает новый обработчик интересов
func NewNewInterestHandler(service *core.BotService, interestService *core.InterestService, bot *tgbotapi.BotAPI, keyboardBuilder *KeyboardBuilder, errorHandler *errors.ErrorHandler) NewInterestHandler {
	return &NewInterestHandlerImpl{
		service:         service,
		interestService: interestService,
		bot:             bot,
		keyboardBuilder: keyboardBuilder,
		errorHandler:    errorHandler,
	}
}

// HandleInterestCategorySelection обрабатывает выбор категории интересов
func (h *NewInterestHandlerImpl) HandleInterestCategorySelection(callback *tgbotapi.CallbackQuery, user *models.User, categoryKey string) error {
	// Получаем категории
	categories, err := h.interestService.GetInterestCategories()
	if err != nil {
		return h.errorHandler.HandleTelegramError(err, callback.Message.Chat.ID, int64(user.ID), "GetInterestCategories")
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
		return h.errorHandler.HandleTelegramError(err, callback.Message.Chat.ID, int64(user.ID), "CategoryNotFound")
	}

	// Получаем интересы в категории
	interests, err := h.interestService.GetInterestsByCategory(selectedCategory.ID)
	if err != nil {
		return h.errorHandler.HandleTelegramError(err, callback.Message.Chat.ID, int64(user.ID), "GetInterestsByCategory")
	}

	// Получаем уже выбранные интересы пользователя
	userSelections, err := h.interestService.GetUserInterestSelections(user.ID)
	if err != nil {
		return h.errorHandler.HandleTelegramError(err, callback.Message.Chat.ID, int64(user.ID), "GetUserInterestSelections")
	}

	// Создаем карту выбранных интересов
	selectedMap := make(map[int]bool)
	for _, selection := range userSelections {
		selectedMap[selection.InterestID] = true
	}

	// Создаем клавиатуру для интересов в категории
	keyboard := h.keyboardBuilder.CreateCategoryInterestsKeyboard(interests, selectedMap, selectedCategory.KeyName, user.InterfaceLanguageCode)

	// Создаем текст сообщения
	categoryName := h.service.Localizer.Get(user.InterfaceLanguageCode, "category_"+categoryKey)
	messageText := categoryName + " - " + h.service.Localizer.Get(user.InterfaceLanguageCode, "choose_interests")

	// Обновляем сообщение
	editMsg := tgbotapi.NewEditMessageTextAndMarkup(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		messageText,
		keyboard,
	)

	_, err = h.bot.Request(editMsg)
	if err != nil {
		return h.errorHandler.HandleTelegramError(err, callback.Message.Chat.ID, int64(user.ID), "EditMessage")
	}

	return nil
}

// HandleInterestSelection обрабатывает выбор интереса в категории
func (h *NewInterestHandlerImpl) HandleInterestSelection(callback *tgbotapi.CallbackQuery, user *models.User, interestIDStr string) error {
	interestID, err := strconv.Atoi(interestIDStr)
	if err != nil {
		return h.errorHandler.HandleTelegramError(err, callback.Message.Chat.ID, int64(user.ID), "ParseInterestID")
	}

	// Получаем текущие выборы пользователя
	userSelections, err := h.interestService.GetUserInterestSelections(user.ID)
	if err != nil {
		return h.errorHandler.HandleTelegramError(err, callback.Message.Chat.ID, int64(user.ID), "GetUserInterestSelections")
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
		err = h.interestService.RemoveUserInterestSelection(user.ID, interestID)
	} else {
		// Добавляем выбор
		err = h.interestService.AddUserInterestSelection(user.ID, interestID, false)
	}

	if err != nil {
		return h.errorHandler.HandleTelegramError(err, callback.Message.Chat.ID, int64(user.ID), "ToggleInterestSelection")
	}

	// Обновляем клавиатуру
	return h.updateCategoryInterestsKeyboard(callback, user, interestIDStr)
}

// HandlePrimaryInterestSelection обрабатывает выбор основного интереса
func (h *NewInterestHandlerImpl) HandlePrimaryInterestSelection(callback *tgbotapi.CallbackQuery, user *models.User, interestIDStr string) error {
	interestID, err := strconv.Atoi(interestIDStr)
	if err != nil {
		return h.errorHandler.HandleTelegramError(err, callback.Message.Chat.ID, int64(user.ID), "ParseInterestID")
	}

	// Получаем текущие выборы пользователя
	userSelections, err := h.interestService.GetUserInterestSelections(user.ID)
	if err != nil {
		return h.errorHandler.HandleTelegramError(err, callback.Message.Chat.ID, int64(user.ID), "GetUserInterestSelections")
	}

	// Находим выбор для этого интереса
	var currentSelection *models.InterestSelection
	for _, selection := range userSelections {
		if selection.InterestID == interestID {
			currentSelection = &selection
			break
		}
	}

	if currentSelection == nil {
		// Интерес не выбран, сначала добавляем его
		err = h.interestService.AddUserInterestSelection(user.ID, interestID, true)
	} else {
		// Переключаем статус основного
		err = h.interestService.SetPrimaryInterest(user.ID, interestID, !currentSelection.IsPrimary)
	}

	if err != nil {
		return h.errorHandler.HandleTelegramError(err, callback.Message.Chat.ID, int64(user.ID), "TogglePrimaryInterest")
	}

	// Обновляем клавиатуру выбора основных интересов
	return h.updatePrimaryInterestsKeyboard(callback, user)
}

// HandleInterestsContinue обрабатывает продолжение после выбора интересов
func (h *NewInterestHandlerImpl) HandleInterestsContinue(callback *tgbotapi.CallbackQuery, user *models.User) error {
	// Получаем выбранные интересы
	userSelections, err := h.interestService.GetUserInterestSelections(user.ID)
	if err != nil {
		return h.errorHandler.HandleTelegramError(err, callback.Message.Chat.ID, int64(user.ID), "GetUserInterestSelections")
	}

	// Проверяем, выбраны ли интересы
	if len(userSelections) == 0 {
		warningMsg := "❗ " + h.service.Localizer.Get(user.InterfaceLanguageCode, "choose_at_least_one_interest")
		if warningMsg == "choose_at_least_one_interest" {
			warningMsg = "❗ Пожалуйста, выберите хотя бы один интерес"
		}

		// Показываем предупреждение и возвращаем к категориям
		keyboard := h.keyboardBuilder.CreateInterestCategoriesKeyboard(user.InterfaceLanguageCode)
		editMsg := tgbotapi.NewEditMessageTextAndMarkup(
			callback.Message.Chat.ID,
			callback.Message.MessageID,
			warningMsg+"\n\n"+h.service.Localizer.Get(user.InterfaceLanguageCode, "choose_interests"),
			keyboard,
		)
		_, err = h.bot.Request(editMsg)
		return err
	}

	// Переходим к выбору основных интересов
	return h.showPrimaryInterestsSelection(callback, user)
}

// HandlePrimaryInterestsContinue обрабатывает завершение выбора основных интересов
func (h *NewInterestHandlerImpl) HandlePrimaryInterestsContinue(callback *tgbotapi.CallbackQuery, user *models.User) error {
	// Получаем выборы пользователя
	userSelections, err := h.interestService.GetUserInterestSelections(user.ID)
	if err != nil {
		return h.errorHandler.HandleTelegramError(err, callback.Message.Chat.ID, int64(user.ID), "GetUserInterestSelections")
	}

	// Подсчитываем основные интересы
	primaryCount := 0
	for _, selection := range userSelections {
		if selection.IsPrimary {
			primaryCount++
		}
	}

	// Получаем конфигурацию лимитов
	limits, err := h.interestService.GetInterestLimitsConfig()
	if err != nil {
		return h.errorHandler.HandleTelegramError(err, callback.Message.Chat.ID, int64(user.ID), "GetInterestLimitsConfig")
	}

	// Проверяем минимальное количество основных интересов
	if primaryCount < limits.MinPrimaryInterests {
		warningMsg := h.service.Localizer.Get(user.InterfaceLanguageCode, "choose_at_least_primary_interests")
		if warningMsg == "choose_at_least_primary_interests" {
			warningMsg = "❗ Пожалуйста, выберите минимум " + strconv.Itoa(limits.MinPrimaryInterests) + " основных интереса"
		}

		// Показываем предупреждение
		editMsg := tgbotapi.NewEditMessageTextAndMarkup(
			callback.Message.Chat.ID,
			callback.Message.MessageID,
			warningMsg+"\n\n"+h.service.Localizer.Get(user.InterfaceLanguageCode, "choose_primary_interests"),
			h.keyboardBuilder.CreatePrimaryInterestsKeyboard(userSelections, user.InterfaceLanguageCode),
		)
		_, err = h.bot.Request(editMsg)
		return err
	}

	// Завершаем настройку профиля
	return h.completeProfileSetup(callback, user)
}

// HandleBackToCategories возвращает к выбору категорий
func (h *NewInterestHandlerImpl) HandleBackToCategories(callback *tgbotapi.CallbackQuery, user *models.User) error {
	keyboard := h.keyboardBuilder.CreateInterestCategoriesKeyboard(user.InterfaceLanguageCode)
	messageText := h.service.Localizer.Get(user.InterfaceLanguageCode, "choose_interests")

	editMsg := tgbotapi.NewEditMessageTextAndMarkup(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		messageText,
		keyboard,
	)

	_, err := h.bot.Request(editMsg)
	return err
}

// HandleBackToInterests возвращает к выбору интересов
func (h *NewInterestHandlerImpl) HandleBackToInterests(callback *tgbotapi.CallbackQuery, user *models.User) error {
	// Показываем выбор основных интересов
	return h.showPrimaryInterestsSelection(callback, user)
}

// showPrimaryInterestsSelection показывает интерфейс выбора основных интересов
func (h *NewInterestHandlerImpl) showPrimaryInterestsSelection(callback *tgbotapi.CallbackQuery, user *models.User) error {
	// Получаем выборы пользователя
	userSelections, err := h.interestService.GetUserInterestSelections(user.ID)
	if err != nil {
		return h.errorHandler.HandleTelegramError(err, callback.Message.Chat.ID, int64(user.ID), "GetUserInterestSelections")
	}

	// Создаем клавиатуру для выбора основных интересов
	keyboard := h.keyboardBuilder.CreatePrimaryInterestsKeyboard(userSelections, user.InterfaceLanguageCode)

	// Создаем текст сообщения
	messageText := h.service.Localizer.Get(user.InterfaceLanguageCode, "choose_primary_interests")

	// Обновляем сообщение
	editMsg := tgbotapi.NewEditMessageTextAndMarkup(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		messageText,
		keyboard,
	)

	_, err = h.bot.Request(editMsg)
	return err
}

// updateCategoryInterestsKeyboard обновляет клавиатуру интересов в категории
func (h *NewInterestHandlerImpl) updateCategoryInterestsKeyboard(callback *tgbotapi.CallbackQuery, user *models.User, categoryKey string) error {
	// Получаем категории
	categories, err := h.interestService.GetInterestCategories()
	if err != nil {
		return err
	}

	// Находим категорию
	var selectedCategory *models.InterestCategory
	for _, category := range categories {
		if category.KeyName == categoryKey {
			selectedCategory = &category
			break
		}
	}

	if selectedCategory == nil {
		return err
	}

	// Получаем интересы в категории
	interests, err := h.interestService.GetInterestsByCategory(selectedCategory.ID)
	if err != nil {
		return err
	}

	// Получаем выборы пользователя
	userSelections, err := h.interestService.GetUserInterestSelections(user.ID)
	if err != nil {
		return err
	}

	// Создаем карту выбранных интересов
	selectedMap := make(map[int]bool)
	for _, selection := range userSelections {
		selectedMap[selection.InterestID] = true
	}

	// Создаем клавиатуру
	keyboard := h.keyboardBuilder.CreateCategoryInterestsKeyboard(interests, selectedMap, categoryKey, user.InterfaceLanguageCode)

	// Обновляем сообщение
	categoryName := h.service.Localizer.Get(user.InterfaceLanguageCode, "category_"+categoryKey)
	messageText := categoryName + " - " + h.service.Localizer.Get(user.InterfaceLanguageCode, "choose_interests")

	editMsg := tgbotapi.NewEditMessageTextAndMarkup(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		messageText,
		keyboard,
	)

	_, err = h.bot.Request(editMsg)
	return err
}

// updatePrimaryInterestsKeyboard обновляет клавиатуру выбора основных интересов
func (h *NewInterestHandlerImpl) updatePrimaryInterestsKeyboard(callback *tgbotapi.CallbackQuery, user *models.User) error {
	// Получаем выборы пользователя
	userSelections, err := h.interestService.GetUserInterestSelections(user.ID)
	if err != nil {
		return err
	}

	// Создаем клавиатуру
	keyboard := h.keyboardBuilder.CreatePrimaryInterestsKeyboard(userSelections, user.InterfaceLanguageCode)

	// Обновляем сообщение
	messageText := h.service.Localizer.Get(user.InterfaceLanguageCode, "choose_primary_interests")

	editMsg := tgbotapi.NewEditMessageTextAndMarkup(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		messageText,
		keyboard,
	)

	_, err = h.bot.Request(editMsg)
	return err
}

// completeProfileSetup завершает настройку профиля
func (h *NewInterestHandlerImpl) completeProfileSetup(callback *tgbotapi.CallbackQuery, user *models.User) error {
	// Получаем сводку интересов пользователя
	summary, err := h.interestService.GetUserInterestSummary(user.ID)
	if err != nil {
		return h.errorHandler.HandleTelegramError(err, callback.Message.Chat.ID, int64(user.ID), "GetUserInterestSummary")
	}

	// Создаем текст с основными и дополнительными интересами
	var primaryText, additionalText strings.Builder

	if len(summary.PrimaryInterests) > 0 {
		primaryText.WriteString("⭐ Основные: ")
		for i, interest := range summary.PrimaryInterests {
			if i > 0 {
				primaryText.WriteString(", ")
			}
			interestName := h.service.Localizer.Get(user.InterfaceLanguageCode, "interest_"+interest.KeyName)
			primaryText.WriteString(interestName)
		}
		primaryText.WriteString("\n")
	}

	if len(summary.AdditionalInterests) > 0 {
		additionalText.WriteString("➕ Дополнительные: ")
		for i, interest := range summary.AdditionalInterests {
			if i > 0 {
				additionalText.WriteString(", ")
			}
			interestName := h.service.Localizer.Get(user.InterfaceLanguageCode, "interest_"+interest.KeyName)
			additionalText.WriteString(interestName)
		}
		additionalText.WriteString("\n")
	}

	// Создаем итоговое сообщение
	completionMsg := h.service.Localizer.Get(user.InterfaceLanguageCode, "interests_selection_complete")
	feedbackSuggestion := h.service.Localizer.Get(user.InterfaceLanguageCode, "interests_feedback_suggestion")

	fullMessage := completionMsg + "\n\n" + primaryText.String() + additionalText.String() + "\n" + feedbackSuggestion

	// Создаем клавиатуру
	keyboard := h.keyboardBuilder.CreateProfileCompletedKeyboard(user.InterfaceLanguageCode)

	// Обновляем сообщение
	editMsg := tgbotapi.NewEditMessageTextAndMarkup(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		fullMessage,
		keyboard,
	)

	_, err = h.bot.Request(editMsg)
	if err != nil {
		return err
	}

	// Обновляем состояние пользователя
	err = h.service.DB.UpdateUserState(user.ID, models.StateActive)
	if err != nil {
		log.Printf("Error updating user state: %v", err)
	}

	err = h.service.DB.UpdateUserStatus(user.ID, models.StatusActive)
	if err != nil {
		log.Printf("Error updating user status: %v", err)
	}

	// Обновляем уровень завершения профиля
	err = h.updateProfileCompletionLevel(user.ID, 100)
	if err != nil {
		log.Printf("Error updating profile completion level: %v", err)
	}

	return nil
}

// updateProfileCompletionLevel обновляет уровень завершения профиля
func (h *NewInterestHandlerImpl) updateProfileCompletionLevel(userID int, completionLevel int) error {
	_, err := h.service.DB.GetConnection().Exec(`
		UPDATE users
		SET profile_completion_level = $1, updated_at = NOW()
		WHERE id = $2
	`, completionLevel, userID)
	return err
}
