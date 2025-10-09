package integration //nolint:testpackage

import (
	"testing"

	"language-exchange-bot/internal/core"
	"language-exchange-bot/tests/helpers"
	"language-exchange-bot/tests/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// E2EUserRegistrationSuite - E2E тест для полного цикла регистрации пользователя.
type E2EUserRegistrationSuite struct {
	suite.Suite

	handler *mocks.TelegramHandlerWrapper
	service *core.BotService
	mockDB  *mocks.DatabaseMock
}

// SetupSuite выполняется один раз перед всеми тестами.
func (s *E2EUserRegistrationSuite) SetupSuite() {
	s.mockDB = mocks.NewDatabaseMock()
	s.handler, s.service = helpers.SetupTestBot(s.mockDB)
}

// SetupTest выполняется перед каждым тестом.
func (s *E2EUserRegistrationSuite) SetupTest() {
	s.mockDB.Reset()
	s.handler.Reset()
}

// TestCompleteUserRegistrationFlow - полный E2E сценарий регистрации пользователя.
func (s *E2EUserRegistrationSuite) TestCompleteUserRegistrationFlow() {
	// Test Data
	userID := int64(123456789)
	username := "newuser"
	telegramLangCode := "en"

	// Step 1: User sends /start command (first interaction)
	message := helpers.CreateTestCommand("start", userID, username)
	message.From.LanguageCode = telegramLangCode // Set Telegram language
	update := helpers.CreateUpdateWithMessage(message)

	// Act: Process the /start command
	err := s.handler.HandleUpdate(update)

	// Assert: No errors during processing
	s.NoError(err, "HandleUpdate should not return error during registration")

	// Assert: User was created in database
	user := s.mockDB.GetUser(userID)
	s.NotNil(user, "User should be created in database")
	s.Equal(userID, user.TelegramID, "User TelegramID should match")
	s.Equal(username, user.Username, "Username should match")
	s.Equal("en", user.InterfaceLanguageCode, "Interface language should be detected from Telegram")

	// Assert: Welcome message was sent
	s.Equal(1, s.handler.GetSentMessagesCount(), "Should send exactly one message")

	welcomeMsg := s.handler.GetLastSentMessage()
	s.NotNil(welcomeMsg, "Welcome message should be sent")
	s.Contains(welcomeMsg.Text, "Test", "Welcome message should contain user's first name")

	// Assert: Message contains expected keyboard buttons
	s.NotNil(welcomeMsg.ReplyMarkup, "Message should contain inline keyboard")

	// Step 2: Verify that subsequent interactions work with the registered user
	// User sends another command to verify user persistence
	statusMessage := helpers.CreateTestCommand("status", userID, username)
	statusUpdate := helpers.CreateUpdateWithMessage(statusMessage)

	err = s.handler.HandleUpdate(statusUpdate)
	s.NoError(err, "Subsequent commands should work with registered user")

	// Should have sent status message
	s.Equal(2, s.handler.GetSentMessagesCount(), "Should have sent status message")
}

// TestUserRegistrationWithRussianLanguage - тест регистрации с русским языком Telegram.
func (s *E2EUserRegistrationSuite) TestUserRegistrationWithRussianLanguage() {
	userID := int64(987654321)
	username := "russianuser"
	telegramLangCode := "ru"

	// Send /start command
	message := helpers.CreateTestCommand("start", userID, username)
	message.From.LanguageCode = telegramLangCode
	update := helpers.CreateUpdateWithMessage(message)

	err := s.handler.HandleUpdate(update)
	s.NoError(err)

	// Verify Russian language detection
	user := s.mockDB.GetUser(userID)
	s.NotNil(user, "User should be created")
	s.Equal("ru", user.InterfaceLanguageCode, "Should detect Russian language from Telegram")
}

// TestUserRegistrationFallbackToDefaultLanguage - тест с неизвестным языком Telegram.
func (s *E2EUserRegistrationSuite) TestUserRegistrationFallbackToDefaultLanguage() {
	userID := int64(555666777)
	username := "unknownlang"
	telegramLangCode := "xx" // Unknown language code

	// Send /start command
	message := helpers.CreateTestCommand("start", userID, username)
	message.From.LanguageCode = telegramLangCode
	update := helpers.CreateUpdateWithMessage(message)

	err := s.handler.HandleUpdate(update)
	s.NoError(err)

	// Verify fallback to default language
	user := s.mockDB.GetUser(userID)
	s.NotNil(user, "User should be created")
	s.Equal("en", user.InterfaceLanguageCode, "Should fallback to English for unknown language")
}

// TestUserRegistrationDatabaseError - тест обработки ошибки базы данных.
func (s *E2EUserRegistrationSuite) TestUserRegistrationDatabaseError() {
	userID := int64(111222333)
	username := "erroruser"

	// Set mock to return error
	s.mockDB.SetError(assert.AnError)

	// Send /start command
	message := helpers.CreateTestCommand("start", userID, username)
	update := helpers.CreateUpdateWithMessage(message)

	// Act: Process should return error when database fails
	err := s.handler.HandleUpdate(update)

	// Assert: Handler should return error when database fails during registration
	s.Error(err, "Handler should return error when database fails")

	// Assert: Error should be recorded in handler
	s.Error(s.handler.LastError, "Error should be recorded in handler")
}

// TestSuite runs the test suite.
func TestE2EUserRegistrationSuite(t *testing.T) {
	suite.Run(t, new(E2EUserRegistrationSuite))
}
