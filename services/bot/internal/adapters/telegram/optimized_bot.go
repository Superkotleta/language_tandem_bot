package telegram

import (
	"context"
	"fmt"
	"time"

	"language-exchange-bot/internal/cache"
	"language-exchange-bot/internal/core"
	"language-exchange-bot/internal/database"
	"language-exchange-bot/internal/logging"
	"language-exchange-bot/internal/monitoring"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// OptimizedTelegramBot оптимизированная версия Telegram бота.
type OptimizedTelegramBot struct {
	api            *tgbotapi.BotAPI
	service        *core.OptimizedBotService
	cache          cache.Cache
	metrics        *monitoring.Metrics
	logger         logging.Logger
	debug          bool
	adminChatIDs   []int64
	adminUsernames []string
	handler        *OptimizedTelegramHandler
}

// NewOptimizedTelegramBot создает оптимизированный Telegram бот.
func NewOptimizedTelegramBot(
	token string,
	db *database.OptimizedDB,
	cacheInstance cache.Cache,
	metrics *monitoring.Metrics,
	logger logging.Logger,
	debug bool,
	adminChatIDs []int64,
	adminUsernames []string,
) (*OptimizedTelegramBot, error) {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, fmt.Errorf("failed to create telegram bot: %w", err)
	}
	bot.Debug = debug

	// Создаем оптимизированный сервис
	service := core.NewOptimizedBotService(db, cacheInstance)

	// Создаем оптимизированный хендлер
	handler := NewOptimizedTelegramHandler(bot, service, cacheInstance, metrics, logger, adminChatIDs, adminUsernames)

	return &OptimizedTelegramBot{
		api:            bot,
		service:        service,
		cache:          cacheInstance,
		metrics:        metrics,
		logger:         logger,
		debug:          debug,
		adminChatIDs:   adminChatIDs,
		adminUsernames: adminUsernames,
		handler:        handler,
	}, nil
}

// Start запускает оптимизированный бот.
func (tb *OptimizedTelegramBot) Start(ctx context.Context) error {
	tb.logger.Info("Starting optimized Telegram bot",
		logging.String("bot_username", tb.api.Self.UserName),
		logging.Int("admin_count", len(tb.adminChatIDs)),
	)

	// Автоматическое переключение между режимами
	// В development используем polling, в production - webhook
	if tb.debug {
		tb.logger.Info("Starting in polling mode (development)")
		return tb.startPolling(ctx)
	}

	tb.logger.Info("Starting in webhook mode (production)")
	return tb.startWebhook(ctx)
}

// startPolling запускает бот в режиме polling (для development).
func (tb *OptimizedTelegramBot) startPolling(ctx context.Context) error {
	tb.logger.Info("Starting bot in polling mode")

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := tb.api.GetUpdatesChan(u)

	for {
		select {
		case update := <-updates:
			go func(upd tgbotapi.Update) {
				start := time.Now()
				if err := tb.handler.HandleUpdate(upd); err != nil {
					tb.logger.Error("Error handling update",
						logging.ErrorField(err),
						logging.Duration("duration", time.Since(start)),
					)
					tb.metrics.RecordHTTPRequest("telegram", "update", "error", time.Since(start))
				} else {
					tb.metrics.RecordHTTPRequest("telegram", "update", "success", time.Since(start))
				}
			}(update)
		case <-ctx.Done():
			tb.logger.Info("Stopping Telegram bot...")
			tb.api.StopReceivingUpdates()
			return nil
		}
	}
}

// startWebhook запускает бот в режиме webhook (для production).
func (tb *OptimizedTelegramBot) startWebhook(ctx context.Context) error {
	tb.logger.Info("Starting bot in webhook mode")

	// В production режиме webhook настраивается через HTTP сервер
	// в main.go, здесь просто логируем
	tb.logger.Info("Webhook mode enabled - webhook will be handled by HTTP server")

	// Возвращаем nil, так как webhook обрабатывается в HTTP сервере
	return nil
}

// Stop останавливает бот.
func (tb *OptimizedTelegramBot) Stop(ctx context.Context) error {
	tb.logger.Info("Stopping optimized Telegram bot")
	tb.api.StopReceivingUpdates()

	// Закрываем сервис
	if err := tb.service.Close(); err != nil {
		tb.logger.Error("Error closing service", logging.ErrorField(err))
	}

	// Закрываем кэш
	if err := tb.cache.Close(); err != nil {
		tb.logger.Error("Error closing cache", logging.ErrorField(err))
	}

	return nil
}

// GetPlatformName возвращает название платформы.
func (tb *OptimizedTelegramBot) GetPlatformName() string {
	return "telegram"
}

// GetService возвращает оптимизированный сервис.
func (tb *OptimizedTelegramBot) GetService() *core.OptimizedBotService {
	return tb.service
}

// GetHandler возвращает обработчик для webhook.
func (tb *OptimizedTelegramBot) GetHandler() *OptimizedTelegramHandler {
	return tb.handler
}

// SendFeedbackNotification отправляет уведомление администраторам о новом отзыве.
func (tb *OptimizedTelegramBot) SendFeedbackNotification(feedbackData map[string]interface{}) error {
	tb.logger.Info("Sending feedback notification to admins",
		logging.Int("admin_count", len(tb.adminChatIDs)),
	)

	// Формируем сообщение для администраторов
	adminMsg := fmt.Sprintf(`
📝 Новый отзыв от пользователя:

👤 Имя: %s
📱 Telegram ID: %d

%s

📝 Отзыв:
%s
`,
		feedbackData["first_name"].(string),
		feedbackData["telegram_id"].(int64),
		func() string {
			if username, ok := feedbackData["username"].(*string); ok && username != nil {
				return fmt.Sprintf("👤 Username: @%s", *username)
			}
			return "👤 Username: отсутствует"
		}(),
		feedbackData["feedback_text"].(string),
	)

	// Добавляем контактную информацию, если есть
	if contactInfo, ok := feedbackData["contact_info"].(*string); ok && contactInfo != nil {
		adminMsg += fmt.Sprintf("\n📞 Контакты: %s", *contactInfo)
	}

	// Отправляем сообщение всем администраторам
	successCount := 0
	for _, adminID := range tb.adminChatIDs {
		msg := tgbotapi.NewMessage(adminID, adminMsg)
		if _, err := tb.api.Send(msg); err != nil {
			tb.logger.Error("Failed to send notification to admin",
				logging.Int64("admin_id", adminID),
				logging.ErrorField(err),
			)
		} else {
			successCount++
			tb.logger.Debug("Notification sent to admin",
				logging.Int64("admin_id", adminID),
			)
		}
	}

	tb.logger.Info("Feedback notifications sent",
		logging.Int("success_count", successCount),
		logging.Int("total_count", len(tb.adminChatIDs)),
	)

	// Записываем метрику
	tb.metrics.RecordFeedbackSubmission()

	return nil
}

// SetAdminChatIDs устанавливает Chat ID администраторов.
func (tb *OptimizedTelegramBot) SetAdminChatIDs(chatIDs []int64) {
	tb.adminChatIDs = chatIDs
	tb.logger.Info("Admin chat IDs updated",
		logging.Int("count", len(chatIDs)),
	)
}

// GetAdminCount возвращает количество настроенных администраторов.
func (tb *OptimizedTelegramBot) GetAdminCount() int {
	return len(tb.adminChatIDs) + len(tb.adminUsernames)
}

// HealthCheck проверяет здоровье бота.
func (tb *OptimizedTelegramBot) HealthCheck() error {
	// Проверяем соединение с Telegram API
	_, err := tb.api.GetMe()
	if err != nil {
		return fmt.Errorf("telegram API health check failed: %w", err)
	}

	// Проверяем здоровье сервиса
	if err := tb.service.HealthCheck(); err != nil {
		return fmt.Errorf("service health check failed: %w", err)
	}

	return nil
}

// GetMetrics возвращает метрики бота.
func (tb *OptimizedTelegramBot) GetMetrics() map[string]interface{} {
	serviceMetrics := tb.service.GetMetrics()
	dbStats := tb.service.GetDBStats()

	return map[string]interface{}{
		"service_metrics": serviceMetrics,
		"database_stats":  dbStats,
		"admin_count":     tb.GetAdminCount(),
		"bot_username":    tb.api.Self.UserName,
	}
}
