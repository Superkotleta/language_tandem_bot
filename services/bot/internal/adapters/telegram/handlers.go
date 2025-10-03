package telegram

import (
	"log"
	"strings"

	"language-exchange-bot/internal/adapters/telegram/handlers"
	"language-exchange-bot/internal/core"
	"language-exchange-bot/internal/errors"
	"language-exchange-bot/internal/models"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Константы для работы с коллбэками и массивами.
const (
	MinPartsForFeedbackNav = 2 // Минимальное количество частей для навигации по отзывам
	MinPartsForNav         = 4 // Минимальное количество частей для навигации

)

// TelegramHandler handles Telegram bot message and callback processing.
// The name includes "Telegram" prefix for clarity, even though it may cause stuttering with the package name.
type TelegramHandler struct {
	bot                    *tgbotapi.BotAPI
	service                *core.BotService
	adminChatIDs           []int64  // Chat ID администраторов
	adminUsernames         []string // Usernames администраторов для проверки доступа
	keyboardBuilder        *handlers.KeyboardBuilder
	menuHandler            *handlers.MenuHandler
	profileHandler         *handlers.ProfileHandlerImpl
	feedbackHandler        *handlers.FeedbackHandlerImpl
	languageHandler        *handlers.LanguageHandlerImpl
	interestHandler        *handlers.NewInterestHandlerImpl
	profileInterestHandler *handlers.ProfileInterestHandler
	adminHandler           *handlers.AdminHandlerImpl
	utilityHandler         *handlers.UtilityHandlerImpl
	errorHandler           *errors.ErrorHandler
}

// NewTelegramHandler создает новый экземпляр TelegramHandler с базовой конфигурацией.
func NewTelegramHandler(
	bot *tgbotapi.BotAPI,
	service *core.BotService,
	adminChatIDs []int64,
	errorHandler *errors.ErrorHandler,
) *TelegramHandler {
	keyboardBuilder := handlers.NewKeyboardBuilder(service)
	menuHandler := handlers.NewMenuHandler(bot, service, keyboardBuilder, errorHandler)
	profileHandler := handlers.NewProfileHandler(bot, service, keyboardBuilder, errorHandler)
	feedbackHandler := handlers.NewFeedbackHandler(
		bot,
		service,
		keyboardBuilder,
		adminChatIDs,
		make([]string, 0),
		errorHandler,
	)
	languageHandler := handlers.NewLanguageHandler(service, bot, keyboardBuilder, errorHandler)
	interestService := core.NewInterestService(service.DB.GetConnection())
	interestHandler := handlers.NewNewInterestHandler(service, interestService, bot, keyboardBuilder, errorHandler)
	profileInterestHandler := handlers.NewProfileInterestHandler(
		service,
		interestService,
		bot,
		keyboardBuilder,
		errorHandler,
	)
	adminHandler := handlers.NewAdminHandler(service, bot, keyboardBuilder, adminChatIDs, make([]string, 0), errorHandler)
	utilityHandler := handlers.NewUtilityHandler(service, bot, errorHandler)

	return &TelegramHandler{
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
		adminHandler:           adminHandler,
		utilityHandler:         utilityHandler,
		errorHandler:           errorHandler,
	}
}

// NewTelegramHandlerWithAdmins создает новый экземпляр TelegramHandler с полной конфигурацией администраторов.
func NewTelegramHandlerWithAdmins(
	bot *tgbotapi.BotAPI,
	service *core.BotService,
	adminChatIDs []int64,
	adminUsernames []string,
	errorHandler *errors.ErrorHandler,
) *TelegramHandler {
	keyboardBuilder := handlers.NewKeyboardBuilder(service)
	menuHandler := handlers.NewMenuHandler(
		bot,
		service,
		keyboardBuilder,
		errorHandler,
	)
	profileHandler := handlers.NewProfileHandler(bot, service, keyboardBuilder, errorHandler)
	feedbackHandler := handlers.NewFeedbackHandler(
		bot,
		service,
		keyboardBuilder,
		adminChatIDs,
		adminUsernames,
		errorHandler,
	)
	languageHandler := handlers.NewLanguageHandler(service, bot, keyboardBuilder, errorHandler)
	interestService := core.NewInterestService(service.DB.GetConnection())
	interestHandler := handlers.NewNewInterestHandler(
		service,
		interestService,
		bot,
		keyboardBuilder,
		errorHandler,
	)
	profileInterestHandler := handlers.NewProfileInterestHandler(
		service,
		interestService,
		bot,
		keyboardBuilder,
		errorHandler,
	)
	adminHandler := handlers.NewAdminHandler(
		service,
		bot,
		keyboardBuilder,
		adminChatIDs,
		adminUsernames,
		errorHandler,
	)
	utilityHandler := handlers.NewUtilityHandler(
		service,
		bot,
		errorHandler,
	)

	return &TelegramHandler{
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
		adminHandler:           adminHandler,
		utilityHandler:         utilityHandler,
		errorHandler:           errorHandler,
	}
}

// HandleUpdate обрабатывает входящие обновления от Telegram API.
func (h *TelegramHandler) HandleUpdate(update tgbotapi.Update) error {
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
	switch user.State {
	case models.StateWaitingLanguage,
		models.StateWaitingInterests,
		models.StateWaitingTime:
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
	user, err := h.service.HandleUserRegistration(
		callback.From.ID,
		callback.From.UserName,
		callback.From.FirstName,
		callback.From.LanguageCode,
	)
	if err != nil {
		log.Printf("Error handling user registration: %v", err)

		return err
	}

	data := callback.Data
	_, _ = h.bot.Request(tgbotapi.NewCallback(callback.ID, ""))

	// Разделяем обработку callback'ов по категориям для уменьшения сложности
	if err := h.handleLanguageCallbacks(callback, user, data); err != nil {
		return err
	}

	if err := h.handleInterestCallbacks(callback, user, data); err != nil {
		return err
	}

	if err := h.handleProfileCallbacks(
		callback,
		user,
		data,
	); err != nil {
		return err
	}

	if err := h.handleMenuCallbacks(
		callback,
		user,
		data,
	); err != nil {
		return err
	}

	if err := h.handleFeedbackCallbacks(callback, user, data); err != nil {
		return err
	}

	// Если callback не был обработан ни одним обработчиком, просто игнорируем
	return nil
}

// isAdmin проверяет, является ли пользователь администратором
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

// handleMainViewProfile делегирует просмотр профиля в menu handler
// TODO: функция может быть использована в будущем для обработки просмотра профиля
//
//nolint:unused
func (h *TelegramHandler) handleMainViewProfile(callback *tgbotapi.CallbackQuery, user *models.User) error {
	return h.menuHandler.HandleMainViewProfile(callback, user, h.profileHandler)
}

// handleMainEditProfile делегирует редактирование профиля в menu handler
// TODO: функция может быть использована в будущем для обработки редактирования профиля
//
//nolint:unused
func (h *TelegramHandler) handleMainEditProfile(callback *tgbotapi.CallbackQuery, user *models.User) error {
	return h.menuHandler.HandleMainEditProfile(callback, user, h.profileHandler)
}

// handleMainFeedback делегирует работу с отзывами в feedback handler
// TODO: функция может быть использована в будущем для обработки обратной связи
//
//nolint:unused
func (h *TelegramHandler) handleMainFeedback(callback *tgbotapi.CallbackQuery, user *models.User) error {
	// Создаем message объект для handleFeedbackCommand
	message := &tgbotapi.Message{
		Chat: callback.Message.Chat,
	}

	return h.feedbackHandler.HandleFeedbackCommand(message, user)
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
	case strings.HasPrefix(data, "edit_level_"):
		levelCode := strings.TrimPrefix(data, "edit_level_")

		return h.profileHandler.HandleEditLevelSelection(
			callback,
			user,
			levelCode,
		)
	case strings.HasPrefix(data, "lang_edit_native_"):
		return h.profileHandler.HandleEditNativeLanguage(callback, user)
	case strings.HasPrefix(data, "lang_edit_target_"):
		return h.profileHandler.HandleEditTargetLanguage(callback, user)
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
	// Обработка редактирования интересов
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
	}

	return nil
}

func (h *TelegramHandler) handleProfileCommands(callback *tgbotapi.CallbackQuery, user *models.User, data string) error {
	switch data {
	case "profile_show":
		return h.profileHandler.HandleProfileShow(callback, user)
	case "profile_reset_ask":
		return h.profileHandler.HandleProfileResetAsk(callback, user)
	case "profile_reset_yes":
		return h.profileHandler.HandleProfileResetYes(callback, user)
	case "profile_reset_no":
		return h.menuHandler.HandleBackToMainMenu(callback, user)
	case "back_to_previous_step":
		return h.profileHandler.HandleProfileShow(callback, user)
	case "edit_interests":
		return h.profileInterestHandler.HandleEditInterestsFromProfile(callback, user)
	case "edit_languages":
		return h.profileHandler.HandleEditLanguages(callback, user)
	case "edit_native_lang":
		return h.profileHandler.HandleEditNativeLang(callback, user)
	case "edit_target_lang":
		return h.profileHandler.HandleEditTargetLang(callback, user)
	case "edit_level":
		return h.profileHandler.HandleEditLevelLang(callback, user)
	}

	return nil
}

// handleMenuCallbacks обрабатывает callback'и связанные с меню.
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
		if len(indexAndType) == MinPartsForFeedbackNav {
			return h.feedbackHandler.HandleFeedbackPrev(callback, user, indexAndType[0], indexAndType[1])
		}

		return nil
	case strings.HasPrefix(data, "feedback_next_"):
		parts := strings.TrimPrefix(data, "feedback_next_")

		indexAndType := strings.Split(parts, "_")
		if len(indexAndType) == MinPartsForFeedbackNav {
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
		if len(parts) >= MinPartsForNav {
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
		if len(parts) >= MinPartsForNav {
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
