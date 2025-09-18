package core

import (
	"database/sql"
	"testing"

	"language-exchange-bot/internal/core"
	"language-exchange-bot/internal/localization"
	"language-exchange-bot/internal/models"

	"github.com/stretchr/testify/assert"
)

func TestBotService_NewBotServiceWithInterface(t *testing.T) {
	// Arrange
	mockDB := &TestDatabase{}
	localizer := localization.NewLocalizer(nil)

	// Act
	service := core.NewBotServiceWithInterface(mockDB, localizer)

	// Assert
	assert.NotNil(t, service)
	assert.Equal(t, mockDB, service.DB)
	assert.Equal(t, localizer, service.Localizer)
}

func TestBotService_GetWelcomeMessage(t *testing.T) {
	// Arrange
	mockDB := &TestDatabase{}
	localizer := localization.NewLocalizer(nil)
	service := core.NewBotServiceWithInterface(mockDB, localizer)

	user := &models.User{
		FirstName:             "Test",
		InterfaceLanguageCode: "en",
	}

	// Act
	welcomeMessage := service.GetWelcomeMessage(user)

	// Assert
	assert.NotEmpty(t, welcomeMessage)
	assert.Contains(t, welcomeMessage, "Test")
}

func TestBotService_IsProfileCompleted(t *testing.T) {
	// Arrange
	mockDB := &TestDatabase{}
	localizer := localization.NewLocalizer(nil)
	service := core.NewBotServiceWithInterface(mockDB, localizer)

	tests := []struct {
		name     string
		user     *models.User
		expected bool
	}{
		{
			name: "Completed profile",
			user: &models.User{
				ID:                     1,
				TelegramID:             12345,
				Username:               "testuser",
				FirstName:              "Test",
				InterfaceLanguageCode:  "en",
				NativeLanguageCode:     "ru",
				TargetLanguageCode:     "en",
				Interests:              []int{1, 2, 3},
				Status:                 models.StatusActive,
				State:                  models.StateActive,
				ProfileCompletionLevel: 100,
			},
			expected: true,
		},
		{
			name: "Incomplete profile",
			user: &models.User{
				ID:                     1,
				TelegramID:             12345,
				Username:               "testuser",
				FirstName:              "Test",
				InterfaceLanguageCode:  "en",
				Status:                 models.StatusNew,
				State:                  models.StateNew,
				ProfileCompletionLevel: 0,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			completed, err := service.IsProfileCompleted(tt.user)

			// Assert
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, completed)
		})
	}
}

func TestBotService_BuildProfileSummary(t *testing.T) {
	// Arrange
	mockDB := &TestDatabase{}
	localizer := localization.NewLocalizer(nil)
	service := core.NewBotServiceWithInterface(mockDB, localizer)

	user := &models.User{
		ID:                     1,
		TelegramID:             12345,
		Username:               "testuser",
		FirstName:              "Test",
		InterfaceLanguageCode:  "en",
		NativeLanguageCode:     "ru",
		TargetLanguageCode:     "en",
		Interests:              []int{1, 2, 3},
		Status:                 models.StatusActive,
		State:                  models.StateActive,
		ProfileCompletionLevel: 100,
	}

	// Act
	summary, err := service.BuildProfileSummary(user)

	// Assert
	assert.NoError(t, err)
	assert.NotEmpty(t, summary)
	assert.Contains(t, summary, "profile")
}

func TestBotService_GetUserDataForFeedback(t *testing.T) {
	// Arrange
	mockDB := &TestDatabase{}
	localizer := localization.NewLocalizer(nil)
	service := core.NewBotServiceWithInterface(mockDB, localizer)

	userID := 1

	// Act
	data, err := service.GetUserDataForFeedback(userID)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, data)
}

func TestBotService_ArchiveFeedback(t *testing.T) {
	// Arrange
	mockDB := &TestDatabase{}
	localizer := localization.NewLocalizer(nil)
	service := core.NewBotServiceWithInterface(mockDB, localizer)

	feedbackID := 1

	// Act
	err := service.ArchiveFeedback(feedbackID)

	// Assert
	assert.NoError(t, err)
}

func TestBotService_DeleteFeedback(t *testing.T) {
	// Arrange
	mockDB := &TestDatabase{}
	localizer := localization.NewLocalizer(nil)
	service := core.NewBotServiceWithInterface(mockDB, localizer)

	feedbackID := 1

	// Act
	err := service.DeleteFeedback(feedbackID)

	// Assert
	assert.NoError(t, err)
}

func TestBotService_UpdateFeedbackStatus(t *testing.T) {
	// Arrange
	mockDB := &TestDatabase{}
	localizer := localization.NewLocalizer(nil)
	service := core.NewBotServiceWithInterface(mockDB, localizer)

	feedbackID := 1
	isProcessed := true

	// Act
	err := service.UpdateFeedbackStatus(feedbackID, isProcessed)

	// Assert
	assert.NoError(t, err)
}

func TestBotService_DeleteAllProcessedFeedbacks(t *testing.T) {
	// Arrange
	mockDB := &TestDatabase{}
	localizer := localization.NewLocalizer(nil)
	service := core.NewBotServiceWithInterface(mockDB, localizer)

	// Act
	count, err := service.DeleteAllProcessedFeedbacks()

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, 0, count)
}

func TestBotService_UnarchiveFeedback(t *testing.T) {
	// Arrange
	mockDB := &TestDatabase{}
	localizer := localization.NewLocalizer(nil)
	service := core.NewBotServiceWithInterface(mockDB, localizer)

	feedbackID := 1

	// Act
	err := service.UnarchiveFeedback(feedbackID)

	// Assert
	assert.NoError(t, err)
}

func TestBotService_GetAllFeedback(t *testing.T) {
	// Arrange
	mockDB := &TestDatabase{}
	localizer := localization.NewLocalizer(nil)
	service := core.NewBotServiceWithInterface(mockDB, localizer)

	// Act
	feedbacks, err := service.GetAllFeedback()

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, feedbacks)
}

// TestDatabase - простая тестовая реализация интерфейса Database.
type TestDatabase struct{}

func (t *TestDatabase) FindOrCreateUser(telegramID int64, username, firstName string) (*models.User, error) {
	return nil, nil
}
func (t *TestDatabase) GetUserByTelegramID(telegramID int64) (*models.User, error)    { return nil, nil }
func (t *TestDatabase) UpdateUser(user *models.User) error                            { return nil }
func (t *TestDatabase) UpdateUserInterfaceLanguage(userID int, language string) error { return nil }
func (t *TestDatabase) UpdateUserState(userID int, state string) error                { return nil }
func (t *TestDatabase) UpdateUserStatus(userID int, status string) error              { return nil }
func (t *TestDatabase) UpdateUserNativeLanguage(userID int, langCode string) error    { return nil }
func (t *TestDatabase) UpdateUserTargetLanguage(userID int, langCode string) error    { return nil }
func (t *TestDatabase) UpdateUserTargetLanguageLevel(userID int, level string) error  { return nil }
func (t *TestDatabase) UpdateUserProfileCompletionLevel(userID int, level int) error  { return nil }
func (t *TestDatabase) ResetUserProfile(userID int) error                             { return nil }
func (t *TestDatabase) GetLanguages() ([]*models.Language, error)                     { return nil, nil }
func (t *TestDatabase) GetLanguageByCode(code string) (*models.Language, error)       { return nil, nil }
func (t *TestDatabase) GetInterests() ([]*models.Interest, error)                     { return nil, nil }
func (t *TestDatabase) GetUserSelectedInterests(userID int) ([]int, error) {
	return []int{1, 2, 3}, nil
}
func (t *TestDatabase) SaveUserInterests(userID int64, interestIDs []int) error       { return nil }
func (t *TestDatabase) SaveUserInterest(userID, interestID int, isPrimary bool) error { return nil }
func (t *TestDatabase) RemoveUserInterest(userID, interestID int) error               { return nil }
func (t *TestDatabase) ClearUserInterests(userID int) error                           { return nil }
func (t *TestDatabase) SaveUserFeedback(userID int, feedbackText string, contactInfo *string) error {
	return nil
}
func (t *TestDatabase) GetUserFeedbackByUserID(userID int) ([]map[string]interface{}, error) {
	return nil, nil
}
func (t *TestDatabase) GetUnprocessedFeedback() ([]map[string]interface{}, error)        { return nil, nil }
func (t *TestDatabase) MarkFeedbackProcessed(feedbackID int, adminResponse string) error { return nil }
func (t *TestDatabase) GetUserDataForFeedback(userID int) (map[string]interface{}, error) {
	return map[string]interface{}{
		"telegram_id": int64(12345),
		"first_name":  "Test User",
		"username":    "testuser",
	}, nil
}
func (t *TestDatabase) GetAllFeedback() ([]map[string]interface{}, error) {
	return []map[string]interface{}{}, nil
}
func (t *TestDatabase) DeleteFeedback(feedbackID int) error                         { return nil }
func (t *TestDatabase) ArchiveFeedback(feedbackID int) error                        { return nil }
func (t *TestDatabase) UnarchiveFeedback(feedbackID int) error                      { return nil }
func (t *TestDatabase) UpdateFeedbackStatus(feedbackID int, isProcessed bool) error { return nil }
func (t *TestDatabase) DeleteAllProcessedFeedbacks() (int, error)                   { return 0, nil }
func (t *TestDatabase) GetConnection() *sql.DB                                      { return nil }
func (t *TestDatabase) Close() error                                                { return nil }
