package errors

import (
	"fmt"
	"log"
)

// AdminNotifierImpl реализует уведомления администраторов.
type AdminNotifierImpl struct {
	adminChatIDs []int64
	botAPI       interface{} // Telegram Bot API для отправки уведомлений
}

// NewAdminNotifier создает новый уведомитель администраторов.
func NewAdminNotifier(adminChatIDs []int64, botAPI interface{}) *AdminNotifierImpl {
	return &AdminNotifierImpl{
		adminChatIDs: adminChatIDs,
		botAPI:       botAPI,
	}
}

// NotifyCriticalError уведомляет администраторов о критической ошибке.
func (n *AdminNotifierImpl) NotifyCriticalError(err *CustomError) {
	message := fmt.Sprintf(`
🚨 **КРИТИЧЕСКАЯ ОШИБКА**

**Тип:** %s
**Время:** %s
**RequestID:** %s
**Сообщение:** %s

**Контекст:**
- User ID: %v
- Chat ID: %v
- Операция: %v

**Дополнительная информация:**
%s
`,
		err.Type.String(),
		err.Timestamp.Format("2006-01-02 15:04:05"),
		err.RequestID,
		err.Message,
		err.Context["user_id"],
		err.Context["chat_id"],
		err.Context["operation"],
		n.formatContext(err.Context),
	)

	n.sendToAdmins(message)
}

// NotifyTelegramAPIError уведомляет администраторов об ошибке Telegram API.
func (n *AdminNotifierImpl) NotifyTelegramAPIError(err *CustomError, chatID int64) {
	message := fmt.Sprintf(`
⚠️ **ОШИБКА TELEGRAM API**

**Время:** %s
**RequestID:** %s
**Chat ID:** %d
**Сообщение:** %s

**Контекст:**
- User ID: %v
- Операция: %v
`,
		err.Timestamp.Format("2006-01-02 15:04:05"),
		err.RequestID,
		chatID,
		err.Message,
		err.Context["user_id"],
		err.Context["operation"],
	)

	n.sendToAdmins(message)
}

// sendToAdmins отправляет сообщение всем администраторам.
func (n *AdminNotifierImpl) sendToAdmins(message string) {
	for _, chatID := range n.adminChatIDs {
		n.sendMessage(chatID, message)
	}
}

// sendMessage отправляет сообщение (заглушка для интеграции с Telegram API).
func (n *AdminNotifierImpl) sendMessage(chatID int64, message string) {
	// Здесь должна быть интеграция с Telegram Bot API
	// Интегрировать с реальным Telegram Bot API
	// Пример:
	// if botAPI, ok := n.botAPI.(*tgbotapi.BotAPI); ok {
	//     msg := tgbotapi.NewMessage(chatID, message)
	//     msg.ParseMode = tgbotapi.ModeMarkdown
	//     botAPI.Send(msg)
	// }
	// Пока что просто логируем
	log.Printf("Admin notification to chat %d: %s", chatID, message)
}

// formatContext форматирует контекст для отображения.
func (n *AdminNotifierImpl) formatContext(context map[string]interface{}) string {
	result := ""

	for key, value := range context {
		if key != "user_id" && key != "chat_id" && key != "operation" {
			result += fmt.Sprintf("- %s: %v\n", key, value)
		}
	}

	if result == "" {
		return "Нет дополнительной информации"
	}

	return result
}

// SetAdminChatIDs обновляет список Chat ID администраторов.
func (n *AdminNotifierImpl) SetAdminChatIDs(chatIDs []int64) {
	n.adminChatIDs = chatIDs
}

// GetAdminChatIDs возвращает список Chat ID администраторов.
func (n *AdminNotifierImpl) GetAdminChatIDs() []int64 {
	return n.adminChatIDs
}
