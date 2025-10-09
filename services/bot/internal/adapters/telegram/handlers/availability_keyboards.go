package handlers

import (
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// createDayTypeSelectionKeyboard создает клавиатуру выбора типа дней.
func (ah *AvailabilityHandlerImpl) createDayTypeSelectionKeyboard(lang string) tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				"📅 "+ah.base.service.Localizer.Get(lang, "time_weekdays"),
				"availability_daytype_weekdays",
			),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				"🏖️ "+ah.base.service.Localizer.Get(lang, "time_weekends"),
				"availability_daytype_weekends",
			),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				"🌟 "+ah.base.service.Localizer.Get(lang, "time_any"),
				"availability_daytype_any",
			),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				"📝 "+ah.base.service.Localizer.Get(lang, "select_specific_days_button"),
				"availability_daytype_specific",
			),
		),
	)
}

// createSpecificDaysSelectionKeyboard создает клавиатуру выбора конкретных дней.
func (ah *AvailabilityHandlerImpl) createSpecificDaysSelectionKeyboard(lang string, selectedDays []string) tgbotapi.InlineKeyboardMarkup {
	days := []string{"monday", "tuesday", "wednesday", "thursday", "friday", "saturday", "sunday"}

	var rows [][]tgbotapi.InlineKeyboardButton

	// Создаем кнопки для каждого дня (2 в ряд)
	for i := 0; i < len(days); i += 2 {
		var row []tgbotapi.InlineKeyboardButton

		// Первая кнопка в ряду
		day1 := days[i]
		day1Name := ah.base.service.Localizer.Get(lang, "day_"+day1)

		prefix1 := "☐"
		if ah.containsDay(selectedDays, day1) {
			prefix1 = "☑"
		}

		row = append(row, tgbotapi.NewInlineKeyboardButtonData(
			fmt.Sprintf("%s %s", prefix1, day1Name),
			"availability_specific_day_"+day1,
		))

		// Вторая кнопка в ряду (если есть)
		if i+1 < len(days) {
			day2 := days[i+1]
			day2Name := ah.base.service.Localizer.Get(lang, "day_"+day2)

			prefix2 := "☐"
			if ah.containsDay(selectedDays, day2) {
				prefix2 = "☑"
			}

			row = append(row, tgbotapi.NewInlineKeyboardButtonData(
				fmt.Sprintf("%s %s", prefix2, day2Name),
				"availability_specific_day_"+day2,
			))
		}

		rows = append(rows, row)
	}

	// Кнопка "Продолжить"
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData(
			"✅ "+ah.base.service.Localizer.Get(lang, "continue_button"),
			"availability_proceed_to_time",
		),
	))

	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

// createTimeSlotSelectionKeyboard создает клавиатуру выбора временного слота.
func (ah *AvailabilityHandlerImpl) createTimeSlotSelectionKeyboard(lang string) tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				"🌅 "+ah.base.service.Localizer.Get(lang, "time_morning"),
				"availability_timeslot_morning",
			),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				"☀️ "+ah.base.service.Localizer.Get(lang, "time_day"),
				"availability_timeslot_day",
			),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				"🌆 "+ah.base.service.Localizer.Get(lang, "time_evening"),
				"availability_timeslot_evening",
			),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				"🌙 "+ah.base.service.Localizer.Get(lang, "time_late"),
				"availability_timeslot_late",
			),
		),
	)
}

// createActivityTypeSelectionKeyboard создает клавиатуру выбора типа активности.
func (ah *AvailabilityHandlerImpl) createActivityTypeSelectionKeyboard(lang string) tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				"🎬 "+ah.base.service.Localizer.Get(lang, "activity_movies"),
				"availability_activity_movies",
			),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				"🎮 "+ah.base.service.Localizer.Get(lang, "activity_games"),
				"availability_activity_games",
			),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				"☕ "+ah.base.service.Localizer.Get(lang, "activity_casual_chat"),
				"availability_activity_casual_chat",
			),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				"🎨 "+ah.base.service.Localizer.Get(lang, "activity_creative"),
				"availability_activity_creative",
			),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				"⚽ "+ah.base.service.Localizer.Get(lang, "activity_active"),
				"availability_activity_active",
			),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				"📚 "+ah.base.service.Localizer.Get(lang, "activity_educational"),
				"availability_activity_educational",
			),
		),
	)
}

// createCommunicationStyleSelectionKeyboard создает клавиатуру выбора стиля общения.
func (ah *AvailabilityHandlerImpl) createCommunicationStyleSelectionKeyboard(lang string) tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				"💬 "+ah.base.service.Localizer.Get(lang, "communication_text"),
				"availability_communication_text",
			),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				"🎤 "+ah.base.service.Localizer.Get(lang, "communication_voice_msg"),
				"availability_communication_voice_msg",
			),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				"📞 "+ah.base.service.Localizer.Get(lang, "communication_audio_call"),
				"availability_communication_audio_call",
			),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				"📹 "+ah.base.service.Localizer.Get(lang, "communication_video_call"),
				"availability_communication_video_call",
			),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				"🤝 "+ah.base.service.Localizer.Get(lang, "communication_meet_person"),
				"availability_communication_meet_person",
			),
		),
	)
}

// createCommunicationFrequencySelectionKeyboard создает клавиатуру выбора частоты общения.
func (ah *AvailabilityHandlerImpl) createCommunicationFrequencySelectionKeyboard(lang string) tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				"⚡ "+ah.base.service.Localizer.Get(lang, "frequency_spontaneous"),
				"availability_frequency_spontaneous",
			),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				"📅 "+ah.base.service.Localizer.Get(lang, "frequency_weekly"),
				"availability_frequency_weekly",
			),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				"📆 "+ah.base.service.Localizer.Get(lang, "frequency_daily"),
				"availability_frequency_daily",
			),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				"🔥 "+ah.base.service.Localizer.Get(lang, "frequency_intensive"),
				"availability_frequency_intensive",
			),
		),
	)
}
