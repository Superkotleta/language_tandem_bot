package handlers

import (
	"log"
	"strconv"
	"strings"
	"sync"

	"language-exchange-bot/internal/core"
	"language-exchange-bot/internal/errors"
	"language-exchange-bot/internal/models"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// TemporaryInterestStorage временное хранилище выборов пользователей
type TemporaryInterestStorage struct {
	mu      sync.RWMutex
	storage map[int][]TemporaryInterestSelection // userID -> selections
}

// NewTemporaryInterestStorage создает новое временное хранилище
func NewTemporaryInterestStorage() *TemporaryInterestStorage {
	return &TemporaryInterestStorage{
		storage: make(map[int][]TemporaryInterestSelection),
	}
}

// AddInterest добавляет интерес во временное хранилище
func (s *TemporaryInterestStorage) AddInterest(userID, interestID int, isPrimary bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Проверяем, не выбран ли уже этот интерес
	for i, selection := range s.storage[userID] {
		if selection.InterestID == interestID {
			// Обновляем существующий выбор
			s.storage[userID][i].IsPrimary = isPrimary
			return
		}
	}

	// Добавляем новый выбор
	nextOrder := len(s.storage[userID]) + 1
	selection := TemporaryInterestSelection{
		InterestID:     interestID,
		IsPrimary:      isPrimary,
		SelectionOrder: nextOrder,
	}
	s.storage[userID] = append(s.storage[userID], selection)
}

// RemoveInterest удаляет интерес из временного хранилища
func (s *TemporaryInterestStorage) RemoveInterest(userID, interestID int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	selections := s.storage[userID]
	for i, selection := range selections {
		if selection.InterestID == interestID {
			// Удаляем из слайса
			s.storage[userID] = append(selections[:i], selections[i+1:]...)
			return
		}
	}
}

// ToggleInterest переключает выбор интереса
func (s *TemporaryInterestStorage) ToggleInterest(userID, interestID int) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	selections := s.storage[userID]
	for i, selection := range selections {
		if selection.InterestID == interestID {
			// Удаляем из временного хранилища
			s.storage[userID] = append(selections[:i], selections[i+1:]...)
			return false // был выбран, теперь не выбран
		}
	}

	// Добавляем в временное хранилище
	nextOrder := len(selections) + 1
	selection := TemporaryInterestSelection{
		InterestID:     interestID,
		IsPrimary:      false, // по умолчанию не основной
		SelectionOrder: nextOrder,
	}
	s.storage[userID] = append(s.storage[userID], selection)
	return true // теперь выбран
}

// TogglePrimary переключает статус основного интереса
func (s *TemporaryInterestStorage) TogglePrimary(userID, interestID int) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	selections := s.storage[userID]
	for i, selection := range selections {
		if selection.InterestID == interestID {
			// Переключаем статус
			s.storage[userID][i].IsPrimary = !selection.IsPrimary
			return s.storage[userID][i].IsPrimary
		}
	}
	return false
}

// GetSelections возвращает выборы пользователя
func (s *TemporaryInterestStorage) GetSelections(userID int) []TemporaryInterestSelection {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.storage[userID]
}

// GetSelectedInterests возвращает ID выбранных интересов
func (s *TemporaryInterestStorage) GetSelectedInterests(userID int) []int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var selected []int
	for _, selection := range s.storage[userID] {
		selected = append(selected, selection.InterestID)
	}
	return selected
}

// GetPrimaryInterests возвращает ID основных интересов
func (s *TemporaryInterestStorage) GetPrimaryInterests(userID int) []int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var primary []int
	for _, selection := range s.storage[userID] {
		if selection.IsPrimary {
			primary = append(primary, selection.InterestID)
		}
	}
	return primary
}

// ClearSelections очищает выборы пользователя
func (s *TemporaryInterestStorage) ClearSelections(userID int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.storage, userID)
}

// SaveToDatabase сохраняет выборы в базу данных
func (s *TemporaryInterestStorage) SaveToDatabase(userID int, interestService *core.InterestService) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	selections := s.storage[userID]
	if len(selections) == 0 {
		return nil
	}

	// Сначала удаляем все существующие выборы пользователя
	// (это делается в рамках транзакции в InterestService)

	// Добавляем новые выборы
	for _, selection := range selections {
		err := interestService.AddUserInterestSelection(userID, selection.InterestID, selection.IsPrimary)
		if err != nil {
			return err
		}
	}

	// Очищаем временное хранилище после успешного сохранения
	delete(s.storage, userID)
	return nil
}

// ImprovedInterestHandler улучшенный обработчик с временным хранением
type ImprovedInterestHandler struct {
	service         *core.BotService
	interestService *core.InterestService
	bot             *tgbotapi.BotAPI
	keyboardBuilder *KeyboardBuilder
	errorHandler    *errors.ErrorHandler
	tempStorage     *TemporaryInterestStorage
}

// NewImprovedInterestHandler создает улучшенный обработчик
func NewImprovedInterestHandler(service *core.BotService, interestService *core.InterestService, bot *tgbotapi.BotAPI, keyboardBuilder *KeyboardBuilder, errorHandler *errors.ErrorHandler) *ImprovedInterestHandler {
	return &ImprovedInterestHandler{
		service:         service,
		interestService: interestService,
		bot:             bot,
		keyboardBuilder: keyboardBuilder,
		errorHandler:    errorHandler,
		tempStorage:     NewTemporaryInterestStorage(),
	}
}

// HandleInterestCategorySelection обрабатывает выбор категории
func (h *ImprovedInterestHandler) HandleInterestCategorySelection(callback *tgbotapi.CallbackQuery, user *models.User, categoryKey string) error {
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

	// Получаем временные выборы пользователя
	tempSelections := h.tempStorage.GetSelections(user.ID)
	selectedMap := make(map[int]bool)
	for _, selection := range tempSelections {
		selectedMap[selection.InterestID] = true
	}

	// Создаем клавиатуру
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
	return err
}

// HandleInterestSelection обрабатывает выбор интереса (только во временном хранилище)
func (h *ImprovedInterestHandler) HandleInterestSelection(callback *tgbotapi.CallbackQuery, user *models.User, interestIDStr string) error {
	interestID, err := strconv.Atoi(interestIDStr)
	if err != nil {
		return h.errorHandler.HandleTelegramError(err, callback.Message.Chat.ID, int64(user.ID), "ParseInterestID")
	}

	// Переключаем выбор во временном хранилище
	isSelected := h.tempStorage.ToggleInterest(user.ID, interestID)

	log.Printf("User %d toggled interest %d: %v", user.ID, interestID, isSelected)

	// Обновляем клавиатуру (получаем categoryKey из callback data)
	// В реальной реализации нужно извлекать categoryKey из контекста
	return h.updateCategoryInterestsKeyboard(callback, user, "entertainment") // упрощенно
}

// HandlePrimaryInterestSelection обрабатывает выбор основного интереса
func (h *ImprovedInterestHandler) HandlePrimaryInterestSelection(callback *tgbotapi.CallbackQuery, user *models.User, interestIDStr string) error {
	interestID, err := strconv.Atoi(interestIDStr)
	if err != nil {
		return h.errorHandler.HandleTelegramError(err, callback.Message.Chat.ID, int64(user.ID), "ParseInterestID")
	}

	// Переключаем статус основного во временном хранилище
	isPrimary := h.tempStorage.TogglePrimary(user.ID, interestID)

	log.Printf("User %d toggled primary status for interest %d: %v", user.ID, interestID, isPrimary)

	// Обновляем клавиатуру выбора основных интересов
	return h.updatePrimaryInterestsKeyboard(callback, user)
}

// HandleInterestsContinue обрабатывает продолжение после выбора интересов
func (h *ImprovedInterestHandler) HandleInterestsContinue(callback *tgbotapi.CallbackQuery, user *models.User) error {
	// Получаем временные выборы
	selectedInterests := h.tempStorage.GetSelectedInterests(user.ID)

	// Проверяем, выбраны ли интересы
	if len(selectedInterests) == 0 {
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
		_, err := h.bot.Request(editMsg)
		return err
	}

	// Переходим к выбору основных интересов
	return h.showPrimaryInterestsSelection(callback, user)
}

// HandlePrimaryInterestsContinue обрабатывает завершение выбора основных интересов
func (h *ImprovedInterestHandler) HandlePrimaryInterestsContinue(callback *tgbotapi.CallbackQuery, user *models.User) error {
	// Получаем временные выборы
	tempSelections := h.tempStorage.GetSelections(user.ID)
	primaryCount := len(h.tempStorage.GetPrimaryInterests(user.ID))

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
			h.keyboardBuilder.CreatePrimaryInterestsKeyboard(tempSelections, user.InterfaceLanguageCode),
		)
		_, err = h.bot.Request(editMsg)
		return err
	}

	// Сохраняем в базу данных
	err = h.tempStorage.SaveToDatabase(user.ID, h.interestService)
	if err != nil {
		return h.errorHandler.HandleTelegramError(err, callback.Message.Chat.ID, int64(user.ID), "SaveToDatabase")
	}

	// Завершаем настройку профиля
	return h.completeProfileSetup(callback, user)
}

// showPrimaryInterestsSelection показывает интерфейс выбора основных интересов
func (h *ImprovedInterestHandler) showPrimaryInterestsSelection(callback *tgbotapi.CallbackQuery, user *models.User) error {
	// Получаем временные выборы
	tempSelections := h.tempStorage.GetSelections(user.ID)

	// Создаем клавиатуру для выбора основных интересов
	keyboard := h.keyboardBuilder.CreatePrimaryInterestsKeyboard(tempSelections, user.InterfaceLanguageCode)

	// Создаем текст сообщения
	messageText := h.service.Localizer.Get(user.InterfaceLanguageCode, "choose_primary_interests")

	// Обновляем сообщение
	editMsg := tgbotapi.NewEditMessageTextAndMarkup(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		messageText,
		keyboard,
	)

	_, err := h.bot.Request(editMsg)
	return err
}

// updateCategoryInterestsKeyboard обновляет клавиатуру интересов в категории
func (h *ImprovedInterestHandler) updateCategoryInterestsKeyboard(callback *tgbotapi.CallbackQuery, user *models.User, categoryKey string) error {
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

	// Получаем временные выборы
	tempSelections := h.tempStorage.GetSelections(user.ID)
	selectedMap := make(map[int]bool)
	for _, selection := range tempSelections {
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
func (h *ImprovedInterestHandler) updatePrimaryInterestsKeyboard(callback *tgbotapi.CallbackQuery, user *models.User) error {
	// Получаем временные выборы
	tempSelections := h.tempStorage.GetSelections(user.ID)

	// Создаем клавиатуру
	keyboard := h.keyboardBuilder.CreatePrimaryInterestsKeyboard(tempSelections, user.InterfaceLanguageCode)

	// Обновляем сообщение
	messageText := h.service.Localizer.Get(user.InterfaceLanguageCode, "choose_primary_interests")

	editMsg := tgbotapi.NewEditMessageTextAndMarkup(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		messageText,
		keyboard,
	)

	_, err := h.bot.Request(editMsg)
	return err
}

// completeProfileSetup завершает настройку профиля
func (h *ImprovedInterestHandler) completeProfileSetup(callback *tgbotapi.CallbackQuery, user *models.User) error {
	// Получаем сводку интересов пользователя из БД (после сохранения)
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
func (h *ImprovedInterestHandler) updateProfileCompletionLevel(userID int, completionLevel int) error {
	_, err := h.service.DB.GetConnection().Exec(`
		UPDATE users
		SET profile_completion_level = $1, updated_at = NOW()
		WHERE id = $2
	`, completionLevel, userID)
	return err
}
