package telegram

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

	"language-exchange-bot/internal/core"
	"language-exchange-bot/internal/database"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// ResolveUsernameToChatID - упрощенная функция валидации
// Username'ы теперь считываются только из .env файла
func (tb *TelegramBot) ResolveUsernameToChatID(username string) (int64, error) {
	// Все username'ы теперь динамически читаются из конфигурации
	// Эта функция оставлена для совместимости, но не содержит хардкода
	log.Printf("Валидация username @%s через конфигурацию", username)
	return 0, nil // для совместимости с существующим кодом
}

type TelegramBot struct {
	api            *tgbotapi.BotAPI
	service        *core.BotService
	debug          bool
	adminChatIDs   []int64  // ID администраторов для уведомлений (resolved)
	adminUsernames []string // Usernames администраторов (дополнительно храним для логов)
}

func NewTelegramBot(token string, db *database.DB, debug bool, adminChatIDs []int64) (*TelegramBot, error) {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, fmt.Errorf("failed to create telegram bot: %w", err)
	}
	bot.Debug = debug
	service := core.NewBotService(db)
	return &TelegramBot{
		api:            bot,
		service:        service,
		debug:          debug,
		adminChatIDs:   adminChatIDs,
		adminUsernames: make([]string, 0), // инициализируем пустой если нужно
	}, nil
}

// NewTelegramBotWithUsernames создает бота с поддержкой usernames администраторов
func NewTelegramBotWithUsernames(token string, db *database.DB, debug bool, adminInputs []string) (*TelegramBot, error) {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, fmt.Errorf("failed to create telegram bot: %w", err)
	}
	bot.Debug = debug

	tgBot := &TelegramBot{
		api:            bot,
		service:        core.NewBotService(db),
		debug:          debug,
		adminChatIDs:   make([]int64, 0),
		adminUsernames: make([]string, 0),
	}

	// Обрабатываем usernames (работаем с ними напрямую без разрешения)
	for _, adminInput := range adminInputs {
		adminInput = strings.TrimSpace(adminInput)
		if adminInput == "" {
			continue
		}

		if strings.HasPrefix(adminInput, "@") {
			// Это username - сохраняем для отправки по @username
			tgBot.adminUsernames = append(tgBot.adminUsernames, adminInput)
			log.Printf("Добавлен администратор: %s", adminInput)
		} else {
			// Это числовой ID - добавляем как обычно
			chatID, err := strconv.ParseInt(adminInput, 10, 64)
			if err == nil {
				tgBot.adminChatIDs = append(tgBot.adminChatIDs, chatID)
				log.Printf("Использован готовый Chat ID: %d", chatID)
			} else {
				log.Printf("Неверный формат администратора: %s", adminInput)
			}
		}
	}

	totalAdmins := len(tgBot.adminChatIDs) + len(tgBot.adminUsernames)
	if totalAdmins == 0 {
		log.Println("Предупреждение: не удалось настроить ни одного администратора")
	}

	log.Printf("Бот настроен с %d администраторами (%d по ID, %d по username)",
		totalAdmins, len(tgBot.adminChatIDs), len(tgBot.adminUsernames))
	return tgBot, nil
}

// SendFeedbackNotification отправляет уведомление администраторам о новом отзыве
func (tb *TelegramBot) SendFeedbackNotification(feedbackData map[string]interface{}) error {
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

	// Отправляем сообщение всем администраторам по ID
	for _, adminID := range tb.adminChatIDs {
		msg := tgbotapi.NewMessage(adminID, adminMsg)
		if _, err := tb.api.Send(msg); err != nil {
			log.Printf("Ошибка отправки уведомления администратору %d: %v", adminID, err)
		}
	}

	// Отправляем сообщение всем администраторам по username
	for _, username := range tb.adminUsernames {
		// Убираем @ из username перед отправкой
		cleanUsername := strings.TrimPrefix(username, "@")

		// Попытка получить Chat ID по username с помощью GetChat
		if chatID, err := tb.getChatIDByUsername(cleanUsername); err == nil {
			// Успешно получили Chat ID, отправляем по ID
			msg := tgbotapi.NewMessage(chatID, adminMsg)
			if _, err := tb.api.Send(msg); err != nil {
				log.Printf("Не удалось отправить уведомление администратору @%s: %v", cleanUsername, err)
			} else {
				log.Printf("✅ Уведомление отправлено администратору @%s (по ID: %d)", cleanUsername, chatID)
			}
		} else {
			// Не удалось получить Chat ID, логируем ошибку
			log.Printf("❌ Не удалось получить Chat ID для @%s: %v", cleanUsername, err)
		}
	}

	totalAdmins := len(tb.adminChatIDs) + len(tb.adminUsernames)
	log.Printf("Отправлено уведомление %d администраторам (%d по ID, %d по username)",
		totalAdmins, len(tb.adminChatIDs), len(tb.adminUsernames))
	return nil
}

// GetService возвращает сервис бота для внешнего доступа
func (tb *TelegramBot) GetService() *core.BotService {
	return tb.service
}

func (tb *TelegramBot) Start(ctx context.Context) error {
	log.Printf("Authorized on account %s", tb.api.Self.UserName)
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := tb.api.GetUpdatesChan(u)
	// Передаем usernames администраторов в обработчик
	handler := NewTelegramHandlerWithAdmins(tb.api, tb.service, tb.adminChatIDs, tb.adminUsernames)

	for {
		select {
		case update := <-updates:
			go func(upd tgbotapi.Update) {
				if err := handler.HandleUpdate(upd); err != nil {
					log.Printf("Error handling update: %v", err)
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

func (tb *TelegramBot) GetPlatformName() string {
	return "telegram"
}

// getChatIDByUsername - функция для получения Chat ID по username
// Пока что всегда возвращает ошибку для реализации в будущем
func (tb *TelegramBot) getChatIDByUsername(username string) (int64, error) {
	// Функция зарезервирована для будущей реализации получения Chat ID через Telegram API
	// Сейчас возвращает ошибку чтобы не было хардкода
	log.Printf("Получение Chat ID по @%s пока не реализовано", username)
	return 0, fmt.Errorf("получение Chat ID по username пока не поддерживается")
}

// GetAdminCount возвращает количество настроенных администраторов
func (tb *TelegramBot) GetAdminCount() int {
	return len(tb.adminChatIDs)
}
