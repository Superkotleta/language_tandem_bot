// Package integration contains end-to-end tests for the Language Exchange Bot
//
// E2E Test Scenarios:
//
// 1. User Registration Flow:
//   - New user sends /start command
//   - Bot detects interface language from Telegram settings
//   - User profile is created with detected language
//   - Bot sends welcome message in appropriate language
//
// 2. Profile Completion Flow:
//   - User selects native language from available options
//   - User selects target language and proficiency level
//   - User selects interests from categorized list
//   - Profile completion status is properly tracked
//   - Bot notifies when profile is complete
//
// 3. Partner Matching Flow:
//   - User with complete profile requests partner search
//   - System finds compatible users based on:
//   - Complementary language pairs (A learns B's native, B learns A's native)
//   - Matching interests
//   - Compatible proficiency levels
//   - Matching results are returned with contact information
//
// 4. Admin Feedback Flow:
//   - User submits feedback through bot interface
//   - Feedback is stored in database with user info
//   - Admin receives notification (if configured)
//   - Admin can view and process feedback via Admin API
//
// 5. Rate Limiting Flow:
//   - User sends multiple messages rapidly
//   - Rate limiter activates after threshold
//   - User receives rate limit warning
//   - Rate limit is properly reset after cooldown period
//
// 6. Error Recovery Flow:
//   - Database connection is lost during operation
//   - Circuit breaker activates
//   - Bot gracefully handles error and provides user feedback
//   - System recovers when database is restored
package integration //nolint:testpackage

import (
	"testing"

	"language-exchange-bot/internal/core"
	"language-exchange-bot/tests/helpers"
	"language-exchange-bot/tests/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// TelegramBotSuite набор тестов для Telegram бота.
type TelegramBotSuite struct {
	suite.Suite

	handler *mocks.TelegramHandlerWrapper
	service *core.BotService
	mockDB  *mocks.DatabaseMock
}

// SetupSuite выполняется один раз перед всеми тестами.
func (s *TelegramBotSuite) SetupSuite() {
	s.mockDB = mocks.NewDatabaseMock()
	s.handler, s.service = helpers.SetupTestBot(s.mockDB)
}

// SetupTest выполняется перед каждым тестом.
func (s *TelegramBotSuite) SetupTest() {
	s.mockDB.Reset()
	s.handler.Reset()
}

// TestStartCommand_NewUser тестирует команду /start для нового пользователя.
func (s *TelegramBotSuite) TestStartCommand_NewUser() {
	// Arrange
	userID, username := helpers.GetTestRegularUser()
	message := helpers.CreateTestCommand("start", userID, username)
	update := helpers.CreateUpdateWithMessage(message)

	// Act
	err := s.handler.HandleUpdate(update)

	// Assert
	s.NoError(err, "HandleUpdate should not return error")

	// Проверяем, что пользователь создан в БД
	user := s.mockDB.GetUser(userID)
	s.NotNil(user, "User should be created in database")
	s.Equal(userID, user.TelegramID, "User Telegram ID should match")
	s.Equal(username, user.Username, "Username should match")
	s.Equal("new", user.Status, "New user should have 'new' status")

	// Проверяем, что отправлено приветственное сообщение
	s.Equal(1, s.handler.GetSentMessagesCount(), "Should send exactly one message")

	lastMessage := s.handler.GetLastSentMessage()
	s.NotNil(lastMessage, "Should have sent a message")
	s.Equal(userID, lastMessage.ChatID, "Message should be sent to correct chat")
	s.Contains(lastMessage.Text, "Test", "Welcome message should contain user's name")

	// Проверяем, что отправлена клавиатура
	s.NotNil(lastMessage.ReplyMarkup, "Should include reply markup")
}

// TestStartCommand_ExistingUser тестирует команду /start для существующего пользователя.
func (s *TelegramBotSuite) TestStartCommand_ExistingUser() {
	// Arrange
	userID, username := helpers.GetTestRegularUser()

	// Создаем пользователя в БД
	existingUser, err := s.mockDB.CreateUser(userID, username, "TestUser", "en")
	s.NoError(err)

	existingUser.Status = "active"
	existingUser.ProfileCompletionLevel = 80
	_ = s.mockDB.UpdateUser(existingUser)

	message := helpers.CreateTestCommand("start", userID, username)
	update := helpers.CreateUpdateWithMessage(message)

	// Act
	err = s.handler.HandleUpdate(update)

	// Assert
	s.NoError(err, "HandleUpdate should not return error")

	// Проверяем, что пользователь не изменился
	user := s.mockDB.GetUser(userID)
	s.Equal("active", user.Status, "Existing user status should not change")
	s.Equal(80, user.ProfileCompletionLevel, "Profile completion should not change")

	// Проверяем, что отправлено сообщение
	s.Equal(1, s.handler.GetSentMessagesCount(), "Should send exactly one message")
}

// TestStartCommand_AdminUser тестирует команду /start для администратора.
func (s *TelegramBotSuite) TestStartCommand_AdminUser() {
	// Arrange
	adminID, adminUsername := helpers.GetTestAdminUser()
	message := helpers.CreateTestCommand("start", adminID, adminUsername)
	// Обновляем FirstName в сообщении для правильной проверки
	message.From.FirstName = "Admin"
	update := helpers.CreateUpdateWithMessage(message)

	// Act
	err := s.handler.HandleUpdate(update)

	// Assert
	s.NoError(err, "HandleUpdate should not return error")

	// Проверяем, что администратор создан
	user := s.mockDB.GetUser(adminID)
	s.NotNil(user, "Admin user should be created")
	s.Equal(adminUsername, user.Username, "Admin username should match")

	// Проверяем, что отправлено сообщение
	s.Equal(1, s.handler.GetSentMessagesCount(), "Should send exactly one message")

	lastMessage := s.handler.GetLastSentMessage()
	s.NotNil(lastMessage, "Should have sent a message")
	s.Contains(lastMessage.Text, "Admin", "Welcome message should contain admin's name")
}

// TestProfileCallback тестирует обработку callback для просмотра профиля.
func (s *TelegramBotSuite) TestProfileCallback() {
	// Arrange
	userID, username := helpers.GetTestRegularUser()

	// Создаем пользователя с заполненным профилем
	user, err := s.mockDB.CreateUser(userID, username, "TestUser", "en")
	s.NoError(err)

	user.NativeLanguageCode = "en"
	user.TargetLanguageCode = "ru"
	user.ProfileCompletionLevel = 60
	_ = s.mockDB.UpdateUser(user)

	callback := helpers.CreateTestCallback("profile_show", userID)
	update := helpers.CreateUpdateWithCallback(callback)

	// Act
	err = s.handler.HandleUpdate(update)

	// Assert
	s.NoError(err, "HandleUpdate should not return error")

	// Проверяем, что отправлен ответ на callback
	s.NotEmpty(s.handler.SentCallbacks, "Should send callback response")

	// Проверяем, что сообщение отредактировано
	s.Len(s.handler.EditedMessages, 1, "Should edit exactly one message")

	editedMessage := s.handler.EditedMessages[0]
	s.Equal(userID, editedMessage.ChatID, "Edited message should be in correct chat")
	s.NotEmpty(editedMessage.Text, "Edited message should have text")
}

// TestUnknownCommand тестирует обработку неизвестной команды.
func (s *TelegramBotSuite) TestUnknownCommand() {
	// Arrange
	userID, username := helpers.GetTestRegularUser()
	message := helpers.CreateTestCommand("unknown", userID, username)
	update := helpers.CreateUpdateWithMessage(message)

	// Act
	err := s.handler.HandleUpdate(update)

	// Assert
	s.NoError(err, "HandleUpdate should not return error")

	// Проверяем, что отправлено сообщение об ошибке
	s.Equal(1, s.handler.GetSentMessagesCount(), "Should send exactly one message")

	lastMessage := s.handler.GetLastSentMessage()
	s.Contains(lastMessage.Text, "Unknown command", "Should send unknown command message")
}

// TestDatabaseError тестирует обработку ошибки базы данных.
func (s *TelegramBotSuite) TestDatabaseError() {
	// Arrange
	s.mockDB.SetError(assert.AnError)

	userID, username := helpers.GetTestRegularUser()
	message := helpers.CreateTestCommand("start", userID, username)
	update := helpers.CreateUpdateWithMessage(message)

	// Act
	err := s.handler.HandleUpdate(update)

	// Assert
	s.Error(err, "Should return error when database fails")
	s.Error(s.handler.LastError, "Should store the database error")

	// Очищаем ошибку для следующих тестов
	s.mockDB.ClearError()
}

// TestTelegramBotSuite запускает весь набор тестов.
func TestTelegramBotSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(TelegramBotSuite))
}
