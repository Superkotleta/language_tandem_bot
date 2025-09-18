package integration

import (
	"testing"

	"language-exchange-bot/internal/core"
	"language-exchange-bot/internal/database"
	"language-exchange-bot/tests/helpers"
	"language-exchange-bot/tests/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// FeedbackSystemSuite тесты для системы отзывов
type FeedbackSystemSuite struct {
	suite.Suite
	db      *database.DB
	service *core.BotService
	mockDB  *mocks.DatabaseMock
}

// SetupSuite выполняется один раз перед всеми тестами
func (s *FeedbackSystemSuite) SetupSuite() {
	s.mockDB = mocks.NewDatabaseMock()
	_, s.service = helpers.SetupTestBot(s.mockDB)
}

// TearDownSuite выполняется один раз после всех тестов
func (s *FeedbackSystemSuite) TearDownSuite() {
	// Ничего не нужно закрывать для моков
}

// SetupTest выполняется перед каждым тестом
func (s *FeedbackSystemSuite) SetupTest() {
	s.mockDB.Reset()
}

// TestFeedbackSubmission тестирует отправку отзыва
func (s *FeedbackSystemSuite) TestFeedbackSubmission() {
	// Arrange
	telegramID := int64(12351)
	username := "feedbackuser1"
	firstName := "FeedbackUser1"

	user, err := s.service.HandleUserRegistration(telegramID, username, firstName, "en")
	assert.NoError(s.T(), err)

	feedbackText := "This is a test feedback message with sufficient length to pass validation."
	contactInfo := "test@example.com"

	// Act - Отправляем отзыв
	err = s.service.SaveUserFeedback(user.ID, feedbackText, &contactInfo, []int64{123456})
	assert.NoError(s.T(), err)

	// Assert - Проверяем, что отзыв сохранен
	feedbacks, err := s.service.DB.GetUserFeedbackByUserID(user.ID)
	assert.NoError(s.T(), err)
	assert.Len(s.T(), feedbacks, 1, "Should have one feedback")

	feedback := feedbacks[0]
	assert.Equal(s.T(), feedbackText, feedback["feedback_text"])

	// contact_info может быть указателем на строку
	contactInfoValue := feedback["contact_info"]
	if contactInfoPtr, ok := contactInfoValue.(*string); ok {
		assert.Equal(s.T(), contactInfo, *contactInfoPtr)
	} else {
		assert.Equal(s.T(), contactInfo, contactInfoValue)
	}

	assert.Equal(s.T(), false, feedback["is_processed"])
}

// TestFeedbackValidation тестирует валидацию отзывов
func (s *FeedbackSystemSuite) TestFeedbackValidation() {
	// Arrange
	telegramID := int64(12352)
	username := "feedbackuser2"
	firstName := "FeedbackUser2"

	_, err := s.service.HandleUserRegistration(telegramID, username, firstName, "en")
	assert.NoError(s.T(), err)

	// Act & Assert - Тестируем слишком короткий отзыв
	shortFeedback := "Short"
	err = s.service.ValidateFeedback(shortFeedback)
	assert.Error(s.T(), err, "Should reject short feedback")
	assert.Contains(s.T(), err.Error(), "too short")

	// Act & Assert - Тестируем слишком длинный отзыв
	longFeedback := string(make([]byte, 1001)) // 1001 символов
	for i := range longFeedback {
		longFeedback = longFeedback[:i] + "a" + longFeedback[i+1:]
	}
	err = s.service.ValidateFeedback(longFeedback)
	assert.Error(s.T(), err, "Should reject long feedback")
	assert.Contains(s.T(), err.Error(), "too long")

	// Act & Assert - Тестируем валидный отзыв
	validFeedback := "This is a valid feedback message with appropriate length."
	err = s.service.ValidateFeedback(validFeedback)
	assert.NoError(s.T(), err, "Should accept valid feedback")
}

// TestFeedbackWithoutContactInfo тестирует отправку отзыва без контактной информации
func (s *FeedbackSystemSuite) TestFeedbackWithoutContactInfo() {
	// Arrange
	telegramID := int64(12353)
	username := "feedbackuser3"
	firstName := "FeedbackUser3"

	user, err := s.service.HandleUserRegistration(telegramID, username, firstName, "en")
	assert.NoError(s.T(), err)

	feedbackText := "This is a test feedback without contact information."

	// Act - Отправляем отзыв без контактной информации
	err = s.service.SaveUserFeedback(user.ID, feedbackText, nil, []int64{123456})
	assert.NoError(s.T(), err)

	// Assert - Проверяем, что отзыв сохранен
	feedbacks, err := s.service.DB.GetUserFeedbackByUserID(user.ID)
	assert.NoError(s.T(), err)
	assert.Len(s.T(), feedbacks, 1, "Should have one feedback")

	feedback := feedbacks[0]
	assert.Equal(s.T(), feedbackText, feedback["feedback_text"])
	assert.Nil(s.T(), feedback["contact_info"])
}

// TestMultipleFeedbacks тестирует отправку нескольких отзывов от одного пользователя
func (s *FeedbackSystemSuite) TestMultipleFeedbacks() {
	// Arrange
	telegramID := int64(12354)
	username := "feedbackuser4"
	firstName := "FeedbackUser4"

	user, err := s.service.HandleUserRegistration(telegramID, username, firstName, "en")
	assert.NoError(s.T(), err)

	// Act - Отправляем несколько отзывов
	feedbacks := []string{
		"First feedback message with sufficient length.",
		"Second feedback message with sufficient length.",
		"Third feedback message with sufficient length.",
	}

	for i, feedbackText := range feedbacks {
		err = s.service.SaveUserFeedback(user.ID, feedbackText, nil, []int64{123456})
		assert.NoError(s.T(), err, "Should save feedback %d", i+1)
	}

	// Assert - Проверяем, что все отзывы сохранены
	savedFeedbacks, err := s.service.DB.GetUserFeedbackByUserID(user.ID)
	assert.NoError(s.T(), err)
	assert.Len(s.T(), savedFeedbacks, 3, "Should have three feedbacks")

	// Проверяем, что отзывы отсортированы по дате создания (новые первыми)
	for i, feedback := range savedFeedbacks {
		assert.Equal(s.T(), feedbacks[len(feedbacks)-1-i], feedback["feedback_text"])
	}
}

// TestAdminFeedbackManagement тестирует управление отзывами администратором
func (s *FeedbackSystemSuite) TestAdminFeedbackManagement() {
	// Arrange
	telegramID := int64(12355)
	username := "feedbackuser5"
	firstName := "FeedbackUser5"

	user, err := s.service.HandleUserRegistration(telegramID, username, firstName, "en")
	assert.NoError(s.T(), err)

	// Создаем отзыв
	feedbackText := "This is a test feedback for admin management."
	err = s.service.SaveUserFeedback(user.ID, feedbackText, nil, []int64{123456})
	assert.NoError(s.T(), err)

	// Act - Получаем необработанные отзывы
	unprocessedFeedbacks, err := s.service.GetAllUnprocessedFeedback()
	assert.NoError(s.T(), err)
	assert.Len(s.T(), unprocessedFeedbacks, 1, "Should have one unprocessed feedback")

	feedbackID := unprocessedFeedbacks[0]["id"].(int)

	// Act - Помечаем отзыв как обработанный
	adminResponse := "Thank you for your feedback. We will review it."
	err = s.service.MarkFeedbackProcessed(feedbackID, adminResponse)
	assert.NoError(s.T(), err)

	// Assert - Проверяем, что отзыв помечен как обработанный
	userFeedbacks, err := s.service.DB.GetUserFeedbackByUserID(user.ID)
	assert.NoError(s.T(), err)
	assert.Len(s.T(), userFeedbacks, 1, "Should have one feedback")

	feedback := userFeedbacks[0]
	assert.Equal(s.T(), true, feedback["is_processed"])
	assert.Equal(s.T(), adminResponse, feedback["admin_response"])

	// Assert - Проверяем, что необработанных отзывов больше нет
	unprocessedFeedbacks, err = s.service.GetAllUnprocessedFeedback()
	assert.NoError(s.T(), err)
	assert.Len(s.T(), unprocessedFeedbacks, 0, "Should have no unprocessed feedbacks")
}

// TestFeedbackStatusUpdate тестирует обновление статуса отзыва
func (s *FeedbackSystemSuite) TestFeedbackStatusUpdate() {
	// Arrange
	telegramID := int64(12356)
	username := "feedbackuser6"
	firstName := "FeedbackUser6"

	user, err := s.service.HandleUserRegistration(telegramID, username, firstName, "en")
	assert.NoError(s.T(), err)

	// Создаем отзыв
	feedbackText := "This is a test feedback for status update."
	err = s.service.SaveUserFeedback(user.ID, feedbackText, nil, []int64{123456})
	assert.NoError(s.T(), err)

	// Получаем ID отзыва
	feedbacks, err := s.service.DB.GetUserFeedbackByUserID(user.ID)
	assert.NoError(s.T(), err)
	feedbackID := feedbacks[0]["id"].(int)

	// Act - Обновляем статус отзыва
	err = s.service.UpdateFeedbackStatus(feedbackID, true)
	assert.NoError(s.T(), err)

	// Assert - Проверяем, что статус обновлен
	updatedFeedbacks, err := s.service.DB.GetUserFeedbackByUserID(user.ID)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), true, updatedFeedbacks[0]["is_processed"])
}

// TestFeedbackArchiving тестирует архивирование отзыва
func (s *FeedbackSystemSuite) TestFeedbackArchiving() {
	// Arrange
	telegramID := int64(12357)
	username := "feedbackuser7"
	firstName := "FeedbackUser7"

	user, err := s.service.HandleUserRegistration(telegramID, username, firstName, "en")
	assert.NoError(s.T(), err)

	// Создаем отзыв
	feedbackText := "This is a test feedback for archiving."
	err = s.service.SaveUserFeedback(user.ID, feedbackText, nil, []int64{123456})
	assert.NoError(s.T(), err)

	// Получаем ID отзыва
	feedbacks, err := s.service.DB.GetUserFeedbackByUserID(user.ID)
	assert.NoError(s.T(), err)
	feedbackID := feedbacks[0]["id"].(int)

	// Act - Архивируем отзыв
	err = s.service.ArchiveFeedback(feedbackID)
	assert.NoError(s.T(), err)

	// Assert - Проверяем, что отзыв заархивирован
	archivedFeedbacks, err := s.service.DB.GetUserFeedbackByUserID(user.ID)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), true, archivedFeedbacks[0]["is_processed"])
}

// TestFeedbackDeletion тестирует удаление отзыва
func (s *FeedbackSystemSuite) TestFeedbackDeletion() {
	// Arrange
	telegramID := int64(12358)
	username := "feedbackuser8"
	firstName := "FeedbackUser8"

	user, err := s.service.HandleUserRegistration(telegramID, username, firstName, "en")
	assert.NoError(s.T(), err)

	// Создаем отзыв
	feedbackText := "This is a test feedback for deletion."
	err = s.service.SaveUserFeedback(user.ID, feedbackText, nil, []int64{123456})
	assert.NoError(s.T(), err)

	// Получаем ID отзыва
	feedbacks, err := s.service.DB.GetUserFeedbackByUserID(user.ID)
	assert.NoError(s.T(), err)
	feedbackID := feedbacks[0]["id"].(int)

	// Act - Удаляем отзыв
	err = s.service.DeleteFeedback(feedbackID)
	assert.NoError(s.T(), err)

	// Assert - Проверяем, что отзыв удален
	deletedFeedbacks, err := s.service.DB.GetUserFeedbackByUserID(user.ID)
	assert.NoError(s.T(), err)
	assert.Len(s.T(), deletedFeedbacks, 0, "Should have no feedbacks after deletion")
}

// TestFeedbackSystemSuite запускает весь набор тестов
func TestFeedbackSystemSuite(t *testing.T) {
	suite.Run(t, new(FeedbackSystemSuite))
}
