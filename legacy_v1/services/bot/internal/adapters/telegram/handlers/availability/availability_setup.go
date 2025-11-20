package availability

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"language-exchange-bot/internal/localization"
	"language-exchange-bot/internal/models"

	"language-exchange-bot/internal/adapters/telegram/handlers/base"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// AvailabilityHandlerImpl handles availability setup and editing
type AvailabilityHandlerImpl struct {
	baseHandler *base.BaseHandler
}

// NewAvailabilityHandler creates a new availability handler
func NewAvailabilityHandler(baseHandler *base.BaseHandler) *AvailabilityHandlerImpl {
	return &AvailabilityHandlerImpl{
		baseHandler: baseHandler,
	}
}

// =============================================================================
// VALIDATION METHODS
// =============================================================================

// validateTimeAvailability validates time availability data
func (h *AvailabilityHandlerImpl) validateTimeAvailability(availability *models.TimeAvailability, lang string) error {
	if availability == nil {
		return fmt.Errorf("availability data is nil")
	}

	// Validate day type
	switch availability.DayType {
	case "weekdays", "weekends", "any":
		// Valid types, no additional validation needed
	case "specific":
		// Must have at least one specific day
		if len(availability.SpecificDays) == 0 {
			return fmt.Errorf("%s", h.baseHandler.Service.Localizer.Get(lang, localization.LocaleErrorNoDaysSelected))
		}
	default:
		return fmt.Errorf("%s", h.baseHandler.Service.Localizer.Get(lang, localization.LocaleErrorInvalidAvailabilityData))
	}

	// Validate time slots - must have at least one
	if len(availability.TimeSlots) == 0 {
		return fmt.Errorf("%s", h.baseHandler.Service.Localizer.Get(lang, localization.LocaleErrorNoTimeSelected))
	}

	// Validate time slot values
	validTimeSlots := map[string]bool{
		"morning": true,
		"day":     true,
		"evening": true,
		"late":    true,
	}

	for _, slot := range availability.TimeSlots {
		if !validTimeSlots[slot] {
			return fmt.Errorf("%s", h.baseHandler.Service.Localizer.Get(lang, localization.LocaleErrorInvalidAvailabilityData))
		}
	}

	return nil
}

// validateFriendshipPreferences validates friendship preferences data
func (h *AvailabilityHandlerImpl) validateFriendshipPreferences(preferences *models.FriendshipPreferences, lang string) error {
	if preferences == nil {
		return fmt.Errorf("friendship preferences data is nil")
	}

	// Validate activity type
	validActivityTypes := map[string]bool{
		"movies":      true,
		"games":       true,
		"educational": true,
		"casual_chat": true,
	}

	if !validActivityTypes[preferences.ActivityType] {
		return fmt.Errorf("%s", h.baseHandler.Service.Localizer.Get(lang, localization.LocaleErrorInvalidAvailabilityData))
	}

	// Validate communication styles - must have at least one
	if len(preferences.CommunicationStyles) == 0 {
		return fmt.Errorf("%s", h.baseHandler.Service.Localizer.Get(lang, localization.LocaleErrorNoCommunicationSelected))
	}

	// Validate communication style values
	validCommStyles := map[string]bool{
		"text":        true,
		"voice_msg":   true,
		"audio_call":  true,
		"video_call":  true,
		"meet_person": true,
	}

	for _, style := range preferences.CommunicationStyles {
		if !validCommStyles[style] {
			return fmt.Errorf("%s", h.baseHandler.Service.Localizer.Get(lang, localization.LocaleErrorInvalidAvailabilityData))
		}
	}

	// Validate communication frequency
	validFrequencies := map[string]bool{
		"multiple_weekly":  true,
		"weekly":           true,
		"multiple_monthly": true,
		"flexible":         true,
	}

	if !validFrequencies[preferences.CommunicationFreq] {
		return fmt.Errorf("%s", h.baseHandler.Service.Localizer.Get(lang, localization.LocaleErrorInvalidAvailabilityData))
	}

	return nil
}

// =============================================================================
// PLACEHOLDER METHODS (to be implemented in next phases)
// =============================================================================

// HandleTimeAvailabilityStart starts the time availability setup process
func (h *AvailabilityHandlerImpl) HandleTimeAvailabilityStart(callback *tgbotapi.CallbackQuery, user *models.User) error {
	loggingService := h.baseHandler.Service.LoggingService.Telegram()
	loggingService.InfoWithContext("Starting availability setup process", "", int64(user.ID), callback.Message.Chat.ID, "HandleTimeAvailabilityStart", map[string]interface{}{
		"user_id":            user.ID,
		"interface_language": user.InterfaceLanguageCode,
	})

	lang := user.InterfaceLanguageCode
	localizer := h.baseHandler.Service.Localizer

	// Показываем приветственное сообщение
	introMessage := fmt.Sprintf("%s\n\n%s",
		localizer.Get(lang, "time_availability_intro"),
		localizer.Get(lang, "select_day_type"),
	)

	// Создаем клавиатуру для выбора типа дней
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				localizer.Get(lang, localization.LocaleTimeWeekdays),
				"availability_daytype_weekdays",
			),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				localizer.Get(lang, localization.LocaleTimeWeekends),
				"availability_daytype_weekends",
			),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				localizer.Get(lang, localization.LocaleTimeAny),
				"availability_daytype_any",
			),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				localizer.Get(lang, "select_specific_days_button"),
				"availability_daytype_specific",
			),
		),
		tgbotapi.NewInlineKeyboardRow(
			h.baseHandler.KeyboardBuilder.CreateBackButton(lang, "back_to_primary_interests"),
		),
	)

	return h.baseHandler.MessageFactory.EditWithKeyboard(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		introMessage,
		&keyboard,
	)
}

// HandleDayTypeSelection handles day type selection
func (h *AvailabilityHandlerImpl) HandleDayTypeSelection(callback *tgbotapi.CallbackQuery, user *models.User, dayType string) error {
	loggingService := h.baseHandler.Service.LoggingService.Telegram()
	loggingService.InfoWithContext("Processing day type selection", "", int64(user.ID), callback.Message.Chat.ID, "HandleDayTypeSelection", map[string]interface{}{
		"user_id":           user.ID,
		"selected_day_type": dayType,
	})

	// Сохраняем выбранный тип дней в состояние пользователя или сессию
	// Для простоты используем Redis для хранения временного состояния настройки

	// Создаем временную запись в cache
	cacheKey := fmt.Sprintf("availability_setup:%d", user.ID)
	setupData := map[string]interface{}{
		"day_type":      dayType,
		"specific_days": []string{},
		"time_slots":    []string{},
		"current_step":  "day_type_selected",
	}

	// Сериализуем в JSON
	setupDataJSON, err := json.Marshal(setupData)
	if err != nil {
		return fmt.Errorf("failed to marshal setup data: %w", err)
	}

	// Сохраняем в cache на 30 минут
	err = h.baseHandler.Service.Cache.Set(context.Background(), cacheKey, string(setupDataJSON), 30*time.Minute)
	if err != nil {
		loggingService.ErrorWithContext("Failed to save setup data to cache", "", int64(user.ID), callback.Message.Chat.ID, "HandleDayTypeSelection", map[string]interface{}{
			"user_id": user.ID,
			"error":   err.Error(),
		})
	}

	// Если выбраны конкретные дни, показываем выбор дней
	if dayType == "specific" {
		return h.ShowSpecificDaysSelection(callback, user)
	}

	// Иначе переходим к выбору времени
	return h.ShowTimeSlotSelection(callback, user)
}

// HandleSpecificDaysSelection handles specific days selection
func (h *AvailabilityHandlerImpl) HandleSpecificDaysSelection(callback *tgbotapi.CallbackQuery, user *models.User, day string) error {
	loggingService := h.baseHandler.Service.LoggingService.Telegram()
	loggingService.InfoWithContext("Processing specific day selection", "", int64(user.ID), callback.Message.Chat.ID, "HandleSpecificDaysSelection", map[string]interface{}{
		"user_id":      user.ID,
		"selected_day": day,
	})

	// Получаем текущие данные настройки
	cacheKey := fmt.Sprintf("availability_setup:%d", user.ID)
	var setupDataStr string
	err := h.baseHandler.Service.Cache.Get(context.Background(), cacheKey, &setupDataStr)
	if err != nil {
		return fmt.Errorf("failed to get setup data: %w", err)
	}

	var setupData map[string]interface{}
	err = json.Unmarshal([]byte(setupDataStr), &setupData)
	if err != nil {
		return fmt.Errorf("failed to unmarshal setup data: %w", err)
	}

	// Получаем текущий список дней
	specificDays := setupData["specific_days"].([]interface{})
	specificDaysStr := make([]string, len(specificDays))
	for i, d := range specificDays {
		specificDaysStr[i] = d.(string)
	}

	// Переключаем день (добавляем или удаляем)
	dayIndex := -1
	for i, d := range specificDaysStr {
		if d == day {
			dayIndex = i
			break
		}
	}

	if dayIndex >= 0 {
		// Удаляем день
		specificDaysStr = append(specificDaysStr[:dayIndex], specificDaysStr[dayIndex+1:]...)
	} else {
		// Добавляем день
		specificDaysStr = append(specificDaysStr, day)
	}

	// Обновляем данные
	setupData["specific_days"] = specificDaysStr
	setupDataJSON, err := json.Marshal(setupData)
	if err != nil {
		return fmt.Errorf("failed to marshal updated setup data: %w", err)
	}

	err = h.baseHandler.Service.Cache.Set(context.Background(), cacheKey, string(setupDataJSON), 30*time.Minute)
	if err != nil {
		loggingService.ErrorWithContext("Failed to update setup data in cache", "", int64(user.ID), callback.Message.Chat.ID, "HandleSpecificDaysSelection", map[string]interface{}{
			"user_id": user.ID,
			"error":   err.Error(),
		})
	}

	// Показываем обновленный интерфейс выбора дней
	return h.ShowSpecificDaysSelection(callback, user)
}

// ShowSpecificDaysSelection shows specific days selection interface
func (h *AvailabilityHandlerImpl) ShowSpecificDaysSelection(callback *tgbotapi.CallbackQuery, user *models.User) error {
	loggingService := h.baseHandler.Service.LoggingService.Telegram()
	lang := user.InterfaceLanguageCode
	localizer := h.baseHandler.Service.Localizer

	// Получаем текущие выбранные дни
	cacheKey := fmt.Sprintf("availability_setup:%d", user.ID)
	var setupDataStr string
	err := h.baseHandler.Service.Cache.Get(context.Background(), cacheKey, &setupDataStr)
	if err != nil {
		return fmt.Errorf("failed to get setup data: %w", err)
	}

	var setupData map[string]interface{}
	err = json.Unmarshal([]byte(setupDataStr), &setupData)
	if err != nil {
		return fmt.Errorf("failed to unmarshal setup data: %w", err)
	}

	specificDays := setupData["specific_days"].([]interface{})
	selectedDays := make([]string, len(specificDays))
	for i, d := range specificDays {
		selectedDays[i] = d.(string)
	}

	// Сортируем дни по порядку недели для отображения
	weekOrder := []string{"monday", "tuesday", "wednesday", "thursday", "friday", "saturday", "sunday"}
	selectedDaysMap := make(map[string]bool)
	for _, day := range selectedDays {
		selectedDaysMap[day] = true
	}

	var sortedDayNames []string
	for _, day := range weekOrder {
		if selectedDaysMap[day] {
			switch day {
			case "monday":
				sortedDayNames = append(sortedDayNames, localizer.Get(lang, "day_monday"))
			case "tuesday":
				sortedDayNames = append(sortedDayNames, localizer.Get(lang, "day_tuesday"))
			case "wednesday":
				sortedDayNames = append(sortedDayNames, localizer.Get(lang, "day_wednesday"))
			case "thursday":
				sortedDayNames = append(sortedDayNames, localizer.Get(lang, "day_thursday"))
			case "friday":
				sortedDayNames = append(sortedDayNames, localizer.Get(lang, "day_friday"))
			case "saturday":
				sortedDayNames = append(sortedDayNames, localizer.Get(lang, "day_saturday"))
			case "sunday":
				sortedDayNames = append(sortedDayNames, localizer.Get(lang, "day_sunday"))
			}
		}
	}

	// Форматируем выбранные дни
	selectedDaysText := localizer.Get(lang, "no_days_selected")
	if len(sortedDayNames) > 0 {
		selectedDaysText = strings.Join(sortedDayNames, ", ")
	}

	message := fmt.Sprintf("%s\n\n%s: %s",
		localizer.Get(lang, "select_specific_days"),
		localizer.Get(lang, "selected_days"),
		selectedDaysText,
	)

	// Создаем клавиатуру с днями недели
	keyboard := tgbotapi.NewInlineKeyboardMarkup()

	// Дни недели в 2 колонки
	days := []string{"monday", "tuesday", "wednesday", "thursday", "friday", "saturday", "sunday"}
	selectedDaysMap = make(map[string]bool)
	for _, day := range selectedDays {
		selectedDaysMap[day] = true
	}

	for i := 0; i < len(days); i += 2 {
		var buttons []tgbotapi.InlineKeyboardButton

		// Первая колонка
		if i < len(days) {
			day := days[i]
			symbol := "☑"
			if selectedDaysMap[day] {
				symbol = "✅"
			}
			var dayName string
			switch day {
			case "monday":
				dayName = localizer.Get(lang, "day_monday")
			case "tuesday":
				dayName = localizer.Get(lang, "day_tuesday")
			case "wednesday":
				dayName = localizer.Get(lang, "day_wednesday")
			case "thursday":
				dayName = localizer.Get(lang, "day_thursday")
			case "friday":
				dayName = localizer.Get(lang, "day_friday")
			case "saturday":
				dayName = localizer.Get(lang, "day_saturday")
			case "sunday":
				dayName = localizer.Get(lang, "day_sunday")
			}
			buttonText := fmt.Sprintf("%s %s", symbol, dayName)
			buttons = append(buttons, tgbotapi.NewInlineKeyboardButtonData(buttonText, fmt.Sprintf("availability_specific_day_%s", day)))
		}

		// Вторая колонка
		if i+1 < len(days) {
			day := days[i+1]
			symbol := "☑"
			if selectedDaysMap[day] {
				symbol = "✅"
			}
			var dayName string
			switch day {
			case "monday":
				dayName = localizer.Get(lang, "day_monday")
			case "tuesday":
				dayName = localizer.Get(lang, "day_tuesday")
			case "wednesday":
				dayName = localizer.Get(lang, "day_wednesday")
			case "thursday":
				dayName = localizer.Get(lang, "day_thursday")
			case "friday":
				dayName = localizer.Get(lang, "day_friday")
			case "saturday":
				dayName = localizer.Get(lang, "day_saturday")
			case "sunday":
				dayName = localizer.Get(lang, "day_sunday")
			}
			buttonText := fmt.Sprintf("%s %s", symbol, dayName)
			buttons = append(buttons, tgbotapi.NewInlineKeyboardButtonData(buttonText, fmt.Sprintf("availability_specific_day_%s", day)))
		}

		keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, buttons)
	}

	// Кнопки "Назад" и "Продолжить"
	keyboard.InlineKeyboard = append(keyboard.InlineKeyboard,
		h.baseHandler.KeyboardBuilder.CreateNavigationRow(
			lang,
			"availability_back_to_daytype",
			"availability_proceed_to_time",
		),
	)

	// Редактируем текущее сообщение вместо отправки нового
	editMsg := tgbotapi.NewEditMessageTextAndMarkup(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		message,
		keyboard,
	)

	_, err = h.baseHandler.Bot.Send(editMsg)
	if err != nil {
		loggingService.ErrorWithContext("Failed to edit message", "", int64(user.ID), callback.Message.Chat.ID, "ShowSpecificDaysSelection", map[string]interface{}{
			"user_id": user.ID,
			"error":   err.Error(),
		})
		return err
	}

	return nil
}

// ShowTimeSlotSelectionWithError shows time slot selection interface with an error message
func (h *AvailabilityHandlerImpl) ShowTimeSlotSelectionWithError(callback *tgbotapi.CallbackQuery, user *models.User, errorKey string) error {
	loggingService := h.baseHandler.Service.LoggingService.Telegram()
	lang := user.InterfaceLanguageCode
	localizer := h.baseHandler.Service.Localizer

	// Получаем данные из кеша
	cacheKey := fmt.Sprintf("availability_setup:%d", user.ID)
	var setupDataStr string
	var setupData map[string]interface{}

	err := h.baseHandler.Service.Cache.Get(context.Background(), cacheKey, &setupDataStr)
	if err == nil {
		// Если данные есть, используем их
		err = json.Unmarshal([]byte(setupDataStr), &setupData)
		if err != nil {
			setupData = make(map[string]interface{})
		}
	} else {
		setupData = make(map[string]interface{})
	}

	var timeSlots []interface{}
	if slots, ok := setupData["time_slots"].([]interface{}); ok {
		timeSlots = slots
	}
	selectedSlots := make([]string, 0)
	for _, t := range timeSlots {
		if t != nil {
			if slot, ok := t.(string); ok && slot != "" {
				selectedSlots = append(selectedSlots, slot)
			}
		}
	}

	// Сортируем слоты по порядку времени суток для отображения
	timeOrder := []string{"morning", "day", "evening", "late"}
	selectedSlotsMap := make(map[string]bool)
	for _, slot := range selectedSlots {
		selectedSlotsMap[slot] = true
	}

	var sortedSlotNames []string
	for _, slot := range timeOrder {
		if selectedSlotsMap[slot] {
			switch slot {
			case "morning":
				sortedSlotNames = append(sortedSlotNames, localizer.Get(lang, localization.LocaleTimeMorning))
			case "day":
				sortedSlotNames = append(sortedSlotNames, localizer.Get(lang, localization.LocaleTimeDay))
			case "evening":
				sortedSlotNames = append(sortedSlotNames, localizer.Get(lang, localization.LocaleTimeEvening))
			case "late":
				sortedSlotNames = append(sortedSlotNames, localizer.Get(lang, localization.LocaleTimeLate))
			}
		}
	}

	// Форматируем выбранные слоты
	selectedSlotsText := localizer.Get(lang, "none_selected")
	if len(sortedSlotNames) > 0 {
		selectedSlotsText = strings.Join(sortedSlotNames, ", ")
	}

	// Формируем сообщение с ошибкой
	errorMessage := localizer.Get(lang, errorKey)
	message := fmt.Sprintf("%s\n\n%s: %s\n\n%s",
		localizer.Get(lang, "select_time_slot"),
		localizer.Get(lang, "selected_slots"),
		selectedSlotsText,
		errorMessage,
	)

	// Создаем клавиатуру с временными слотами
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				fmt.Sprintf("%s %s", getSlotSymbol("morning", selectedSlotsMap), localizer.Get(lang, localization.LocaleTimeMorning)),
				"availability_timeslot_morning",
			),
			tgbotapi.NewInlineKeyboardButtonData(
				fmt.Sprintf("%s %s", getSlotSymbol("day", selectedSlotsMap), localizer.Get(lang, localization.LocaleTimeDay)),
				"availability_timeslot_day",
			),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				fmt.Sprintf("%s %s", getSlotSymbol("evening", selectedSlotsMap), localizer.Get(lang, localization.LocaleTimeEvening)),
				"availability_timeslot_evening",
			),
			tgbotapi.NewInlineKeyboardButtonData(
				fmt.Sprintf("%s %s", getSlotSymbol("late", selectedSlotsMap), localizer.Get(lang, localization.LocaleTimeLate)),
				"availability_timeslot_late",
			),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				fmt.Sprintf("✅ %s", localizer.Get(lang, "select_all")),
				"availability_timeslot_select_all",
			),
		),
		h.baseHandler.KeyboardBuilder.CreateNavigationRow(
			lang,
			"availability_back_to_days",
			"availability_proceed_to_communication",
		),
	)

	// Редактируем текущее сообщение
	editMsg := tgbotapi.NewEditMessageTextAndMarkup(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		message,
		keyboard,
	)

	_, err = h.baseHandler.Bot.Send(editMsg)
	if err != nil {
		loggingService.ErrorWithContext("Failed to edit message with error", "", int64(user.ID), callback.Message.Chat.ID, "ShowTimeSlotSelectionWithError", map[string]interface{}{
			"user_id": user.ID,
			"error":   err.Error(),
		})
		return err
	}

	return nil
}

// ShowTimeSlotSelection shows time slot selection interface
func (h *AvailabilityHandlerImpl) ShowTimeSlotSelection(callback *tgbotapi.CallbackQuery, user *models.User) error {
	loggingService := h.baseHandler.Service.LoggingService.Telegram()
	loggingService.InfoWithContext("Showing time slot selection", "", int64(user.ID), callback.Message.Chat.ID, "ShowTimeSlotSelection", map[string]interface{}{
		"user_id": user.ID,
	})

	// Проверяем, что пользователь выбрал тип дней
	cacheKey := fmt.Sprintf("availability_setup:%d", user.ID)
	var setupDataStr string
	err := h.baseHandler.Service.Cache.Get(context.Background(), cacheKey, &setupDataStr)
	if err != nil {
		lang := user.InterfaceLanguageCode
		localizer := h.baseHandler.Service.Localizer
		return h.baseHandler.MessageFactory.EditText(
			callback.Message.Chat.ID,
			callback.Message.MessageID,
			localizer.Get(lang, "error_no_days_selected"),
		)
	}

	var setupData map[string]interface{}
	err = json.Unmarshal([]byte(setupDataStr), &setupData)
	if err != nil {
		return fmt.Errorf("failed to unmarshal setup data: %w", err)
	}

	dayType := setupData["day_type"].(string)
	specificDays := setupData["specific_days"].([]interface{})

	// Проверяем валидность выбора дней
	if dayType == "specific" && len(specificDays) == 0 {
		lang := user.InterfaceLanguageCode
		localizer := h.baseHandler.Service.Localizer
		return h.baseHandler.MessageFactory.EditText(
			callback.Message.Chat.ID,
			callback.Message.MessageID,
			localizer.Get(lang, "error_no_days_selected"),
		)
	}

	// Продолжаем работать с уже загруженными данными
	lang := user.InterfaceLanguageCode
	localizer := h.baseHandler.Service.Localizer

	timeSlots := setupData["time_slots"].([]interface{})
	selectedSlots := make([]string, len(timeSlots))
	for i, t := range timeSlots {
		selectedSlots[i] = t.(string)
	}

	// Сортируем слоты по порядку времени суток для отображения
	timeOrder := []string{"morning", "day", "evening", "late"}
	selectedSlotsMap := make(map[string]bool)
	for _, slot := range selectedSlots {
		selectedSlotsMap[slot] = true
	}

	var sortedSlotNames []string
	for _, slot := range timeOrder {
		if selectedSlotsMap[slot] {
			switch slot {
			case "morning":
				sortedSlotNames = append(sortedSlotNames, localizer.Get(lang, localization.LocaleTimeMorning))
			case "day":
				sortedSlotNames = append(sortedSlotNames, localizer.Get(lang, localization.LocaleTimeDay))
			case "evening":
				sortedSlotNames = append(sortedSlotNames, localizer.Get(lang, localization.LocaleTimeEvening))
			case "late":
				sortedSlotNames = append(sortedSlotNames, localizer.Get(lang, localization.LocaleTimeLate))
			}
		}
	}

	// Форматируем выбранные слоты
	selectedSlotsText := localizer.Get(lang, "none_selected")
	if len(sortedSlotNames) > 0 {
		selectedSlotsText = strings.Join(sortedSlotNames, ", ")
	}

	message := fmt.Sprintf("%s\n\n%s: %s",
		localizer.Get(lang, "select_time_slot"),
		localizer.Get(lang, "selected_slots"),
		selectedSlotsText,
	)

	// Создаем клавиатуру с временными слотами
	selectedSlotsMap = make(map[string]bool)
	for _, slot := range selectedSlots {
		selectedSlotsMap[slot] = true
	}

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				fmt.Sprintf("%s %s", getSlotSymbol("morning", selectedSlotsMap), localizer.Get(lang, localization.LocaleTimeMorning)),
				"availability_timeslot_morning",
			),
			tgbotapi.NewInlineKeyboardButtonData(
				fmt.Sprintf("%s %s", getSlotSymbol("day", selectedSlotsMap), localizer.Get(lang, localization.LocaleTimeDay)),
				"availability_timeslot_day",
			),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				fmt.Sprintf("%s %s", getSlotSymbol("evening", selectedSlotsMap), localizer.Get(lang, localization.LocaleTimeEvening)),
				"availability_timeslot_evening",
			),
			tgbotapi.NewInlineKeyboardButtonData(
				fmt.Sprintf("%s %s", getSlotSymbol("late", selectedSlotsMap), localizer.Get(lang, localization.LocaleTimeLate)),
				"availability_timeslot_late",
			),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				fmt.Sprintf("✅ %s", localizer.Get(lang, "select_all")),
				"availability_timeslot_select_all",
			),
		),
		h.baseHandler.KeyboardBuilder.CreateNavigationRow(
			lang,
			"availability_back_to_days",
			"availability_proceed_to_communication",
		),
	)

	// Редактируем текущее сообщение вместо отправки нового
	editMsg := tgbotapi.NewEditMessageTextAndMarkup(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		message,
		keyboard,
	)

	_, err = h.baseHandler.Bot.Send(editMsg)
	if err != nil {
		loggingService.ErrorWithContext("Failed to edit message", "", int64(user.ID), callback.Message.Chat.ID, "ShowTimeSlotSelection", map[string]interface{}{
			"user_id": user.ID,
			"error":   err.Error(),
		})
		return err
	}

	return nil
}

// Helper function to get checkbox symbol for time slots
func getSlotSymbol(slot string, selectedSlots map[string]bool) string {
	if selectedSlots[slot] {
		return "✅"
	}
	return "☑"
}

// HandleTimeSlotSelection handles time slot selection
func (h *AvailabilityHandlerImpl) HandleTimeSlotSelection(callback *tgbotapi.CallbackQuery, user *models.User, timeSlot string) error {
	loggingService := h.baseHandler.Service.LoggingService.Telegram()
	loggingService.InfoWithContext("Processing time slot selection", "", int64(user.ID), callback.Message.Chat.ID, "HandleTimeSlotSelection", map[string]interface{}{
		"user_id":            user.ID,
		"selected_time_slot": timeSlot,
	})

	// Получаем текущие данные настройки
	cacheKey := fmt.Sprintf("availability_setup:%d", user.ID)
	var setupDataStr string
	err := h.baseHandler.Service.Cache.Get(context.Background(), cacheKey, &setupDataStr)
	if err != nil {
		return fmt.Errorf("failed to get setup data: %w", err)
	}

	var setupData map[string]interface{}
	err = json.Unmarshal([]byte(setupDataStr), &setupData)
	if err != nil {
		return fmt.Errorf("failed to unmarshal setup data: %w", err)
	}

	// Получаем текущий список временных слотов
	timeSlots := setupData["time_slots"].([]interface{})
	timeSlotsStr := make([]string, len(timeSlots))
	for i, t := range timeSlots {
		timeSlotsStr[i] = t.(string)
	}

	// Переключаем слот (добавляем или удаляем)
	slotIndex := -1
	for i, s := range timeSlotsStr {
		if s == timeSlot {
			slotIndex = i
			break
		}
	}

	if slotIndex >= 0 {
		// Удаляем слот
		timeSlotsStr = append(timeSlotsStr[:slotIndex], timeSlotsStr[slotIndex+1:]...)
	} else {
		// Добавляем слот
		timeSlotsStr = append(timeSlotsStr, timeSlot)
	}

	// Обновляем данные
	setupData["time_slots"] = timeSlotsStr
	setupDataJSON, err := json.Marshal(setupData)
	if err != nil {
		return fmt.Errorf("failed to marshal updated setup data: %w", err)
	}

	err = h.baseHandler.Service.Cache.Set(context.Background(), cacheKey, string(setupDataJSON), 30*time.Minute)
	if err != nil {
		loggingService.ErrorWithContext("Failed to update setup data in cache", "", int64(user.ID), callback.Message.Chat.ID, "HandleTimeSlotSelection", map[string]interface{}{
			"user_id": user.ID,
			"error":   err.Error(),
		})
	}

	// Показываем обновленный интерфейс выбора времени
	return h.ShowTimeSlotSelection(callback, user)
}

// HandleFriendshipPreferencesStart starts friendship preferences setup
func (h *AvailabilityHandlerImpl) HandleFriendshipPreferencesStart(callback *tgbotapi.CallbackQuery, user *models.User) error {
	loggingService := h.baseHandler.Service.LoggingService.Telegram()
	loggingService.InfoWithContext("Starting friendship preferences setup", "", int64(user.ID), callback.Message.Chat.ID, "HandleFriendshipPreferencesStart", map[string]interface{}{
		"user_id": user.ID,
	})

	// Проверяем, что пользователь выбрал хотя бы один временной слот
	cacheKey := fmt.Sprintf("availability_setup:%d", user.ID)
	var setupDataStr string
	err := h.baseHandler.Service.Cache.Get(context.Background(), cacheKey, &setupDataStr)
	if err != nil {
		// Если нет данных в кеше, показываем ошибку в том же окне
		return h.ShowTimeSlotSelectionWithError(callback, user, "error_no_time_selected")
	}

	var setupData map[string]interface{}
	err = json.Unmarshal([]byte(setupDataStr), &setupData)
	if err != nil {
		return fmt.Errorf("failed to unmarshal setup data: %w", err)
	}

	timeSlots := setupData["time_slots"].([]interface{})
	if len(timeSlots) == 0 {
		// Показываем ошибку в том же окне, сохраняя клавиатуру
		return h.ShowTimeSlotSelectionWithError(callback, user, "error_no_time_selected")
	}

	// Инициализируем communication_styles в cache если их там нет (пустой массив)
	if _, exists := setupData["communication_styles"]; !exists {
		setupData["communication_styles"] = []string{} // Пустой массив, без выбранных по умолчанию
		setupDataJSON, _ := json.Marshal(setupData)
		h.baseHandler.Service.Cache.Set(context.Background(), cacheKey, string(setupDataJSON), 30*time.Minute)
	}

	// Показываем интерфейс выбора способов общения
	return h.ShowCommunicationStyleSelection(callback, user)
}

// HandleSelectAllTimeSlots handles selecting all time slots
func (h *AvailabilityHandlerImpl) HandleSelectAllTimeSlots(callback *tgbotapi.CallbackQuery, user *models.User) error {
	loggingService := h.baseHandler.Service.LoggingService.Telegram()
	loggingService.InfoWithContext("Selecting all time slots", "", int64(user.ID), callback.Message.Chat.ID, "HandleSelectAllTimeSlots", map[string]interface{}{
		"user_id": user.ID,
	})

	// Получаем текущие данные настройки
	cacheKey := fmt.Sprintf("availability_setup:%d", user.ID)
	var setupDataStr string
	err := h.baseHandler.Service.Cache.Get(context.Background(), cacheKey, &setupDataStr)
	if err != nil {
		return fmt.Errorf("failed to get setup data: %w", err)
	}

	var setupData map[string]interface{}
	err = json.Unmarshal([]byte(setupDataStr), &setupData)
	if err != nil {
		return fmt.Errorf("failed to unmarshal setup data: %w", err)
	}

	// Устанавливаем все временные слоты
	allTimeSlots := []string{"morning", "day", "evening", "late"}
	setupData["time_slots"] = allTimeSlots

	// Сохраняем обновленные данные
	setupDataJSON, err := json.Marshal(setupData)
	if err != nil {
		return fmt.Errorf("failed to marshal updated setup data: %w", err)
	}

	err = h.baseHandler.Service.Cache.Set(context.Background(), cacheKey, string(setupDataJSON), 30*time.Minute)
	if err != nil {
		loggingService.ErrorWithContext("Failed to update setup data in cache", "", int64(user.ID), callback.Message.Chat.ID, "HandleSelectAllTimeSlots", map[string]interface{}{
			"user_id": user.ID,
			"error":   err.Error(),
		})
	}

	// Показываем обновленный интерфейс выбора времени
	return h.ShowTimeSlotSelection(callback, user)
}

// HandleCommunicationStyleSelection handles individual communication style selection
func (h *AvailabilityHandlerImpl) HandleCommunicationStyleSelection(callback *tgbotapi.CallbackQuery, user *models.User, style string) error {
	loggingService := h.baseHandler.Service.LoggingService.Telegram()
	loggingService.InfoWithContext("Processing communication style selection", "", int64(user.ID), callback.Message.Chat.ID, "HandleCommunicationStyleSelection", map[string]interface{}{
		"user_id":        user.ID,
		"selected_style": style,
	})

	// Получаем текущие данные настройки
	cacheKey := fmt.Sprintf("availability_setup:%d", user.ID)
	var setupDataStr string
	err := h.baseHandler.Service.Cache.Get(context.Background(), cacheKey, &setupDataStr)
	if err != nil {
		return fmt.Errorf("failed to get setup data: %w", err)
	}

	var setupData map[string]interface{}
	err = json.Unmarshal([]byte(setupDataStr), &setupData)
	if err != nil {
		return fmt.Errorf("failed to unmarshal setup data: %w", err)
	}

	// Получаем текущий список стилей общения
	communicationStyles := setupData["communication_styles"].([]interface{})
	stylesStr := make([]string, len(communicationStyles))
	for i, s := range communicationStyles {
		stylesStr[i] = s.(string)
	}

	// Переключаем стиль (добавляем или удаляем)
	styleIndex := -1
	for i, s := range stylesStr {
		if s == style {
			styleIndex = i
			break
		}
	}

	if styleIndex >= 0 {
		// Удаляем стиль
		stylesStr = append(stylesStr[:styleIndex], stylesStr[styleIndex+1:]...)
	} else {
		// Добавляем стиль
		stylesStr = append(stylesStr, style)
	}

	// Обновляем данные
	setupData["communication_styles"] = stylesStr
	setupDataJSON, err := json.Marshal(setupData)
	if err != nil {
		return fmt.Errorf("failed to marshal updated setup data: %w", err)
	}

	loggingService.InfoWithContext("DEBUG: Saving communication styles", "", int64(user.ID), callback.Message.Chat.ID, "HandleCommunicationStyleSelection", map[string]interface{}{
		"debug":  true,
		"step":   "saving_styles",
		"styles": stylesStr,
	})

	err = h.baseHandler.Service.Cache.Set(context.Background(), cacheKey, string(setupDataJSON), 30*time.Minute)
	if err != nil {
		loggingService.ErrorWithContext("Failed to update setup data in cache", "", int64(user.ID), callback.Message.Chat.ID, "HandleCommunicationStyleSelection", map[string]interface{}{
			"user_id": user.ID,
			"error":   err.Error(),
		})
	} else {
		loggingService.InfoWithContext("DEBUG: Successfully saved communication styles", "", int64(user.ID), callback.Message.Chat.ID, "HandleCommunicationStyleSelection", map[string]interface{}{
			"debug": true,
			"step":  "saved_successfully",
		})
	}

	// Показываем обновленный интерфейс выбора общения
	return h.ShowCommunicationStyleSelection(callback, user)
}

// HandleSelectAllCommunication handles selecting all communication methods
func (h *AvailabilityHandlerImpl) HandleSelectAllCommunication(callback *tgbotapi.CallbackQuery, user *models.User) error {
	loggingService := h.baseHandler.Service.LoggingService.Telegram()
	loggingService.InfoWithContext("Selecting all communication methods", "", int64(user.ID), callback.Message.Chat.ID, "HandleSelectAllCommunication", map[string]interface{}{
		"user_id": user.ID,
	})

	// Получаем текущие данные настройки
	cacheKey := fmt.Sprintf("availability_setup:%d", user.ID)
	var setupDataStr string
	err := h.baseHandler.Service.Cache.Get(context.Background(), cacheKey, &setupDataStr)
	if err != nil {
		return fmt.Errorf("failed to get setup data: %w", err)
	}

	var setupData map[string]interface{}
	err = json.Unmarshal([]byte(setupDataStr), &setupData)
	if err != nil {
		return fmt.Errorf("failed to unmarshal setup data: %w", err)
	}

	// Устанавливаем все способы общения
	allCommunicationStyles := []string{"text", "voice_msg", "audio_call", "video_call", "meet_person"}
	setupData["communication_styles"] = allCommunicationStyles

	// Сохраняем обновленные данные
	setupDataJSON, err := json.Marshal(setupData)
	if err != nil {
		return fmt.Errorf("failed to marshal updated setup data: %w", err)
	}

	err = h.baseHandler.Service.Cache.Set(context.Background(), cacheKey, string(setupDataJSON), 30*time.Minute)
	if err != nil {
		loggingService.ErrorWithContext("Failed to update setup data in cache", "", int64(user.ID), callback.Message.Chat.ID, "HandleSelectAllCommunication", map[string]interface{}{
			"user_id": user.ID,
			"error":   err.Error(),
		})
	}

	// Показываем обновленный интерфейс выбора общения
	return h.ShowCommunicationStyleSelection(callback, user)
}

// ShowCommunicationStyleSelection shows communication style selection interface
func (h *AvailabilityHandlerImpl) ShowCommunicationStyleSelection(callback *tgbotapi.CallbackQuery, user *models.User) error {
	loggingService := h.baseHandler.Service.LoggingService.Telegram()
	lang := user.InterfaceLanguageCode
	localizer := h.baseHandler.Service.Localizer

	// Получаем текущие выбранные способы общения
	cacheKey := fmt.Sprintf("availability_setup:%d", user.ID)
	var setupDataStr string
	err := h.baseHandler.Service.Cache.Get(context.Background(), cacheKey, &setupDataStr)
	if err != nil {
		return fmt.Errorf("failed to get setup data: %w", err)
	}

	var setupData map[string]interface{}
	err = json.Unmarshal([]byte(setupDataStr), &setupData)
	if err != nil {
		return fmt.Errorf("failed to unmarshal setup data: %w", err)
	}

	communicationStyles := setupData["communication_styles"].([]interface{})
	selectedStyles := make([]string, len(communicationStyles))
	for i, s := range communicationStyles {
		selectedStyles[i] = s.(string)
	}

	// Форматируем выбранные способы общения
	selectedStylesText := localizer.Get(lang, "none_selected")
	if len(selectedStyles) > 0 {
		styleNames := make([]string, len(selectedStyles))
		for i, style := range selectedStyles {
			switch style {
			case "text":
				styleNames[i] = localizer.Get(lang, localization.LocaleCommText)
			case "voice_msg":
				styleNames[i] = localizer.Get(lang, localization.LocaleCommVoice)
			case "audio_call":
				styleNames[i] = localizer.Get(lang, localization.LocaleCommAudio)
			case "video_call":
				styleNames[i] = localizer.Get(lang, localization.LocaleCommVideo)
			case "meet_person":
				styleNames[i] = localizer.Get(lang, localization.LocaleCommMeet)
			default:
				styleNames[i] = style
			}
		}
		selectedStylesText = strings.Join(styleNames, ", ")
	}

	message := fmt.Sprintf("%s\n\n%s: %s",
		localizer.Get(lang, "select_communication_style"),
		localizer.Get(lang, "selected_styles"),
		selectedStylesText,
	)

	// Создаем клавиатуру с чекбоксами
	selectedStylesMap := make(map[string]bool)
	for _, style := range selectedStyles {
		selectedStylesMap[style] = true
	}

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				fmt.Sprintf("%s %s", getStyleSymbol("text", selectedStylesMap), localizer.Get(lang, localization.LocaleCommText)),
				"availability_communication_text",
			),
			tgbotapi.NewInlineKeyboardButtonData(
				fmt.Sprintf("%s %s", getStyleSymbol("voice_msg", selectedStylesMap), localizer.Get(lang, localization.LocaleCommVoice)),
				"availability_communication_voice_msg",
			),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				fmt.Sprintf("%s %s", getStyleSymbol("audio_call", selectedStylesMap), localizer.Get(lang, localization.LocaleCommAudio)),
				"availability_communication_audio_call",
			),
			tgbotapi.NewInlineKeyboardButtonData(
				fmt.Sprintf("%s %s", getStyleSymbol("video_call", selectedStylesMap), localizer.Get(lang, localization.LocaleCommVideo)),
				"availability_communication_video_call",
			),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				fmt.Sprintf("%s %s", getStyleSymbol("meet_person", selectedStylesMap), localizer.Get(lang, localization.LocaleCommMeet)),
				"availability_communication_meet_person",
			),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				fmt.Sprintf("✅ %s", localizer.Get(lang, "select_all")),
				"availability_communication_select_all",
			),
		),
		h.baseHandler.KeyboardBuilder.CreateNavigationRow(
			lang,
			"availability_back_to_time",
			"availability_proceed_to_frequency",
		),
	)

	// Редактируем текущее сообщение вместо отправки нового
	editMsg := tgbotapi.NewEditMessageTextAndMarkup(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		message,
		keyboard,
	)

	_, err = h.baseHandler.Bot.Send(editMsg)
	if err != nil {
		loggingService.ErrorWithContext("Failed to edit message", "", int64(user.ID), callback.Message.Chat.ID, "ShowCommunicationStyleSelection", map[string]interface{}{
			"user_id": user.ID,
			"error":   err.Error(),
		})
		return err
	}

	return nil
}

// Helper function to get checkbox symbol for communication styles
func getStyleSymbol(style string, selectedStyles map[string]bool) string {
	if selectedStyles[style] {
		return "✅"
	}
	return "☑"
}

// CompleteAvailabilitySetup завершает настройку доступности и сохраняет данные
func (h *AvailabilityHandlerImpl) CompleteAvailabilitySetup(callback *tgbotapi.CallbackQuery, user *models.User) error {
	loggingService := h.baseHandler.Service.LoggingService.Telegram()
	loggingService.InfoWithContext("Completing availability setup", "", int64(user.ID), callback.Message.Chat.ID, "CompleteAvailabilitySetup", map[string]interface{}{
		"user_id": user.ID,
	})

	// Дополнительное логирование для отладки
	loggingService.InfoWithContext("DEBUG: CompleteAvailabilitySetup called", "", int64(user.ID), callback.Message.Chat.ID, "CompleteAvailabilitySetup", map[string]interface{}{
		"debug":   true,
		"step":    "method_called",
		"chat_id": callback.Message.Chat.ID,
		"user_id": user.ID,
	})

	// НЕМЕДЛЕННО редактируем сообщение об успехе, чтобы пользователь увидел реакцию
	lang := user.InterfaceLanguageCode
	localizer := h.baseHandler.Service.Localizer

	successMessage := fmt.Sprintf("%s\n\n%s",
		localizer.Get(lang, "availability_setup_complete"),
		localizer.Get(lang, "profile_completed"),
	)
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		h.baseHandler.KeyboardBuilder.CreateProfileActionsRow(lang),
	)

	editResult := h.baseHandler.MessageFactory.EditWithKeyboard(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		successMessage,
		&keyboard,
	)
	if editResult != nil {
		loggingService.ErrorWithContext("Failed to edit message with success", "", int64(user.ID), callback.Message.Chat.ID, "CompleteAvailabilitySetup", map[string]interface{}{
			"error": editResult.Error(),
		})
		// Если редактирование не удалось, отправляем новое сообщение
		sendResult := h.baseHandler.MessageFactory.SendWithKeyboard(
			callback.Message.Chat.ID,
			successMessage,
			keyboard,
		)
		if sendResult != nil {
			loggingService.ErrorWithContext("Failed to send success message as fallback", "", int64(user.ID), callback.Message.Chat.ID, "CompleteAvailabilitySetup", map[string]interface{}{
				"error": sendResult.Error(),
			})
		}
	} else {
		loggingService.InfoWithContext("Success message edited successfully", "", int64(user.ID), callback.Message.Chat.ID, "CompleteAvailabilitySetup", nil)
	}

	// Проверяем, что пользователь выбрал хотя бы один способ общения
	cacheKey := fmt.Sprintf("availability_setup:%d", user.ID)
	var setupDataStr string
	err := h.baseHandler.Service.Cache.Get(context.Background(), cacheKey, &setupDataStr)
	if err != nil {
		lang := user.InterfaceLanguageCode
		localizer := h.baseHandler.Service.Localizer
		return h.baseHandler.MessageFactory.EditText(
			callback.Message.Chat.ID,
			callback.Message.MessageID,
			localizer.Get(lang, "error_no_communication_selected"),
		)
	}

	var setupData map[string]interface{}
	err = json.Unmarshal([]byte(setupDataStr), &setupData)
	if err != nil {
		return fmt.Errorf("failed to unmarshal setup data: %w", err)
	}

	communicationStyles := setupData["communication_styles"].([]interface{})
	loggingService.InfoWithContext("DEBUG: Loaded communication styles", "", int64(user.ID), callback.Message.Chat.ID, "CompleteAvailabilitySetup", map[string]interface{}{
		"debug":        true,
		"step":         "loaded_styles",
		"styles_count": len(communicationStyles),
		"styles":       communicationStyles,
	})

	if len(communicationStyles) == 0 {
		lang := user.InterfaceLanguageCode
		localizer := h.baseHandler.Service.Localizer
		return h.baseHandler.MessageFactory.EditText(
			callback.Message.Chat.ID,
			callback.Message.MessageID,
			localizer.Get(lang, "error_no_communication_selected"),
		)
	}

	// Создаем объекты для сохранения
	timeAvailability := &models.TimeAvailability{
		DayType: setupData["day_type"].(string),
	}

	// Обрабатываем specific_days
	if specificDays, ok := setupData["specific_days"].([]interface{}); ok {
		timeAvailability.SpecificDays = make([]string, len(specificDays))
		for i, d := range specificDays {
			timeAvailability.SpecificDays[i] = d.(string)
		}
	}

	// Обрабатываем time_slots
	if timeSlots, ok := setupData["time_slots"].([]interface{}); ok {
		timeAvailability.TimeSlots = make([]string, len(timeSlots))
		for i, t := range timeSlots {
			timeAvailability.TimeSlots[i] = t.(string)
		}
	}

	// Валидация данных перед сохранением
	if len(timeAvailability.TimeSlots) == 0 {
		fmt.Printf("DEBUG: ERROR - time_slots is empty for user %d\n", user.ID)
		loggingService.ErrorWithContext("Time slots is empty", "", int64(user.ID), callback.Message.Chat.ID, "CompleteAvailabilitySetup", nil)
		return fmt.Errorf("time slots cannot be empty")
	}

	if timeAvailability.DayType == "specific" && len(timeAvailability.SpecificDays) == 0 {
		fmt.Printf("DEBUG: ERROR - specific_days is empty for user %d with day_type=specific\n", user.ID)
		loggingService.ErrorWithContext("Specific days is empty for specific day type", "", int64(user.ID), callback.Message.Chat.ID, "CompleteAvailabilitySetup", nil)
		return fmt.Errorf("specific days cannot be empty when day_type is specific")
	}

	// Создаем предпочтения общения на основе выбранных данных
	selectedCommunicationStyles := setupData["communication_styles"].([]interface{})
	communicationStylesStr := make([]string, len(selectedCommunicationStyles))
	for i, style := range selectedCommunicationStyles {
		communicationStylesStr[i] = style.(string)
	}

	// Валидация communication styles
	if len(communicationStylesStr) == 0 {
		fmt.Printf("DEBUG: ERROR - communication_styles is empty for user %d\n", user.ID)
		loggingService.ErrorWithContext("Communication styles is empty", "", int64(user.ID), callback.Message.Chat.ID, "CompleteAvailabilitySetup", nil)
		return fmt.Errorf("communication styles cannot be empty")
	}

	// Устанавливаем частоту по умолчанию, так как UI для выбора частоты пока не реализован
	defaultFrequency := "weekly"
	if freq, exists := setupData["communication_frequency"]; exists && freq != nil {
		if freqStr, ok := freq.(string); ok && freqStr != "" {
			defaultFrequency = freqStr
		}
	}

	friendshipPreferences := &models.FriendshipPreferences{
		ActivityType:        "casual_chat",
		CommunicationStyles: communicationStylesStr,
		CommunicationFreq:   defaultFrequency,
	}

	// Сохраняем данные в базу
	fmt.Printf("DEBUG: Saving time availability for user %d: %+v\n", user.ID, timeAvailability)
	err = h.baseHandler.Service.SaveTimeAvailability(user.ID, timeAvailability)
	if err != nil {
		fmt.Printf("DEBUG: ERROR saving time availability for user %d: %v\n", user.ID, err)
		loggingService.ErrorWithContext("Failed to save time availability", "", int64(user.ID), callback.Message.Chat.ID, "CompleteAvailabilitySetup", map[string]interface{}{
			"user_id": user.ID,
			"error":   err.Error(),
			"data":    timeAvailability,
		})
		// Продолжаем, не возвращаем ошибку
		loggingService.InfoWithContext("Continuing despite time availability save error", "", int64(user.ID), callback.Message.Chat.ID, "CompleteAvailabilitySetup", nil)
	} else {
		fmt.Printf("DEBUG: Successfully saved time availability for user %d\n", user.ID)
	}

	fmt.Printf("DEBUG: Saving friendship preferences for user %d: %+v\n", user.ID, friendshipPreferences)
	err = h.baseHandler.Service.SaveFriendshipPreferences(user.ID, friendshipPreferences)
	if err != nil {
		fmt.Printf("DEBUG: ERROR saving friendship preferences for user %d: %v\n", user.ID, err)
		loggingService.ErrorWithContext("Failed to save friendship preferences", "", int64(user.ID), callback.Message.Chat.ID, "CompleteAvailabilitySetup", map[string]interface{}{
			"user_id": user.ID,
			"error":   err.Error(),
			"data":    friendshipPreferences,
		})
		// Продолжаем, не возвращаем ошибку
		loggingService.InfoWithContext("Continuing despite friendship preferences save error", "", int64(user.ID), callback.Message.Chat.ID, "CompleteAvailabilitySetup", nil)
	}

	// Обновляем статус пользователя
	err = h.baseHandler.Service.UpdateUserState(user.ID, models.StateActive)
	if err != nil {
		loggingService.ErrorWithContext("Failed to update user state", "", int64(user.ID), callback.Message.Chat.ID, "CompleteAvailabilitySetup", map[string]interface{}{
			"user_id": user.ID,
			"error":   err.Error(),
		})
		// Продолжаем, не возвращаем ошибку
		loggingService.InfoWithContext("Continuing despite user state update error", "", int64(user.ID), callback.Message.Chat.ID, "CompleteAvailabilitySetup", nil)
	} else {
		// Обновляем статус в объекте пользователя в памяти
		user.State = models.StateActive
		// Обновляем кэш пользователя напрямую
		h.baseHandler.Service.Cache.SetUser(context.Background(), user)
	}

	// Обновляем статус профиля на активный (важно для отображения профиля)
	err = h.baseHandler.Service.DB.UpdateUserStatus(user.ID, models.StatusActive)
	if err != nil {
		loggingService.ErrorWithContext("Failed to update user status", "", int64(user.ID), callback.Message.Chat.ID, "CompleteAvailabilitySetup", map[string]interface{}{
			"user_id": user.ID,
			"error":   err.Error(),
		})
		// Продолжаем, не возвращаем ошибку
		loggingService.InfoWithContext("Continuing despite user status update error", "", int64(user.ID), callback.Message.Chat.ID, "CompleteAvailabilitySetup", nil)
	} else {
		// Обновляем статус в объекте пользователя в памяти
		user.Status = models.StatusActive
		// Обновляем кэш пользователя напрямую
		h.baseHandler.Service.Cache.SetUser(context.Background(), user)
	}

	// Обновляем уровень завершения профиля (после настройки доступности профиль полностью завершен)
	err = h.updateProfileCompletionLevel(user.ID, 100)
	if err != nil {
		loggingService.ErrorWithContext("Failed to update profile completion level", "", int64(user.ID), callback.Message.Chat.ID, "CompleteAvailabilitySetup", map[string]interface{}{
			"user_id": user.ID,
			"error":   err.Error(),
		})
		// Продолжаем, не возвращаем ошибку
		loggingService.InfoWithContext("Continuing despite profile completion level update error", "", int64(user.ID), callback.Message.Chat.ID, "CompleteAvailabilitySetup", nil)
	} else {
		// Обновляем уровень в объекте пользователя в памяти
		user.ProfileCompletionLevel = 100
		// Обновляем кэш пользователя напрямую вместо инвалидации
		h.baseHandler.Service.Cache.SetUser(context.Background(), user)
	}

	// Очищаем временные данные
	h.baseHandler.Service.Cache.Delete(context.Background(), cacheKey)

	loggingService.InfoWithContext("Availability setup completed successfully", "", int64(user.ID), callback.Message.Chat.ID, "CompleteAvailabilitySetup", map[string]interface{}{
		"user_id":          user.ID,
		"setup_completed":  true,
		"day_type":         timeAvailability.DayType,
		"time_slots_count": len(timeAvailability.TimeSlots),
	})

	// Сообщение уже отправлено в начале метода
	return nil
}
