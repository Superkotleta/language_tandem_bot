package database

import (
	"database/sql"
	"testing"

	"language-exchange-bot/internal/database"
	"language-exchange-bot/internal/models"

	"github.com/stretchr/testify/assert"
)

func TestDatabaseInterface(t *testing.T) {
	// Тест для проверки, что интерфейс Database содержит все необходимые методы
	var db database.Database = &TestDatabase{}
	assert.NotNil(t, db)
}

func TestDatabase_UserOperations(t *testing.T) {
	// Arrange
	db := &TestDatabase{}

	// Act & Assert - FindOrCreateUser
	user, err := db.FindOrCreateUser(12345, "testuser", "Test User")
	assert.NoError(t, err)
	assert.Nil(t, user) // TestDatabase возвращает nil

	// Act & Assert - GetUserByTelegramID
	user, err = db.GetUserByTelegramID(12345)
	assert.NoError(t, err)
	assert.Nil(t, user)

	// Act & Assert - UpdateUser
	testUser := &models.User{
		ID:         1,
		TelegramID: 12345,
		Username:   "testuser",
		FirstName:  "Test User",
	}
	err = db.UpdateUser(testUser)
	assert.NoError(t, err)
}

func TestDatabase_UserLanguageOperations(t *testing.T) {
	// Arrange
	db := &TestDatabase{}

	// Act & Assert - UpdateUserInterfaceLanguage
	err := db.UpdateUserInterfaceLanguage(1, "en")
	assert.NoError(t, err)

	// Act & Assert - UpdateUserState
	err = db.UpdateUserState(1, "active")
	assert.NoError(t, err)

	// Act & Assert - UpdateUserStatus
	err = db.UpdateUserStatus(1, "active")
	assert.NoError(t, err)

	// Act & Assert - UpdateUserNativeLanguage
	err = db.UpdateUserNativeLanguage(1, "ru")
	assert.NoError(t, err)

	// Act & Assert - UpdateUserTargetLanguage
	err = db.UpdateUserTargetLanguage(1, "en")
	assert.NoError(t, err)

	// Act & Assert - UpdateUserTargetLanguageLevel
	err = db.UpdateUserTargetLanguageLevel(1, "intermediate")
	assert.NoError(t, err)

	// Act & Assert - UpdateUserProfileCompletionLevel
	err = db.UpdateUserProfileCompletionLevel(1, 100)
	assert.NoError(t, err)

	// Act & Assert - ResetUserProfile
	err = db.ResetUserProfile(1)
	assert.NoError(t, err)
}

func TestDatabase_LanguageOperations(t *testing.T) {
	// Arrange
	db := &TestDatabase{}

	// Act & Assert - GetLanguages
	languages, err := db.GetLanguages()
	assert.NoError(t, err)
	assert.Nil(t, languages)

	// Act & Assert - GetLanguageByCode
	language, err := db.GetLanguageByCode("en")
	assert.NoError(t, err)
	assert.Nil(t, language)
}

func TestDatabase_InterestOperations(t *testing.T) {
	// Arrange
	db := &TestDatabase{}

	// Act & Assert - GetInterests
	interests, err := db.GetInterests()
	assert.NoError(t, err)
	assert.Nil(t, interests)

	// Act & Assert - GetUserSelectedInterests
	userInterests, err := db.GetUserSelectedInterests(1)
	assert.NoError(t, err)
	assert.Equal(t, []int{1, 2, 3}, userInterests)

	// Act & Assert - SaveUserInterests
	err = db.SaveUserInterests(12345, []int{1, 2, 3})
	assert.NoError(t, err)

	// Act & Assert - SaveUserInterest
	err = db.SaveUserInterest(1, 1, true)
	assert.NoError(t, err)

	// Act & Assert - RemoveUserInterest
	err = db.RemoveUserInterest(1, 1)
	assert.NoError(t, err)

	// Act & Assert - ClearUserInterests
	err = db.ClearUserInterests(1)
	assert.NoError(t, err)
}

func TestDatabase_FeedbackOperations(t *testing.T) {
	// Arrange
	db := &TestDatabase{}

	// Act & Assert - SaveUserFeedback
	contactInfo := "test@example.com"
	err := db.SaveUserFeedback(1, "Test feedback", &contactInfo)
	assert.NoError(t, err)

	// Act & Assert - SaveUserFeedback without contact info
	err = db.SaveUserFeedback(1, "Test feedback", nil)
	assert.NoError(t, err)

	// Act & Assert - GetUserFeedbackByUserID
	feedbacks, err := db.GetUserFeedbackByUserID(1)
	assert.NoError(t, err)
	assert.Nil(t, feedbacks)

	// Act & Assert - GetUnprocessedFeedback
	unprocessedFeedbacks, err := db.GetUnprocessedFeedback()
	assert.NoError(t, err)
	assert.Nil(t, unprocessedFeedbacks)

	// Act & Assert - MarkFeedbackProcessed
	err = db.MarkFeedbackProcessed(1, "Processed")
	assert.NoError(t, err)

	// Act & Assert - GetUserDataForFeedback
	userData, err := db.GetUserDataForFeedback(1)
	assert.NoError(t, err)
	assert.NotNil(t, userData)

	// Act & Assert - GetAllFeedback
	allFeedbacks, err := db.GetAllFeedback()
	assert.NoError(t, err)
	assert.NotNil(t, allFeedbacks)

	// Act & Assert - DeleteFeedback
	err = db.DeleteFeedback(1)
	assert.NoError(t, err)

	// Act & Assert - ArchiveFeedback
	err = db.ArchiveFeedback(1)
	assert.NoError(t, err)

	// Act & Assert - UnarchiveFeedback
	err = db.UnarchiveFeedback(1)
	assert.NoError(t, err)

	// Act & Assert - UpdateFeedbackStatus
	err = db.UpdateFeedbackStatus(1, true)
	assert.NoError(t, err)

	// Act & Assert - DeleteAllProcessedFeedbacks
	count, err := db.DeleteAllProcessedFeedbacks()
	assert.NoError(t, err)
	assert.Equal(t, 0, count)
}

func TestDatabase_ConnectionOperations(t *testing.T) {
	// Arrange
	db := &TestDatabase{}

	// Act & Assert - GetConnection
	conn := db.GetConnection()
	assert.Nil(t, conn)

	// Act & Assert - Close
	err := db.Close()
	assert.NoError(t, err)
}

// TestDatabase - тестовая реализация интерфейса Database.
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
