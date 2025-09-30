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

type TelegramHandler struct {
	bot                    *tgbotapi.BotAPI
	service                *core.BotService
	adminChatIDs           []int64  // Chat ID администраторов
	adminUsernames         []string // Usernames администраторов для проверки доступа
	keyboardBuilder        *handlers.KeyboardBuilder
	menuHandler            *handlers.MenuHandler
	profileHandler         *handlers.ProfileHandlerImpl
	feedbackHandler        handlers.FeedbackHandler
	languageHandler        handlers.LanguageHandler
	interestHandler        handlers.NewInterestHandler
	profileInterestHandler *handlers.ProfileInterestHandler
	adminHandler           handlers.AdminHandler
	utilityHandler         handlers.UtilityHandler
	errorHandler           *errors.ErrorHandler
}

// NewTelegramHandler создает новый экземпляр TelegramHandler с базовой конфигурацией
func NewTelegramHandler(bot *tgbotapi.BotAPI, service *core.BotService, adminChatIDs []int64, errorHandler *errors.ErrorHandler) *TelegramHandler {
	keyboardBuilder := handlers.NewKeyboardBuilder(service)
	menuHandler := handlers.NewMenuHandler(bot, service, keyboardBuilder, errorHandler)
	profileHandler := handlers.NewProfileHandler(bot, service, keyboardBuilder, errorHandler)
	feedbackHandler := handlers.NewFeedbackHandler(bot, service, keyboardBuilder, adminChatIDs, make([]string, 0), errorHandler)
	languageHandler := handlers.NewLanguageHandler(service, bot, keyboardBuilder, errorHandler)
	interestService := core.NewInterestService(service.DB.GetConnection())
	interestHandler := handlers.NewNewInterestHandler(service, interestService, bot, keyboardBuilder, errorHandler)
	profileInterestHandler := handlers.NewProfileInterestHandler(service, interestService, bot, keyboardBuilder, errorHandler)
	adminHandler := handlers.NewAdminHandler(service, bot, keyboardBuilder, adminChatIDs, make([]string, 0), errorHandler)
	utilityHandler := handlers.NewUtilityHandler(service, bot, errorHandler)

	return &TelegramHandler{
		bot:                    bot,
		service:                service,
		adminChatIDs:           adminChatIDs,
		adminUsernames:         make([]string, 0), // Пустой список, нет хардкода
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

// NewTelegramHandlerWithAdmins создает новый экземпляр TelegramHandler с полной конфигурацией администраторов
func NewTelegramHandlerWithAdmins(bot *tgbotapi.BotAPI, service *core.BotService, adminChatIDs []int64, adminUsernames []string, errorHandler *errors.ErrorHandler) *TelegramHandler {
	keyboardBuilder := handlers.NewKeyboardBuilder(service)
	menuHandler := handlers.NewMenuHandler(bot, service, keyboardBuilder, errorHandler)
	profileHandler := handlers.NewProfileHandler(bot, service, keyboardBuilder, errorHandler)
	feedbackHandler := handlers.NewFeedbackHandler(bot, service, keyboardBuilder, adminChatIDs, adminUsernames, errorHandler)
	languageHandler := handlers.NewLanguageHandler(service, bot, keyboardBuilder, errorHandler)
	interestService := core.NewInterestService(service.DB.GetConnection())
	interestHandler := handlers.NewNewInterestHandler(service, interestService, bot, keyboardBuilder, errorHandler)
	profileInterestHandler := handlers.NewProfileInterestHandler(service, interestService, bot, keyboardBuilder, errorHandler)
	adminHandler := handlers.NewAdminHandler(service, bot, keyboardBuilder, adminChatIDs, adminUsernames, errorHandler)
	utilityHandler := handlers.NewUtilityHandler(service, bot, errorHandler)

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

// HandleUpdate обрабатывает входящие обновления от Telegram API
func (h *TelegramHandler) HandleUpdate(update tgbotapi.Update) error {
	if update.Message != nil {
		return h.handleMessage(update.Message)
	}
	if update.CallbackQuery != nil {
		return h.handleCallbackQuery(update.CallbackQuery)
	}
	return nil
}

// handleMessage обрабатывает входящие текстовые сообщения
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

// handleCommand обрабатывает команды пользователя (начинающиеся с /)
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
		return h.feedbackHandler.HandleFeedbackCommand(message, user)
	case "feedbacks":
		return h.feedbackHandler.HandleFeedbacksCommand(message, user)
	default:
		log.Printf("Unknown command: %s", message.Command())
		return h.utilityHandler.SendMessage(message.Chat.ID, h.service.Localizer.Get(user.InterfaceLanguageCode, "unknown_command"))
	}
}

// handleState обрабатывает сообщения в зависимости от текущего состояния пользователя
func (h *TelegramHandler) handleState(message *tgbotapi.Message, user *models.User) error {
	switch user.State {
	case models.StateWaitingLanguage,
		models.StateWaitingInterests,
		models.StateWaitingTime:
		return h.utilityHandler.SendMessage(message.Chat.ID, h.service.Localizer.Get(user.InterfaceLanguageCode, "use_menu_above"))
	case models.StateWaitingFeedback:
		return h.feedbackHandler.HandleFeedbackMessage(message, user)
	case models.StateWaitingFeedbackContact:
		return h.feedbackHandler.HandleFeedbackContactMessage(message, user)
	default:
		return h.utilityHandler.SendMessage(message.Chat.ID, h.service.Localizer.Get(user.InterfaceLanguageCode, "unknown_command"))
	}
}

// handleCallbackQuery обрабатывает нажатия на inline-кнопки
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

	switch {
	case strings.HasPrefix(data, "lang_native_"):
		return h.languageHandler.HandleNativeLanguageCallback(callback, user)
	case strings.HasPrefix(data, "lang_target_"):
		return h.languageHandler.HandleTargetLanguageCallback(callback, user)
	case strings.HasPrefix(data, "lang_edit_native_"):
		return h.profileHandler.HandleEditNativeLanguage(callback, user)
	case strings.HasPrefix(data, "lang_edit_target_"):
		return h.profileHandler.HandleEditTargetLanguage(callback, user)
	case strings.HasPrefix(data, "lang_interface_"):
		langCode := strings.TrimPrefix(data, "lang_interface_")
		return h.languageHandler.HandleInterfaceLanguageSelection(callback, user, langCode)
	case data == "back_to_categories":
		return h.interestHandler.HandleBackToCategories(callback, user)
	case data == "back_to_language_level":
		return h.languageHandler.HandleBackToLanguageLevel(callback, user)
	case data == "interests_continue":
		return h.interestHandler.HandleInterestsContinue(callback, user)
	case strings.HasPrefix(data, "interest_category_"):
		categoryKey := strings.TrimPrefix(data, "interest_category_")
		return h.interestHandler.HandleInterestCategorySelection(callback, user, categoryKey)
	case strings.HasPrefix(data, "interest_select_"):
		interestID := strings.TrimPrefix(data, "interest_select_")
		return h.interestHandler.HandleInterestSelection(callback, user, interestID)
	case strings.HasPrefix(data, "primary_interest_"):
		interestID := strings.TrimPrefix(data, "primary_interest_")
		return h.interestHandler.HandlePrimaryInterestSelection(callback, user, interestID)
	case data == "primary_interests_continue":
		return h.interestHandler.HandlePrimaryInterestsContinue(callback, user)
	case data == "back_to_interests":
		return h.interestHandler.HandleBackToInterests(callback, user)
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
	case data == "profile_show":
		return h.profileHandler.HandleProfileShow(callback, user)
	case data == "profile_reset_ask":
		return h.profileHandler.HandleProfileResetAsk(callback, user)
	case data == "profile_reset_yes":
		return h.profileHandler.HandleProfileResetYes(callback, user)
	case data == "profile_reset_no":
		return h.menuHandler.HandleBackToMainMenu(callback, user)
	case data == "languages_continue_filling":
		return h.languageHandler.HandleLanguagesContinueFilling(callback, user)
	case data == "languages_reselect":
		return h.languageHandler.HandleLanguagesReselect(callback, user)
	case strings.HasPrefix(data, "level_"):
		levelCode := strings.TrimPrefix(data, "level_")
		return h.languageHandler.HandleLanguageLevelSelection(callback, user, levelCode)
	case strings.HasPrefix(data, "edit_level_"):
		levelCode := strings.TrimPrefix(data, "edit_level_")
		return h.profileHandler.HandleEditLevelSelection(callback, user, levelCode)
	case data == "back_to_previous_step":
		// Возвращаемся к просмотру профиля
		return h.profileHandler.HandleProfileShow(callback, user)
	case data == "main_change_language":
		return h.menuHandler.HandleMainChangeLanguage(callback, user)
	case data == "main_view_profile":
		return h.menuHandler.HandleMainViewProfile(callback, user, h.profileHandler)
	case data == "main_edit_profile":
		return h.menuHandler.HandleMainEditProfile(callback, user, h.profileHandler)
	case data == "main_feedback":
		return h.menuHandler.HandleMainFeedback(callback, user, h.feedbackHandler)
	case data == "feedback_help":
		return h.menuHandler.HandleFeedbackHelp(callback, user)
	case data == "start_profile_setup":
		return h.profileHandler.StartProfileSetup(callback, user)
	case data == "back_to_main_menu":
		return h.menuHandler.HandleBackToMainMenu(callback, user)
	case data == "edit_interests":
		return h.profileInterestHandler.HandleEditInterestsFromProfile(callback, user)
	case data == "edit_languages":
		return h.profileHandler.HandleEditLanguages(callback, user)
	case data == "edit_native_lang":
		return h.profileHandler.HandleEditNativeLang(callback, user)
	case data == "edit_target_lang":
		return h.profileHandler.HandleEditTargetLang(callback, user)
	case data == "edit_level":
		return h.profileHandler.HandleEditLevelLang(callback, user)
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
		if len(indexAndType) == 2 {
			return h.feedbackHandler.HandleFeedbackPrev(callback, user, indexAndType[0], indexAndType[1])
		}
		return nil
	case strings.HasPrefix(data, "feedback_next_"):
		parts := strings.TrimPrefix(data, "feedback_next_")
		indexAndType := strings.Split(parts, "_")
		if len(indexAndType) == 2 {
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
		// Обработка навигации: nav_active_feedback_0, nav_archive_feedback_1, etc.
		parts := strings.Split(data, "_")
		if len(parts) >= 4 {
			feedbackType := parts[1] // active, archive, all
			indexStr := parts[3]     // 0, 1, 2, etc.
			return h.feedbackHandler.HandleNavigateFeedback(callback, user, feedbackType, indexStr)
		}
		return nil
	case strings.HasPrefix(data, "archive_feedback_"):
		// Обработка архивирования: archive_feedback_0
		indexStr := strings.TrimPrefix(data, "archive_feedback_")
		return h.feedbackHandler.HandleArchiveFeedback(callback, user, indexStr)
	case strings.HasPrefix(data, "back_to_"):
		// Обработка возврата к списку: back_to_active_feedbacks, back_to_archive_feedbacks, etc.
		parts := strings.Split(data, "_")
		if len(parts) >= 4 {
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
	case data == "delete_all_archive_feedbacks":
		// Обработка удаления всех обработанных отзывов
		return h.feedbackHandler.HandleDeleteAllArchiveFeedbacks(callback, user)
	case data == "confirm_delete_all_archive":
		// Обработка подтверждения удаления всех обработанных отзывов
		return h.feedbackHandler.HandleConfirmDeleteAllArchive(callback, user)
	case data == "back_to_archive_feedbacks":
		// Обработка возврата к архивным отзывам
		return h.feedbackHandler.HandleShowArchiveFeedbacks(callback, user)
	case strings.HasPrefix(data, "unarchive_feedback_"):
		// Обработка возврата отзыва в активные: unarchive_feedback_0
		indexStr := strings.TrimPrefix(data, "unarchive_feedback_")
		return h.feedbackHandler.HandleUnarchiveFeedback(callback, user, indexStr)
	default:
		return nil
	}
}

// handleMainViewProfile делегирует просмотр профиля в menu handler
func (h *TelegramHandler) handleMainViewProfile(callback *tgbotapi.CallbackQuery, user *models.User) error {
	return h.menuHandler.HandleMainViewProfile(callback, user, h.profileHandler)
}

// handleMainEditProfile делегирует редактирование профиля в menu handler
func (h *TelegramHandler) handleMainEditProfile(callback *tgbotapi.CallbackQuery, user *models.User) error {
	return h.menuHandler.HandleMainEditProfile(callback, user, h.profileHandler)
}

// handleMainFeedback делегирует работу с отзывами в feedback handler
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
