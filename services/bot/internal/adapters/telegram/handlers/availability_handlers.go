package handlers

import (
	"fmt"
	"strings"

	"language-exchange-bot/internal/models"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// AvailabilityHandlerImpl реализует обработчики для настройки доступности пользователя
type AvailabilityHandlerImpl struct {
	base *BaseHandler
}

// NewAvailabilityHandler создает новый обработчик доступности
func NewAvailabilityHandler(base *BaseHandler) *AvailabilityHandlerImpl {
	return &AvailabilityHandlerImpl{
		base: base,
	}
}

// HandleTimeAvailabilityStart начинает настройку временной доступности
func (ah *AvailabilityHandlerImpl) HandleTimeAvailabilityStart(callback *tgbotapi.CallbackQuery, user *models.User) error {
	text := ah.base.service.Localizer.Get(user.InterfaceLanguageCode, "time_availability_intro")
	keyboard := ah.createDayTypeSelectionKeyboard(user.InterfaceLanguageCode)

	editMsg := tgbotapi.NewEditMessageTextAndMarkup(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		text,
		keyboard,
	)

	_, err := ah.base.bot.Request(editMsg)
	if err != nil {
		return fmt.Errorf("failed to send time availability start message: %w", err)
	}

	// Переводим пользователя в состояние ожидания выбора типа дней
	return ah.base.service.DB.UpdateUserState(user.ID, models.StateWaitingTimeAvailability)
}

// HandleDayTypeSelection обрабатывает выбор типа дней (weekdays/weekends/any/specific)
func (ah *AvailabilityHandlerImpl) HandleDayTypeSelection(callback *tgbotapi.CallbackQuery, user *models.User, dayType string) error {
	// Сохраняем выбранный тип дней
	availability := &models.TimeAvailability{
		DayType:      dayType,
		SpecificDays: []string{}, // Пока пустой
		TimeSlot:     "",         // Будет выбран позже
	}

	// Сохраняем в БД
	err := ah.base.service.DB.SaveTimeAvailability(user.ID, availability)
	if err != nil {
		return fmt.Errorf("failed to save time availability: %w", err)
	}

	// Если выбран specific, показываем выбор конкретных дней
	if dayType == "specific" {
		return ah.ShowSpecificDaysSelection(callback, user)
	}

	// Иначе переходим к выбору времени дня
	return ah.ShowTimeSlotSelection(callback, user)
}

// HandleSpecificDaysSelection обрабатывает выбор конкретных дней недели
func (ah *AvailabilityHandlerImpl) HandleSpecificDaysSelection(callback *tgbotapi.CallbackQuery, user *models.User, day string) error {
	// Получаем текущую доступность
	availability, err := ah.base.service.DB.GetTimeAvailability(user.ID)
	if err != nil {
		return fmt.Errorf("failed to get time availability: %w", err)
	}

	// Добавляем или убираем день из списка
	if ah.containsDay(availability.SpecificDays, day) {
		// Убираем день
		availability.SpecificDays = ah.removeDay(availability.SpecificDays, day)
	} else {
		// Добавляем день
		availability.SpecificDays = append(availability.SpecificDays, day)
	}

	// Сохраняем обновленную доступность
	err = ah.base.service.DB.SaveTimeAvailability(user.ID, availability)
	if err != nil {
		return fmt.Errorf("failed to save time availability: %w", err)
	}

	// Обновляем клавиатуру
	return ah.ShowSpecificDaysSelection(callback, user)
}

// HandleTimeSlotSelection обрабатывает выбор временного слота
func (ah *AvailabilityHandlerImpl) HandleTimeSlotSelection(callback *tgbotapi.CallbackQuery, user *models.User, timeSlot string) error {
	// Получаем текущую доступность
	availability, err := ah.base.service.DB.GetTimeAvailability(user.ID)
	if err != nil {
		return fmt.Errorf("failed to get time availability: %w", err)
	}

	// Устанавливаем временной слот
	availability.TimeSlot = timeSlot

	// Сохраняем в БД
	err = ah.base.service.DB.SaveTimeAvailability(user.ID, availability)
	if err != nil {
		return fmt.Errorf("failed to save time availability: %w", err)
	}

	// Переходим к настройке предпочтений общения
	return ah.startFriendshipPreferencesSetup(callback, user)
}

// HandleFriendshipPreferencesStart начинает настройку предпочтений общения
func (ah *AvailabilityHandlerImpl) HandleFriendshipPreferencesStart(callback *tgbotapi.CallbackQuery, user *models.User) error {
	text := ah.base.service.Localizer.Get(user.InterfaceLanguageCode, "friendship_preferences_intro")
	keyboard := ah.createActivityTypeSelectionKeyboard(user.InterfaceLanguageCode)

	editMsg := tgbotapi.NewEditMessageTextAndMarkup(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		text,
		keyboard,
	)

	_, err := ah.base.bot.Request(editMsg)
	if err != nil {
		return fmt.Errorf("failed to send friendship preferences start message: %w", err)
	}

	// Переводим пользователя в состояние ожидания выбора типа активности
	return ah.base.service.DB.UpdateUserState(user.ID, models.StateWaitingFriendshipPreferences)
}

// HandleActivityTypeSelection обрабатывает выбор типа активности
func (ah *AvailabilityHandlerImpl) HandleActivityTypeSelection(callback *tgbotapi.CallbackQuery, user *models.User, activityType string) error {
	// Сохраняем выбранный тип активности
	preferences := &models.FriendshipPreferences{
		ActivityType:       activityType,
		CommunicationStyle: "", // Будет выбран позже
		CommunicationFreq:  "", // Будет выбран позже
	}

	// Сохраняем в БД
	err := ah.base.service.DB.SaveFriendshipPreferences(user.ID, preferences)
	if err != nil {
		return fmt.Errorf("failed to save friendship preferences: %w", err)
	}

	// Переходим к выбору стиля общения
	return ah.showCommunicationStyleSelection(callback, user)
}

// HandleCommunicationStyleSelection обрабатывает выбор стиля общения
func (ah *AvailabilityHandlerImpl) HandleCommunicationStyleSelection(callback *tgbotapi.CallbackQuery, user *models.User, communicationStyle string) error {
	// Получаем текущие предпочтения
	preferences, err := ah.base.service.DB.GetFriendshipPreferences(user.ID)
	if err != nil {
		return fmt.Errorf("failed to get friendship preferences: %w", err)
	}

	// Устанавливаем стиль общения
	preferences.CommunicationStyle = communicationStyle

	// Сохраняем в БД
	err = ah.base.service.DB.SaveFriendshipPreferences(user.ID, preferences)
	if err != nil {
		return fmt.Errorf("failed to save friendship preferences: %w", err)
	}

	// Переходим к выбору частоты общения
	return ah.showCommunicationFrequencySelection(callback, user)
}

// HandleCommunicationFrequencySelection обрабатывает выбор частоты общения
func (ah *AvailabilityHandlerImpl) HandleCommunicationFrequencySelection(callback *tgbotapi.CallbackQuery, user *models.User, frequency string) error {
	// Получаем текущие предпочтения
	preferences, err := ah.base.service.DB.GetFriendshipPreferences(user.ID)
	if err != nil {
		return fmt.Errorf("failed to get friendship preferences: %w", err)
	}

	// Устанавливаем частоту общения
	preferences.CommunicationFreq = frequency

	// Сохраняем в БД
	err = ah.base.service.DB.SaveFriendshipPreferences(user.ID, preferences)
	if err != nil {
		return fmt.Errorf("failed to save friendship preferences: %w", err)
	}

	// Завершаем настройку доступности
	return ah.completeAvailabilitySetup(callback, user)
}

// completeAvailabilitySetup завершает настройку доступности и переводит пользователя в активное состояние
func (ah *AvailabilityHandlerImpl) completeAvailabilitySetup(callback *tgbotapi.CallbackQuery, user *models.User) error {
	// Обновляем уровень завершения профиля
	err := ah.base.service.DB.UpdateUserProfileCompletionLevel(user.ID, 100)
	if err != nil {
		return fmt.Errorf("failed to update profile completion level: %w", err)
	}

	// Переводим пользователя в активное состояние
	err = ah.base.service.DB.UpdateUserState(user.ID, models.StateActive)
	if err != nil {
		return fmt.Errorf("failed to update user state: %w", err)
	}

	// Показываем финальное сообщение
	text := ah.base.service.Localizer.Get(user.InterfaceLanguageCode, "availability_setup_complete")
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				ah.base.service.Localizer.Get(user.InterfaceLanguageCode, "back_to_main_menu"),
				"back_to_main_menu",
			),
		),
	)

	editMsg := tgbotapi.NewEditMessageTextAndMarkup(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		text,
		keyboard,
	)

	_, err = ah.base.bot.Request(editMsg)
	return err
}

// === Вспомогательные методы ===

// ShowSpecificDaysSelection показывает интерфейс выбора конкретных дней недели
func (ah *AvailabilityHandlerImpl) ShowSpecificDaysSelection(callback *tgbotapi.CallbackQuery, user *models.User) error {
	availability, err := ah.base.service.DB.GetTimeAvailability(user.ID)
	if err != nil {
		return fmt.Errorf("failed to get time availability: %w", err)
	}

	text := fmt.Sprintf(
		"%s\n\n%s: %s",
		ah.base.service.Localizer.Get(user.InterfaceLanguageCode, "select_specific_days"),
		ah.base.service.Localizer.Get(user.InterfaceLanguageCode, "selected_days"),
		ah.formatSelectedDays(availability.SpecificDays, user.InterfaceLanguageCode),
	)

	keyboard := ah.createSpecificDaysSelectionKeyboard(user.InterfaceLanguageCode, availability.SpecificDays)

	editMsg := tgbotapi.NewEditMessageTextAndMarkup(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		text,
		keyboard,
	)

	_, err = ah.base.bot.Request(editMsg)
	return err
}

// ShowTimeSlotSelection показывает интерфейс выбора временного слота
func (ah *AvailabilityHandlerImpl) ShowTimeSlotSelection(callback *tgbotapi.CallbackQuery, user *models.User) error {
	text := ah.base.service.Localizer.Get(user.InterfaceLanguageCode, "select_time_slot")
	keyboard := ah.createTimeSlotSelectionKeyboard(user.InterfaceLanguageCode)

	editMsg := tgbotapi.NewEditMessageTextAndMarkup(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		text,
		keyboard,
	)

	_, err := ah.base.bot.Request(editMsg)
	return err
}

// startFriendshipPreferencesSetup начинает настройку предпочтений общения
func (ah *AvailabilityHandlerImpl) startFriendshipPreferencesSetup(callback *tgbotapi.CallbackQuery, user *models.User) error {
	text := ah.base.service.Localizer.Get(user.InterfaceLanguageCode, "friendship_preferences_intro")
	keyboard := ah.createActivityTypeSelectionKeyboard(user.InterfaceLanguageCode)

	editMsg := tgbotapi.NewEditMessageTextAndMarkup(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		text,
		keyboard,
	)

	_, err := ah.base.bot.Request(editMsg)
	if err != nil {
		return fmt.Errorf("failed to send friendship preferences start message: %w", err)
	}

	// Переводим пользователя в состояние ожидания выбора типа активности
	return ah.base.service.DB.UpdateUserState(user.ID, models.StateWaitingFriendshipPreferences)
}

// showCommunicationStyleSelection показывает интерфейс выбора стиля общения
func (ah *AvailabilityHandlerImpl) showCommunicationStyleSelection(callback *tgbotapi.CallbackQuery, user *models.User) error {
	text := ah.base.service.Localizer.Get(user.InterfaceLanguageCode, "select_communication_style")
	keyboard := ah.createCommunicationStyleSelectionKeyboard(user.InterfaceLanguageCode)

	editMsg := tgbotapi.NewEditMessageTextAndMarkup(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		text,
		keyboard,
	)

	_, err := ah.base.bot.Request(editMsg)
	return err
}

// showCommunicationFrequencySelection показывает интерфейс выбора частоты общения
func (ah *AvailabilityHandlerImpl) showCommunicationFrequencySelection(callback *tgbotapi.CallbackQuery, user *models.User) error {
	text := ah.base.service.Localizer.Get(user.InterfaceLanguageCode, "select_communication_frequency")
	keyboard := ah.createCommunicationFrequencySelectionKeyboard(user.InterfaceLanguageCode)

	editMsg := tgbotapi.NewEditMessageTextAndMarkup(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		text,
		keyboard,
	)

	_, err := ah.base.bot.Request(editMsg)
	return err
}

// === Методы для работы с массивами дней ===

func (ah *AvailabilityHandlerImpl) containsDay(days []string, day string) bool {
	for _, d := range days {
		if d == day {
			return true
		}
	}
	return false
}

func (ah *AvailabilityHandlerImpl) removeDay(days []string, day string) []string {
	var result []string
	for _, d := range days {
		if d != day {
			result = append(result, d)
		}
	}
	return result
}

func (ah *AvailabilityHandlerImpl) formatSelectedDays(days []string, lang string) string {
	if len(days) == 0 {
		return ah.base.service.Localizer.Get(lang, "no_days_selected")
	}

	// Сортируем дни для последовательного отображения
	dayOrder := []string{"monday", "tuesday", "wednesday", "thursday", "friday", "saturday", "sunday"}
	var sortedDays []string

	for _, day := range dayOrder {
		if ah.containsDay(days, day) {
			dayName := ah.base.service.Localizer.Get(lang, "day_"+day)
			sortedDays = append(sortedDays, dayName)
		}
	}

	return strings.Join(sortedDays, ", ")
}
