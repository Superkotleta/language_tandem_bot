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

// AdminFunctionsSuite тесты для административных функций.
type AdminFunctionsSuite struct {
	suite.Suite
	db      *database.DB
	service *core.BotService
	mockDB  *mocks.DatabaseMock
}

// SetupSuite выполняется один раз перед всеми тестами.
func (s *AdminFunctionsSuite) SetupSuite() {
	s.mockDB = mocks.NewDatabaseMock()
	_, s.service = helpers.SetupTestBot(s.mockDB)
}

// TearDownSuite выполняется один раз после всех тестов.
func (s *AdminFunctionsSuite) TearDownSuite() {
	// Ничего не нужно закрывать для моков
}

// SetupTest выполняется перед каждым тестом.
func (s *AdminFunctionsSuite) SetupTest() {
	s.mockDB.Reset()
}

// TestAdminFeedbackOverview тестирует обзор отзывов для администратора.
func (s *AdminFunctionsSuite) TestAdminFeedbackOverview() {
	// Arrange - Создаем несколько пользователей с отзывами
	users := []struct {
		telegramID int64
		username   string
		firstName  string
		feedback   string
	}{
		{12361, "user1", "User1", "First feedback message with sufficient length."},
		{12362, "user2", "User2", "Second feedback message with sufficient length."},
		{12363, "user3", "User3", "Third feedback message with sufficient length."},
	}

	for _, userData := range users {
		user, err := s.service.HandleUserRegistration(userData.telegramID, userData.username, userData.firstName, "en")
		assert.NoError(s.T(), err)

		err = s.service.SaveUserFeedback(user.ID, userData.feedback, nil, []int64{123456})
		assert.NoError(s.T(), err)
	}

	// Act - Получаем все отзывы для администратора
	allFeedbacks, err := s.service.GetAllFeedback()
	assert.NoError(s.T(), err)

	// Assert - Проверяем, что получены все отзывы
	assert.Len(s.T(), allFeedbacks, 3, "Should have three feedbacks")

	// Проверяем структуру отзыва
	feedback := allFeedbacks[0]
	assert.Contains(s.T(), feedback, "id")
	assert.Contains(s.T(), feedback, "feedback_text")
	assert.Contains(s.T(), feedback, "created_at")
	assert.Contains(s.T(), feedback, "telegram_id")
	assert.Contains(s.T(), feedback, "first_name")
	assert.Contains(s.T(), feedback, "is_processed")
}

// TestAdminFeedbackProcessing тестирует обработку отзывов администратором.
func (s *AdminFunctionsSuite) TestAdminFeedbackProcessing() {
	// Arrange
	telegramID := int64(12364)
	username := "adminuser1"
	firstName := "AdminUser1"

	user, err := s.service.HandleUserRegistration(telegramID, username, firstName, "en")
	assert.NoError(s.T(), err)

	// Создаем отзыв
	feedbackText := "This is a test feedback for admin processing."
	err = s.service.SaveUserFeedback(user.ID, feedbackText, nil, []int64{123456})
	assert.NoError(s.T(), err)

	// Получаем необработанные отзывы
	unprocessedFeedbacks, err := s.service.GetAllUnprocessedFeedback()
	assert.NoError(s.T(), err)
	assert.Len(s.T(), unprocessedFeedbacks, 1, "Should have one unprocessed feedback")

	feedbackID := unprocessedFeedbacks[0]["id"].(int)

	// Act - Обрабатываем отзыв
	adminResponse := "Thank you for your feedback. We have received it and will review it."
	err = s.service.MarkFeedbackProcessed(feedbackID, adminResponse)
	assert.NoError(s.T(), err)

	// Assert - Проверяем, что отзыв обработан
	processedFeedbacks, err := s.service.GetAllFeedback()
	assert.NoError(s.T(), err)
	assert.Len(s.T(), processedFeedbacks, 1, "Should have one feedback")

	feedback := processedFeedbacks[0]
	assert.Equal(s.T(), true, feedback["is_processed"])
	assert.Equal(s.T(), adminResponse, feedback["admin_response"])

	// Проверяем, что необработанных отзывов больше нет
	unprocessedFeedbacks, err = s.service.GetAllUnprocessedFeedback()
	assert.NoError(s.T(), err)
	assert.Len(s.T(), unprocessedFeedbacks, 0, "Should have no unprocessed feedbacks")
}

// TestAdminBulkOperations тестирует массовые операции администратора.
func (s *AdminFunctionsSuite) TestAdminBulkOperations() {
	// Arrange - Создаем несколько обработанных отзывов
	users := []struct {
		telegramID int64
		username   string
		firstName  string
		feedback   string
	}{
		{12365, "bulkuser1", "BulkUser1", "First processed feedback message with sufficient length."},
		{12366, "bulkuser2", "BulkUser2", "Second processed feedback message with sufficient length."},
		{12367, "bulkuser3", "BulkUser3", "Third processed feedback message with sufficient length."},
	}

	var feedbackIDs []int
	for _, userData := range users {
		user, err := s.service.HandleUserRegistration(userData.telegramID, userData.username, userData.firstName, "en")
		assert.NoError(s.T(), err)

		err = s.service.SaveUserFeedback(user.ID, userData.feedback, nil, []int64{123456})
		assert.NoError(s.T(), err)

		// Помечаем отзыв как обработанный
		feedbacks, err := s.service.DB.GetUserFeedbackByUserID(user.ID)
		assert.NoError(s.T(), err)
		feedbackID := feedbacks[0]["id"].(int)
		feedbackIDs = append(feedbackIDs, feedbackID)

		err = s.service.MarkFeedbackProcessed(feedbackID, "Processed by admin")
		assert.NoError(s.T(), err)
	}

	// Act - Удаляем все обработанные отзывы
	deletedCount, err := s.service.DeleteAllProcessedFeedbacks()
	assert.NoError(s.T(), err)

	// Assert - Проверяем, что все обработанные отзывы удалены
	assert.Equal(s.T(), 3, deletedCount, "Should delete three processed feedbacks")

	// Проверяем, что отзывов больше нет
	allFeedbacks, err := s.service.GetAllFeedback()
	assert.NoError(s.T(), err)
	assert.Len(s.T(), allFeedbacks, 0, "Should have no feedbacks after bulk deletion")
}

// TestAdminFeedbackUnarchive тестирует возврат отзыва в активные.
func (s *AdminFunctionsSuite) TestAdminFeedbackUnarchive() {
	// Arrange
	telegramID := int64(12368)
	username := "unarchiveuser1"
	firstName := "UnarchiveUser1"

	user, err := s.service.HandleUserRegistration(telegramID, username, firstName, "en")
	assert.NoError(s.T(), err)

	// Создаем отзыв
	feedbackText := "This is a test feedback for unarchiving."
	err = s.service.SaveUserFeedback(user.ID, feedbackText, nil, []int64{123456})
	assert.NoError(s.T(), err)

	// Получаем ID отзыва
	feedbacks, err := s.service.DB.GetUserFeedbackByUserID(user.ID)
	assert.NoError(s.T(), err)
	feedbackID := feedbacks[0]["id"].(int)

	// Архивируем отзыв
	err = s.service.ArchiveFeedback(feedbackID)
	assert.NoError(s.T(), err)

	// Act - Возвращаем отзыв в активные
	err = s.service.UnarchiveFeedback(feedbackID)
	assert.NoError(s.T(), err)

	// Assert - Проверяем, что отзыв возвращен в активные
	unarchivedFeedbacks, err := s.service.DB.GetUserFeedbackByUserID(user.ID)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), false, unarchivedFeedbacks[0]["is_processed"])

	// Проверяем, что отзыв снова в списке необработанных
	unprocessedFeedbacks, err := s.service.GetAllUnprocessedFeedback()
	assert.NoError(s.T(), err)
	assert.Len(s.T(), unprocessedFeedbacks, 1, "Should have one unprocessed feedback after unarchive")
}

// TestAdminFeedbackNotFound тестирует обработку несуществующих отзывов.
func (s *AdminFunctionsSuite) TestAdminFeedbackNotFound() {
	// Act & Assert - Пытаемся обработать несуществующий отзыв
	nonExistentID := 99999
	err := s.service.MarkFeedbackProcessed(nonExistentID, "Response")
	assert.Error(s.T(), err, "Should return error for non-existent feedback")

	// Act & Assert - Пытаемся удалить несуществующий отзыв
	err = s.service.DeleteFeedback(nonExistentID)
	assert.Error(s.T(), err, "Should return error for non-existent feedback")

	// Act & Assert - Пытаемся архивировать несуществующий отзыв
	err = s.service.ArchiveFeedback(nonExistentID)
	assert.Error(s.T(), err, "Should return error for non-existent feedback")

	// Act & Assert - Пытаемся вернуть в активные несуществующий отзыв
	err = s.service.UnarchiveFeedback(nonExistentID)
	assert.Error(s.T(), err, "Should return error for non-existent feedback")
}

// TestAdminFeedbackStatusUpdate тестирует обновление статуса отзыва.
func (s *AdminFunctionsSuite) TestAdminFeedbackStatusUpdate() {
	// Arrange
	telegramID := int64(12369)
	username := "statususer1"
	firstName := "StatusUser1"

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

	// Act - Обновляем статус на обработанный
	err = s.service.UpdateFeedbackStatus(feedbackID, true)
	assert.NoError(s.T(), err)

	// Assert - Проверяем, что статус обновлен
	updatedFeedbacks, err := s.service.DB.GetUserFeedbackByUserID(user.ID)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), true, updatedFeedbacks[0]["is_processed"])

	// Act - Обновляем статус на необработанный
	err = s.service.UpdateFeedbackStatus(feedbackID, false)
	assert.NoError(s.T(), err)

	// Assert - Проверяем, что статус обновлен
	updatedFeedbacks, err = s.service.DB.GetUserFeedbackByUserID(user.ID)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), false, updatedFeedbacks[0]["is_processed"])
}

// TestAdminFunctionsSuite запускает весь набор тестов.
func TestAdminFunctionsSuite(t *testing.T) {
	suite.Run(t, new(AdminFunctionsSuite))
}
