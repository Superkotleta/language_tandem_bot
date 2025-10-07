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

// ProfileCompletionSuite тесты для заполнения профиля пользователя.
type ProfileCompletionSuite struct {
	suite.Suite
	db      *database.DB
	service *core.BotService
	mockDB  *mocks.DatabaseMock // Используем мок для изоляции
}

// SetupSuite выполняется один раз перед всеми тестами.
func (s *ProfileCompletionSuite) SetupSuite() {
	s.mockDB = mocks.NewDatabaseMock()
	_, s.service = helpers.SetupTestBot(s.mockDB)
}

// TearDownSuite выполняется один раз после всех тестов.
func (s *ProfileCompletionSuite) TearDownSuite() {
	// Ничего не нужно закрывать для моков
}

// SetupTest выполняется перед каждым тестом.
func (s *ProfileCompletionSuite) SetupTest() {
	s.mockDB.Reset()
}

// TestCompleteProfileFlow тестирует полный процесс заполнения профиля.
func (s *ProfileCompletionSuite) TestCompleteProfileFlow() {
	// Arrange
	telegramID := int64(12345)
	username := "testuser"
	firstName := "Test"

	// Act & Assert - Создание пользователя
	user, err := s.service.HandleUserRegistration(telegramID, username, firstName, "ru")
	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), user)
	assert.Equal(s.T(), telegramID, user.TelegramID)
	assert.Equal(s.T(), "new", user.Status)

	// Act & Assert - Установка родного языка
	err = s.service.DB.UpdateUserNativeLanguage(user.ID, "ru")
	assert.NoError(s.T(), err)

	// Act & Assert - Установка изучаемого языка
	err = s.service.DB.UpdateUserTargetLanguage(user.ID, "en")
	assert.NoError(s.T(), err)

	// Act & Assert - Установка уровня языка
	err = s.service.DB.UpdateUserTargetLanguageLevel(user.ID, "intermediate")
	assert.NoError(s.T(), err)

	// Act & Assert - Добавление интересов
	interestIDs := []int{1, 2, 3}
	for _, interestID := range interestIDs {
		err = s.service.DB.SaveUserInterest(user.ID, interestID, false)
		assert.NoError(s.T(), err)
	}

	// Act & Assert - Получаем свежие данные пользователя
	updatedUser := s.mockDB.GetUserByID(user.ID)
	assert.NotNil(s.T(), updatedUser, "User should exist")

	// Act & Assert - Проверка завершенности профиля
	completed, err := s.service.IsProfileCompleted(updatedUser)
	assert.NoError(s.T(), err)
	assert.True(s.T(), completed, "Profile should be completed")

	// Act & Assert - Проверка резюме профиля
	summary, err := s.service.BuildProfileSummary(updatedUser)
	assert.NoError(s.T(), err)
	assert.Contains(s.T(), summary, "Русский", "Should contain native language")
	assert.Contains(s.T(), summary, "Английский", "Should contain target language")
	assert.Contains(s.T(), summary, "3", "Should contain interests count")
}

// TestProfileReset тестирует сброс профиля пользователя.
func (s *ProfileCompletionSuite) TestProfileReset() {
	// Arrange
	telegramID := int64(12346)
	username := "testuser2"
	firstName := "Test2"

	user, err := s.service.HandleUserRegistration(telegramID, username, firstName, "en")
	assert.NoError(s.T(), err)

	// Заполняем профиль
	_ = s.service.DB.UpdateUserNativeLanguage(user.ID, "en")
	_ = s.service.DB.UpdateUserTargetLanguage(user.ID, "ru")
	_ = s.service.DB.SaveUserInterest(user.ID, 1, false)
	_ = s.service.DB.SaveUserInterest(user.ID, 2, false)

	// Act - Сбрасываем профиль
	err = s.service.DB.ResetUserProfile(user.ID)
	assert.NoError(s.T(), err)

	// Assert - Проверяем, что профиль сброшен
	completed, err := s.service.IsProfileCompleted(user)
	assert.NoError(s.T(), err)
	assert.False(s.T(), completed, "Profile should not be completed after reset")

	// Проверяем, что интересы удалены
	interests, err := s.service.DB.GetUserSelectedInterests(user.ID)
	assert.NoError(s.T(), err)
	assert.Empty(s.T(), interests, "Interests should be cleared")
}

// TestLanguageSelection тестирует выбор языков.
func (s *ProfileCompletionSuite) TestLanguageSelection() {
	// Arrange
	telegramID := int64(12347)
	username := "testuser3"
	firstName := "Test3"

	user, err := s.service.HandleUserRegistration(telegramID, username, firstName, "es")
	assert.NoError(s.T(), err)

	// Act & Assert - Тестируем локализованные названия языков
	nativeName := s.service.GetLocalizedLanguageName("ru", "es")
	assert.NotEmpty(s.T(), nativeName, "Should return localized language name")

	targetName := s.service.GetLocalizedLanguageName("en", "es")
	assert.NotEmpty(s.T(), targetName, "Should return localized language name")

	// Act & Assert - Тестируем промпты для выбора языков
	nativePrompt := s.service.GetLanguagePrompt(user, "native")
	assert.Contains(s.T(), nativePrompt, "idioma", "Should contain language selection prompt")

	targetPrompt := s.service.GetLanguagePrompt(user, "target")
	assert.Contains(s.T(), targetPrompt, "aprendiendo", "Should contain learning language prompt")
}

// TestInterestManagement тестирует управление интересами.
func (s *ProfileCompletionSuite) TestInterestManagement() {
	// Arrange
	telegramID := int64(12348)
	username := "testuser4"
	firstName := "Test4"

	user, err := s.service.HandleUserRegistration(telegramID, username, firstName, "en")
	assert.NoError(s.T(), err)

	// Act & Assert - Добавление интересов
	interestIDs := []int{1, 2, 3, 4}
	for _, interestID := range interestIDs {
		err = s.service.DB.SaveUserInterest(user.ID, interestID, false)
		assert.NoError(s.T(), err)
	}

	// Проверяем, что интересы добавлены
	interests, err := s.service.DB.GetUserSelectedInterests(user.ID)
	assert.NoError(s.T(), err)
	assert.Len(s.T(), interests, 4, "Should have 4 interests")

	// Act & Assert - Удаление одного интереса
	err = s.service.DB.RemoveUserInterest(user.ID, 2)
	assert.NoError(s.T(), err)

	// Проверяем, что интерес удален
	interests, err = s.service.DB.GetUserSelectedInterests(user.ID)
	assert.NoError(s.T(), err)
	assert.Len(s.T(), interests, 3, "Should have 3 interests after removal")

	// Act & Assert - Очистка всех интересов
	err = s.service.DB.ClearUserInterests(user.ID)
	assert.NoError(s.T(), err)

	// Проверяем, что все интересы удалены
	interests, err = s.service.DB.GetUserSelectedInterests(user.ID)
	assert.NoError(s.T(), err)
	assert.Empty(s.T(), interests, "Should have no interests after clear")
}

// TestLocalizedInterests тестирует локализованные интересы.
func (s *ProfileCompletionSuite) TestLocalizedInterests() {
	// Arrange
	telegramID := int64(12349)
	username := "testuser5"
	firstName := "Test5"

	_, err := s.service.HandleUserRegistration(telegramID, username, firstName, "ru")
	assert.NoError(s.T(), err)

	// Act - Получаем локализованные интересы
	interests, err := s.service.GetLocalizedInterests("ru")
	assert.NoError(s.T(), err)

	// Assert - Проверяем, что интересы локализованы
	assert.NotEmpty(s.T(), interests, "Should return localized interests")

	// Проверяем, что есть хотя бы один интерес
	assert.Greater(s.T(), len(interests), 0, "Should have at least one interest")
}

// TestProfileCompletionLevel тестирует уровень завершенности профиля.
func (s *ProfileCompletionSuite) TestProfileCompletionLevel() {
	// Arrange
	telegramID := int64(12350)
	username := "testuser6"
	firstName := "Test6"

	user, err := s.service.HandleUserRegistration(telegramID, username, firstName, "en")
	assert.NoError(s.T(), err)

	// Act & Assert - Проверяем начальное состояние
	completed, err := s.service.IsProfileCompleted(user)
	assert.NoError(s.T(), err)
	assert.False(s.T(), completed, "New user profile should not be completed")

	// Act & Assert - Добавляем только родной язык
	err = s.service.DB.UpdateUserNativeLanguage(user.ID, "en")
	assert.NoError(s.T(), err)

	completed, err = s.service.IsProfileCompleted(user)
	assert.NoError(s.T(), err)
	assert.False(s.T(), completed, "Profile with only native language should not be completed")

	// Act & Assert - Добавляем изучаемый язык
	err = s.service.DB.UpdateUserTargetLanguage(user.ID, "ru")
	assert.NoError(s.T(), err)

	completed, err = s.service.IsProfileCompleted(user)
	assert.NoError(s.T(), err)
	assert.False(s.T(), completed, "Profile with languages but no interests should not be completed")

	// Act & Assert - Добавляем интерес
	err = s.service.DB.SaveUserInterest(user.ID, 1, false)
	assert.NoError(s.T(), err)

	completed, err = s.service.IsProfileCompleted(user)
	assert.NoError(s.T(), err)
	assert.True(s.T(), completed, "Profile with languages and interests should be completed")
}

// TestProfileCompletionSuite запускает весь набор тестов.
func TestProfileCompletionSuite(t *testing.T) {
	suite.Run(t, new(ProfileCompletionSuite))
}
