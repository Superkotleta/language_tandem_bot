package core

import (
	"database/sql"
	"testing"

	"language-exchange-bot/internal/localization"
	"language-exchange-bot/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockDatabase - мок для интерфейса Database
type MockDatabase struct {
	mock.Mock
}

func (m *MockDatabase) FindOrCreateUser(telegramID int64, username, firstName string) (*models.User, error) {
	args := m.Called(telegramID, username, firstName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockDatabase) GetUserByTelegramID(telegramID int64) (*models.User, error) {
	args := m.Called(telegramID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockDatabase) UpdateUser(user *models.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockDatabase) UpdateUserInterfaceLanguage(userID int, language string) error {
	args := m.Called(userID, language)
	return args.Error(0)
}

func (m *MockDatabase) UpdateUserState(userID int, state string) error {
	args := m.Called(userID, state)
	return args.Error(0)
}

func (m *MockDatabase) UpdateUserStatus(userID int, status string) error {
	args := m.Called(userID, status)
	return args.Error(0)
}

func (m *MockDatabase) UpdateUserProfileCompletionLevel(userID int, level int) error {
	args := m.Called(userID, level)
	return args.Error(0)
}

func (m *MockDatabase) UpdateUserNativeLanguage(userID int, langCode string) error {
	args := m.Called(userID, langCode)
	return args.Error(0)
}

func (m *MockDatabase) UpdateUserTargetLanguage(userID int, langCode string) error {
	args := m.Called(userID, langCode)
	return args.Error(0)
}

func (m *MockDatabase) UpdateUserTargetLanguageLevel(userID int, level string) error {
	args := m.Called(userID, level)
	return args.Error(0)
}

func (m *MockDatabase) ResetUserProfile(userID int) error {
	args := m.Called(userID)
	return args.Error(0)
}

func (m *MockDatabase) GetLanguages() ([]*models.Language, error) {
	args := m.Called()
	return args.Get(0).([]*models.Language), args.Error(1)
}

func (m *MockDatabase) GetLanguageByCode(code string) (*models.Language, error) {
	args := m.Called(code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Language), args.Error(1)
}

func (m *MockDatabase) GetInterests() ([]*models.Interest, error) {
	args := m.Called()
	return args.Get(0).([]*models.Interest), args.Error(1)
}

func (m *MockDatabase) GetUserSelectedInterests(userID int) ([]int, error) {
	args := m.Called(userID)
	return args.Get(0).([]int), args.Error(1)
}

func (m *MockDatabase) SaveUserInterests(userID int, interestIDs []int) error {
	args := m.Called(userID, interestIDs)
	return args.Error(0)
}

func (m *MockDatabase) SaveUserInterest(userID, interestID int, isPrimary bool) error {
	args := m.Called(userID, interestID, isPrimary)
	return args.Error(0)
}

func (m *MockDatabase) RemoveUserInterest(userID, interestID int) error {
	args := m.Called(userID, interestID)
	return args.Error(0)
}

func (m *MockDatabase) ClearUserInterests(userID int) error {
	args := m.Called(userID)
	return args.Error(0)
}

func (m *MockDatabase) GetUserInterestSelections(userID int) ([]models.InterestSelection, error) {
	args := m.Called(userID)
	return args.Get(0).([]models.InterestSelection), args.Error(1)
}

func (m *MockDatabase) GetInterestByID(interestID int) (*models.Interest, error) {
	args := m.Called(interestID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Interest), args.Error(1)
}

func (m *MockDatabase) SaveUserFeedback(userID int, feedbackText string, contactInfo *string) error {
	args := m.Called(userID, feedbackText, contactInfo)
	return args.Error(0)
}

func (m *MockDatabase) GetUnprocessedFeedback() ([]map[string]interface{}, error) {
	args := m.Called()
	return args.Get(0).([]map[string]interface{}), args.Error(1)
}

func (m *MockDatabase) MarkFeedbackProcessed(feedbackID int, adminResponse string) error {
	args := m.Called(feedbackID, adminResponse)
	return args.Error(0)
}

func (m *MockDatabase) GetConnection() *sql.DB {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*sql.DB)
}

func (m *MockDatabase) Close() error {
	args := m.Called()
	return args.Error(0)
}

// Методы для работы с доступностью
func (m *MockDatabase) SaveTimeAvailability(userID int, availability *models.TimeAvailability) error {
	args := m.Called(userID, availability)
	return args.Error(0)
}

func (m *MockDatabase) GetTimeAvailability(userID int) (*models.TimeAvailability, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.TimeAvailability), args.Error(1)
}

func (m *MockDatabase) SaveFriendshipPreferences(userID int, preferences *models.FriendshipPreferences) error {
	args := m.Called(userID, preferences)
	return args.Error(0)
}

func (m *MockDatabase) GetFriendshipPreferences(userID int) (*models.FriendshipPreferences, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.FriendshipPreferences), args.Error(1)
}

func TestHandleUserRegistration(t *testing.T) {
	mockDB := new(MockDatabase)
	mockLocalizer := &localization.Localizer{}

	service := NewBotServiceWithInterface(mockDB, mockLocalizer)

	telegramID := int64(12345)
	username := "testuser"
	firstName := "Test"
	telegramLangCode := "ru"

	// Ожидаем вызов FindOrCreateUser
	expectedUser := &models.User{
		ID:                    1,
		TelegramID:            telegramID,
		Username:              username,
		FirstName:             firstName,
		InterfaceLanguageCode: "ru",
		Status:                models.StatusNew,
	}

	mockDB.On("FindOrCreateUser", telegramID, username, firstName).Return(expectedUser, nil)
	mockDB.On("UpdateUserInterfaceLanguage", 1, "ru").Return(nil)

	user, err := service.HandleUserRegistration(telegramID, username, firstName, telegramLangCode)

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, telegramID, user.TelegramID)
	assert.Equal(t, "ru", user.InterfaceLanguageCode)

	mockDB.AssertExpectations(t)
}

func TestDetectLanguage(t *testing.T) {
	tests := []struct {
		name             string
		telegramLangCode string
		expected         string
	}{
		{"Russian", "ru", "ru"},
		{"Russian RU", "ru-RU", "ru"},
		{"English", "en", "en"},
		{"Spanish", "es", "es"},
		{"Chinese", "zh", "zh"},
		{"Unknown", "fr", "en"},
		{"Empty", "", "en"},
	}

	service := &BotService{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.DetectLanguage(tt.telegramLangCode)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsProfileCompleted(t *testing.T) {
	t.Run("Profile completed", func(t *testing.T) {
		mockDB := new(MockDatabase)
		mockLocalizer := &localization.Localizer{}

		service := NewBotServiceWithInterface(mockDB, mockLocalizer)

		user := &models.User{
			ID:                 1,
			NativeLanguageCode: "ru",
			TargetLanguageCode: "en",
		}

		mockDB.On("GetUserSelectedInterests", 1).Return([]int{1, 2, 3}, nil)

		completed, err := service.IsProfileCompleted(user)

		assert.NoError(t, err)
		assert.True(t, completed)
		mockDB.AssertExpectations(t)
	})

	t.Run("Profile not completed - no languages", func(t *testing.T) {
		mockDB := new(MockDatabase)
		mockLocalizer := &localization.Localizer{}

		service := NewBotServiceWithInterface(mockDB, mockLocalizer)

		user := &models.User{
			ID:                 1,
			NativeLanguageCode: "",
			TargetLanguageCode: "en",
		}

		completed, err := service.IsProfileCompleted(user)

		assert.NoError(t, err)
		assert.False(t, completed)
	})

	t.Run("Profile not completed - no interests", func(t *testing.T) {
		mockDB := new(MockDatabase)
		mockLocalizer := &localization.Localizer{}

		service := NewBotServiceWithInterface(mockDB, mockLocalizer)

		user := &models.User{
			ID:                 1,
			NativeLanguageCode: "ru",
			TargetLanguageCode: "en",
		}

		mockDB.On("GetUserSelectedInterests", 1).Return([]int{}, nil)

		completed, err := service.IsProfileCompleted(user)

		assert.NoError(t, err)
		assert.False(t, completed)
		mockDB.AssertExpectations(t)
	})
}

func TestValidateFeedback(t *testing.T) {
	service := &BotService{}

	t.Run("Valid feedback", func(t *testing.T) {
		err := service.ValidateFeedback("This is a valid feedback text with enough characters")
		assert.NoError(t, err)
	})

	t.Run("Too short feedback", func(t *testing.T) {
		err := service.ValidateFeedback("Hi")
		assert.Error(t, err)
	})

	t.Run("Too long feedback", func(t *testing.T) {
		longText := string(make([]byte, 1001))
		err := service.ValidateFeedback(longText)
		assert.Error(t, err)
	})
}

func TestFormatLanguageLevel(t *testing.T) {
	service := &BotService{}

	tests := []struct {
		name     string
		level    string
		expected string
	}{
		{"Beginner", "beginner", "A1-A2"},
		{"Elementary", "elementary", "A2-B1"},
		{"Intermediate", "intermediate", "B1-B2"},
		{"Upper Intermediate", "upper_intermediate", "B2-C1"},
		{"Advanced", "advanced", "C1-C2"},
		{"Unknown", "unknown", "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.formatLanguageLevel(tt.level)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetDisplayName(t *testing.T) {
	service := &BotService{}

	t.Run("Regular user", func(t *testing.T) {
		user := &models.User{
			FirstName: "John",
			Username:  "john_doe",
		}

		result := service.getDisplayName(user)
		assert.Equal(t, "John", result)
	})

	t.Run("Special user", func(t *testing.T) {
		user := &models.User{
			FirstName: "Test",
			Username:  "madam_di_5",
		}

		result := service.getDisplayName(user)
		assert.Equal(t, "Лисёнок 🦊", result)
	})
}

func TestGetWelcomeMessage(t *testing.T) {
	mockLocalizer := &localization.Localizer{}

	// Mock для тестирования локализации
	service := &BotService{
		Localizer: mockLocalizer,
	}

	user := &models.User{
		ID:                    1,
		FirstName:             "John",
		InterfaceLanguageCode: "en",
	}

	// Поскольку мы не можем легко замокать Localizer.GetWithParams,
	// тестируем что метод не падает и возвращает не пустую строку
	message := service.GetWelcomeMessage(user)
	assert.NotEmpty(t, message)
}

func TestGetLanguagePrompt(t *testing.T) {
	mockLocalizer := &localization.Localizer{}

	service := &BotService{
		Localizer: mockLocalizer,
	}

	user := &models.User{
		ID:                    1,
		InterfaceLanguageCode: "en",
	}

	// Test native language prompt
	nativePrompt := service.GetLanguagePrompt(user, "native")
	assert.NotEmpty(t, nativePrompt)

	// Test target language prompt
	targetPrompt := service.GetLanguagePrompt(user, "target")
	assert.NotEmpty(t, targetPrompt)

	// Test default prompt
	defaultPrompt := service.GetLanguagePrompt(user, "unknown")
	assert.NotEmpty(t, defaultPrompt)
}

func TestGetLocalizedLanguageName(t *testing.T) {
	mockLocalizer := &localization.Localizer{}

	service := &BotService{
		Localizer: mockLocalizer,
	}

	// Test with existing language
	name := service.GetLocalizedLanguageName("en", "ru")
	assert.NotEmpty(t, name)

	// Test with non-existing language
	name = service.GetLocalizedLanguageName("unknown", "en")
	assert.NotEmpty(t, name)
}

func TestSetFeedbackNotificationFunc(t *testing.T) {
	service := &BotService{}

	// Test setting notification function
	called := false
	testFunc := func(data map[string]interface{}) error {
		called = true
		return nil
	}

	service.SetFeedbackNotificationFunc(testFunc)
	assert.NotNil(t, service.FeedbackNotificationFunc)

	// Test calling the function
	err := service.FeedbackNotificationFunc(map[string]interface{}{"test": "data"})
	assert.NoError(t, err)
	assert.True(t, called)
}

func TestExecuteWithCircuitBreakers(t *testing.T) {
	service := &BotService{}

	t.Run("ExecuteWithTelegramCircuitBreaker - no circuit breaker", func(t *testing.T) {
		result, err := service.ExecuteWithTelegramCircuitBreaker(func() (interface{}, error) {
			return "success", nil
		})
		assert.NoError(t, err)
		assert.Equal(t, "success", result)
	})

	t.Run("ExecuteWithDatabaseCircuitBreaker - no circuit breaker", func(t *testing.T) {
		result, err := service.ExecuteWithDatabaseCircuitBreaker(func() (interface{}, error) {
			return "success", nil
		})
		assert.NoError(t, err)
		assert.Equal(t, "success", result)
	})

	t.Run("ExecuteWithRedisCircuitBreaker - no circuit breaker", func(t *testing.T) {
		result, err := service.ExecuteWithRedisCircuitBreaker(func() (interface{}, error) {
			return "success", nil
		})
		assert.NoError(t, err)
		assert.Equal(t, "success", result)
	})
}

func TestGetCircuitBreakerStates(t *testing.T) {
	service := &BotService{}

	// Test with no circuit breakers initialized - should return empty map
	states := service.GetCircuitBreakerStates()
	assert.NotNil(t, states)
	assert.Empty(t, states) // No circuit breakers initialized in empty service
}

func TestGetCircuitBreakerCounts(t *testing.T) {
	service := &BotService{}

	// Test with no circuit breakers initialized - should return empty map
	counts := service.GetCircuitBreakerCounts()
	assert.NotNil(t, counts)
	assert.Empty(t, counts) // No circuit breakers initialized in empty service
}

func TestStopCache(t *testing.T) {
	service := &BotService{}

	// Test that StopCache doesn't panic when cache is nil
	// The method should check if cache is not nil before calling Stop()
	assert.NotPanics(t, func() {
		service.StopCache()
	})
}

func TestGetConfig(t *testing.T) {
	service := &BotService{}

	// Test that GetConfig doesn't panic (may return nil for empty service)
	assert.NotPanics(t, func() {
		_ = service.GetConfig()
	})
}
