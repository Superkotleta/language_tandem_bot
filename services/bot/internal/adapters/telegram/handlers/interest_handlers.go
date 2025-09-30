package handlers

import (
	"log"
	"strconv"

	"language-exchange-bot/internal/core"
	"language-exchange-bot/internal/errors"
	"language-exchange-bot/internal/models"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// InterestHandler интерфейс для обработки интересов
type InterestHandler interface {
	HandleInterestSelection(callback *tgbotapi.CallbackQuery, user *models.User, interestIDStr string) error
	HandleInterestsContinue(callback *tgbotapi.CallbackQuery, user *models.User) error
}

// InterestHandlerImpl реализация обработчика интересов
type InterestHandlerImpl struct {
	service         *core.BotService
	bot             *tgbotapi.BotAPI
	keyboardBuilder *KeyboardBuilder
	errorHandler    *errors.ErrorHandler
}

// NewInterestHandlerLegacy создает новый обработчик интересов (legacy)
func NewInterestHandlerLegacy(service *core.BotService, bot *tgbotapi.BotAPI, keyboardBuilder *KeyboardBuilder, errorHandler *errors.ErrorHandler) InterestHandler {
	return &InterestHandlerImpl{
		service:         service,
		bot:             bot,
		keyboardBuilder: keyboardBuilder,
		errorHandler:    errorHandler,
	}
}

// HandleInterestSelection обрабатывает выбор интереса
func (h *InterestHandlerImpl) HandleInterestSelection(callback *tgbotapi.CallbackQuery, user *models.User, interestIDStr string) error {
	interestID, err := strconv.Atoi(interestIDStr)
	if err != nil {
		log.Printf("Error parsing interest ID: %v", err)
		return err
	}

	// Получаем текущие выбранные интересы пользователя через Batch Loading
	selectedInterestsMap, err := h.service.BatchLoadUserInterests([]int{user.ID})
	var selectedInterests []int
	if err != nil {
		log.Printf("Error getting user interests, using empty list: %v", err)
		selectedInterests = []int{} // fallback
	} else {
		selectedInterests = selectedInterestsMap[user.ID]
	}

	// Переключаем интерес (toggle)
	isCurrentlySelected := false
	for i, id := range selectedInterests {
		if id == interestID {
			// Убираем из списка
			selectedInterests = append(selectedInterests[:i], selectedInterests[i+1:]...)
			isCurrentlySelected = true
			break
		}
	}

	if !isCurrentlySelected {
		// Добавляем в список
		selectedInterests = append(selectedInterests, interestID)
		err = h.service.DB.SaveUserInterest(user.ID, interestID, false)
	} else {
		// Удаляем из базы данных
		err = h.service.DB.RemoveUserInterest(user.ID, interestID)
	}

	if err != nil {
		log.Printf("Error saving/removing user interest: %v", err)
		return err
	}

	// Обновляем клавиатуру с новым состоянием
	interests, err := h.service.GetCachedInterests(user.InterfaceLanguageCode)
	if err != nil {
		log.Printf("Error getting interests: %v", err)
		return err
	}

	keyboard := h.keyboardBuilder.CreateInterestsKeyboard(interests, selectedInterests, user.InterfaceLanguageCode)
	editMsg := tgbotapi.NewEditMessageReplyMarkup(callback.Message.Chat.ID, callback.Message.MessageID, keyboard)
	_, err = h.bot.Request(editMsg)
	return err
}

// HandleInterestsContinue обрабатывает продолжение после выбора интересов
func (h *InterestHandlerImpl) HandleInterestsContinue(callback *tgbotapi.CallbackQuery, user *models.User) error {
	log.Printf("InterestHandler.HandleInterestsContinue called for user ID: %d, Telegram ID: %d", user.ID, user.TelegramID)

	// Проверяем, выбраны ли интересы через Batch Loading
	selectedInterestsMap, err := h.service.BatchLoadUserInterests([]int{user.ID})
	if err != nil {
		log.Printf("Error getting selected interests: %v", err)
		return err
	}
	selectedInterests := selectedInterestsMap[user.ID]

	log.Printf("User %d has %d selected interests: %v", user.ID, len(selectedInterests), selectedInterests)

	// Если не выбрано ни одного интереса, сообщаем пользователю и оставляем клавиатуру
	if len(selectedInterests) == 0 {
		warningMsg := "❗ " + h.service.Localizer.Get(user.InterfaceLanguageCode, "choose_at_least_one_interest")
		if warningMsg == "choose_at_least_one_interest" { // fallback if key doesn't exist
			warningMsg = "❗ Пожалуйста, выберите хотя бы один интерес"
		}

		// Добавляем оригинальный текст с предупреждением
		chooseInterestsText := h.service.Localizer.Get(user.InterfaceLanguageCode, "choose_interests")
		fullText := warningMsg + "\n\n" + chooseInterestsText

		// Получаем интересы и оставляем клавиатуру с интересами видимой, обновляя только текст
		interests, _ := h.service.GetCachedInterests(user.InterfaceLanguageCode)
		keyboard := h.keyboardBuilder.CreateInterestsKeyboard(interests, []int{}, user.InterfaceLanguageCode)
		editMsg := tgbotapi.NewEditMessageTextAndMarkup(
			callback.Message.Chat.ID,
			callback.Message.MessageID,
			fullText,
			keyboard,
		)
		_, err := h.bot.Request(editMsg)
		return err
	}

	// Если интересы выбраны, завершаем настройку профиля
	completionMsg := h.service.Localizer.Get(user.InterfaceLanguageCode, "profile_completed")
	keyboard := h.keyboardBuilder.CreateProfileCompletedKeyboard(user.InterfaceLanguageCode)
	editMsg := tgbotapi.NewEditMessageTextAndMarkup(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		completionMsg,
		keyboard,
	)
	_, err = h.bot.Request(editMsg)
	if err != nil {
		return err
	}

	// Обновляем состояние пользователя
	log.Printf("Updating user %d state to active", user.ID)
	err = h.service.DB.UpdateUserState(user.ID, models.StateActive)
	if err != nil {
		log.Printf("Error updating user state: %v", err)
		return err
	}

	log.Printf("Updating user %d status to active", user.ID)
	err = h.service.DB.UpdateUserStatus(user.ID, models.StatusActive)
	if err != nil {
		log.Printf("Error updating user status: %v", err)
		return err
	}

	// Обновляем уровень завершения профиля до 100%
	log.Printf("Updating user %d profile completion level to 100%%", user.ID)
	err = h.updateProfileCompletionLevel(user.ID, 100)
	if err != nil {
		log.Printf("Error updating profile completion level: %v", err)
		return err
	}

	log.Printf("Successfully completed profile for user %d", user.ID)

	return nil
}

// updateProfileCompletionLevel обновляет уровень завершения профиля от 0 до 100
func (h *InterestHandlerImpl) updateProfileCompletionLevel(userID int, completionLevel int) error {
	log.Printf("Executing updateProfileCompletionLevel: userID=%d, level=%d", userID, completionLevel)

	result, err := h.service.DB.GetConnection().Exec(`
		UPDATE users
		SET profile_completion_level = $1, updated_at = NOW()
		WHERE id = $2
	`, completionLevel, userID)
	if err != nil {
		log.Printf("Error in updateProfileCompletionLevel: %v", err)
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	log.Printf("updateProfileCompletionLevel: %d rows affected for user %d", rowsAffected, userID)

	return nil
}
