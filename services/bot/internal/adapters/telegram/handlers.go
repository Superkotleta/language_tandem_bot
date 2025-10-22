package telegram

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"

	"language-exchange-bot/internal/adapters/telegram/handlers/admin"
	"language-exchange-bot/internal/adapters/telegram/handlers/availability"
	"language-exchange-bot/internal/adapters/telegram/handlers/base"
	"language-exchange-bot/internal/adapters/telegram/handlers/feedback"
	"language-exchange-bot/internal/adapters/telegram/handlers/interests"
	"language-exchange-bot/internal/adapters/telegram/handlers/language"
	"language-exchange-bot/internal/adapters/telegram/handlers/menu"
	"language-exchange-bot/internal/adapters/telegram/handlers/profile"
	"language-exchange-bot/internal/adapters/telegram/handlers/utility"
	"language-exchange-bot/internal/core"
	errorsPkg "language-exchange-bot/internal/errors"
	"language-exchange-bot/internal/localization"
	"language-exchange-bot/internal/models"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Константы для работы с коллбэками и массивами.

// TelegramHandler handles Telegram bot message and callback processing.
// The name includes "Telegram" prefix for clarity, even though it may cause stuttering with the package name.
type TelegramHandler struct {
	bot                    *tgbotapi.BotAPI
	service                *core.BotService
	adminChatIDs           []int64  // Chat ID администраторов
	adminUsernames         []string // Usernames администраторов для проверки доступа
	keyboardBuilder        *base.KeyboardBuilder
	menuHandler            *menu.MenuHandler
	profileHandler         *profile.ProfileHandlerImpl
	feedbackHandler        *feedback.FeedbackHandlerImpl
	languageHandler        *language.LanguageHandlerImpl
	interestHandler        *interests.NewInterestHandlerImpl
	profileInterestHandler *interests.ProfileInterestHandler
	isolatedInterestEditor *interests.IsolatedInterestEditor
	isolatedLanguageEditor *language.IsolatedLanguageEditor
	availabilityHandler    *availability.AvailabilityHandlerImpl
	availabilityEditor     availability.IsolatedAvailabilityEditor
	adminHandler           *admin.AdminHandlerImpl
	utilityHandler         *utility.UtilityHandlerImpl
	errorHandler           *errorsPkg.ErrorHandler
	isolatedRouter         *CallbackRouter // Роутер для изолированных callback'ов
	rateLimiter            *RateLimiter    // Rate limiter для защиты от спама
	messageFactory         *base.MessageFactory
}

// NewTelegramHandler создает новый экземпляр TelegramHandler с базовой конфигурацией.
func NewTelegramHandler(
	bot *tgbotapi.BotAPI,
	service *core.BotService,
	adminChatIDs []int64,
	errorHandler *errorsPkg.ErrorHandler,
) *TelegramHandler {
	keyboardBuilder := base.NewKeyboardBuilder(service)
	messageFactory := base.NewMessageFactory(bot, errorHandler, service.LoggingService)

	// Создаем общий BaseHandler
	baseHandler := base.NewBaseHandler(
		bot,
		service,
		keyboardBuilder,
		errorHandler,
		messageFactory,
	)

	menuHandler := menu.NewMenuHandler(baseHandler)
	profileHandler := profile.NewProfileHandler(baseHandler)
	feedbackHandler := feedback.NewFeedbackHandler(
		baseHandler,
		adminChatIDs,
		make([]string, 0),
	)
	languageHandler := language.NewLanguageHandler(baseHandler)

	var interestService *core.InterestService
	if service.DB != nil {
		interestService = core.NewInterestService(service.DB.GetConnection())
	} else {
		interestService = nil // Для тестов без DB
	}

	interestHandler := interests.NewNewInterestHandler(baseHandler, interestService)
	profileInterestHandler := interests.NewProfileInterestHandler(
		service,
		interestService,
		bot,
		keyboardBuilder,
		errorHandler,
	)
	isolatedInterestEditor := interests.NewIsolatedInterestEditor(
		service,
		interestService,
		bot,
		keyboardBuilder,
		errorHandler,
		service.Cache,
	)
	isolatedLanguageEditor := language.NewIsolatedLanguageEditor(baseHandler)
	availabilityHandler := availability.NewAvailabilityHandler(baseHandler)
	availabilityEditor := *availability.NewIsolatedAvailabilityEditor(baseHandler)
	adminHandler := admin.NewAdminHandler(baseHandler, adminChatIDs, make([]string, 0))
	utilityHandler := utility.NewUtilityHandler(baseHandler)

	// Создаем rate limiter для защиты от спама
	rateLimiter := NewRateLimiter(DefaultRateLimitConfig())

	// Создаем и настраиваем роутер для изолированных callback'ов
	isolatedRouter := NewCallbackRouter()
	handler := &TelegramHandler{
		bot:                    bot,
		service:                service,
		adminChatIDs:           adminChatIDs,
		adminUsernames:         make([]string, 0),
		keyboardBuilder:        keyboardBuilder,
		menuHandler:            menuHandler,
		profileHandler:         profileHandler,
		feedbackHandler:        feedbackHandler,
		languageHandler:        languageHandler,
		interestHandler:        interestHandler,
		profileInterestHandler: profileInterestHandler,
		isolatedInterestEditor: isolatedInterestEditor,
		isolatedLanguageEditor: isolatedLanguageEditor,
		availabilityHandler:    availabilityHandler,
		availabilityEditor:     availabilityEditor,
		adminHandler:           adminHandler,
		utilityHandler:         utilityHandler,
		errorHandler:           errorHandler,
		isolatedRouter:         isolatedRouter,
		rateLimiter:            rateLimiter,
		messageFactory:         messageFactory,
	}

	// Настраиваем маршруты для изолированных callback'ов
	if err := isolatedRouter.SetupIsolatedRoutes(handler); err != nil {
		panic(fmt.Sprintf("failed to setup isolated routes: %v", err))
	}

	return handler
}

// NewTelegramHandlerWithAdmins создает новый экземпляр TelegramHandler с полной конфигурацией администраторов.
func NewTelegramHandlerWithAdmins(
	bot *tgbotapi.BotAPI,
	service *core.BotService,
	adminChatIDs []int64,
	adminUsernames []string,
	errorHandler *errorsPkg.ErrorHandler,
) *TelegramHandler {
	keyboardBuilder := base.NewKeyboardBuilder(service)
	messageFactory := base.NewMessageFactory(bot, errorHandler, service.LoggingService)

	// Создаем общий BaseHandler
	baseHandler := base.NewBaseHandler(
		bot,
		service,
		keyboardBuilder,
		errorHandler,
		messageFactory,
	)

	menuHandler := menu.NewMenuHandler(baseHandler)
	profileHandler := profile.NewProfileHandler(baseHandler)
	feedbackHandler := feedback.NewFeedbackHandler(
		baseHandler,
		adminChatIDs,
		adminUsernames,
	)
	languageHandler := language.NewLanguageHandler(baseHandler)
	interestService := core.NewInterestService(service.DB.GetConnection())
	interestHandler := interests.NewNewInterestHandler(baseHandler, interestService)
	profileInterestHandler := interests.NewProfileInterestHandler(
		service,
		interestService,
		bot,
		keyboardBuilder,
		errorHandler,
	)
	isolatedInterestEditor := interests.NewIsolatedInterestEditor(
		service,
		interestService,
		bot,
		keyboardBuilder,
		errorHandler,
		service.Cache,
	)
	isolatedLanguageEditor := language.NewIsolatedLanguageEditor(baseHandler)
	availabilityHandler := availability.NewAvailabilityHandler(baseHandler)
	availabilityEditor := *availability.NewIsolatedAvailabilityEditor(baseHandler)
	adminHandler := admin.NewAdminHandler(baseHandler, adminChatIDs, adminUsernames)
	utilityHandler := utility.NewUtilityHandler(baseHandler)

	// Создаем rate limiter для защиты от спама
	rateLimiter := NewRateLimiter(DefaultRateLimitConfig())

	// Создаем и настраиваем роутер для изолированных callback'ов
	isolatedRouter := NewCallbackRouter()
	handler := &TelegramHandler{
		bot:                    bot,
		service:                service,
		adminChatIDs:           adminChatIDs,
		adminUsernames:         adminUsernames,
		keyboardBuilder:        keyboardBuilder,
		menuHandler:            menuHandler,
		profileHandler:         profileHandler,
		feedbackHandler:        feedbackHandler,
		languageHandler:        languageHandler,
		interestHandler:        interestHandler,
		profileInterestHandler: profileInterestHandler,
		isolatedInterestEditor: isolatedInterestEditor,
		isolatedLanguageEditor: isolatedLanguageEditor,
		availabilityHandler:    availabilityHandler,
		availabilityEditor:     availabilityEditor,
		adminHandler:           adminHandler,
		utilityHandler:         utilityHandler,
		errorHandler:           errorHandler,
		isolatedRouter:         isolatedRouter,
		rateLimiter:            rateLimiter,
		messageFactory:         messageFactory,
	}

	// Настраиваем маршруты для изолированных callback'ов
	if err := isolatedRouter.SetupIsolatedRoutes(handler); err != nil {
		panic(fmt.Sprintf("failed to setup isolated routes: %v", err))
	}

	return handler
}

// HandleUpdate обрабатывает входящие обновления от Telegram API.
func (h *TelegramHandler) HandleUpdate(update tgbotapi.Update) error {
	// Получаем ID пользователя
	var userID int64
	if update.Message != nil {
		userID = update.Message.From.ID
	} else if update.CallbackQuery != nil {
		userID = update.CallbackQuery.From.ID
	} else {
		// Неизвестный тип обновления, пропускаем
		return nil
	}

	// Проверяем rate limit
	if err := h.rateLimiter.CheckRateLimit(userID); err != nil {
		// Отправляем сообщение о превышении лимита
		h.sendRateLimitMessage(userID, err)

		return nil // Не возвращаем ошибку, чтобы не логировать её как системную
	}

	if update.Message != nil {
		return h.handleMessage(update.Message)
	}

	if update.CallbackQuery != nil {
		return h.handleCallbackQuery(update.CallbackQuery)
	}

	return nil
}

// handleMessage обрабатывает входящие текстовые сообщения.
func (h *TelegramHandler) handleMessage(message *tgbotapi.Message) error {
	if h.service == nil {
		return errors.New("service not initialized")
	}

	user, err := h.service.HandleUserRegistration(
		message.From.ID,
		message.From.UserName,
		message.From.FirstName,
		message.From.LanguageCode,
	)
	if err != nil {
		// Используем новую систему обработки ошибок
		if h.errorHandler != nil {
			userID := int64(0)
			if user != nil {
				userID = int64(user.ID)
			}

			return h.errorHandler.HandleDatabaseError(
				err,
				userID,
				message.Chat.ID,
				"HandleUserRegistration",
			)
		}
		// Fallback к простому логированию если errorHandler не инициализирован
		log.Printf("Database error in HandleUserRegistration: %v", err)

		return nil
	}

	if message.IsCommand() {
		return h.handleCommand(message, user)
	}

	return h.handleState(message, user)
}

// handleCommand обрабатывает команды пользователя (начинающиеся с /).
func (h *TelegramHandler) handleCommand(message *tgbotapi.Message, user *models.User) error {
	if h.service == nil {
		return errors.New("service not initialized")
	}

	switch message.Command() {
	case "start":
		return h.menuHandler.HandleStartCommand(message, user)
	case "status":
		return h.menuHandler.HandleStatusCommand(message, user)
	case "reset":
		return h.menuHandler.HandleResetCommand(message, user)
	case "language":
		return h.menuHandler.HandleLanguageCommand(message, user)
	case "profile":
		return h.profileHandler.HandleProfileCommand(message, user)
	case "feedback":
		return h.feedbackHandler.HandleFeedbackCommand(
			message,
			user,
		)
	case "feedbacks":
		return h.feedbackHandler.HandleFeedbacksCommand(
			message,
			user,
		)
	default:
		log.Printf("Unknown command: %s", message.Command())

		return h.utilityHandler.SendMessage(
			message.Chat.ID,
			h.service.Localizer.Get(user.InterfaceLanguageCode, "unknown_command"),
		)
	}
}

// handleState обрабатывает сообщения в зависимости от текущего состояния пользователя.
func (h *TelegramHandler) handleState(message *tgbotapi.Message, user *models.User) error {
	if h.service == nil {
		return errors.New("service not initialized")
	}

	switch user.State {
	case models.StateWaitingLanguage,
		models.StateWaitingInterests,
		models.StateWaitingTime,
		models.StateWaitingFriendshipPreferences:
		return h.utilityHandler.SendMessage(
			message.Chat.ID,
			h.service.Localizer.Get(user.InterfaceLanguageCode, "use_menu_above"),
		)
	case models.StateWaitingFeedback:
		return h.feedbackHandler.HandleFeedbackMessage(message, user)
	case models.StateWaitingFeedbackContact:
		return h.feedbackHandler.HandleFeedbackContactMessage(message, user)
	default:
		// Игнорируем текстовые сообщения, если пользователь не в специальном состоянии
		// Пользователь должен использовать кнопки меню
		return nil
	}
}

// handleCallbackQuery обрабатывает нажатия на inline-кнопки.
func (h *TelegramHandler) handleCallbackQuery(callback *tgbotapi.CallbackQuery) error {
	if h.service == nil {
		return errors.New("service not initialized")
	}

	if callback == nil {
		return errors.New("callback is nil")
	}

	if callback.From == nil {
		return errors.New("callback.From is nil")
	}

	log.Printf("DEBUG: handleCallbackQuery called with data: '%s' from user %d", callback.Data, callback.From.ID)

	user, err := h.service.HandleUserRegistration(
		callback.From.ID,
		callback.From.UserName,
		callback.From.FirstName,
		callback.From.LanguageCode,
	)
	if err != nil {
		log.Printf("ERROR: Failed to handle user registration for user %d: %v", callback.From.ID, err)

		return err
	}

	log.Printf("DEBUG: User found: ID=%d, State=%s, InterfaceLang=%s", user.ID, user.State, user.InterfaceLanguageCode)

	data := callback.Data
	_, _ = h.bot.Request(tgbotapi.NewCallback(callback.ID, ""))

	// Разделяем обработку callback'ов по категориям для уменьшения сложности
	if err := h.handleLanguageCallbacks(callback, user, data); err != nil {
		log.Printf("DEBUG: handleLanguageCallbacks returned error: %v", err)

		return err
	}

	if err := h.handleInterestCallbacks(callback, user, data); err != nil {
		log.Printf("DEBUG: handleInterestCallbacks returned error: %v", err)

		return err
	}

	if err := h.handleProfileCallbacks(
		callback,
		user,
		data,
	); err != nil {
		log.Printf("DEBUG: handleProfileCallbacks returned error: %v", err)

		return err
	}

	if err := h.handleMenuCallbacks(
		callback,
		user,
		data,
	); err != nil {
		log.Printf("DEBUG: handleMenuCallbacks returned error: %v", err)

		return err
	}

	if err := h.handleAvailabilityCallbacks(callback, user, data); err != nil {
		log.Printf("DEBUG: handleAvailabilityCallbacks returned error: %v", err)

		return err
	}

	if err := h.handleFeedbackCallbacks(callback, user, data); err != nil {
		log.Printf("DEBUG: handleFeedbackCallbacks returned error: %v", err)

		return err
	}

	// Если callback не был обработан ни одним обработчиком, просто игнорируем
	log.Printf("DEBUG: No handler processed callback data: '%s'", data)

	return nil
}

// isAdmin проверяет, является ли пользователь администратором.
func (h *TelegramHandler) isAdmin(userID int64, username string) bool {
	// Проверяем по Chat ID
	for _, adminID := range h.adminChatIDs {
		if userID == adminID {
			return true
		}
	}

	// Проверяем по username
	if username != "" {
		for _, adminUsername := range h.adminUsernames {
			if username == adminUsername {
				return true
			}
		}
	}

	return false
}

// === ОБРАБОТЧИКИ ВИДОВ ОТЗЫВОВ ===

// вспомогательная функция для сохранения состояния навигации отзывов

// === ОБРАБОТЧИКИ КОНТРОЛЯ ОТЗЫВОВ ===

// === НОВЫЕ МЕТОДЫ ДЛЯ СНИЖЕНИЯ ЦИКЛОМАТИЧЕСКОЙ СЛОЖНОСТИ ===

// handleLanguageCallbacks обрабатывает callback'и связанные с языками.
func (h *TelegramHandler) handleLanguageCallbacks(callback *tgbotapi.CallbackQuery, user *models.User, data string) error {
	// Обработка выбора языков
	if strings.HasPrefix(data, "lang_") {
		return h.handleLanguageSelection(
			callback,
			user,
			data,
		)
	}

	// Обработка уровней языка
	if strings.HasPrefix(data, "level_") {
		levelCode := strings.TrimPrefix(data, "level_")

		return h.languageHandler.HandleLanguageLevelSelection(callback, user, levelCode)
	}

	// Обработка редактирования
	if strings.HasPrefix(data, "edit_") {
		return h.handleLanguageEditing(
			callback,
			user,
			data,
		)
	}

	// Обработка специальных команд
	return h.handleLanguageSpecialCommands(
		callback,
		user,
		data,
	)
}

func (h *TelegramHandler) handleLanguageSelection(callback *tgbotapi.CallbackQuery, user *models.User, data string) error {
	switch {
	case strings.HasPrefix(data, "lang_native_"):
		return h.languageHandler.HandleNativeLanguageCallback(callback, user)
	case strings.HasPrefix(data, "lang_target_"):
		return h.languageHandler.HandleTargetLanguageCallback(callback, user)
	case strings.HasPrefix(data, "lang_interface_"):
		langCode := strings.TrimPrefix(data, "lang_interface_")

		return h.languageHandler.HandleInterfaceLanguageSelection(callback, user, langCode)
	}

	return nil
}

func (h *TelegramHandler) handleLanguageEditing(callback *tgbotapi.CallbackQuery, user *models.User, data string) error {
	switch {
	// Removed deprecated edit_level_, lang_edit_native_, lang_edit_target_ callbacks
	// Use isolated language editor via "edit_languages" callback instead
	}

	return nil
}

func (h *TelegramHandler) handleLanguageSpecialCommands(callback *tgbotapi.CallbackQuery, user *models.User, data string) error {
	switch data {
	case "back_to_language_level":
		return h.languageHandler.HandleBackToLanguageLevel(
			callback,
			user,
		)
	case "languages_continue_filling":
		return h.languageHandler.HandleLanguagesContinueFilling(callback, user)
	case "languages_reselect":
		return h.languageHandler.HandleLanguagesReselect(
			callback,
			user,
		)
	}

	return nil
}

// handleInterestCallbacks обрабатывает callback'и связанные с интересами.
func (h *TelegramHandler) handleInterestCallbacks(callback *tgbotapi.CallbackQuery, user *models.User, data string) error {
	log.Printf("DEBUG: handleInterestCallbacks called with data: '%s' for user %d", data, user.ID)

	switch {
	case data == "back_to_categories":
		return h.interestHandler.HandleBackToCategories(callback, user)
	case data == "interests_continue":
		return h.interestHandler.HandleInterestsContinue(
			callback,
			user,
		)
	case strings.HasPrefix(data, "interest_category_"):
		categoryKey := strings.TrimPrefix(data, "interest_category_")

		return h.interestHandler.HandleInterestCategorySelection(callback, user, categoryKey)
	case strings.HasPrefix(data, "interest_select_"):
		interestID := strings.TrimPrefix(data, "interest_select_")

		return h.interestHandler.HandleInterestSelection(
			callback,
			user,
			interestID,
		)
	case strings.HasPrefix(data, "primary_interest_"):
		interestID := strings.TrimPrefix(data, "primary_interest_")

		return h.interestHandler.HandlePrimaryInterestSelection(callback, user, interestID)
	case data == "primary_interests_continue":
		return h.interestHandler.HandlePrimaryInterestsContinue(callback, user)
	case data == "back_to_interests":
		return h.interestHandler.HandleBackToInterests(callback, user)
	}

	return nil
}

// handleProfileCallbacks обрабатывает callback'и связанные с профилем.
func (h *TelegramHandler) handleProfileCallbacks(callback *tgbotapi.CallbackQuery, user *models.User, data string) error {
	// Обработка изолированного редактирования интересов
	if strings.HasPrefix(data, "isolated_") {
		return h.handleIsolatedCallbacks(callback, user, data)
	}

	// Обработка редактирования интересов (только для совместимости с существующими сессиями)
	if strings.HasPrefix(data, "edit_interest") || data == "save_interest_edits" {
		return h.handleProfileInterestEditing(
			callback,
			user,
			data,
		)
	}

	// Обработка команд профиля
	if strings.HasPrefix(data, "profile_") ||
		strings.HasPrefix(data, "edit_") ||
		data == "back_to_previous_step" {
		return h.handleProfileCommands(
			callback,
			user,
			data,
		)
	}

	return nil
}

func (h *TelegramHandler) handleProfileInterestEditing(callback *tgbotapi.CallbackQuery, user *models.User, data string) error {
	log.Printf("DEBUG: handleProfileInterestEditing called with data: '%s' for user %d", data, user.ID)

	switch {
	case strings.HasPrefix(data, "edit_interest_category_"):
		categoryKey := strings.TrimPrefix(data, "edit_interest_category_")

		return h.profileInterestHandler.HandleEditInterestCategoryFromProfile(callback, user, categoryKey)
	case strings.HasPrefix(data, "edit_interest_select_"):
		interestID := strings.TrimPrefix(data, "edit_interest_select_")

		return h.profileInterestHandler.HandleEditInterestSelectionFromProfile(callback, user, interestID)
	case data == "edit_primary_interests":
		return h.profileInterestHandler.HandleEditPrimaryInterestsFromProfile(callback, user)
	case strings.HasPrefix(data, "edit_primary_interest_"):
		interestID := strings.TrimPrefix(data, "edit_primary_interest_")

		return h.profileInterestHandler.HandleEditPrimaryInterestSelectionFromProfile(callback, user, interestID)
	case data == "save_interest_edits":
		return h.profileInterestHandler.HandleSaveInterestEditsFromProfile(callback, user)
	case data == "back_to_categories":
		return h.profileInterestHandler.HandleEditInterestsFromProfile(callback, user)
	case data == "back_to_edit_categories":
		return h.profileInterestHandler.HandleEditInterestsFromProfile(callback, user)
	case data == "back_to_profile":
		return h.profileHandler.HandleProfileShow(callback, user)
	}

	log.Printf("DEBUG: No handler found in handleProfileInterestEditing for data: '%s'", data)

	return nil
}

// HandleIsolatedEditStart начинает изолированное редактирование интересов.
func (h *TelegramHandler) HandleIsolatedEditStart(callback *tgbotapi.CallbackQuery, user *models.User) error {
	log.Printf("Starting isolated edit session for user %d", user.ID)

	return h.isolatedInterestEditor.StartEditSession(callback, user)
}

// HandleIsolatedMainMenu обрабатывает главное меню изолированного редактирования.
func (h *TelegramHandler) HandleIsolatedMainMenu(callback *tgbotapi.CallbackQuery, user *models.User) error {
	log.Printf("Showing isolated main menu for user %d", user.ID)

	session, err := h.isolatedInterestEditor.GetEditSession(user.ID)
	if err != nil {
		return h.errorHandler.HandleTelegramError(err, callback.Message.Chat.ID, int64(user.ID), "GetEditSession")
	}

	return h.isolatedInterestEditor.ShowEditMainMenu(callback, user, session)
}

// HandleIsolatedEditCategories обрабатывает меню категорий.
func (h *TelegramHandler) HandleIsolatedEditCategories(callback *tgbotapi.CallbackQuery, user *models.User) error {
	log.Printf("Showing isolated categories menu for user %d", user.ID)

	session, err := h.isolatedInterestEditor.GetEditSession(user.ID)
	if err != nil {
		return h.errorHandler.HandleTelegramError(err, callback.Message.Chat.ID, int64(user.ID), "GetEditSession")
	}

	return h.isolatedInterestEditor.ShowEditCategoriesMenu(callback, user, session)
}

// HandleIsolatedEditCategory обрабатывает выбор категории для редактирования.
func (h *TelegramHandler) HandleIsolatedEditCategory(callback *tgbotapi.CallbackQuery, user *models.User, categoryKey string) error {
	log.Printf("Showing isolated category interests for user %d, category: %s", user.ID, categoryKey)

	session, err := h.isolatedInterestEditor.GetEditSession(user.ID)
	if err != nil {
		return h.errorHandler.HandleTelegramError(err, callback.Message.Chat.ID, int64(user.ID), "GetEditSession")
	}

	return h.isolatedInterestEditor.ShowEditCategoryInterests(callback, user, session, categoryKey)
}

// HandleIsolatedEditPrimary обрабатывает редактирование основных интересов.
func (h *TelegramHandler) HandleIsolatedEditPrimary(callback *tgbotapi.CallbackQuery, user *models.User) error {
	session, err := h.isolatedInterestEditor.GetEditSession(user.ID)
	if err != nil {
		return h.errorHandler.HandleTelegramError(err, callback.Message.Chat.ID, int64(user.ID), "GetEditSession")
	}

	return h.isolatedInterestEditor.ShowEditPrimaryInterests(callback, user, session)
}

// HandleIsolatedToggleInterest обрабатывает переключение выбора интереса.
func (h *TelegramHandler) HandleIsolatedToggleInterest(callback *tgbotapi.CallbackQuery, user *models.User, interestIDStr string) error {
	interestID, err := strconv.Atoi(interestIDStr)
	if err != nil {
		return h.errorHandler.HandleTelegramError(err, callback.Message.Chat.ID, int64(user.ID), "ParseInterestID")
	}

	log.Printf("Toggling interest %d for user %d", interestID, user.ID)

	return h.isolatedInterestEditor.ToggleInterestSelection(callback, user, interestID)
}

// HandleIsolatedTogglePrimary обрабатывает переключение основного интереса.
func (h *TelegramHandler) HandleIsolatedTogglePrimary(callback *tgbotapi.CallbackQuery, user *models.User, interestIDStr string) error {
	interestID, err := strconv.Atoi(interestIDStr)
	if err != nil {
		return h.errorHandler.HandleTelegramError(err, callback.Message.Chat.ID, int64(user.ID), "ParseInterestID")
	}

	log.Printf("Toggling primary status for interest %d for user %d", interestID, user.ID)

	return h.isolatedInterestEditor.TogglePrimaryInterest(callback, user, interestID)
}

// HandleIsolatedPreviewChanges обрабатывает предварительный просмотр изменений.
func (h *TelegramHandler) HandleIsolatedPreviewChanges(callback *tgbotapi.CallbackQuery, user *models.User) error {
	log.Printf("Showing changes preview for user %d", user.ID)

	session, err := h.isolatedInterestEditor.GetEditSession(user.ID)
	if err != nil {
		return h.errorHandler.HandleTelegramError(err, callback.Message.Chat.ID, int64(user.ID), "GetEditSession")
	}

	return h.isolatedInterestEditor.ShowChangesPreview(callback, user, session)
}

// HandleIsolatedSaveChanges обрабатывает сохранение изменений.
func (h *TelegramHandler) HandleIsolatedSaveChanges(callback *tgbotapi.CallbackQuery, user *models.User) error {
	log.Printf("Saving changes for user %d", user.ID)

	session, err := h.isolatedInterestEditor.GetEditSession(user.ID)
	if err != nil {
		return h.errorHandler.HandleTelegramError(err, callback.Message.Chat.ID, int64(user.ID), "GetEditSession")
	}

	return h.isolatedInterestEditor.SaveChanges(callback, user, session)
}

// HandleIsolatedCancelEdit обрабатывает отмену редактирования.
func (h *TelegramHandler) HandleIsolatedCancelEdit(callback *tgbotapi.CallbackQuery, user *models.User) error {
	log.Printf("Canceling edit for user %d", user.ID)

	return h.isolatedInterestEditor.CancelEdit(callback, user)
}

// HandleIsolatedMassSelect обрабатывает массовый выбор в категории.
func (h *TelegramHandler) HandleIsolatedMassSelect(callback *tgbotapi.CallbackQuery, user *models.User, categoryKey string) error {
	log.Printf("Mass selecting all interests in category %s for user %d", categoryKey, user.ID)

	session, err := h.isolatedInterestEditor.GetEditSession(user.ID)
	if err != nil {
		return h.errorHandler.HandleTelegramError(err, callback.Message.Chat.ID, int64(user.ID), "GetEditSession")
	}

	return h.isolatedInterestEditor.MassSelectCategory(callback, user, session, categoryKey)
}

// HandleIsolatedMassClear обрабатывает массовую очистку категории.
func (h *TelegramHandler) HandleIsolatedMassClear(callback *tgbotapi.CallbackQuery, user *models.User, categoryKey string) error {
	log.Printf("Mass clearing all interests in category %s for user %d", categoryKey, user.ID)

	session, err := h.isolatedInterestEditor.GetEditSession(user.ID)
	if err != nil {
		return h.errorHandler.HandleTelegramError(err, callback.Message.Chat.ID, int64(user.ID), "GetEditSession")
	}

	return h.isolatedInterestEditor.MassClearCategory(callback, user, session, categoryKey)
}

// HandleIsolatedUndoLast обрабатывает отмену последнего действия.
func (h *TelegramHandler) HandleIsolatedUndoLast(callback *tgbotapi.CallbackQuery, user *models.User) error {
	log.Printf("Undoing last action for user %d", user.ID)

	session, err := h.isolatedInterestEditor.GetEditSession(user.ID)
	if err != nil {
		return h.errorHandler.HandleTelegramError(err, callback.Message.Chat.ID, int64(user.ID), "GetEditSession")
	}

	return h.isolatedInterestEditor.UndoLastChange(callback, user, session)
}

// HandleIsolatedShowStats обрабатывает показ статистики.
func (h *TelegramHandler) HandleIsolatedShowStats(callback *tgbotapi.CallbackQuery, user *models.User) error {
	log.Printf("Showing edit statistics for user %d", user.ID)

	session, err := h.isolatedInterestEditor.GetEditSession(user.ID)
	if err != nil {
		return h.errorHandler.HandleTelegramError(err, callback.Message.Chat.ID, int64(user.ID), "GetEditSession")
	}

	return h.isolatedInterestEditor.ShowEditStatistics(callback, user, session)
}

// =============================================================================
// ISOLATED LANGUAGE EDITOR HANDLERS
// =============================================================================

// HandleIsolatedLangEditNative начинает редактирование родного языка.
func (h *TelegramHandler) HandleIsolatedLangEditNative(callback *tgbotapi.CallbackQuery, user *models.User) error {
	log.Printf("Starting native language edit for user %d", user.ID)
	return h.isolatedLanguageEditor.HandleEditNativeLanguage(callback, user)
}

// HandleIsolatedLangEditTarget начинает редактирование изучаемого языка.
func (h *TelegramHandler) HandleIsolatedLangEditTarget(callback *tgbotapi.CallbackQuery, user *models.User) error {
	log.Printf("Starting target language edit for user %d", user.ID)
	return h.isolatedLanguageEditor.HandleEditTargetLanguage(callback, user)
}

// HandleIsolatedLangEditLevel начинает редактирование уровня владения языком.
func (h *TelegramHandler) HandleIsolatedLangEditLevel(callback *tgbotapi.CallbackQuery, user *models.User) error {
	log.Printf("Starting language level edit for user %d", user.ID)
	return h.isolatedLanguageEditor.HandleEditLanguageLevel(callback, user)
}

// HandleIsolatedLangPreview показывает предпросмотр изменений языков.
func (h *TelegramHandler) HandleIsolatedLangPreview(callback *tgbotapi.CallbackQuery, user *models.User) error {
	log.Printf("Showing language changes preview for user %d", user.ID)
	return h.isolatedLanguageEditor.HandlePreviewChanges(callback, user)
}

// HandleIsolatedLangUndoLast отменяет последнее изменение языка.
func (h *TelegramHandler) HandleIsolatedLangUndoLast(callback *tgbotapi.CallbackQuery, user *models.User) error {
	log.Printf("Undoing last language change for user %d", user.ID)
	return h.isolatedLanguageEditor.HandleUndoLastChange(callback, user)
}

// HandleIsolatedLangBackToMenu возвращает в главное меню редактора языков.
func (h *TelegramHandler) HandleIsolatedLangBackToMenu(callback *tgbotapi.CallbackQuery, user *models.User) error {
	log.Printf("Returning to language edit menu for user %d", user.ID)
	return h.isolatedLanguageEditor.HandleBackToMenu(callback, user)
}

// HandleIsolatedLangNativeSelection обрабатывает выбор родного языка.
func (h *TelegramHandler) HandleIsolatedLangNativeSelection(callback *tgbotapi.CallbackQuery, user *models.User, langCode string) error {
	log.Printf("Native language selected: %s for user %d", langCode, user.ID)
	return h.isolatedLanguageEditor.HandleNativeLanguageSelection(callback, user, langCode)
}

// HandleIsolatedLangTargetSelection обрабатывает выбор изучаемого языка.
func (h *TelegramHandler) HandleIsolatedLangTargetSelection(callback *tgbotapi.CallbackQuery, user *models.User, langCode string) error {
	log.Printf("Target language selected: %s for user %d", langCode, user.ID)
	return h.isolatedLanguageEditor.HandleTargetLanguageSelection(callback, user, langCode)
}

// HandleIsolatedLangLevelSelection обрабатывает выбор уровня владения языком.
func (h *TelegramHandler) HandleIsolatedLangLevelSelection(callback *tgbotapi.CallbackQuery, user *models.User, levelCode string) error {
	log.Printf("Language level selected: %s for user %d", levelCode, user.ID)
	return h.isolatedLanguageEditor.HandleLanguageLevelSelection(callback, user, levelCode)
}

// HandleIsolatedLangSave обрабатывает сохранение изменений языков.
func (h *TelegramHandler) HandleIsolatedLangSave(callback *tgbotapi.CallbackQuery, user *models.User) error {
	log.Printf("Saving language changes for user %d", user.ID)
	return h.isolatedLanguageEditor.HandleSaveChanges(callback, user)
}

// HandleIsolatedLangCancel обрабатывает отмену редактирования языков.
func (h *TelegramHandler) HandleIsolatedLangCancel(callback *tgbotapi.CallbackQuery, user *models.User) error {
	log.Printf("Cancelling language edit for user %d", user.ID)
	return h.isolatedLanguageEditor.HandleCancelEdit(callback, user)
}

// handleIsolatedCallbacks обрабатывает все callback'и изолированной системы через роутер.
func (h *TelegramHandler) handleIsolatedCallbacks(callback *tgbotapi.CallbackQuery, user *models.User, data string) error {
	// Используем роутер для обработки callback'а
	return h.isolatedRouter.Handle(callback, user)
}

func (h *TelegramHandler) handleProfileCommands(callback *tgbotapi.CallbackQuery, user *models.User, data string) error {
	log.Printf("DEBUG: handleProfileCommands called with data: '%s' for user %d", data, user.ID)

	switch data {
	case "profile_show":
		log.Printf("DEBUG: Handling profile_show for user %d", user.ID)

		return h.profileHandler.HandleProfileShow(callback, user)
	case "profile_reset_ask":
		log.Printf("DEBUG: Handling profile_reset_ask for user %d", user.ID)

		return h.profileHandler.HandleProfileResetAsk(callback, user)
	case "profile_reset_yes":
		log.Printf("DEBUG: Handling profile_reset_yes for user %d", user.ID)

		return h.profileHandler.HandleProfileResetYes(callback, user)
	case "profile_reset_no":
		log.Printf("DEBUG: Handling profile_reset_no for user %d", user.ID)

		return h.menuHandler.HandleBackToMainMenu(callback, user)
	case "back_to_previous_step":
		log.Printf("DEBUG: Handling back_to_previous_step for user %d", user.ID)

		return h.profileHandler.HandleProfileShow(callback, user)
	case "edit_languages":
		log.Printf("DEBUG: Handling edit_languages for user %d", user.ID)

		return h.profileHandler.HandleEditLanguages(callback, user)
	case "edit_availability":
		log.Printf("DEBUG: Handling edit_availability for user %d", user.ID)

		return h.availabilityHandler.HandleTimeAvailabilityStart(callback, user)
		// Removed deprecated language edit callbacks (edit_native_lang, edit_target_lang, edit_level)
		// Use isolated language editor via "edit_languages" callback instead
	}

	log.Printf("DEBUG: No handler found for data: '%s' for user %d", data, user.ID)

	return nil
}

// handleMenuCallbacks обрабатывает callback'и связанные с меню.
// handleAvailabilityCallbacks обрабатывает callback'и связанные с настройкой доступности
func (h *TelegramHandler) handleAvailabilityCallbacks(callback *tgbotapi.CallbackQuery, user *models.User, data string) error {
	// Обработка callback'а для перехода к настройке доступности
	if data == "continue_to_availability" {
		return h.availabilityHandler.HandleTimeAvailabilityStart(callback, user)
	}

	// Обработка callback'ов редактора доступности (isolated editor)
	if strings.HasPrefix(data, "availability_") ||
		strings.HasPrefix(data, "avail_edit_") ||
		strings.HasPrefix(data, "avail_") {

		// Логируем все availability callback'и для отладки
		h.service.LoggingService.Telegram().InfoWithContext("Processing availability callback", "", int64(user.ID), callback.Message.Chat.ID, "AvailabilityCallback", map[string]interface{}{
			"user_id":       user.ID,
			"callback_data": data,
		})

		// Обработка callback'ов редактора
		switch {
		case data == "avail_edit_days":
			return h.availabilityEditor.EditDays(callback, user)
		case data == "avail_edit_time":
			return h.availabilityEditor.EditTimeSlots(callback, user)
		case data == "avail_edit_communication":
			return h.availabilityEditor.EditCommunication(callback, user)
		case data == "avail_edit_frequency":
			return h.availabilityEditor.EditFrequency(callback, user)
		case data == "avail_save_changes":
			return h.availabilityEditor.SaveChanges(callback, user)
		case data == "avail_cancel_edit":
			return h.availabilityEditor.CancelEdit(callback, user)
		case data == "avail_back_to_edit_menu":
			session, err := h.availabilityEditor.GetEditSession(user.ID)
			if err != nil {
				return err
			}
			return h.availabilityEditor.ShowEditMenu(callback, session, user)

		// Обработка выбора типа дней
		case strings.HasPrefix(data, "avail_edit_daytype_"):
			dayType := strings.TrimPrefix(data, "avail_edit_daytype_")
			return h.availabilityEditor.HandleDayTypeSelection(callback, user, dayType)

		// Обработка выбора конкретных дней
		case strings.HasPrefix(data, "avail_edit_day_"):
			day := strings.TrimPrefix(data, "avail_edit_day_")
			return h.availabilityEditor.ToggleSpecificDay(callback, user, day)

		// Обработка применения дней
		case data == "avail_apply_days":
			session, err := h.availabilityEditor.GetEditSession(user.ID)
			if err != nil {
				return err
			}
			return h.availabilityEditor.ShowEditMenu(callback, session, user)

		// Обработка выбора временных слотов
		case strings.HasPrefix(data, "avail_edit_timeslot_"):
			slot := strings.TrimPrefix(data, "avail_edit_timeslot_")
			return h.availabilityEditor.ToggleTimeSlot(callback, user, slot)

		// Обработка применения времени
		case data == "avail_apply_time":
			session, err := h.availabilityEditor.GetEditSession(user.ID)
			if err != nil {
				return err
			}
			return h.availabilityEditor.ShowEditMenu(callback, session, user)

		// Обработка выбора стилей общения
		case strings.HasPrefix(data, "avail_edit_commstyle_"):
			style := strings.TrimPrefix(data, "avail_edit_commstyle_")
			return h.availabilityEditor.ToggleCommunicationStyle(callback, user, style)

		// Обработка применения общения
		case data == "avail_apply_communication":
			session, err := h.availabilityEditor.GetEditSession(user.ID)
			if err != nil {
				return err
			}
			return h.availabilityEditor.ShowEditMenu(callback, session, user)

		// Обработка выбора частоты
		case strings.HasPrefix(data, "avail_edit_freq_"):
			freq := strings.TrimPrefix(data, "avail_edit_freq_")
			return h.availabilityEditor.HandleFrequencySelection(callback, user, freq)

		// Обработка setup flow
		case data == "availability_proceed_to_time":
			return h.availabilityHandler.ShowTimeSlotSelection(callback, user)
		case data == "availability_proceed_to_communication":
			return h.availabilityHandler.HandleFriendshipPreferencesStart(callback, user)
		case data == "availability_proceed_to_frequency":
			log.Printf("DEBUG: Processing availability_proceed_to_frequency for user %d", user.ID)
			return h.availabilityHandler.CompleteAvailabilitySetup(callback, user)

		// Обработка кнопок "Назад"
		case data == "availability_back_to_daytype":
			// Возврат к началу setup (выбор типа дней)
			return h.availabilityHandler.HandleTimeAvailabilityStart(callback, user)
		case data == "availability_back_to_days":
			// Возврат к выбору типа дней (то же что и начало setup)
			return h.availabilityHandler.HandleTimeAvailabilityStart(callback, user)
		case data == "availability_back_to_time":
			// Возврат к выбору времени - проверяем, есть ли данные
			cacheKey := fmt.Sprintf("availability_setup:%d", user.ID)
			err := h.service.Cache.Get(context.Background(), cacheKey, nil)
			if err != nil {
				// Если данных нет, начинаем setup заново
				return h.availabilityHandler.HandleTimeAvailabilityStart(callback, user)
			}
			return h.availabilityHandler.ShowTimeSlotSelection(callback, user)

		// Обработка выбора дней в setup
		case strings.HasPrefix(data, "availability_specific_day_"):
			day := strings.TrimPrefix(data, "availability_specific_day_")
			return h.availabilityHandler.HandleSpecificDaysSelection(callback, user, day)

		// Обработка выбора времени в setup
		case strings.HasPrefix(data, "availability_timeslot_"):
			slot := strings.TrimPrefix(data, "availability_timeslot_")
			if slot == "select_all" {
				return h.availabilityHandler.HandleSelectAllTimeSlots(callback, user)
			}
			return h.availabilityHandler.HandleTimeSlotSelection(callback, user, slot)

		// Обработка выбора типа дней в setup
		case strings.HasPrefix(data, "availability_daytype_"):
			dayType := strings.TrimPrefix(data, "availability_daytype_")
			return h.availabilityHandler.HandleDayTypeSelection(callback, user, dayType)

		// Обработка кнопки "выбрать всё" для способов общения
		case data == "availability_communication_select_all":
			return h.availabilityHandler.HandleSelectAllCommunication(callback, user)

		// Обработка выбора индивидуальных способов общения
		case strings.HasPrefix(data, "availability_communication_"):
			style := strings.TrimPrefix(data, "availability_communication_")
			return h.availabilityHandler.HandleCommunicationStyleSelection(callback, user, style)
		case data == "view_profile":
			// Показать профиль пользователя
			return h.menuHandler.HandleMainViewProfile(callback, user, h.profileHandler)
		case data == "back_to_main_menu":
			// Возврат в главное меню
			return h.menuHandler.HandleBackToMainMenu(callback, user)
		}

		return nil // Callback не обработан этой функцией
	}

	// Если callback не связан с availability, возвращаем nil для продолжения обработки другими handlers
	return nil
}

func (h *TelegramHandler) handleMenuCallbacks(callback *tgbotapi.CallbackQuery, user *models.User, data string) error {
	switch data {
	case "main_change_language":
		return h.menuHandler.HandleMainChangeLanguage(callback, user)
	case "main_view_profile":
		return h.menuHandler.HandleMainViewProfile(callback, user, h.profileHandler)
	case "main_edit_profile":
		return h.menuHandler.HandleMainEditProfile(callback, user, h.profileHandler)
	case "main_feedback":
		return h.menuHandler.HandleMainFeedback(callback, user, h.feedbackHandler)
	case "feedback_help":
		return h.menuHandler.HandleFeedbackHelp(callback, user)
	case "start_profile_setup":
		return h.profileHandler.StartProfileSetup(callback, user)
	case "show_profile_setup_features":
		return h.profileHandler.ShowProfileSetupFeatures(callback, user)
	case "profile_setup_continue":
		return h.profileHandler.StartProfileSetup(callback, user)
	case "back_to_main_menu":
		return h.menuHandler.HandleBackToMainMenu(callback, user)
	}

	return nil
}

// handleFeedbackCallbacks обрабатывает callback'и связанные с отзывами.
func (h *TelegramHandler) handleFeedbackCallbacks(callback *tgbotapi.CallbackQuery, user *models.User, data string) error {
	// Проверяем права администратора для доступа к отзывам
	if !h.isAdmin(callback.From.ID, callback.From.UserName) {
		// Если это не администратор, игнорируем callback
		return nil
	}

	switch {
	case strings.HasPrefix(data, "fb_process_"):
		feedbackIDStr := strings.TrimPrefix(data, "fb_process_")

		return h.feedbackHandler.HandleFeedbackProcess(callback, user, feedbackIDStr)
	case strings.HasPrefix(data, "fb_unprocess_"):
		feedbackIDStr := strings.TrimPrefix(data, "fb_unprocess_")

		return h.feedbackHandler.HandleFeedbackUnprocess(callback, user, feedbackIDStr)
	case strings.HasPrefix(data, "fb_delete_"):
		feedbackIDStr := strings.TrimPrefix(data, "fb_delete_")

		return h.feedbackHandler.HandleFeedbackDelete(callback, user, feedbackIDStr)
	case strings.HasPrefix(data, "browse_active_feedbacks_"):
		indexStr := strings.TrimPrefix(data, "browse_active_feedbacks_")

		return h.feedbackHandler.HandleBrowseActiveFeedbacks(callback, user, indexStr)
	case strings.HasPrefix(data, "browse_archive_feedbacks_"):
		indexStr := strings.TrimPrefix(data, "browse_archive_feedbacks_")

		return h.feedbackHandler.HandleBrowseArchiveFeedbacks(callback, user, indexStr)
	case strings.HasPrefix(data, "browse_all_feedbacks_"):
		indexStr := strings.TrimPrefix(data, "browse_all_feedbacks_")

		return h.feedbackHandler.HandleBrowseAllFeedbacks(callback, user, indexStr)
	case strings.HasPrefix(data, "feedback_prev_"):
		parts := strings.TrimPrefix(data, "feedback_prev_")

		indexAndType := strings.Split(parts, "_")
		if len(indexAndType) == localization.MinPartsForFeedbackNav {
			return h.feedbackHandler.HandleFeedbackPrev(callback, user, indexAndType[0], indexAndType[1])
		}

		return nil
	case strings.HasPrefix(data, "feedback_next_"):
		parts := strings.TrimPrefix(data, "feedback_next_")

		indexAndType := strings.Split(parts, "_")
		if len(indexAndType) == localization.MinPartsForFeedbackNav {
			return h.feedbackHandler.HandleFeedbackNext(callback, user, indexAndType[0], indexAndType[1])
		}

		return nil
	case strings.HasPrefix(data, "feedback_back_"):
		feedbackType := strings.TrimPrefix(data, "feedback_back_")

		return h.feedbackHandler.HandleFeedbackBack(callback, user, feedbackType)
	case data == "show_active_feedbacks":
		return h.feedbackHandler.HandleShowActiveFeedbacks(callback, user)
	case data == "show_archive_feedbacks":
		return h.feedbackHandler.HandleShowArchiveFeedbacks(callback, user)
	case data == "show_all_feedbacks":
		return h.feedbackHandler.HandleShowAllFeedbacks(callback, user)
	case strings.HasPrefix(data, "nav_"):
		parts := strings.Split(data, "_")
		if len(parts) >= localization.MinPartsForNav {
			feedbackType := parts[1] // active, archive, all
			indexStr := parts[3]     // 0, 1, 2, etc.

			return h.feedbackHandler.HandleNavigateFeedback(callback, user, feedbackType, indexStr)
		}

		return nil
	case strings.HasPrefix(data, "archive_feedback_"):
		// Обработка архивирования: archive_feedback_0
		indexStr := strings.TrimPrefix(data, "archive_feedback_")

		return h.feedbackHandler.HandleArchiveFeedback(callback, user, indexStr)
	case strings.HasPrefix(data, "back_to_active_feedbacks") ||
		strings.HasPrefix(data, "back_to_archive_feedbacks") ||
		strings.HasPrefix(data, "back_to_all_feedbacks"):
		// Обработка возврата к списку отзывов: back_to_active_feedbacks, back_to_archive_feedbacks, etc.
		parts := strings.Split(data, "_")
		if len(parts) >= localization.MinPartsForNav {
			feedbackType := parts[2] // active, archive, all

			return h.feedbackHandler.HandleBackToFeedbacks(callback, user, feedbackType)
		}

		return nil
	case data == "back_to_feedback_stats":
		// Обработка возврата к статистике отзывов
		return h.feedbackHandler.HandleBackToFeedbackStats(callback, user)
	case strings.HasPrefix(data, "delete_current_feedback_"):
		// Обработка удаления текущего отзыва: delete_current_feedback_0
		indexStr := strings.TrimPrefix(data, "delete_current_feedback_")

		return h.feedbackHandler.HandleDeleteCurrentFeedback(callback, user, indexStr)
	}

	return nil
}

// sendRateLimitMessage отправляет сообщение о превышении лимита запросов.
func (h *TelegramHandler) sendRateLimitMessage(userID int64, err error) {
	// Получаем локализацию (используем английский по умолчанию)
	message := "Too many requests. Please try again later."

	// В будущем можно добавить локализацию для сообщений rate limiting

	// Отправляем сообщение пользователю
	msg := tgbotapi.NewMessage(userID, message)
	msg.ParseMode = localization.ParseModeHTML

	if _, sendErr := h.bot.Send(msg); sendErr != nil {
		// Логируем ошибку отправки, но не возвращаем её
		log.Printf("Failed to send rate limit message to user %d: %v", userID, sendErr)
	}
}

// GetRateLimiterStats возвращает статистику rate limiter'а (для администраторов).
func (h *TelegramHandler) GetRateLimiterStats() map[string]interface{} {
	if h.rateLimiter != nil {
		return h.rateLimiter.GetStats()
	}

	return map[string]interface{}{"error": "rate limiter not initialized"}
}

// Stop останавливает все компоненты handler'а.
func (h *TelegramHandler) Stop() {
	if h.rateLimiter != nil {
		h.rateLimiter.Stop()
	}
}

// SetService устанавливает сервис для handler'а.
func (h *TelegramHandler) SetService(service *core.BotService) {
	h.service = service
}

// SetBotAPI устанавливает BotAPI для handler'а.
func (h *TelegramHandler) SetBotAPI(bot *tgbotapi.BotAPI) {
	h.bot = bot
}

// GetService возвращает сервис handler'а.
func (h *TelegramHandler) GetService() *core.BotService {
	return h.service
}

// GetBotAPI возвращает BotAPI handler'а.
func (h *TelegramHandler) GetBotAPI() *tgbotapi.BotAPI {
	return h.bot
}
