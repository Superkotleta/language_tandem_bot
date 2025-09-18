package telegram

import (
	"context"
	"strings"
	"time"

	"language-exchange-bot/internal/adapters/telegram/handlers"
	"language-exchange-bot/internal/cache"
	"language-exchange-bot/internal/core"
	"language-exchange-bot/internal/logging"
	"language-exchange-bot/internal/models"
	"language-exchange-bot/internal/monitoring"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// OptimizedTelegramHandler оптимизированный обработчик Telegram обновлений.
type OptimizedTelegramHandler struct {
	bot             *tgbotapi.BotAPI
	service         *core.OptimizedBotService
	cache           cache.Cache
	metrics         *monitoring.Metrics
	logger          logging.Logger
	adminChatIDs    []int64
	adminUsernames  []string
	keyboardBuilder *handlers.KeyboardBuilder
	menuHandler     *handlers.MenuHandler
	profileHandler  *handlers.ProfileHandlerImpl
	feedbackHandler handlers.FeedbackHandler
	languageHandler handlers.LanguageHandler
	interestHandler handlers.InterestHandler
	adminHandler    handlers.AdminHandler
	utilityHandler  handlers.UtilityHandler
}

// NewOptimizedTelegramHandler создает оптимизированный обработчик.
func NewOptimizedTelegramHandler(
	bot *tgbotapi.BotAPI,
	service *core.OptimizedBotService,
	cacheInstance cache.Cache,
	metrics *monitoring.Metrics,
	logger logging.Logger,
	adminChatIDs []int64,
	adminUsernames []string,
) *OptimizedTelegramHandler {
	// Создаем адаптер для совместимости с существующими хендлерами
	// Для упрощения используем nil, так как хендлеры будут работать напрямую с OptimizedBotService
	legacyService := &core.BotService{
		DB:        nil, // Будем использовать OptimizedBotService напрямую
		Localizer: service.GetLocalizer(),
	}

	keyboardBuilder := handlers.NewKeyboardBuilder(legacyService)
	menuHandler := handlers.NewMenuHandler(bot, legacyService, keyboardBuilder)
	profileHandler := handlers.NewProfileHandler(bot, legacyService, keyboardBuilder)
	feedbackHandler := handlers.NewFeedbackHandler(bot, legacyService, keyboardBuilder, adminChatIDs, adminUsernames)
	languageHandler := handlers.NewLanguageHandler(legacyService, bot, keyboardBuilder)
	interestHandler := handlers.NewInterestHandler(legacyService, bot, keyboardBuilder)
	adminHandler := handlers.NewAdminHandler(legacyService, bot, keyboardBuilder, adminChatIDs, adminUsernames)
	utilityHandler := handlers.NewUtilityHandler(legacyService, bot)

	return &OptimizedTelegramHandler{
		bot:             bot,
		service:         service,
		cache:           cacheInstance,
		metrics:         metrics,
		logger:          logger,
		adminChatIDs:    adminChatIDs,
		adminUsernames:  adminUsernames,
		keyboardBuilder: keyboardBuilder,
		menuHandler:     menuHandler,
		profileHandler:  profileHandler,
		feedbackHandler: feedbackHandler,
		languageHandler: languageHandler,
		interestHandler: interestHandler,
		adminHandler:    adminHandler,
		utilityHandler:  utilityHandler,
	}
}

// HandleUpdate обрабатывает входящие обновления от Telegram API.
func (h *OptimizedTelegramHandler) HandleUpdate(update tgbotapi.Update) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	start := time.Now()
	defer func() {
		h.metrics.RecordHTTPRequest("telegram", "update", "success", time.Since(start))
	}()

	if update.Message != nil {
		return h.handleMessage(ctx, update.Message)
	}
	if update.CallbackQuery != nil {
		return h.handleCallbackQuery(ctx, update.CallbackQuery)
	}
	return nil
}

// handleMessage обрабатывает входящие текстовые сообщения.
func (h *OptimizedTelegramHandler) handleMessage(ctx context.Context, message *tgbotapi.Message) error {
	// Используем оптимизированный сервис для регистрации пользователя
	user, err := h.service.HandleUserRegistration(
		message.From.ID,
		message.From.UserName,
		message.From.FirstName,
		message.From.LanguageCode,
	)
	if err != nil {
		h.logger.Error("Error handling user registration",
			logging.Int64("user_id", message.From.ID),
			logging.ErrorField(err),
		)
		return err
	}

	// Записываем метрику регистрации пользователя
	h.metrics.RecordUserRegistration()

	if message.IsCommand() {
		return h.handleCommand(ctx, message, user)
	}
	return h.handleState(ctx, message, user)
}

// handleCommand обрабатывает команды пользователя.
func (h *OptimizedTelegramHandler) handleCommand(ctx context.Context, message *tgbotapi.Message, user *models.User) error {
	start := time.Now()
	defer func() {
		h.metrics.RecordHTTPRequest("telegram", "command", "success", time.Since(start))
	}()

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
		h.logger.Warn("Unknown command received",
			logging.String("command", message.Command()),
			logging.Int64("user_id", message.From.ID),
		)
		return h.utilityHandler.SendMessage(message.Chat.ID, h.service.GetLocalizer().Get(user.InterfaceLanguageCode, "unknown_command"))
	}
}

// handleState обрабатывает сообщения в зависимости от состояния пользователя.
func (h *OptimizedTelegramHandler) handleState(ctx context.Context, message *tgbotapi.Message, user *models.User) error {
	switch user.State {
	case models.StateWaitingLanguage,
		models.StateWaitingInterests,
		models.StateWaitingTime:
		return h.utilityHandler.SendMessage(message.Chat.ID, h.service.GetLocalizer().Get(user.InterfaceLanguageCode, "use_menu_above"))
	case models.StateWaitingFeedback:
		return h.feedbackHandler.HandleFeedbackMessage(message, user)
	case models.StateWaitingFeedbackContact:
		return h.feedbackHandler.HandleFeedbackContactMessage(message, user)
	default:
		return h.utilityHandler.SendMessage(message.Chat.ID, h.service.GetLocalizer().Get(user.InterfaceLanguageCode, "unknown_command"))
	}
}

// handleCallbackQuery обрабатывает нажатия на inline-кнопки.
func (h *OptimizedTelegramHandler) handleCallbackQuery(ctx context.Context, callback *tgbotapi.CallbackQuery) error {
	start := time.Now()
	defer func() {
		h.metrics.RecordHTTPRequest("telegram", "callback", "success", time.Since(start))
	}()

	// Используем оптимизированный сервис для регистрации пользователя
	user, err := h.service.HandleUserRegistration(
		callback.From.ID,
		callback.From.UserName,
		callback.From.FirstName,
		callback.From.LanguageCode,
	)
	if err != nil {
		h.logger.Error("Error handling user registration in callback",
			logging.Int64("user_id", callback.From.ID),
			logging.ErrorField(err),
		)
		return err
	}

	data := callback.Data
	_, _ = h.bot.Request(tgbotapi.NewCallback(callback.ID, ""))

	// Обрабатываем callback в зависимости от типа
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
	case strings.HasPrefix(data, "interest_"):
		interestID := strings.TrimPrefix(data, "interest_")
		return h.interestHandler.HandleInterestSelection(callback, user, interestID)
	case strings.HasPrefix(data, "edit_interest_"):
		interestID := strings.TrimPrefix(data, "edit_interest_")
		return h.profileHandler.HandleEditInterestSelection(callback, user, interestID)
	case data == "profile_show":
		return h.profileHandler.HandleProfileShow(callback, user)
	case data == "profile_reset_ask":
		return h.profileHandler.HandleProfileResetAsk(callback, user)
	case data == "profile_reset_yes":
		return h.profileHandler.HandleProfileResetYes(callback, user)
	case data == "profile_reset_no":
		return h.menuHandler.HandleBackToMainMenu(callback, user)
	case data == "interests_continue":
		return h.interestHandler.HandleInterestsContinue(callback, user)
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
		return h.profileHandler.HandleProfileShow(callback, user)
	case data == "main_change_language":
		return h.menuHandler.HandleMainChangeLanguage(callback, user)
	case data == "main_view_profile":
		return h.menuHandler.HandleMainViewProfile(callback, user, h.profileHandler)
	case data == "main_edit_profile":
		return h.menuHandler.HandleMainEditProfile(callback, user, h.profileHandler)
	case data == "main_feedback":
		return h.feedbackHandler.HandleMainFeedback(callback, user)
	case data == "start_profile_setup":
		return h.profileHandler.StartProfileSetup(callback, user)
	case data == "back_to_main_menu":
		return h.menuHandler.HandleBackToMainMenu(callback, user)
	case data == "edit_interests":
		return h.profileHandler.HandleEditInterests(callback, user)
	case data == "edit_languages":
		return h.profileHandler.HandleEditLanguages(callback, user)
	case data == "save_edits":
		return h.profileHandler.HandleSaveEdits(callback, user)
	case data == "cancel_edits":
		return h.profileHandler.HandleCancelEdits(callback, user)
	case data == "edit_native_lang":
		return h.profileHandler.HandleEditNativeLang(callback, user)
	case data == "edit_target_lang":
		return h.profileHandler.HandleEditTargetLang(callback, user)
	case data == "edit_level":
		return h.profileHandler.HandleEditLevelLang(callback, user)
	// Обработка отзывов
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
		parts := strings.Split(data, "_")
		if len(parts) >= 4 {
			feedbackType := parts[1]
			indexStr := parts[3]
			return h.feedbackHandler.HandleNavigateFeedback(callback, user, feedbackType, indexStr)
		}
		return nil
	case strings.HasPrefix(data, "archive_feedback_"):
		indexStr := strings.TrimPrefix(data, "archive_feedback_")
		return h.feedbackHandler.HandleArchiveFeedback(callback, user, indexStr)
	case strings.HasPrefix(data, "back_to_"):
		parts := strings.Split(data, "_")
		if len(parts) >= 4 {
			feedbackType := parts[2]
			return h.feedbackHandler.HandleBackToFeedbacks(callback, user, feedbackType)
		}
		return nil
	case data == "back_to_feedback_stats":
		return h.feedbackHandler.HandleBackToFeedbackStats(callback, user)
	case strings.HasPrefix(data, "delete_current_feedback_"):
		indexStr := strings.TrimPrefix(data, "delete_current_feedback_")
		return h.feedbackHandler.HandleDeleteCurrentFeedback(callback, user, indexStr)
	case data == "delete_all_archive_feedbacks":
		return h.feedbackHandler.HandleDeleteAllArchiveFeedbacks(callback, user)
	case data == "confirm_delete_all_archive":
		return h.feedbackHandler.HandleConfirmDeleteAllArchive(callback, user)
	case data == "back_to_archive_feedbacks":
		return h.feedbackHandler.HandleShowArchiveFeedbacks(callback, user)
	case strings.HasPrefix(data, "unarchive_feedback_"):
		indexStr := strings.TrimPrefix(data, "unarchive_feedback_")
		return h.feedbackHandler.HandleUnarchiveFeedback(callback, user, indexStr)
	default:
		h.logger.Debug("Unknown callback data",
			logging.String("data", data),
			logging.Int64("user_id", callback.From.ID),
		)
		return nil
	}
}

// HealthCheck проверяет здоровье обработчика.
func (h *OptimizedTelegramHandler) HealthCheck() error {
	// Проверяем здоровье сервиса
	if err := h.service.HealthCheck(); err != nil {
		return err
	}

	// Проверяем здоровье кэша
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := h.cache.Set(ctx, "health_check", "ok", time.Second); err != nil {
		return err
	}

	return nil
}
