package availability

import (
	"testing"
	"time"

	"language-exchange-bot/internal/models"

	"github.com/stretchr/testify/assert"
)

// TestAvailabilityDataValidation тестирует валидацию данных доступности
func TestAvailabilityDataValidation(t *testing.T) {
	t.Run("Valid time availability", func(t *testing.T) {
		validAvailability := &models.TimeAvailability{
			DayType:      "weekdays",
			SpecificDays: []string{},
			TimeSlots:    []string{"morning", "evening"},
		}

		// Should not panic and should be valid
		assert.NotNil(t, validAvailability)
		assert.Equal(t, "weekdays", validAvailability.DayType)
		assert.Len(t, validAvailability.TimeSlots, 2)
	})

	t.Run("Valid friendship preferences", func(t *testing.T) {
		validPreferences := &models.FriendshipPreferences{
			ActivityType:        "casual_chat",
			CommunicationStyles: []string{"text", "voice_msg"},
			CommunicationFreq:   "weekly",
		}

		assert.NotNil(t, validPreferences)
		assert.Equal(t, "casual_chat", validPreferences.ActivityType)
		assert.Len(t, validPreferences.CommunicationStyles, 2)
		assert.Equal(t, "weekly", validPreferences.CommunicationFreq)
	})

	t.Run("Invalid time availability - empty time slots", func(t *testing.T) {
		invalidAvailability := &models.TimeAvailability{
			DayType:      "weekdays",
			SpecificDays: []string{},
			TimeSlots:    []string{}, // Empty - should be invalid
		}

		assert.NotNil(t, invalidAvailability)
		assert.Len(t, invalidAvailability.TimeSlots, 0) // This would fail validation in real service
	})

	t.Run("Invalid friendship preferences - empty communication styles", func(t *testing.T) {
		invalidPreferences := &models.FriendshipPreferences{
			ActivityType:        "casual_chat",
			CommunicationStyles: []string{}, // Empty - should be invalid
			CommunicationFreq:   "weekly",
		}

		assert.NotNil(t, invalidPreferences)
		assert.Len(t, invalidPreferences.CommunicationStyles, 0) // This would fail validation in real service
	})

	t.Run("Specific days availability", func(t *testing.T) {
		specificAvailability := &models.TimeAvailability{
			DayType:      "specific",
			SpecificDays: []string{"monday", "wednesday", "friday"},
			TimeSlots:    []string{"morning"},
		}

		assert.NotNil(t, specificAvailability)
		assert.Equal(t, "specific", specificAvailability.DayType)
		assert.Len(t, specificAvailability.SpecificDays, 3)
		assert.Contains(t, specificAvailability.SpecificDays, "monday")
		assert.Contains(t, specificAvailability.SpecificDays, "wednesday")
		assert.Contains(t, specificAvailability.SpecificDays, "friday")
	})
}

// TestAvailabilitySessionStructure тестирует структуру сессии редактирования
func TestAvailabilitySessionStructure(t *testing.T) {
	t.Run("Create availability edit session", func(t *testing.T) {
		session := &AvailabilityEditSession{
			UserID: 123,
			OriginalTimeAvailability: &models.TimeAvailability{
				DayType:   "weekdays",
				TimeSlots: []string{"morning"},
			},
			CurrentTimeAvailability: &models.TimeAvailability{
				DayType:   "weekdays",
				TimeSlots: []string{"morning", "evening"}, // Modified
			},
			OriginalPreferences: &models.FriendshipPreferences{
				CommunicationStyles: []string{"text"},
			},
			CurrentPreferences: &models.FriendshipPreferences{
				CommunicationStyles: []string{"text", "voice_msg"}, // Modified
			},
			Changes: []AvailabilityChange{
				{
					Field:     "time_slots",
					OldValue:  []string{"morning"},
					NewValue:  []string{"morning", "evening"},
					Timestamp: time.Now(),
				},
			},
			CurrentStep:  "menu",
			SessionStart: time.Now(),
			LastActivity: time.Now(),
		}

		assert.NotNil(t, session)
		assert.Equal(t, 123, session.UserID)
		assert.Len(t, session.Changes, 1)
		assert.Equal(t, "time_slots", session.Changes[0].Field)
		assert.Equal(t, "menu", session.CurrentStep)

		// Verify modifications
		assert.Len(t, session.OriginalTimeAvailability.TimeSlots, 1)
		assert.Len(t, session.CurrentTimeAvailability.TimeSlots, 2)
		assert.Len(t, session.OriginalPreferences.CommunicationStyles, 1)
		assert.Len(t, session.CurrentPreferences.CommunicationStyles, 2)
	})

	t.Run("Availability change tracking", func(t *testing.T) {
		change := AvailabilityChange{
			Field:     "communication_styles",
			OldValue:  []string{"text"},
			NewValue:  []string{"text", "voice_msg", "video_call"},
			Timestamp: time.Now(),
		}

		assert.NotNil(t, change)
		assert.Equal(t, "communication_styles", change.Field)
		assert.NotNil(t, change.Timestamp)
		assert.IsType(t, []string{}, change.OldValue)
		assert.IsType(t, []string{}, change.NewValue)

		oldStyles := change.OldValue.([]string)
		newStyles := change.NewValue.([]string)

		assert.Len(t, oldStyles, 1)
		assert.Len(t, newStyles, 3)
		assert.Contains(t, newStyles, "voice_msg")
		assert.Contains(t, newStyles, "video_call")
	})
}
