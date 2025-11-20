package availability

import (
	"strings"

	"language-exchange-bot/internal/localization"
	"language-exchange-bot/internal/models"
)

// =============================================================================
// UTILITY METHODS
// =============================================================================

// formatTimeAvailabilityForDisplay formats time availability for display in profile
func (h *AvailabilityHandlerImpl) formatTimeAvailabilityForDisplay(availability *models.TimeAvailability, lang string) string {
	if availability == nil {
		return h.baseHandler.Service.Localizer.Get(lang, localization.LocaleErrorInvalidAvailabilityData)
	}

	var parts []string

	// Format day type
	switch availability.DayType {
	case "weekdays":
		parts = append(parts, h.baseHandler.Service.Localizer.Get(lang, localization.LocaleTimeWeekdays))
	case "weekends":
		parts = append(parts, h.baseHandler.Service.Localizer.Get(lang, localization.LocaleTimeWeekends))
	case "any":
		parts = append(parts, h.baseHandler.Service.Localizer.Get(lang, localization.LocaleTimeAny))
	case "specific":
		if len(availability.SpecificDays) > 0 {
			parts = append(parts, strings.Join(availability.SpecificDays, ", "))
		} else {
			parts = append(parts, h.baseHandler.Service.Localizer.Get(lang, localization.LocaleTimeAny))
		}
	}

	// Format time slots
	if len(availability.TimeSlots) > 0 {
		var timeParts []string
		for _, slot := range availability.TimeSlots {
			switch slot {
			case "morning":
				timeParts = append(timeParts, h.baseHandler.Service.Localizer.Get(lang, localization.LocaleTimeMorning))
			case "day":
				timeParts = append(timeParts, h.baseHandler.Service.Localizer.Get(lang, localization.LocaleTimeDay))
			case "evening":
				timeParts = append(timeParts, h.baseHandler.Service.Localizer.Get(lang, localization.LocaleTimeEvening))
			case "late":
				timeParts = append(timeParts, h.baseHandler.Service.Localizer.Get(lang, localization.LocaleTimeLate))
			}
		}
		if len(timeParts) > 0 {
			parts = append(parts, strings.Join(timeParts, ", "))
		}
	}

	if len(parts) == 0 {
		return h.baseHandler.Service.Localizer.Get(lang, localization.LocaleErrorInvalidAvailabilityData)
	}

	return strings.Join(parts, ", ")
}

// formatFriendshipPreferencesForDisplay formats friendship preferences for display in profile
func (h *AvailabilityHandlerImpl) formatFriendshipPreferencesForDisplay(preferences *models.FriendshipPreferences, lang string) string {
	if preferences == nil {
		return h.baseHandler.Service.Localizer.Get(lang, localization.LocaleErrorInvalidAvailabilityData)
	}

	var parts []string

	// Format communication styles
	if len(preferences.CommunicationStyles) > 0 {
		var styleParts []string
		for _, style := range preferences.CommunicationStyles {
			switch style {
			case "text":
				styleParts = append(styleParts, h.baseHandler.Service.Localizer.Get(lang, localization.LocaleCommText))
			case "voice_msg":
				styleParts = append(styleParts, h.baseHandler.Service.Localizer.Get(lang, localization.LocaleCommVoice))
			case "audio_call":
				styleParts = append(styleParts, h.baseHandler.Service.Localizer.Get(lang, localization.LocaleCommAudio))
			case "video_call":
				styleParts = append(styleParts, h.baseHandler.Service.Localizer.Get(lang, localization.LocaleCommVideo))
			case "meet_person":
				styleParts = append(styleParts, h.baseHandler.Service.Localizer.Get(lang, localization.LocaleCommMeet))
			}
		}
		if len(styleParts) > 0 {
			parts = append(parts, strings.Join(styleParts, ", "))
		}
	}

	// Format frequency
	switch preferences.CommunicationFreq {
	case "multiple_weekly":
		parts = append(parts, h.baseHandler.Service.Localizer.Get(lang, localization.LocaleFreqMultipleWeekly))
	case "weekly":
		parts = append(parts, h.baseHandler.Service.Localizer.Get(lang, localization.LocaleFreqWeekly))
	case "multiple_monthly":
		parts = append(parts, h.baseHandler.Service.Localizer.Get(lang, localization.LocaleFreqMultipleMonthly))
	case "flexible":
		parts = append(parts, h.baseHandler.Service.Localizer.Get(lang, localization.LocaleFreqFlexible))
	}

	if len(parts) == 0 {
		return h.baseHandler.Service.Localizer.Get(lang, localization.LocaleErrorInvalidAvailabilityData)
	}

	return strings.Join(parts, ", ")
}

// updateProfileCompletionLevel обновляет уровень завершения профиля до указанного значения (0-100).
func (h *AvailabilityHandlerImpl) updateProfileCompletionLevel(userID int, completionLevel int) error {
	_, err := h.baseHandler.Service.DB.GetConnection().Exec(`
		UPDATE users
		SET profile_completion_level = $1, updated_at = NOW()
		WHERE id = $2
	`, completionLevel, userID)

	return err
}
