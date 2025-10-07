package integration

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
	userID, username := helpers.GetTestRegularUserID()
	message := helpers.CreateTestCommand("start", userID, username)
	update := helpers.CreateUpdateWithMessage(message)

	// Act
	err := s.handler.HandleUpdate(update)

	// Assert
	assert.NoError(s.T(), err, "HandleUpdate should not return error")

	// Проверяем, что пользователь создан в БД
	user := s.mockDB.GetUser(userID)
	assert.NotNil(s.T(), user, "User should be created in database")
	assert.Equal(s.T(), userID, user.TelegramID, "User Telegram ID should match")
	assert.Equal(s.T(), username, user.Username, "Username should match")
	assert.Equal(s.T(), "new", user.Status, "New user should have 'new' status")

	// Проверяем, что отправлено приветственное сообщение
	assert.Equal(s.T(), 1, s.handler.GetSentMessagesCount(), "Should send exactly one message")

	lastMessage := s.handler.GetLastSentMessage()
	assert.NotNil(s.T(), lastMessage, "Should have sent a message")
	assert.Equal(s.T(), userID, lastMessage.ChatID, "Message should be sent to correct chat")
	assert.Contains(s.T(), lastMessage.Text, "Test", "Welcome message should contain user's name")

	// Проверяем, что отправлена клавиатура
	assert.NotNil(s.T(), lastMessage.ReplyMarkup, "Should include reply markup")
}

// TestStartCommand_ExistingUser тестирует команду /start для существующего пользователя.
func (s *TelegramBotSuite) TestStartCommand_ExistingUser() {
	// Arrange
	userID, username := helpers.GetTestRegularUserID()

	// Создаем пользователя в БД
	existingUser, err := s.mockDB.CreateUser(userID, username, "TestUser", "en")
	assert.NoError(s.T(), err)
	existingUser.Status = "active"
	existingUser.ProfileCompletionLevel = 80
	_ = s.mockDB.UpdateUser(existingUser)

	message := helpers.CreateTestCommand("start", userID, username)
	update := helpers.CreateUpdateWithMessage(message)

	// Act
	err = s.handler.HandleUpdate(update)

	// Assert
	assert.NoError(s.T(), err, "HandleUpdate should not return error")

	// Проверяем, что пользователь не изменился
	user := s.mockDB.GetUser(userID)
	assert.Equal(s.T(), "active", user.Status, "Existing user status should not change")
	assert.Equal(s.T(), 80, user.ProfileCompletionLevel, "Profile completion should not change")

	// Проверяем, что отправлено сообщение
	assert.Equal(s.T(), 1, s.handler.GetSentMessagesCount(), "Should send exactly one message")
}

// TestStartCommand_AdminUser тестирует команду /start для администратора.
func (s *TelegramBotSuite) TestStartCommand_AdminUser() {
	// Arrange
	adminID, adminUsername := helpers.GetTestAdminUserID()
	message := helpers.CreateTestCommand("start", adminID, adminUsername)
	// Обновляем FirstName в сообщении для правильной проверки
	message.From.FirstName = "Admin"
	update := helpers.CreateUpdateWithMessage(message)

	// Act
	err := s.handler.HandleUpdate(update)

	// Assert
	assert.NoError(s.T(), err, "HandleUpdate should not return error")

	// Проверяем, что администратор создан
	user := s.mockDB.GetUser(adminID)
	assert.NotNil(s.T(), user, "Admin user should be created")
	assert.Equal(s.T(), adminUsername, user.Username, "Admin username should match")

	// Проверяем, что отправлено сообщение
	assert.Equal(s.T(), 1, s.handler.GetSentMessagesCount(), "Should send exactly one message")

	lastMessage := s.handler.GetLastSentMessage()
	assert.NotNil(s.T(), lastMessage, "Should have sent a message")
	assert.Contains(s.T(), lastMessage.Text, "Admin", "Welcome message should contain admin's name")
}

// TestProfileCallback тестирует обработку callback для просмотра профиля.
func (s *TelegramBotSuite) TestProfileCallback() {
	// Arrange
	userID, username := helpers.GetTestRegularUserID()

	// Создаем пользователя с заполненным профилем
	user, err := s.mockDB.CreateUser(userID, username, "TestUser", "en")
	assert.NoError(s.T(), err)
	user.NativeLanguageCode = "en"
	user.TargetLanguageCode = "ru"
	user.ProfileCompletionLevel = 60
	_ = s.mockDB.UpdateUser(user)

	callback := helpers.CreateTestCallback("profile_show", userID)
	update := helpers.CreateUpdateWithCallback(callback)

	// Act
	err = s.handler.HandleUpdate(update)

	// Assert
	assert.NoError(s.T(), err, "HandleUpdate should not return error")

	// Проверяем, что отправлен ответ на callback
	assert.Greater(s.T(), len(s.handler.SentCallbacks), 0, "Should send callback response")

	// Проверяем, что сообщение отредактировано
	assert.Equal(s.T(), 1, len(s.handler.EditedMessages), "Should edit exactly one message")

	editedMessage := s.handler.EditedMessages[0]
	assert.Equal(s.T(), userID, editedMessage.ChatID, "Edited message should be in correct chat")
	assert.NotEmpty(s.T(), editedMessage.Text, "Edited message should have text")
}

// TestUnknownCommand тестирует обработку неизвестной команды.
func (s *TelegramBotSuite) TestUnknownCommand() {
	// Arrange
	userID, username := helpers.GetTestRegularUserID()
	message := helpers.CreateTestCommand("unknown", userID, username)
	update := helpers.CreateUpdateWithMessage(message)

	// Act
	err := s.handler.HandleUpdate(update)

	// Assert
	assert.NoError(s.T(), err, "HandleUpdate should not return error")

	// Проверяем, что отправлено сообщение об ошибке
	assert.Equal(s.T(), 1, s.handler.GetSentMessagesCount(), "Should send exactly one message")

	lastMessage := s.handler.GetLastSentMessage()
	assert.Contains(s.T(), lastMessage.Text, "Unknown command", "Should send unknown command message")
}

// TestDatabaseError тестирует обработку ошибки базы данных.
func (s *TelegramBotSuite) TestDatabaseError() {
	// Arrange
	s.mockDB.SetError(assert.AnError)

	userID, username := helpers.GetTestRegularUserID()
	message := helpers.CreateTestCommand("start", userID, username)
	update := helpers.CreateUpdateWithMessage(message)

	// Act
	err := s.handler.HandleUpdate(update)

	// Assert
	assert.Error(s.T(), err, "Should return error when database fails")
	assert.Equal(s.T(), assert.AnError, s.handler.LastError, "Should store the database error")

	// Очищаем ошибку для следующих тестов
	s.mockDB.ClearError()
}

// TestTelegramBotSuite запускает весь набор тестов.
func TestTelegramBotSuite(t *testing.T) {
	suite.Run(t, new(TelegramBotSuite))
}
