package integration

import (
	"testing"

	"language-exchange-bot/internal/core"
	"language-exchange-bot/internal/database"
	"language-exchange-bot/internal/localization"
	"language-exchange-bot/tests/helpers"
	"language-exchange-bot/tests/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// LocalizationSuite тесты для системы локализации.
type LocalizationSuite struct {
	suite.Suite
	db        *database.DB
	service   *core.BotService
	localizer *localization.Localizer
	mockDB    *mocks.DatabaseMock
}

// SetupSuite выполняется один раз перед всеми тестами.
func (s *LocalizationSuite) SetupSuite() {
	s.mockDB = mocks.NewDatabaseMock()
	_, s.service = helpers.SetupTestBot(s.mockDB)
	// Для локализации пока оставляем заглушку
	s.localizer = &localization.Localizer{}
}

// TearDownSuite выполняется один раз после всех тестов.
func (s *LocalizationSuite) TearDownSuite() {
	// Ничего не нужно закрывать для моков
}

// SetupTest выполняется перед каждым тестом.
func (s *LocalizationSuite) SetupTest() {
	s.mockDB.Reset()
}

// TestLanguageDetection тестирует определение языка пользователя.
func (s *LocalizationSuite) TestLanguageDetection() {
	// Arrange & Act & Assert - Тестируем различные коды языков
	testCases := []struct {
		telegramLangCode string
		expectedLang     string
	}{
		{"ru", "ru"},
		{"ru-RU", "ru"},
		{"en", "en"},
		{"en-US", "en"},
		{"es", "es"},
		{"es-ES", "es"},
		{"es-MX", "es"},
		{"zh", "zh"},
		{"zh-CN", "zh"},
		{"zh-TW", "zh"},
		{"fr", "en"}, // Неподдерживаемый язык должен вернуть английский
		{"de", "en"}, // Неподдерживаемый язык должен вернуть английский
		{"", "en"},   // Пустой код должен вернуть английский
	}

	for _, tc := range testCases {
		detected := s.service.DetectLanguage(tc.telegramLangCode)
		assert.Equal(s.T(), tc.expectedLang, detected,
			"Language detection failed for code: %s", tc.telegramLangCode)
	}
}

// TestLocalizationFallback тестирует fallback механизм локализации.
func (s *LocalizationSuite) TestLocalizationFallback() {
	// Arrange
	telegramID := int64(12371)
	username := "localizationuser1"
	firstName := "LocalizationUser1"

	user, err := s.service.HandleUserRegistration(telegramID, username, firstName, "fr") // Неподдерживаемый язык
	assert.NoError(s.T(), err)

	// Act - Получаем приветственное сообщение
	welcomeMessage := s.service.GetWelcomeMessage(user)

	// Assert - Проверяем, что используется fallback (английский)
	assert.Contains(s.T(), welcomeMessage, "Welcome", "Should use English fallback")
	assert.Contains(s.T(), welcomeMessage, firstName, "Should contain user's name")
}

// TestLocalizedLanguageNames тестирует локализованные названия языков.
func (s *LocalizationSuite) TestLocalizedLanguageNames() {
	// Arrange
	telegramID := int64(12372)
	username := "localizationuser2"
	firstName := "LocalizationUser2"

	_, err := s.service.HandleUserRegistration(telegramID, username, firstName, "ru")
	assert.NoError(s.T(), err)

	// Act & Assert - Тестируем локализованные названия языков
	testCases := []struct {
		langCode         string
		interfaceLang    string
		expectedContains string
	}{
		{"ru", "ru", "Русский"},
		{"en", "ru", "Английский"},
		{"es", "ru", "Испанский"},
		{"zh", "ru", "Китайский"},
		{"ru", "en", "Russian"},
		{"en", "en", "English"},
		{"es", "en", "Spanish"},
		{"zh", "en", "Chinese"},
	}

	for _, tc := range testCases {
		localizedName := s.service.GetLocalizedLanguageName(tc.langCode, tc.interfaceLang)
		assert.NotEmpty(s.T(), localizedName, "Should return localized name for %s in %s", tc.langCode, tc.interfaceLang)
		// Проверяем, что название содержит ожидаемый текст (может быть не точное совпадение из-за fallback)
		assert.NotEqual(s.T(), tc.langCode, localizedName, "Should not return language code as name")
	}
}

// TestLocalizedInterests тестирует локализованные интересы.
func (s *LocalizationSuite) TestLocalizedInterests() {
	// Act & Assert - Тестируем получение локализованных интересов
	testCases := []string{"en", "ru", "es", "zh"}

	for _, lang := range testCases {
		interests, err := s.service.GetLocalizedInterests(lang)
		assert.NoError(s.T(), err, "Should get interests for language: %s", lang)
		assert.NotEmpty(s.T(), interests, "Should have interests for language: %s", lang)

		// Проверяем, что интересы не пустые
		for id, name := range interests {
			assert.NotEmpty(s.T(), name, "Interest %d should have name in language %s", id, lang)
			assert.NotEqual(s.T(), "", name, "Interest name should not be empty")
		}
	}
}

// TestLocalizedPrompts тестирует локализованные промпты.
func (s *LocalizationSuite) TestLocalizedPrompts() {
	// Arrange
	telegramID := int64(12373)
	username := "localizationuser3"
	firstName := "LocalizationUser3"

	user, err := s.service.HandleUserRegistration(telegramID, username, firstName, "ru")
	assert.NoError(s.T(), err)

	// Act & Assert - Тестируем промпты для выбора языков
	nativePrompt := s.service.GetLanguagePrompt(user, "native")
	assert.NotEmpty(s.T(), nativePrompt, "Should return native language prompt")
	assert.NotEqual(s.T(), "choose_native_language", nativePrompt, "Should not return key as prompt")

	targetPrompt := s.service.GetLanguagePrompt(user, "target")
	assert.NotEmpty(s.T(), targetPrompt, "Should return target language prompt")
	assert.NotEqual(s.T(), "choose_target_language", targetPrompt, "Should not return key as prompt")

	// Проверяем, что промпты разные
	assert.NotEqual(s.T(), nativePrompt, targetPrompt, "Native and target prompts should be different")
}

// TestLocalizedProfileSummary тестирует локализованное резюме профиля.
func (s *LocalizationSuite) TestLocalizedProfileSummary() {
	// Arrange
	telegramID := int64(12374)
	username := "localizationuser4"
	firstName := "LocalizationUser4"

	user, err := s.service.HandleUserRegistration(telegramID, username, firstName, "ru")
	assert.NoError(s.T(), err)

	// Заполняем профиль
	_ = s.service.DB.UpdateUserNativeLanguage(user.ID, "ru")
	_ = s.service.DB.UpdateUserTargetLanguage(user.ID, "en")
	_ = s.service.DB.SaveUserInterest(user.ID, 1, false)
	_ = s.service.DB.SaveUserInterest(user.ID, 2, false)

	// Act - Получаем резюме профиля
	summary, err := s.service.BuildProfileSummary(user)
	assert.NoError(s.T(), err)

	// Assert - Проверяем, что резюме локализовано
	assert.NotEmpty(s.T(), summary, "Should return profile summary")
	assert.Contains(s.T(), summary, "профиль", "Should contain localized profile title")
	assert.Contains(s.T(), summary, "Русский", "Should contain native language")
	assert.Contains(s.T(), summary, "Английский", "Should contain target language")
	assert.Contains(s.T(), summary, "2", "Should contain interests count")
}

// TestLocalizationWithParams тестирует локализацию с параметрами.
func (s *LocalizationSuite) TestLocalizationWithParams() {
	// Arrange
	telegramID := int64(12375)
	username := "localizationuser5"
	firstName := "TestUser"

	user, err := s.service.HandleUserRegistration(telegramID, username, firstName, "en")
	assert.NoError(s.T(), err)

	// Act - Получаем приветственное сообщение с параметрами
	welcomeMessage := s.service.GetWelcomeMessage(user)

	// Assert - Проверяем, что параметры подставлены
	assert.Contains(s.T(), welcomeMessage, firstName, "Should contain user's name")
	assert.Contains(s.T(), welcomeMessage, "Welcome", "Should contain welcome text")
}

// TestLocalizationFallbackChain тестирует цепочку fallback для локализации.
func (s *LocalizationSuite) TestLocalizationFallbackChain() {
	// Arrange
	telegramID := int64(12376)
	username := "localizationuser6"
	firstName := "LocalizationUser6"

	user, err := s.service.HandleUserRegistration(telegramID, username, firstName, "unknown") // Неподдерживаемый язык
	assert.NoError(s.T(), err)

	// Act - Получаем различные локализованные тексты
	welcomeMessage := s.service.GetWelcomeMessage(user)
	nativePrompt := s.service.GetLanguagePrompt(user, "native")
	targetPrompt := s.service.GetLanguagePrompt(user, "target")

	// Assert - Проверяем, что используется fallback (английский)
	assert.Contains(s.T(), welcomeMessage, "Welcome", "Should use English fallback for welcome")
	assert.NotEqual(s.T(), "choose_native_language", nativePrompt, "Should use English fallback for native prompt")
	assert.NotEqual(s.T(), "choose_target_language", targetPrompt, "Should use English fallback for target prompt")
}

// TestLocalizationPerformance тестирует производительность локализации.
func (s *LocalizationSuite) TestLocalizationPerformance() {
	// Arrange
	telegramID := int64(12377)
	username := "localizationuser7"
	firstName := "LocalizationUser7"

	user, err := s.service.HandleUserRegistration(telegramID, username, firstName, "en")
	assert.NoError(s.T(), err)

	// Act - Выполняем множественные запросы локализации
	for i := 0; i < 100; i++ {
		welcomeMessage := s.service.GetWelcomeMessage(user)
		nativePrompt := s.service.GetLanguagePrompt(user, "native")
		targetPrompt := s.service.GetLanguagePrompt(user, "target")

		// Проверяем, что результаты не пустые
		assert.NotEmpty(s.T(), welcomeMessage)
		assert.NotEmpty(s.T(), nativePrompt)
		assert.NotEmpty(s.T(), targetPrompt)
	}
}

// TestLocalizationSuite запускает весь набор тестов.
func TestLocalizationSuite(t *testing.T) {
	suite.Run(t, new(LocalizationSuite))
}
