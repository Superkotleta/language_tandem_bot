package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"language-exchange-bot/internal/localization"
	"language-exchange-bot/internal/models"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// IsolatedAvailabilityEditor управляет изолированным редактированием доступности
type IsolatedAvailabilityEditor struct {
	baseHandler *BaseHandler
}

// AvailabilityEditSession представляет сессию редактирования доступности
type AvailabilityEditSession struct {
	UserID                   int                           `json:"user_id"`
	OriginalTimeAvailability *models.TimeAvailability      `json:"original_time_availability"`
	CurrentTimeAvailability  *models.TimeAvailability      `json:"current_time_availability"`
	OriginalPreferences      *models.FriendshipPreferences `json:"original_preferences"`
	CurrentPreferences       *models.FriendshipPreferences `json:"current_preferences"`
	Changes                  []AvailabilityChange          `json:"changes"`
	CurrentStep              string                        `json:"current_step"` // "time", "communication", "frequency"
	SessionStart             time.Time                     `json:"session_start"`
	LastActivity             time.Time                     `json:"last_activity"`
}

// AvailabilityChange представляет изменение в настройках доступности
type AvailabilityChange struct {
	Field     string      `json:"field"` // "day_type", "time_slots", "communication_styles", "frequency"
	OldValue  interface{} `json:"old_value"`
	NewValue  interface{} `json:"new_value"`
	Timestamp time.Time   `json:"timestamp"`
}

// NewIsolatedAvailabilityEditor создает новый изолированный редактор доступности
func NewIsolatedAvailabilityEditor(baseHandler *BaseHandler) *IsolatedAvailabilityEditor {
	return &IsolatedAvailabilityEditor{
		baseHandler: baseHandler,
	}
}

// =============================================================================
// ОСНОВНЫЕ МЕТОДЫ УПРАВЛЕНИЯ СЕССИЯМИ
// =============================================================================

// StartEditSession начинает сессию редактирования доступности
func (e *IsolatedAvailabilityEditor) StartEditSession(callback *tgbotapi.CallbackQuery, user *models.User) error {
	loggingService := e.baseHandler.service.LoggingService

	// Детальное логирование начала сессии
	loggingService.LogRequestStart("", int64(user.ID), callback.Message.Chat.ID, "StartEditSession")
	loggingService.Telegram().InfoWithContext("Starting availability edit session", "", int64(user.ID), callback.Message.Chat.ID, "StartEditSession", map[string]interface{}{
		"user_id":            user.ID,
		"operation":          "start_edit_session",
		"interface_language": user.InterfaceLanguageCode,
		"current_status":     user.Status,
	})

	// Получаем текущие данные пользователя
	timeAvailability, err := e.baseHandler.service.GetTimeAvailability(user.ID)
	if err != nil {
		loggingService.Telegram().ErrorWithContext("Failed to get time availability for edit session", "", int64(user.ID), callback.Message.Chat.ID, "StartEditSession", map[string]interface{}{
			"user_id":    user.ID,
			"error":      err.Error(),
			"error_type": "database_error",
		})
		loggingService.LogErrorWithContext(err, "", int64(user.ID), callback.Message.Chat.ID, "StartEditSession", "database")
		return fmt.Errorf("failed to get time availability: %w", err)
	}

	friendshipPreferences, err := e.baseHandler.service.GetFriendshipPreferences(user.ID)
	if err != nil {
		loggingService.Telegram().ErrorWithContext("Failed to get friendship preferences for edit session", "", int64(user.ID), callback.Message.Chat.ID, "StartEditSession", map[string]interface{}{
			"user_id":    user.ID,
			"error":      err.Error(),
			"error_type": "database_error",
		})
		loggingService.LogErrorWithContext(err, "", int64(user.ID), callback.Message.Chat.ID, "StartEditSession", "database")
		return fmt.Errorf("failed to get friendship preferences: %w", err)
	}

	// Логируем текущее состояние данных
	loggingService.Telegram().InfoWithContext("Current user availability data loaded", "", int64(user.ID), callback.Message.Chat.ID, "StartEditSession", map[string]interface{}{
		"user_id": user.ID,
		"current_time_availability": map[string]interface{}{
			"day_type":            timeAvailability.DayType,
			"specific_days_count": len(timeAvailability.SpecificDays),
			"time_slots_count":    len(timeAvailability.TimeSlots),
		},
		"current_preferences": map[string]interface{}{
			"activity_type":              friendshipPreferences.ActivityType,
			"communication_styles_count": len(friendshipPreferences.CommunicationStyles),
			"communication_freq":         friendshipPreferences.CommunicationFreq,
		},
	})

	// Создаем сессию
	session := &AvailabilityEditSession{
		UserID:                   user.ID,
		OriginalTimeAvailability: e.deepCopyTimeAvailability(timeAvailability),
		CurrentTimeAvailability:  e.deepCopyTimeAvailability(timeAvailability),
		OriginalPreferences:      e.deepCopyFriendshipPreferences(friendshipPreferences),
		CurrentPreferences:       e.deepCopyFriendshipPreferences(friendshipPreferences),
		Changes:                  []AvailabilityChange{},
		CurrentStep:              "menu",
		SessionStart:             time.Now(),
		LastActivity:             time.Now(),
	}

	// Сохраняем сессию в кеше
	if err := e.saveEditSession(session); err != nil {
		loggingService.Telegram().ErrorWithContext("Failed to save edit session to cache", "", int64(user.ID), callback.Message.Chat.ID, "StartEditSession", map[string]interface{}{
			"user_id":    user.ID,
			"error":      err.Error(),
			"error_type": "cache_error",
		})
		loggingService.LogErrorWithContext(err, "", int64(user.ID), callback.Message.Chat.ID, "StartEditSession", "cache")
		return err
	}

	loggingService.Telegram().InfoWithContext("Availability edit session created successfully", "", int64(user.ID), callback.Message.Chat.ID, "StartEditSession", map[string]interface{}{
		"user_id":           user.ID,
		"session_id":        fmt.Sprintf("session_%d_%d", user.ID, session.SessionStart.Unix()),
		"cache_ttl_minutes": 30,
	})

	// Показываем главное меню редактирования
	return e.ShowEditMenu(callback, session, user)
}

// ShowEditMenu показывает главное меню редактирования доступности
func (e *IsolatedAvailabilityEditor) ShowEditMenu(callback *tgbotapi.CallbackQuery, session *AvailabilityEditSession, user *models.User) error {
	lang := user.InterfaceLanguageCode
	localizer := e.baseHandler.service.Localizer

	// Форматируем текущие настройки для отображения
	timeDisplay := e.formatCurrentTimeAvailability(session.CurrentTimeAvailability, lang)
	commDisplay := e.formatCurrentCommunicationPreferences(session.CurrentPreferences, lang)
	freqDisplay := e.formatCurrentFrequency(session.CurrentPreferences, lang)

	message := fmt.Sprintf("%s\n\n%s\n%s\n%s\n%s",
		localizer.Get(lang, "edit_availability"),
		localizer.Get(lang, "current_settings"),
		timeDisplay,
		commDisplay,
		freqDisplay,
	)

	keyboard := e.createEditMenuKeyboard(session, lang)

	return e.baseHandler.messageFactory.EditWithKeyboard(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		message,
		&keyboard,
	)
}

// =============================================================================
// МЕТОДЫ РЕДАКТИРОВАНИЯ ДНЕЙ
// =============================================================================

// EditDays переходит к редактированию дней
func (e *IsolatedAvailabilityEditor) EditDays(callback *tgbotapi.CallbackQuery, user *models.User) error {
	session, err := e.getEditSession(user.ID)
	if err != nil {
		return err
	}

	session.CurrentStep = "days"
	session.LastActivity = time.Now()
	e.saveEditSession(session)

	return e.ShowDayTypeSelection(callback, session, user)
}

// ShowDayTypeSelection показывает выбор типа дней
func (e *IsolatedAvailabilityEditor) ShowDayTypeSelection(callback *tgbotapi.CallbackQuery, session *AvailabilityEditSession, user *models.User) error {
	lang := user.InterfaceLanguageCode
	localizer := e.baseHandler.service.Localizer

	message := localizer.Get(lang, "select_day_type")

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				localizer.Get(lang, localization.LocaleTimeWeekdays),
				localization.CallbackAvailEditDayTypeWeekdays,
			),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				localizer.Get(lang, localization.LocaleTimeWeekends),
				localization.CallbackAvailEditDayTypeWeekends,
			),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				localizer.Get(lang, localization.LocaleTimeAny),
				localization.CallbackAvailEditDayTypeAny,
			),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				localizer.Get(lang, "select_specific_days_button"),
				localization.CallbackAvailEditDayTypeSpecific,
			),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				localizer.Get(lang, localization.LocaleBackToEditMenu),
				localization.CallbackAvailBackToEditMenu,
			),
		),
	)

	return e.baseHandler.messageFactory.EditWithKeyboard(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		message,
		&keyboard,
	)
}

// HandleDayTypeSelection обрабатывает выбор типа дней
func (e *IsolatedAvailabilityEditor) HandleDayTypeSelection(callback *tgbotapi.CallbackQuery, user *models.User, dayType string) error {
	session, err := e.getEditSession(user.ID)
	if err != nil {
		return err
	}

	// Записываем изменение
	e.recordChange(session, "day_type", session.CurrentTimeAvailability.DayType, dayType)

	// Обновляем сессию
	session.CurrentTimeAvailability.DayType = dayType
	session.LastActivity = time.Now()

	// Если выбраны конкретные дни, показываем выбор дней
	if dayType == "specific" {
		session.CurrentTimeAvailability.SpecificDays = []string{} // Сбрасываем выбор
		e.saveEditSession(session)
		return e.ShowSpecificDaysSelection(callback, session, user)
	}

	e.saveEditSession(session)
	return e.ShowEditMenu(callback, session, user)
}

// ShowSpecificDaysSelection показывает выбор конкретных дней
func (e *IsolatedAvailabilityEditor) ShowSpecificDaysSelection(callback *tgbotapi.CallbackQuery, session *AvailabilityEditSession, user *models.User) error {
	lang := user.InterfaceLanguageCode
	localizer := e.baseHandler.service.Localizer

	// Форматируем выбранные дни
	selectedDays := e.formatSelectedDays(session.CurrentTimeAvailability.SpecificDays, lang)

	message := fmt.Sprintf("%s\n\n%s: %s",
		localizer.Get(lang, "select_specific_days"),
		localizer.Get(lang, "selected_days"),
		selectedDays,
	)

	keyboard := e.createSpecificDaysKeyboard(session, lang)

	return e.baseHandler.messageFactory.EditWithKeyboard(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		message,
		&keyboard,
	)
}

// ToggleSpecificDay переключает выбор конкретного дня
func (e *IsolatedAvailabilityEditor) ToggleSpecificDay(callback *tgbotapi.CallbackQuery, user *models.User, day string) error {
	session, err := e.getEditSession(user.ID)
	if err != nil {
		return err
	}

	// Переключаем день в массиве
	days := session.CurrentTimeAvailability.SpecificDays
	dayIndex := -1
	for i, d := range days {
		if d == day {
			dayIndex = i
			break
		}
	}

	if dayIndex >= 0 {
		// Удаляем день
		session.CurrentTimeAvailability.SpecificDays = append(days[:dayIndex], days[dayIndex+1:]...)
	} else {
		// Добавляем день
		session.CurrentTimeAvailability.SpecificDays = append(days, day)
	}

	session.LastActivity = time.Now()
	e.saveEditSession(session)

	return e.ShowSpecificDaysSelection(callback, session, user)
}

// =============================================================================
// МЕТОДЫ РЕДАКТИРОВАНИЯ ВРЕМЕНИ
// =============================================================================

// EditTimeSlots переходит к редактированию временных слотов
func (e *IsolatedAvailabilityEditor) EditTimeSlots(callback *tgbotapi.CallbackQuery, user *models.User) error {
	session, err := e.getEditSession(user.ID)
	if err != nil {
		return err
	}

	session.CurrentStep = "time"
	session.LastActivity = time.Now()
	e.saveEditSession(session)

	return e.ShowTimeSlotsSelection(callback, session, user)
}

// ShowTimeSlotsSelection показывает выбор временных слотов
func (e *IsolatedAvailabilityEditor) ShowTimeSlotsSelection(callback *tgbotapi.CallbackQuery, session *AvailabilityEditSession, user *models.User) error {
	lang := user.InterfaceLanguageCode
	localizer := e.baseHandler.service.Localizer

	// Форматируем выбранные слоты
	selectedSlots := e.formatSelectedTimeSlots(session.CurrentTimeAvailability.TimeSlots, lang)

	message := fmt.Sprintf("%s\n\n%s: %s",
		localizer.Get(lang, "select_time_slot"),
		localizer.Get(lang, "selected_slots"),
		selectedSlots,
	)

	keyboard := e.createTimeSlotsKeyboard(session, lang)

	return e.baseHandler.messageFactory.EditWithKeyboard(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		message,
		&keyboard,
	)
}

// ToggleTimeSlot переключает выбор временного слота
func (e *IsolatedAvailabilityEditor) ToggleTimeSlot(callback *tgbotapi.CallbackQuery, user *models.User, slot string) error {
	session, err := e.getEditSession(user.ID)
	if err != nil {
		return err
	}

	// Переключаем слот в массиве
	slots := session.CurrentTimeAvailability.TimeSlots
	slotIndex := -1
	for i, s := range slots {
		if s == slot {
			slotIndex = i
			break
		}
	}

	if slotIndex >= 0 {
		// Удаляем слот
		session.CurrentTimeAvailability.TimeSlots = append(slots[:slotIndex], slots[slotIndex+1:]...)
	} else {
		// Добавляем слот
		session.CurrentTimeAvailability.TimeSlots = append(slots, slot)
	}

	session.LastActivity = time.Now()
	e.saveEditSession(session)

	return e.ShowTimeSlotsSelection(callback, session, user)
}

// =============================================================================
// МЕТОДЫ РЕДАКТИРОВАНИЯ СПОСОБОВ ОБЩЕНИЯ
// =============================================================================

// EditCommunication переходит к редактированию способов общения
func (e *IsolatedAvailabilityEditor) EditCommunication(callback *tgbotapi.CallbackQuery, user *models.User) error {
	session, err := e.getEditSession(user.ID)
	if err != nil {
		return err
	}

	session.CurrentStep = "communication"
	session.LastActivity = time.Now()
	e.saveEditSession(session)

	return e.ShowCommunicationSelection(callback, session, user)
}

// ShowCommunicationSelection показывает выбор способов общения
func (e *IsolatedAvailabilityEditor) ShowCommunicationSelection(callback *tgbotapi.CallbackQuery, session *AvailabilityEditSession, user *models.User) error {
	lang := user.InterfaceLanguageCode
	localizer := e.baseHandler.service.Localizer

	// Форматируем выбранные способы
	selectedStyles := e.formatSelectedCommunicationStyles(session.CurrentPreferences.CommunicationStyles, lang)

	message := fmt.Sprintf("%s\n\n%s: %s",
		localizer.Get(lang, "select_communication_style"),
		localizer.Get(lang, "selected_styles"),
		selectedStyles,
	)

	keyboard := e.createCommunicationKeyboard(session, lang)

	return e.baseHandler.messageFactory.EditWithKeyboard(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		message,
		&keyboard,
	)
}

// ToggleCommunicationStyle переключает выбор способа общения
func (e *IsolatedAvailabilityEditor) ToggleCommunicationStyle(callback *tgbotapi.CallbackQuery, user *models.User, style string) error {
	session, err := e.getEditSession(user.ID)
	if err != nil {
		return err
	}

	// Переключаем стиль в массиве
	styles := session.CurrentPreferences.CommunicationStyles
	styleIndex := -1
	for i, s := range styles {
		if s == style {
			styleIndex = i
			break
		}
	}

	if styleIndex >= 0 {
		// Удаляем стиль
		session.CurrentPreferences.CommunicationStyles = append(styles[:styleIndex], styles[styleIndex+1:]...)
	} else {
		// Добавляем стиль
		session.CurrentPreferences.CommunicationStyles = append(styles, style)
	}

	session.LastActivity = time.Now()
	e.saveEditSession(session)

	return e.ShowCommunicationSelection(callback, session, user)
}

// =============================================================================
// МЕТОДЫ РЕДАКТИРОВАНИЯ ЧАСТОТЫ
// =============================================================================

// EditFrequency переходит к редактированию частоты общения
func (e *IsolatedAvailabilityEditor) EditFrequency(callback *tgbotapi.CallbackQuery, user *models.User) error {
	session, err := e.getEditSession(user.ID)
	if err != nil {
		return err
	}

	session.CurrentStep = "frequency"
	session.LastActivity = time.Now()
	e.saveEditSession(session)

	return e.ShowFrequencySelection(callback, session, user)
}

// ShowFrequencySelection показывает выбор частоты общения
func (e *IsolatedAvailabilityEditor) ShowFrequencySelection(callback *tgbotapi.CallbackQuery, session *AvailabilityEditSession, user *models.User) error {
	lang := user.InterfaceLanguageCode
	localizer := e.baseHandler.service.Localizer

	message := localizer.Get(lang, "select_communication_frequency")

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				localizer.Get(lang, localization.LocaleFreqMultipleWeekly),
				localization.CallbackAvailEditFreqMultipleWeekly,
			),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				localizer.Get(lang, localization.LocaleFreqWeekly),
				localization.CallbackAvailEditFreqWeekly,
			),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				localizer.Get(lang, localization.LocaleFreqMultipleMonthly),
				localization.CallbackAvailEditFreqMultipleMonthly,
			),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				localizer.Get(lang, localization.LocaleFreqFlexible),
				localization.CallbackAvailEditFreqFlexible,
			),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				localizer.Get(lang, localization.LocaleBackToEditMenu),
				localization.CallbackAvailBackToEditMenu,
			),
		),
	)

	return e.baseHandler.messageFactory.EditWithKeyboard(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		message,
		&keyboard,
	)
}

// HandleFrequencySelection обрабатывает выбор частоты
func (e *IsolatedAvailabilityEditor) HandleFrequencySelection(callback *tgbotapi.CallbackQuery, user *models.User, frequency string) error {
	session, err := e.getEditSession(user.ID)
	if err != nil {
		return err
	}

	// Записываем изменение
	e.recordChange(session, "frequency", session.CurrentPreferences.CommunicationFreq, frequency)

	// Обновляем сессию
	session.CurrentPreferences.CommunicationFreq = frequency
	session.LastActivity = time.Now()
	e.saveEditSession(session)

	return e.ShowEditMenu(callback, session, user)
}

// =============================================================================
// МЕТОДЫ СОХРАНЕНИЯ И ОТМЕНЫ
// =============================================================================

// SaveChanges сохраняет все изменения
func (e *IsolatedAvailabilityEditor) SaveChanges(callback *tgbotapi.CallbackQuery, user *models.User) error {
	loggingService := e.baseHandler.service.LoggingService
	telegramLogger := loggingService.Telegram()

	session, err := e.getEditSession(user.ID)
	if err != nil {
		telegramLogger.ErrorWithContext("Failed to retrieve edit session for saving", "", int64(user.ID), callback.Message.Chat.ID, "SaveChanges", map[string]interface{}{
			"user_id": user.ID,
			"error":   err.Error(),
		})
		return err
	}

	telegramLogger.InfoWithContext("Starting save operation", "", int64(user.ID), callback.Message.Chat.ID, "SaveChanges", map[string]interface{}{
		"user_id":                  user.ID,
		"changes_count":            len(session.Changes),
		"session_duration_seconds": time.Since(session.SessionStart).Seconds(),
		"current_step":             session.CurrentStep,
	})

	// Валидируем данные перед сохранением
	if err := e.validateSessionData(session, user.InterfaceLanguageCode); err != nil {
		telegramLogger.ErrorWithContext("Validation failed during save", "", int64(user.ID), callback.Message.Chat.ID, "SaveChanges", map[string]interface{}{
			"user_id":          user.ID,
			"error":            err.Error(),
			"validation_error": true,
		})
		return err
	}

	telegramLogger.InfoWithContext("Validation passed, proceeding with save", "", int64(user.ID), callback.Message.Chat.ID, "SaveChanges", map[string]interface{}{
		"user_id": user.ID,
		"final_time_availability": map[string]interface{}{
			"day_type":            session.CurrentTimeAvailability.DayType,
			"specific_days_count": len(session.CurrentTimeAvailability.SpecificDays),
			"time_slots_count":    len(session.CurrentTimeAvailability.TimeSlots),
		},
		"final_preferences": map[string]interface{}{
			"activity_type":              session.CurrentPreferences.ActivityType,
			"communication_styles_count": len(session.CurrentPreferences.CommunicationStyles),
			"communication_freq":         session.CurrentPreferences.CommunicationFreq,
		},
	})

	// Сохраняем данные в базу
	if err := e.baseHandler.service.SaveTimeAvailability(user.ID, session.CurrentTimeAvailability); err != nil {
		loggingService.LogErrorWithContext(err, "", int64(user.ID), callback.Message.Chat.ID, "SaveChanges", "database")
		return err
	}

	if err := e.baseHandler.service.SaveFriendshipPreferences(user.ID, session.CurrentPreferences); err != nil {
		loggingService.LogErrorWithContext(err, "", int64(user.ID), callback.Message.Chat.ID, "SaveChanges", "database")
		return err
	}

	// Обновляем статус пользователя
	if err := e.baseHandler.service.UpdateUserState(user.ID, models.StateActive); err != nil {
		telegramLogger.ErrorWithContext("Failed to update user state after successful save", "", int64(user.ID), callback.Message.Chat.ID, "SaveChanges", map[string]interface{}{
			"user_id":              user.ID,
			"error":                err.Error(),
			"error_type":           "state_update_error",
			"data_save_successful": true,
		})
		loggingService.LogErrorWithContext(err, "", int64(user.ID), callback.Message.Chat.ID, "SaveChanges", "database")
		// Не возвращаем ошибку, так как основные данные уже сохранены
	}

	telegramLogger.InfoWithContext("Availability data saved successfully", "", int64(user.ID), callback.Message.Chat.ID, "SaveChanges", map[string]interface{}{
		"user_id":                        user.ID,
		"changes_applied":                len(session.Changes),
		"total_session_duration_seconds": time.Since(session.SessionStart).Seconds(),
		"data_persisted":                 true,
		"user_state_updated":             true,
	})

	// Очищаем сессию
	e.clearEditSession(user.ID)

	telegramLogger.InfoWithContext("Edit session cleaned up after successful save", "", int64(user.ID), callback.Message.Chat.ID, "SaveChanges", map[string]interface{}{
		"user_id":       user.ID,
		"cache_cleared": true,
	})

	loggingService.LogRequestEnd("", int64(user.ID), callback.Message.Chat.ID, "SaveChanges", true)

	// Показываем подтверждение
	return e.ShowSaveConfirmation(callback, session, user)
}

// CancelEdit отменяет редактирование
func (e *IsolatedAvailabilityEditor) CancelEdit(callback *tgbotapi.CallbackQuery, user *models.User) error {
	loggingService := e.baseHandler.service.LoggingService
	telegramLogger := loggingService.Telegram()

	// Получаем информацию о сессии перед очисткой для логирования
	session, err := e.getEditSession(user.ID)
	sessionInfo := map[string]interface{}{
		"user_id":           user.ID,
		"session_retrieved": err == nil,
	}
	if err == nil {
		sessionInfo["changes_count"] = len(session.Changes)
		sessionInfo["session_duration_seconds"] = time.Since(session.SessionStart).Seconds()
		sessionInfo["current_step"] = session.CurrentStep
	}

	telegramLogger.InfoWithContext("Cancelling availability edit session", "", int64(user.ID), callback.Message.Chat.ID, "CancelEdit", sessionInfo)

	// Очищаем сессию
	e.clearEditSession(user.ID)

	telegramLogger.InfoWithContext("Availability edit session cancelled successfully", "", int64(user.ID), callback.Message.Chat.ID, "CancelEdit", map[string]interface{}{
		"user_id":       user.ID,
		"cache_cleared": true,
	})

	loggingService.LogRequestEnd("", int64(user.ID), callback.Message.Chat.ID, "CancelEdit", false)

	lang := user.InterfaceLanguageCode
	localizer := e.baseHandler.service.Localizer

	message := fmt.Sprintf("%s\n\n%s",
		localizer.Get(lang, "edit_cancelled"),
		localizer.Get(lang, "changes_not_saved"),
	)

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				localizer.Get(lang, "back_to_profile"),
				"view_profile",
			),
		),
	)

	return e.baseHandler.messageFactory.EditWithKeyboard(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		message,
		&keyboard,
	)
}

// ShowSaveConfirmation показывает подтверждение сохранения
func (e *IsolatedAvailabilityEditor) ShowSaveConfirmation(callback *tgbotapi.CallbackQuery, session *AvailabilityEditSession, user *models.User) error {
	lang := user.InterfaceLanguageCode
	localizer := e.baseHandler.service.Localizer

	changesSummary := e.formatChangesSummary(session, lang)

	message := fmt.Sprintf("%s\n\n%s\n\n%s",
		localizer.Get(lang, "changes_saved_successfully"),
		changesSummary,
		localizer.Get(lang, "redirecting_to_profile"),
	)

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				localizer.Get(lang, "profile_show"),
				"view_profile",
			),
		),
	)

	return e.baseHandler.messageFactory.EditWithKeyboard(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		message,
		&keyboard,
	)
}

// =============================================================================
// ВСПОМОГАТЕЛЬНЫЕ МЕТОДЫ
// =============================================================================

// GetEditSession получает сессию из кеша (публичный метод)
func (e *IsolatedAvailabilityEditor) GetEditSession(userID int) (*AvailabilityEditSession, error) {
	return e.getEditSession(userID)
}

// getEditSession получает сессию из кеша
func (e *IsolatedAvailabilityEditor) getEditSession(userID int) (*AvailabilityEditSession, error) {
	cacheKey := fmt.Sprintf("availability_edit_session:%d", userID)

	var data string
	err := e.baseHandler.service.Cache.Get(context.Background(), cacheKey, &data)
	if err != nil {
		return nil, fmt.Errorf("failed to get edit session from cache: %w", err)
	}

	if data == "" {
		return nil, fmt.Errorf("edit session not found")
	}

	var session AvailabilityEditSession
	if err := json.Unmarshal([]byte(data), &session); err != nil {
		return nil, fmt.Errorf("failed to unmarshal edit session: %w", err)
	}

	return &session, nil
}

// saveEditSession сохраняет сессию в кеш
func (e *IsolatedAvailabilityEditor) saveEditSession(session *AvailabilityEditSession) error {
	cacheKey := fmt.Sprintf("availability_edit_session:%d", session.UserID)

	return e.baseHandler.service.Cache.Set(context.Background(), cacheKey, session, 30*time.Minute)
}

// clearEditSession очищает сессию из кеша
func (e *IsolatedAvailabilityEditor) clearEditSession(userID int) {
	cacheKey := fmt.Sprintf("availability_edit_session:%d", userID)
	e.baseHandler.service.Cache.Delete(context.Background(), cacheKey)
}

// recordChange записывает изменение в сессию
func (e *IsolatedAvailabilityEditor) recordChange(session *AvailabilityEditSession, field string, oldValue, newValue interface{}) {
	change := AvailabilityChange{
		Field:     field,
		OldValue:  oldValue,
		NewValue:  newValue,
		Timestamp: time.Now(),
	}

	session.Changes = append(session.Changes, change)

	// Логируем каждое изменение
	loggingService := e.baseHandler.service.LoggingService.Telegram()
	loggingService.InfoWithContext("Availability data changed", "", int64(session.UserID), 0, "RecordChange", map[string]interface{}{
		"user_id":                  session.UserID,
		"field":                    field,
		"old_value":                oldValue,
		"new_value":                newValue,
		"change_index":             len(session.Changes),
		"session_duration_seconds": time.Since(session.SessionStart).Seconds(),
	})
}

// validateSessionData валидирует данные сессии
func (e *IsolatedAvailabilityEditor) validateSessionData(session *AvailabilityEditSession, lang string) error {
	// Валидируем временную доступность
	if err := e.baseHandler.service.ValidateTimeAvailability(session.CurrentTimeAvailability, lang); err != nil {
		return err
	}

	// Валидируем предпочтения общения
	if err := e.baseHandler.service.ValidateFriendshipPreferences(session.CurrentPreferences, lang); err != nil {
		return err
	}

	return nil
}

// deepCopyTimeAvailability создает глубокую копию TimeAvailability
func (e *IsolatedAvailabilityEditor) deepCopyTimeAvailability(original *models.TimeAvailability) *models.TimeAvailability {
	if original == nil {
		return &models.TimeAvailability{
			DayType:      "any",
			SpecificDays: []string{},
			TimeSlots:    []string{"any"},
		}
	}

	specificDays := make([]string, len(original.SpecificDays))
	copy(specificDays, original.SpecificDays)

	timeSlots := make([]string, len(original.TimeSlots))
	copy(timeSlots, original.TimeSlots)

	return &models.TimeAvailability{
		DayType:      original.DayType,
		SpecificDays: specificDays,
		TimeSlots:    timeSlots,
	}
}

// deepCopyFriendshipPreferences создает глубокую копию FriendshipPreferences
func (e *IsolatedAvailabilityEditor) deepCopyFriendshipPreferences(original *models.FriendshipPreferences) *models.FriendshipPreferences {
	if original == nil {
		return &models.FriendshipPreferences{
			ActivityType:        "casual_chat",
			CommunicationStyles: []string{"text"},
			CommunicationFreq:   "weekly",
		}
	}

	styles := make([]string, len(original.CommunicationStyles))
	copy(styles, original.CommunicationStyles)

	return &models.FriendshipPreferences{
		ActivityType:        original.ActivityType,
		CommunicationStyles: styles,
		CommunicationFreq:   original.CommunicationFreq,
	}
}

// =============================================================================
// МЕТОДЫ ФОРМАТИРОВАНИЯ ДЛЯ ОТОБРАЖЕНИЯ
// =============================================================================

// formatCurrentTimeAvailability форматирует текущую временную доступность
func (e *IsolatedAvailabilityEditor) formatCurrentTimeAvailability(availability *models.TimeAvailability, lang string) string {
	if availability == nil {
		return "⏰ " + e.baseHandler.service.Localizer.Get(lang, localization.LocaleErrorInvalidAvailabilityData)
	}

	parts := []string{"⏰"}

	// Форматируем дни
	switch availability.DayType {
	case "weekdays":
		parts = append(parts, e.baseHandler.service.Localizer.Get(lang, localization.LocaleTimeWeekdays))
	case "weekends":
		parts = append(parts, e.baseHandler.service.Localizer.Get(lang, localization.LocaleTimeWeekends))
	case "any":
		parts = append(parts, e.baseHandler.service.Localizer.Get(lang, localization.LocaleTimeAny))
	case "specific":
		if len(availability.SpecificDays) > 0 {
			days := make([]string, len(availability.SpecificDays))
			for i, day := range availability.SpecificDays {
				days[i] = e.formatDayName(day, lang)
			}
			parts = append(parts, strings.Join(days, ", "))
		} else {
			parts = append(parts, e.baseHandler.service.Localizer.Get(lang, localization.LocaleTimeAny))
		}
	}

	// Форматируем время
	if len(availability.TimeSlots) > 0 {
		timeParts := make([]string, len(availability.TimeSlots))
		for i, slot := range availability.TimeSlots {
			switch slot {
			case "morning":
				timeParts[i] = e.baseHandler.service.Localizer.Get(lang, localization.LocaleTimeMorning)
			case "day":
				timeParts[i] = e.baseHandler.service.Localizer.Get(lang, localization.LocaleTimeDay)
			case "evening":
				timeParts[i] = e.baseHandler.service.Localizer.Get(lang, localization.LocaleTimeEvening)
			case "late":
				timeParts[i] = e.baseHandler.service.Localizer.Get(lang, localization.LocaleTimeLate)
			}
		}
		parts = append(parts, strings.Join(timeParts, ", "))
	}

	return strings.Join(parts, " ")
}

// formatCurrentCommunicationPreferences форматирует текущие предпочтения общения
func (e *IsolatedAvailabilityEditor) formatCurrentCommunicationPreferences(preferences *models.FriendshipPreferences, lang string) string {
	if preferences == nil {
		return "💬 " + e.baseHandler.service.Localizer.Get(lang, localization.LocaleErrorInvalidAvailabilityData)
	}

	parts := []string{"💬"}

	// Форматируем способы общения
	if len(preferences.CommunicationStyles) > 0 {
		styleParts := make([]string, len(preferences.CommunicationStyles))
		for i, style := range preferences.CommunicationStyles {
			switch style {
			case "text":
				styleParts[i] = e.baseHandler.service.Localizer.Get(lang, localization.LocaleCommText)
			case "voice_msg":
				styleParts[i] = e.baseHandler.service.Localizer.Get(lang, localization.LocaleCommVoice)
			case "audio_call":
				styleParts[i] = e.baseHandler.service.Localizer.Get(lang, localization.LocaleCommAudio)
			case "video_call":
				styleParts[i] = e.baseHandler.service.Localizer.Get(lang, localization.LocaleCommVideo)
			case "meet_person":
				styleParts[i] = e.baseHandler.service.Localizer.Get(lang, localization.LocaleCommMeet)
			}
		}
		parts = append(parts, strings.Join(styleParts, ", "))
	}

	return strings.Join(parts, " ")
}

// formatCurrentFrequency форматирует текущую частоту
func (e *IsolatedAvailabilityEditor) formatCurrentFrequency(preferences *models.FriendshipPreferences, lang string) string {
	if preferences == nil {
		return "📊 " + e.baseHandler.service.Localizer.Get(lang, localization.LocaleErrorInvalidAvailabilityData)
	}

	freqText := "📊 "
	switch preferences.CommunicationFreq {
	case "multiple_weekly":
		freqText += e.baseHandler.service.Localizer.Get(lang, localization.LocaleFreqMultipleWeekly)
	case "weekly":
		freqText += e.baseHandler.service.Localizer.Get(lang, localization.LocaleFreqWeekly)
	case "multiple_monthly":
		freqText += e.baseHandler.service.Localizer.Get(lang, localization.LocaleFreqMultipleMonthly)
	case "flexible":
		freqText += e.baseHandler.service.Localizer.Get(lang, localization.LocaleFreqFlexible)
	default:
		freqText += e.baseHandler.service.Localizer.Get(lang, localization.LocaleFreqWeekly)
	}

	return freqText
}

// formatSelectedDays форматирует выбранные дни
func (e *IsolatedAvailabilityEditor) formatSelectedDays(days []string, lang string) string {
	if len(days) == 0 {
		return e.baseHandler.service.Localizer.Get(lang, "no_days_selected")
	}

	dayNames := make([]string, len(days))
	for i, day := range days {
		dayNames[i] = e.formatDayName(day, lang)
	}

	return strings.Join(dayNames, ", ")
}

// formatDayName форматирует название дня
func (e *IsolatedAvailabilityEditor) formatDayName(day, lang string) string {
	switch day {
	case "monday":
		return e.baseHandler.service.Localizer.Get(lang, "day_monday")
	case "tuesday":
		return e.baseHandler.service.Localizer.Get(lang, "day_tuesday")
	case "wednesday":
		return e.baseHandler.service.Localizer.Get(lang, "day_wednesday")
	case "thursday":
		return e.baseHandler.service.Localizer.Get(lang, "day_thursday")
	case "friday":
		return e.baseHandler.service.Localizer.Get(lang, "day_friday")
	case "saturday":
		return e.baseHandler.service.Localizer.Get(lang, "day_saturday")
	case "sunday":
		return e.baseHandler.service.Localizer.Get(lang, "day_sunday")
	default:
		return day
	}
}

// formatSelectedTimeSlots форматирует выбранные временные слоты
func (e *IsolatedAvailabilityEditor) formatSelectedTimeSlots(slots []string, lang string) string {
	if len(slots) == 0 {
		return e.baseHandler.service.Localizer.Get(lang, "none_selected")
	}

	slotNames := make([]string, len(slots))
	for i, slot := range slots {
		switch slot {
		case "morning":
			slotNames[i] = e.baseHandler.service.Localizer.Get(lang, localization.LocaleTimeMorning)
		case "day":
			slotNames[i] = e.baseHandler.service.Localizer.Get(lang, localization.LocaleTimeDay)
		case "evening":
			slotNames[i] = e.baseHandler.service.Localizer.Get(lang, localization.LocaleTimeEvening)
		case "late":
			slotNames[i] = e.baseHandler.service.Localizer.Get(lang, localization.LocaleTimeLate)
		default:
			slotNames[i] = slot
		}
	}

	return strings.Join(slotNames, ", ")
}

// formatSelectedCommunicationStyles форматирует выбранные способы общения
func (e *IsolatedAvailabilityEditor) formatSelectedCommunicationStyles(styles []string, lang string) string {
	if len(styles) == 0 {
		return e.baseHandler.service.Localizer.Get(lang, "none_selected")
	}

	styleNames := make([]string, len(styles))
	for i, style := range styles {
		switch style {
		case "text":
			styleNames[i] = e.baseHandler.service.Localizer.Get(lang, localization.LocaleCommText)
		case "voice_msg":
			styleNames[i] = e.baseHandler.service.Localizer.Get(lang, localization.LocaleCommVoice)
		case "audio_call":
			styleNames[i] = e.baseHandler.service.Localizer.Get(lang, localization.LocaleCommAudio)
		case "video_call":
			styleNames[i] = e.baseHandler.service.Localizer.Get(lang, localization.LocaleCommVideo)
		case "meet_person":
			styleNames[i] = e.baseHandler.service.Localizer.Get(lang, localization.LocaleCommMeet)
		default:
			styleNames[i] = style
		}
	}

	return strings.Join(styleNames, ", ")
}

// formatChangesSummary форматирует сводку изменений
func (e *IsolatedAvailabilityEditor) formatChangesSummary(session *AvailabilityEditSession, lang string) string {
	if len(session.Changes) == 0 {
		return e.baseHandler.service.Localizer.Get(lang, "no_changes_made")
	}

	changes := make([]string, len(session.Changes))
	for i, change := range session.Changes {
		fieldName := e.formatFieldName(change.Field, lang)
		changes[i] = fmt.Sprintf("• %s: %v → %v", fieldName, change.OldValue, change.NewValue)
	}

	return strings.Join(changes, "\n")
}

// formatFieldName форматирует название поля для отображения
func (e *IsolatedAvailabilityEditor) formatFieldName(field, lang string) string {
	switch field {
	case "day_type":
		return e.baseHandler.service.Localizer.Get(lang, "time_weekdays") // Generic day field
	case "time_slots":
		return e.baseHandler.service.Localizer.Get(lang, "select_time_slot")
	case "communication_styles":
		return e.baseHandler.service.Localizer.Get(lang, "select_communication_style")
	case "frequency":
		return e.baseHandler.service.Localizer.Get(lang, "select_communication_frequency")
	default:
		return field
	}
}

// =============================================================================
// МЕТОДЫ СОЗДАНИЯ КЛАВИАТУР
// =============================================================================

// createEditMenuKeyboard создает клавиатуру главного меню редактирования
func (e *IsolatedAvailabilityEditor) createEditMenuKeyboard(session *AvailabilityEditSession, lang string) tgbotapi.InlineKeyboardMarkup {
	localizer := e.baseHandler.service.Localizer

	var rows [][]tgbotapi.InlineKeyboardButton

	// Кнопки редактирования
	rows = append(rows, []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData(
			"📅 "+localizer.Get(lang, "edit_days"),
			localization.CallbackAvailEditDays,
		),
	})

	rows = append(rows, []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData(
			"🕐 "+localizer.Get(lang, "edit_time"),
			localization.CallbackAvailEditTime,
		),
	})

	rows = append(rows, []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData(
			"💬 "+localizer.Get(lang, "edit_communication"),
			localization.CallbackAvailEditCommunication,
		),
	})

	rows = append(rows, []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData(
			"📊 "+localizer.Get(lang, "edit_frequency"),
			localization.CallbackAvailEditFrequency,
		),
	})

	// Кнопка сохранения (только если есть изменения)
	if len(session.Changes) > 0 {
		rows = append(rows, []tgbotapi.InlineKeyboardButton{
			tgbotapi.NewInlineKeyboardButtonData(
				"✅ "+localizer.Get(lang, localization.LocaleSaveChanges),
				localization.CallbackAvailSaveChanges,
			),
		})
	}

	// Кнопка отмены
	rows = append(rows, []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData(
			"❌ "+localizer.Get(lang, localization.LocaleCancelEdit),
			localization.CallbackAvailCancelEdit,
		),
	})

	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

// createSpecificDaysKeyboard создает клавиатуру выбора конкретных дней
func (e *IsolatedAvailabilityEditor) createSpecificDaysKeyboard(session *AvailabilityEditSession, lang string) tgbotapi.InlineKeyboardMarkup {
	localizer := e.baseHandler.service.Localizer
	selectedDays := make(map[string]bool)
	for _, day := range session.CurrentTimeAvailability.SpecificDays {
		selectedDays[day] = true
	}

	var rows [][]tgbotapi.InlineKeyboardButton

	// Дни недели (2 колонки)
	days := []string{"monday", "tuesday", "wednesday", "thursday", "friday", "saturday", "sunday"}
	for i := 0; i < len(days); i += 2 {
		var row []tgbotapi.InlineKeyboardButton

		// Первая колонка
		if i < len(days) {
			day := days[i]
			symbol := "☐"
			if selectedDays[day] {
				symbol = "☑"
			}
			buttonText := fmt.Sprintf("%s %s", symbol, e.formatDayName(day, lang))
			callbackData := fmt.Sprintf("%s_%s", localization.CallbackPrefixAvailEditDay, day)
			row = append(row, tgbotapi.NewInlineKeyboardButtonData(buttonText, callbackData))
		}

		// Вторая колонка
		if i+1 < len(days) {
			day := days[i+1]
			symbol := "☐"
			if selectedDays[day] {
				symbol = "☑"
			}
			buttonText := fmt.Sprintf("%s %s", symbol, e.formatDayName(day, lang))
			callbackData := fmt.Sprintf("%s_%s", localization.CallbackPrefixAvailEditDay, day)
			row = append(row, tgbotapi.NewInlineKeyboardButtonData(buttonText, callbackData))
		}

		rows = append(rows, row)
	}

	// Кнопки действий
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData(
			"✅ "+localizer.Get(lang, localization.LocaleSaveChanges),
			localization.CallbackAvailApplyDays,
		),
		tgbotapi.NewInlineKeyboardButtonData(
			localizer.Get(lang, localization.LocaleBackToEditMenu),
			localization.CallbackAvailBackToEditMenu,
		),
	))

	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

// createTimeSlotsKeyboard создает клавиатуру выбора временных слотов
func (e *IsolatedAvailabilityEditor) createTimeSlotsKeyboard(session *AvailabilityEditSession, lang string) tgbotapi.InlineKeyboardMarkup {
	localizer := e.baseHandler.service.Localizer
	selectedSlots := make(map[string]bool)
	for _, slot := range session.CurrentTimeAvailability.TimeSlots {
		selectedSlots[slot] = true
	}

	var rows [][]tgbotapi.InlineKeyboardButton

	// Временные слоты
	slots := []string{"morning", "day", "evening", "late"}
	for _, slot := range slots {
		symbol := "☐"
		if selectedSlots[slot] {
			symbol = "☑"
		}

		var slotText string
		switch slot {
		case "morning":
			slotText = localizer.Get(lang, localization.LocaleTimeMorning)
		case "day":
			slotText = localizer.Get(lang, localization.LocaleTimeDay)
		case "evening":
			slotText = localizer.Get(lang, localization.LocaleTimeEvening)
		case "late":
			slotText = localizer.Get(lang, localization.LocaleTimeLate)
		}

		buttonText := fmt.Sprintf("%s 🌅 %s", symbol, slotText)
		callbackData := fmt.Sprintf("%s_%s", localization.CallbackPrefixAvailEditTimeSlot, slot)

		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(buttonText, callbackData),
		))
	}

	// Кнопки действий
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData(
			"✅ "+localizer.Get(lang, localization.LocaleSaveChanges),
			localization.CallbackAvailApplyTime,
		),
		tgbotapi.NewInlineKeyboardButtonData(
			localizer.Get(lang, localization.LocaleBackToEditMenu),
			localization.CallbackAvailBackToEditMenu,
		),
	))

	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

// createCommunicationKeyboard создает клавиатуру выбора способов общения
func (e *IsolatedAvailabilityEditor) createCommunicationKeyboard(session *AvailabilityEditSession, lang string) tgbotapi.InlineKeyboardMarkup {
	localizer := e.baseHandler.service.Localizer
	selectedStyles := make(map[string]bool)
	for _, style := range session.CurrentPreferences.CommunicationStyles {
		selectedStyles[style] = true
	}

	var rows [][]tgbotapi.InlineKeyboardButton

	// Способы общения
	styles := []string{"text", "voice_msg", "audio_call", "video_call", "meet_person"}
	for _, style := range styles {
		symbol := "☐"
		if selectedStyles[style] {
			symbol = "☑"
		}

		var styleText string
		switch style {
		case "text":
			styleText = localizer.Get(lang, localization.LocaleCommText)
		case "voice_msg":
			styleText = localizer.Get(lang, localization.LocaleCommVoice)
		case "audio_call":
			styleText = localizer.Get(lang, localization.LocaleCommAudio)
		case "video_call":
			styleText = localizer.Get(lang, localization.LocaleCommVideo)
		case "meet_person":
			styleText = localizer.Get(lang, localization.LocaleCommMeet)
		}

		var emoji string
		switch style {
		case "text":
			emoji = "💬"
		case "voice_msg":
			emoji = "🎤"
		case "audio_call":
			emoji = "📞"
		case "video_call":
			emoji = "📹"
		case "meet_person":
			emoji = "🤝"
		}

		buttonText := fmt.Sprintf("%s %s %s", symbol, emoji, styleText)
		callbackData := fmt.Sprintf("%s_%s", localization.CallbackPrefixAvailEditCommStyle, style)

		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(buttonText, callbackData),
		))
	}

	// Кнопки действий
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData(
			"✅ "+localizer.Get(lang, localization.LocaleSaveChanges),
			localization.CallbackAvailApplyCommunication,
		),
		tgbotapi.NewInlineKeyboardButtonData(
			localizer.Get(lang, localization.LocaleBackToEditMenu),
			localization.CallbackAvailBackToEditMenu,
		),
	))

	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}
