package integration //nolint:testpackage

import (
	"testing"

	"language-exchange-bot/internal/core"
	"language-exchange-bot/internal/models"
	"language-exchange-bot/tests/helpers"
	"language-exchange-bot/tests/mocks"

	"github.com/stretchr/testify/suite"
)

// AvailabilitySystemSuite тестирует систему доступности.
type AvailabilitySystemSuite struct {
	suite.Suite

	handler *mocks.TelegramHandlerWrapper
	service *core.BotService
	mockDB  *mocks.DatabaseMock
}

// SetupSuite выполняется один раз перед всеми тестами.
func (s *AvailabilitySystemSuite) SetupSuite() {
	s.mockDB = mocks.NewDatabaseMock()
	s.handler, s.service = helpers.SetupTestBot(s.mockDB)
}

// SetupTest выполняется перед каждым тестом.
func (s *AvailabilitySystemSuite) SetupTest() {
	s.mockDB.Reset()
}

// TestSaveAndGetTimeAvailability тестирует сохранение и получение временной доступности.
func (s *AvailabilitySystemSuite) TestSaveAndGetTimeAvailability() {
	// Создаем пользователя
	user := &models.User{
		ID:                     1,
		TelegramID:             123456789,
		Username:               "testuser",
		FirstName:              "Test",
		InterfaceLanguageCode:  "ru",
		State:                  models.StateActive,
		Status:                 models.StatusActive,
		ProfileCompletionLevel: 100,
	}

	// Сохраняем пользователя
	err := s.mockDB.UpdateUser(user)
	s.NoError(err)

	// Тестируем сохранение временной доступности
	availability := &models.TimeAvailability{
		DayType:      "weekdays",
		SpecificDays: []string{},
		TimeSlot:     "morning",
	}

	err = s.mockDB.SaveTimeAvailability(user.ID, availability)
	s.NoError(err)

	// Проверяем, что данные сохранились
	savedAvailability, err := s.mockDB.GetTimeAvailability(user.ID)
	s.NoError(err)
	s.Equal("weekdays", savedAvailability.DayType)
	s.Equal("morning", savedAvailability.TimeSlot)
}

// TestSaveAndGetFriendshipPreferences тестирует сохранение и получение предпочтений общения.
func (s *AvailabilitySystemSuite) TestSaveAndGetFriendshipPreferences() {
	// Создаем пользователя
	user := &models.User{
		ID:                    1,
		TelegramID:            123456789,
		Username:              "testuser",
		FirstName:             "Test",
		InterfaceLanguageCode: "ru",
	}

	// Сохраняем пользователя
	err := s.mockDB.UpdateUser(user)
	s.NoError(err)

	// Тестируем сохранение предпочтений общения
	preferences := &models.FriendshipPreferences{
		ActivityType:       "movies",
		CommunicationStyle: "text",
		CommunicationFreq:  "weekly",
	}

	err = s.mockDB.SaveFriendshipPreferences(user.ID, preferences)
	s.NoError(err)

	// Проверяем, что данные сохранились
	savedPreferences, err := s.mockDB.GetFriendshipPreferences(user.ID)
	s.NoError(err)
	s.Equal("movies", savedPreferences.ActivityType)
	s.Equal("text", savedPreferences.CommunicationStyle)
	s.Equal("weekly", savedPreferences.CommunicationFreq)
}

// TestSpecificDaysSelection тестирует выбор конкретных дней недели.
func (s *AvailabilitySystemSuite) TestSpecificDaysSelection() {
	// Создаем пользователя
	user := &models.User{
		ID:         1,
		TelegramID: 123456789,
		Username:   "testuser",
		FirstName:  "Test",
	}

	// Сохраняем пользователя
	err := s.mockDB.UpdateUser(user)
	s.NoError(err)

	// Тестируем выбор конкретных дней
	specificAvailability := &models.TimeAvailability{
		DayType:      "specific",
		SpecificDays: []string{"monday", "wednesday", "friday"},
		TimeSlot:     "day",
	}

	err = s.mockDB.SaveTimeAvailability(user.ID, specificAvailability)
	s.NoError(err)

	specificSaved, err := s.mockDB.GetTimeAvailability(user.ID)
	s.NoError(err)
	s.Equal("specific", specificSaved.DayType)
	s.Contains(specificSaved.SpecificDays, "monday")
	s.Contains(specificSaved.SpecificDays, "wednesday")
	s.Contains(specificSaved.SpecificDays, "friday")
	s.Equal("day", specificSaved.TimeSlot)
}

// TestDefaultValues тестирует значения по умолчанию для новых пользователей.
func (s *AvailabilitySystemSuite) TestDefaultValues() {
	// Создаем пользователя
	user := &models.User{
		ID:         1,
		TelegramID: 123456789,
		Username:   "testuser",
		FirstName:  "Test",
	}

	// Сохраняем пользователя
	err := s.mockDB.UpdateUser(user)
	s.NoError(err)

	// Проверяем значения по умолчанию для доступности
	availability, err := s.mockDB.GetTimeAvailability(user.ID)
	s.NoError(err)
	s.Equal("any", availability.DayType)
	s.Empty(availability.SpecificDays)
	s.Equal("any", availability.TimeSlot)

	// Проверяем значения по умолчанию для предпочтений
	preferences, err := s.mockDB.GetFriendshipPreferences(user.ID)
	s.NoError(err)
	s.Equal("casual_chat", preferences.ActivityType)
	s.Equal("text", preferences.CommunicationStyle)
	s.Equal("weekly", preferences.CommunicationFreq)
}

// TestAvailabilitySystemIntegration запускает все тесты системы доступности.
func TestAvailabilitySystemIntegration(t *testing.T) {
	suite.Run(t, new(AvailabilitySystemSuite))
}
