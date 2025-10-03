// Package telegram provides Telegram Bot API integration and message handling.
package telegram

import (
	"context"
	"fmt"
	"log"
	"strings"

	"language-exchange-bot/internal/core"
	"language-exchange-bot/internal/database"
	"language-exchange-bot/internal/errors"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// ResolveUsernameToChatID - упрощенная функция валидации
// Username'ы теперь считываются только из .env файла.
func (tb *TelegramBot) ResolveUsernameToChatID(username string) (int64, error) {
	// Все username'ы теперь динамически читаются из конфигурации
	// Эта функция оставлена для совместимости, но не содержит хардкода
	log.Printf("Валидация username @%s через конфигурацию", username)

	return 0, nil // для совместимости с существующим кодом
}

// TelegramBot represents a Telegram bot instance with message handling capabilities.
// The name includes "Telegram" prefix for clarity, even though it may cause stuttering with the package name.
type TelegramBot struct {
	api            *tgbotapi.BotAPI
	service        *core.BotService
	debug          bool
	adminChatIDs   []int64  // ID администраторов для уведомлений (resolved)
	adminUsernames []string // Usernames администраторов (дополнительно храним для логов)
	errorHandler   *errors.ErrorHandler
}

// NewTelegramBot creates a new Telegram bot instance with the provided configuration.
func NewTelegramBot(token string, db *database.DB, debug bool, adminChatIDs []int64) (*TelegramBot, error) {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, fmt.Errorf("failed to create telegram bot: %w", err)
	}

	bot.Debug = debug
	service := core.NewBotService(db, nil)

	return &TelegramBot{
		api:            bot,
		service:        service,
		debug:          debug,
		adminChatIDs:   adminChatIDs,
		adminUsernames: make([]string, 0), // инициализируем пустой если нужно
	}, nil
}

// NewTelegramBotWithUsernames создает бота с поддержкой usernames администраторов.
func NewTelegramBotWithUsernames(
	token string,
	db *database.DB,
	debug bool,
	adminUsernames []string,
) (*TelegramBot, error) {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, fmt.Errorf("failed to create telegram bot: %w", err)
	}

	bot.Debug = debug

	tgBot := &TelegramBot{
		api:            bot,
		service:        core.NewBotService(db, nil),
		debug:          debug,
		adminChatIDs:   make([]int64, 0), // Будет установлен позже через SetAdminChatIDs
		adminUsernames: make([]string, 0),
	}

	// Обрабатываем usernames для проверки прав
	for _, username := range adminUsernames {
		username = strings.TrimSpace(username)
		if username == "" {
			continue
		}

		// Убираем @ если есть
		username = strings.TrimPrefix(username, "@")

		tgBot.adminUsernames = append(tgBot.adminUsernames, username)
		log.Printf("Добавлен администратор для проверки прав: @%s", username)
	}

	if len(tgBot.adminUsernames) == 0 {
		log.Println("Предупреждение: не настроено ни одного администратора для проверки прав")
	}

	log.Printf("Бот настроен с %d администраторами для проверки прав", len(tgBot.adminUsernames))

	return tgBot, nil
}

// SendFeedbackNotification отправляет уведомление администраторам о новом отзыве.
func (tb *TelegramBot) SendFeedbackNotification(feedbackData map[string]interface{}) error {
	log.Printf("Отправляем уведомление о новом отзыве администраторам...")
	log.Printf("Администраторы по ID: %v", tb.adminChatIDs)
	log.Printf("Администраторы по username: %v", tb.adminUsernames)
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
				return "👤 Username: @" + *username
			}

			return "👤 Username: отсутствует"
		}(),
		feedbackData["feedback_text"].(string),
	)

	// Добавляем контактную информацию, если есть
	if contactInfo, ok := feedbackData["contact_info"].(*string); ok && contactInfo != nil {
		adminMsg += "\n📞 Контакты: " + *contactInfo
	}

	// Отправляем сообщение всем администраторам по ID
	log.Printf("Отправляем уведомления %d администраторам по ID", len(tb.adminChatIDs))

	for _, adminID := range tb.adminChatIDs {
		msg := tgbotapi.NewMessage(adminID, adminMsg)
		if _, err := tb.api.Send(msg); err != nil {
			log.Printf("Ошибка отправки уведомления администратору %d: %v", adminID, err)
		} else {
			log.Printf("Уведомление отправлено администратору %d", adminID)
		}
	}

	// Username администраторы используются только для проверки прав, не для уведомлений
	log.Printf("Username администраторы (%d) используются только для проверки прав", len(tb.adminUsernames))

	totalAdmins := len(tb.adminChatIDs)
	log.Printf("Отправлено уведомление %d администраторам по Chat ID", totalAdmins)

	return nil
}

// GetService возвращает сервис бота для внешнего доступа.
func (tb *TelegramBot) GetService() *core.BotService {
	return tb.service
}

// Start begins the Telegram bot message processing loop.
func (tb *TelegramBot) Start(ctx context.Context) error {
	log.Printf("Authorized on account %s", tb.api.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := tb.api.GetUpdatesChan(u)
	// Передаем usernames администраторов в обработчик
	handler := NewTelegramHandlerWithAdmins(tb.api, tb.service, tb.adminChatIDs, tb.adminUsernames, tb.errorHandler)

	for {
		select {
		case update := <-updates:
			go func(upd tgbotapi.Update) {
				if err := handler.HandleUpdate(upd); err != nil {
					// Используем новую систему обработки ошибок
					if tb.errorHandler != nil {
						if handlerErr := tb.errorHandler.HandleTelegramError(err, 0, 0, "HandleUpdate"); handlerErr != nil {
							log.Printf("Error in error handler: %v", handlerErr)
						}
					} else {
						log.Printf("Error handling update: %v", err)
					}
				}
			}(update)
		case <-ctx.Done():
			log.Println("Stopping Telegram bot...")
			tb.api.StopReceivingUpdates()

			return nil
		}
	}
}

func (tb *TelegramBot) Stop(ctx context.Context) error {
	tb.api.StopReceivingUpdates()

	return nil
}

// GetPlatformName returns the platform name for the bot.
func (tb *TelegramBot) GetPlatformName() string {
	return "telegram"
}

// getChatIDByUsername - функция для получения Chat ID по username
// TODO: функция может быть использована в будущем для получения chat ID по username
//
//nolint:unused
func (tb *TelegramBot) getChatIDByUsername(username string) (int64, error) {
	log.Printf("Пытаемся получить Chat ID для username: @%s", username)

	// Используем Telegram API для получения информации о чате по username
	chatConfig := tgbotapi.ChatInfoConfig{
		ChatConfig: tgbotapi.ChatConfig{
			SuperGroupUsername: "@" + username,
		},
	}

	chat, err := tb.api.GetChat(chatConfig)
	if err != nil {
		log.Printf("Ошибка получения Chat ID для @%s: %v", username, err)

		return 0, fmt.Errorf("не удалось получить информацию о чате @%s: %w", username, err)
	}

	log.Printf("Получен Chat ID %d для username @%s", chat.ID, username)

	return chat.ID, nil
}

// SetAdminChatIDs устанавливает Chat ID администраторов для уведомлений.
func (tb *TelegramBot) SetAdminChatIDs(chatIDs []int64) {
	tb.adminChatIDs = chatIDs
	log.Printf("Установлены Chat ID администраторов для уведомлений: %v", chatIDs)
}

// GetAdminCount возвращает количество настроенных администраторов.
func (tb *TelegramBot) GetAdminCount() int {
	// Возвращаем максимум из двух списков, так как это одни и те же люди
	chatCount := len(tb.adminChatIDs)
	usernameCount := len(tb.adminUsernames)

	if chatCount > usernameCount {
		return chatCount
	}

	return usernameCount
}

// NewTelegramBotWithService создает бота с готовым сервисом (для Redis кэша).
func NewTelegramBotWithService(
	token string,
	service *core.BotService,
	debug bool,
	adminUsernames []string,
) (*TelegramBot, error) {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, fmt.Errorf("failed to create telegram bot: %w", err)
	}

	bot.Debug = debug

	tgBot := &TelegramBot{
		api:            bot,
		service:        service,
		debug:          debug,
		adminChatIDs:   make([]int64, 0), // Будет установлен позже через SetAdminChatIDs
		adminUsernames: make([]string, 0),
	}

	// Обрабатываем usernames для проверки прав
	for _, username := range adminUsernames {
		username = strings.TrimSpace(username)
		if username == "" {
			continue
		}

		// Убираем @ если есть
		username = strings.TrimPrefix(username, "@")

		tgBot.adminUsernames = append(tgBot.adminUsernames, username)
		log.Printf("Добавлен username администратора: @%s", username)
	}

	return tgBot, nil
}

// SetErrorHandler устанавливает обработчик ошибок.
func (tb *TelegramBot) SetErrorHandler(errorHandler *errors.ErrorHandler) {
	tb.errorHandler = errorHandler

	log.Printf("Error handler установлен для TelegramBot")
}
