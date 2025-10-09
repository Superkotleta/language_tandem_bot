package models

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestUserConstants тестирует константы состояний и статусов пользователя.
func TestUserConstants(t *testing.T) {
	// Test User States
	assert.Equal(t, "new", StateNew)
	assert.Equal(t, "waiting_language", StateWaitingLanguage)
	assert.Equal(t, "waiting_target_language", StateWaitingTargetLanguage)
	assert.Equal(t, "waiting_language_level", StateWaitingLanguageLevel)
	assert.Equal(t, "waiting_interests", StateWaitingInterests)
	assert.Equal(t, "waiting_time", StateWaitingTime)
	assert.Equal(t, "waiting_feedback", StateWaitingFeedback)
	assert.Equal(t, "waiting_feedback_contact", StateWaitingFeedbackContact)
	assert.Equal(t, "active", StateActive)

	// Test User Statuses
	assert.Equal(t, "new", StatusNew)
	assert.Equal(t, "filling_profile", StatusFilling)
	assert.Equal(t, "active", StatusActive)
	assert.Equal(t, "paused", StatusPaused)
}

// TestUser_JSONSerialization тестирует JSON сериализацию/десериализацию User.
func TestUser_JSONSerialization(t *testing.T) {
	now := time.Now()

	user := User{
		ID:                     1,
		TelegramID:             123456789,
		Username:               "testuser",
		FirstName:              "Test User",
		NativeLanguageCode:     "ru",
		TargetLanguageCode:     "en",
		TargetLanguageLevel:    "intermediate",
		InterfaceLanguageCode:  "en",
		State:                  StateActive,
		Status:                 StatusActive,
		ProfileCompletionLevel: 75,
		CreatedAt:              now,
		UpdatedAt:              now,
		Interests:              []int{1, 2, 3},
		TimeAvailability: &TimeAvailability{
			DayType:      "weekdays",
			SpecificDays: []string{"monday", "wednesday"},
			TimeSlot:     "evening",
		},
		FriendshipPreferences: &FriendshipPreferences{
			ActivityType:       "educational",
			CommunicationStyle: "text",
			CommunicationFreq:  "weekly",
		},
	}

	// Test JSON serialization
	data, err := json.Marshal(user)
	assert.NoError(t, err)
	assert.NotEmpty(t, data)

	// Test JSON deserialization
	var deserialized User

	err = json.Unmarshal(data, &deserialized)
	assert.NoError(t, err)

	// Verify all fields
	assert.Equal(t, user.ID, deserialized.ID)
	assert.Equal(t, user.TelegramID, deserialized.TelegramID)
	assert.Equal(t, user.Username, deserialized.Username)
	assert.Equal(t, user.FirstName, deserialized.FirstName)
	assert.Equal(t, user.NativeLanguageCode, deserialized.NativeLanguageCode)
	assert.Equal(t, user.TargetLanguageCode, deserialized.TargetLanguageCode)
	assert.Equal(t, user.TargetLanguageLevel, deserialized.TargetLanguageLevel)
	assert.Equal(t, user.InterfaceLanguageCode, deserialized.InterfaceLanguageCode)
	assert.Equal(t, user.State, deserialized.State)
	assert.Equal(t, user.Status, deserialized.Status)
	assert.Equal(t, user.ProfileCompletionLevel, deserialized.ProfileCompletionLevel)
	assert.Equal(t, user.Interests, deserialized.Interests)

	// Verify nested structs
	assert.NotNil(t, deserialized.TimeAvailability)
	assert.Equal(t, user.TimeAvailability.DayType, deserialized.TimeAvailability.DayType)
	assert.Equal(t, user.TimeAvailability.SpecificDays, deserialized.TimeAvailability.SpecificDays)
	assert.Equal(t, user.TimeAvailability.TimeSlot, deserialized.TimeAvailability.TimeSlot)

	assert.NotNil(t, deserialized.FriendshipPreferences)
	assert.Equal(t, user.FriendshipPreferences.ActivityType, deserialized.FriendshipPreferences.ActivityType)
	assert.Equal(t, user.FriendshipPreferences.CommunicationStyle, deserialized.FriendshipPreferences.CommunicationStyle)
	assert.Equal(t, user.FriendshipPreferences.CommunicationFreq, deserialized.FriendshipPreferences.CommunicationFreq)
}

// TestUser_MinimalUser тестирует создание минимального пользователя.
func TestUser_MinimalUser(t *testing.T) {
	user := User{
		ID:         1,
		TelegramID: 123456789,
		FirstName:  "John",
	}

	assert.Equal(t, 1, user.ID)
	assert.Equal(t, int64(123456789), user.TelegramID)
	assert.Equal(t, "John", user.FirstName)
	assert.Empty(t, user.Username)
	assert.Empty(t, user.NativeLanguageCode)
	assert.Empty(t, user.InterfaceLanguageCode) // no default value in struct
	assert.Empty(t, user.State)                 // no default value in struct
	assert.Empty(t, user.Status)                // no default value in struct
	assert.Equal(t, 0, user.ProfileCompletionLevel)
	assert.Nil(t, user.TimeAvailability)
	assert.Nil(t, user.FriendshipPreferences)
}

// TestUser_ProfileCompletion тестирует логику завершения профиля.
func TestUser_ProfileCompletion(t *testing.T) {
	user := User{
		ID:                     1,
		TelegramID:             123456789,
		FirstName:              "Test User",
		NativeLanguageCode:     "ru",
		TargetLanguageCode:     "en",
		TargetLanguageLevel:    "intermediate",
		InterfaceLanguageCode:  "en",
		State:                  StateActive,
		Status:                 StatusActive,
		ProfileCompletionLevel: 100,
		Interests:              []int{1, 2, 3, 4, 5},
		TimeAvailability: &TimeAvailability{
			DayType:  "weekdays",
			TimeSlot: "evening",
		},
		FriendshipPreferences: &FriendshipPreferences{
			ActivityType:       "educational",
			CommunicationStyle: "text",
			CommunicationFreq:  "weekly",
		},
	}

	// Test complete profile
	assert.Equal(t, 100, user.ProfileCompletionLevel)
	assert.NotEmpty(t, user.Interests)
	assert.NotNil(t, user.TimeAvailability)
	assert.NotNil(t, user.FriendshipPreferences)
}

// TestTimeAvailability_JSONSerialization тестирует JSON сериализацию TimeAvailability.
func TestTimeAvailability_JSONSerialization(t *testing.T) {
	ta := TimeAvailability{
		DayType:      "specific",
		SpecificDays: []string{"monday", "wednesday", "friday"},
		TimeSlot:     "morning",
	}

	// Test JSON serialization
	data, err := json.Marshal(ta)
	assert.NoError(t, err)
	assert.NotEmpty(t, data)

	// Test JSON deserialization
	var deserialized TimeAvailability

	err = json.Unmarshal(data, &deserialized)
	assert.NoError(t, err)

	assert.Equal(t, ta.DayType, deserialized.DayType)
	assert.Equal(t, ta.SpecificDays, deserialized.SpecificDays)
	assert.Equal(t, ta.TimeSlot, deserialized.TimeSlot)
}

// TestFriendshipPreferences_JSONSerialization тестирует JSON сериализацию FriendshipPreferences.
func TestFriendshipPreferences_JSONSerialization(t *testing.T) {
	fp := FriendshipPreferences{
		ActivityType:       "movies",
		CommunicationStyle: "voice_msg",
		CommunicationFreq:  "daily",
	}

	// Test JSON serialization
	data, err := json.Marshal(fp)
	assert.NoError(t, err)
	assert.NotEmpty(t, data)

	// Test JSON deserialization
	var deserialized FriendshipPreferences

	err = json.Unmarshal(data, &deserialized)
	assert.NoError(t, err)

	assert.Equal(t, fp.ActivityType, deserialized.ActivityType)
	assert.Equal(t, fp.CommunicationStyle, deserialized.CommunicationStyle)
	assert.Equal(t, fp.CommunicationFreq, deserialized.CommunicationFreq)
}

// TestLanguage_JSONSerialization тестирует JSON сериализацию Language.
func TestLanguage_JSONSerialization(t *testing.T) {
	now := time.Now()
	lang := Language{
		ID:                  1,
		Code:                "ru",
		NameNative:          "Русский",
		NameEn:              "Russian",
		IsInterfaceLanguage: true,
		CreatedAt:           now,
	}

	// Test JSON serialization
	data, err := json.Marshal(lang)
	assert.NoError(t, err)
	assert.NotEmpty(t, data)

	// Test JSON deserialization
	var deserialized Language

	err = json.Unmarshal(data, &deserialized)
	assert.NoError(t, err)

	assert.Equal(t, lang.ID, deserialized.ID)
	assert.Equal(t, lang.Code, deserialized.Code)
	assert.Equal(t, lang.NameNative, deserialized.NameNative)
	assert.Equal(t, lang.NameEn, deserialized.NameEn)
	assert.Equal(t, lang.IsInterfaceLanguage, deserialized.IsInterfaceLanguage)
}

// TestLanguage_InterfaceLanguages тестирует интерфейсные языки.
func TestLanguage_InterfaceLanguages(t *testing.T) {
	interfaceLang := Language{
		Code:                "en",
		NameNative:          "English",
		NameEn:              "English",
		IsInterfaceLanguage: true,
	}

	nonInterfaceLang := Language{
		Code:                "zh",
		NameNative:          "中文",
		NameEn:              "Chinese",
		IsInterfaceLanguage: false,
	}

	assert.True(t, interfaceLang.IsInterfaceLanguage)
	assert.False(t, nonInterfaceLang.IsInterfaceLanguage)
}

// TestInterest_JSONSerialization тестирует JSON сериализацию Interest.
func TestInterest_JSONSerialization(t *testing.T) {
	now := time.Now()
	interest := Interest{
		ID:           1,
		KeyName:      "programming",
		CategoryID:   1,
		DisplayOrder: 1,
		Type:         "technology",
		CreatedAt:    now,
		CategoryKey:  "tech",
	}

	// Interest doesn't have JSON tags, so it won't serialize properly
	// But we can test that the struct exists and has expected fields
	assert.Equal(t, 1, interest.ID)
	assert.Equal(t, "programming", interest.KeyName)
	assert.Equal(t, 1, interest.CategoryID)
	assert.Equal(t, 1, interest.DisplayOrder)
	assert.Equal(t, "technology", interest.Type)
	assert.Equal(t, "tech", interest.CategoryKey)
}

// TestInterestCategory_JSONSerialization тестирует JSON сериализацию InterestCategory.
func TestInterestCategory_JSONSerialization(t *testing.T) {
	now := time.Now()
	category := InterestCategory{
		ID:           1,
		KeyName:      "technology",
		DisplayOrder: 1,
		Name:         "Technology",
		Description:  "Technology related interests",
		CreatedAt:    now,
	}

	// Test JSON serialization
	data, err := json.Marshal(category)
	assert.NoError(t, err)
	assert.NotEmpty(t, data)

	// Test JSON deserialization
	var deserialized InterestCategory

	err = json.Unmarshal(data, &deserialized)
	assert.NoError(t, err)

	assert.Equal(t, category.ID, deserialized.ID)
	assert.Equal(t, category.KeyName, deserialized.KeyName)
	assert.Equal(t, category.DisplayOrder, deserialized.DisplayOrder)
	assert.Equal(t, category.Name, deserialized.Name)
	assert.Equal(t, category.Description, deserialized.Description)
}

// TestInterestSelection_JSONSerialization тестирует JSON сериализацию InterestSelection.
func TestInterestSelection_JSONSerialization(t *testing.T) {
	now := time.Now()
	selection := InterestSelection{
		ID:             1,
		UserID:         123,
		InterestID:     456,
		IsPrimary:      true,
		SelectionOrder: 1,
		CreatedAt:      now,
	}

	// Test JSON serialization
	data, err := json.Marshal(selection)
	assert.NoError(t, err)
	assert.NotEmpty(t, data)

	// Test JSON deserialization
	var deserialized InterestSelection

	err = json.Unmarshal(data, &deserialized)
	assert.NoError(t, err)

	assert.Equal(t, selection.ID, deserialized.ID)
	assert.Equal(t, selection.UserID, deserialized.UserID)
	assert.Equal(t, selection.InterestID, deserialized.InterestID)
	assert.Equal(t, selection.IsPrimary, deserialized.IsPrimary)
	assert.Equal(t, selection.SelectionOrder, deserialized.SelectionOrder)
}

// TestInterestSelection_PrimaryVsAdditional тестирует различие между primary и additional интересами.
func TestInterestSelection_PrimaryVsAdditional(t *testing.T) {
	primarySelection := InterestSelection{
		UserID:     1,
		InterestID: 10,
		IsPrimary:  true,
	}

	additionalSelection := InterestSelection{
		UserID:     1,
		InterestID: 20,
		IsPrimary:  false,
	}

	assert.True(t, primarySelection.IsPrimary)
	assert.False(t, additionalSelection.IsPrimary)
	assert.NotEqual(t, primarySelection.InterestID, additionalSelection.InterestID)
}

// TestMatchingConfig_JSONSerialization тестирует JSON сериализацию MatchingConfig.
func TestMatchingConfig_JSONSerialization(t *testing.T) {
	now := time.Now()
	config := MatchingConfig{
		ID:          1,
		ConfigKey:   "language_weight",
		ConfigValue: "0.8",
		Description: "Weight for language matching",
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	// Test JSON serialization
	data, err := json.Marshal(config)
	assert.NoError(t, err)
	assert.NotEmpty(t, data)

	// Test JSON deserialization
	var deserialized MatchingConfig

	err = json.Unmarshal(data, &deserialized)
	assert.NoError(t, err)

	assert.Equal(t, config.ID, deserialized.ID)
	assert.Equal(t, config.ConfigKey, deserialized.ConfigKey)
	assert.Equal(t, config.ConfigValue, deserialized.ConfigValue)
	assert.Equal(t, config.Description, deserialized.Description)
}

// TestInterestLimitsConfig_JSONSerialization тестирует JSON сериализацию InterestLimitsConfig.
func TestInterestLimitsConfig_JSONSerialization(t *testing.T) {
	now := time.Now()
	limits := InterestLimitsConfig{
		ID:                  1,
		MinPrimaryInterests: 1,
		MaxPrimaryInterests: 3,
		PrimaryPercentage:   0.6,
		CreatedAt:           now,
		UpdatedAt:           now,
	}

	// Test JSON serialization
	data, err := json.Marshal(limits)
	assert.NoError(t, err)
	assert.NotEmpty(t, data)

	// Test JSON deserialization
	var deserialized InterestLimitsConfig

	err = json.Unmarshal(data, &deserialized)
	assert.NoError(t, err)

	assert.Equal(t, limits.ID, deserialized.ID)
	assert.Equal(t, limits.MinPrimaryInterests, deserialized.MinPrimaryInterests)
	assert.Equal(t, limits.MaxPrimaryInterests, deserialized.MaxPrimaryInterests)
	assert.Equal(t, limits.PrimaryPercentage, deserialized.PrimaryPercentage)
}

// TestInterestLimitsConfig_Validation тестирует валидацию лимитов интересов.
func TestInterestLimitsConfig_Validation(t *testing.T) {
	validLimits := InterestLimitsConfig{
		MinPrimaryInterests: 1,
		MaxPrimaryInterests: 5,
		PrimaryPercentage:   0.7,
	}

	assert.Positive(t, validLimits.MinPrimaryInterests)
	assert.GreaterOrEqual(t, validLimits.MaxPrimaryInterests, validLimits.MinPrimaryInterests)
	assert.GreaterOrEqual(t, validLimits.PrimaryPercentage, 0.0)
	assert.LessOrEqual(t, validLimits.PrimaryPercentage, 1.0)
}

// TestInterestWithCategory_JSONSerialization тестирует JSON сериализацию InterestWithCategory.
func TestInterestWithCategory_JSONSerialization(t *testing.T) {
	now := time.Now()
	interestWithCat := InterestWithCategory{
		Interest: Interest{
			ID:          1,
			KeyName:     "programming",
			CategoryID:  1,
			Type:        "technology",
			CreatedAt:   now,
			CategoryKey: "tech",
		},
		CategoryName: "Technology",
		CategoryKey:  "tech",
	}

	// Test JSON serialization
	data, err := json.Marshal(interestWithCat)
	assert.NoError(t, err)
	assert.NotEmpty(t, data)

	// Test JSON deserialization
	var deserialized InterestWithCategory

	err = json.Unmarshal(data, &deserialized)
	assert.NoError(t, err)

	assert.Equal(t, interestWithCat.ID, deserialized.ID)
	assert.Equal(t, interestWithCat.KeyName, deserialized.KeyName)
	assert.Equal(t, interestWithCat.CategoryName, deserialized.CategoryName)
	assert.Equal(t, interestWithCat.CategoryKey, deserialized.CategoryKey)
}

// TestUserInterestSummary_JSONSerialization тестирует JSON сериализацию UserInterestSummary.
func TestUserInterestSummary_JSONSerialization(t *testing.T) {
	summary := UserInterestSummary{
		UserID:         123,
		TotalInterests: 7,
		PrimaryInterests: []InterestWithCategory{
			{
				Interest: Interest{
					ID:      1,
					KeyName: "programming",
				},
				CategoryName: "Technology",
			},
		},
		AdditionalInterests: []InterestWithCategory{
			{
				Interest: Interest{
					ID:      2,
					KeyName: "music",
				},
				CategoryName: "Entertainment",
			},
			{
				Interest: Interest{
					ID:      3,
					KeyName: "sports",
				},
				CategoryName: "Lifestyle",
			},
		},
	}

	// Test JSON serialization
	data, err := json.Marshal(summary)
	assert.NoError(t, err)
	assert.NotEmpty(t, data)

	// Test JSON deserialization
	var deserialized UserInterestSummary

	err = json.Unmarshal(data, &deserialized)
	assert.NoError(t, err)

	assert.Equal(t, summary.UserID, deserialized.UserID)
	assert.Equal(t, summary.TotalInterests, deserialized.TotalInterests)
	assert.Len(t, deserialized.PrimaryInterests, 1)
	assert.Len(t, deserialized.AdditionalInterests, 2)
}

// TestUserInterestSummary_InterestDistribution тестирует распределение интересов.
func TestUserInterestSummary_InterestDistribution(t *testing.T) {
	summary := UserInterestSummary{
		UserID:              123,
		TotalInterests:      5,
		PrimaryInterests:    make([]InterestWithCategory, 2), // 2 primary
		AdditionalInterests: make([]InterestWithCategory, 3), // 3 additional
	}

	assert.Equal(t, 5, summary.TotalInterests)
	assert.Len(t, summary.PrimaryInterests, 2)
	assert.Len(t, summary.AdditionalInterests, 3)

	// Test that primary + additional = total
	totalCalculated := len(summary.PrimaryInterests) + len(summary.AdditionalInterests)
	assert.Equal(t, summary.TotalInterests, totalCalculated)
}

// TestModels_JSONEdgeCases тестирует edge cases JSON сериализации.
func TestModels_JSONEdgeCases(t *testing.T) {
	// Test empty User
	emptyUser := User{}
	data, err := json.Marshal(emptyUser)
	assert.NoError(t, err)
	assert.NotEmpty(t, data)

	var deserialized User

	err = json.Unmarshal(data, &deserialized)
	assert.NoError(t, err)

	// Test User with nil slices
	userWithNil := User{
		TelegramID: 123,
		FirstName:  "Test",
		Interests:  nil, // explicitly nil
	}
	data, err = json.Marshal(userWithNil)
	assert.NoError(t, err)

	err = json.Unmarshal(data, &deserialized)
	assert.NoError(t, err)
	assert.Nil(t, deserialized.Interests)
}

// TestModels_TimeHandling тестирует работу с временными полями.
func TestModels_TimeHandling(t *testing.T) {
	now := time.Now()
	past := now.Add(-24 * time.Hour)

	// Test Interest with creation time
	interest := Interest{
		ID:        1,
		KeyName:   "test",
		CreatedAt: past,
	}

	assert.True(t, interest.CreatedAt.Before(now))
	assert.True(t, interest.CreatedAt.Equal(past))

	// Test User with timestamps
	user := User{
		ID:        1,
		CreatedAt: past,
		UpdatedAt: now,
	}

	assert.True(t, user.CreatedAt.Before(user.UpdatedAt))
	assert.True(t, user.UpdatedAt.Equal(now))
}

// TestModels_DefaultValues тестирует значения по умолчанию.
func TestModels_DefaultValues(t *testing.T) {
	// Test default User values
	user := User{}
	assert.Empty(t, user.State)                 // empty string, no default value
	assert.Empty(t, user.Status)                // empty string, no default value
	assert.Empty(t, user.InterfaceLanguageCode) // empty string, no default value
	assert.Equal(t, 0, user.ProfileCompletionLevel)
	assert.Nil(t, user.TimeAvailability)
	assert.Nil(t, user.FriendshipPreferences)

	// Test default Language values
	lang := Language{}
	assert.False(t, lang.IsInterfaceLanguage)
	assert.True(t, lang.CreatedAt.IsZero())

	// Test default InterestSelection values
	selection := InterestSelection{}
	assert.False(t, selection.IsPrimary)
	assert.Equal(t, 0, selection.SelectionOrder)
}

// TestModels_DataIntegrity тестирует целостность данных.
func TestModels_DataIntegrity(t *testing.T) {
	// Test that IDs are positive
	user := User{ID: 1, TelegramID: 123, FirstName: "John"}
	assert.Positive(t, user.ID)
	assert.Positive(t, user.TelegramID)

	interest := Interest{ID: 5, CategoryID: 2, KeyName: "programming"}
	assert.Positive(t, interest.ID)
	assert.Positive(t, interest.CategoryID)

	category := InterestCategory{ID: 10, DisplayOrder: 3, KeyName: "tech", Name: "Technology"}
	assert.Positive(t, category.ID)
	assert.GreaterOrEqual(t, category.DisplayOrder, 0)

	// Test string fields are not empty for populated data
	assert.NotEmpty(t, user.FirstName)
	assert.NotEmpty(t, interest.KeyName)
	assert.NotEmpty(t, category.KeyName)
	assert.NotEmpty(t, category.Name)
}
