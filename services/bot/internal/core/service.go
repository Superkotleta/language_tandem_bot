package core

import (
	"language-exchange-bot/internal/database"
	"language-exchange-bot/internal/localization"
	"language-exchange-bot/internal/models"
)

type BotService struct {
	DB        *database.DB
	Localizer *localization.Localizer
}

func NewBotService(db *database.DB) *BotService {
	return &BotService{
		DB:        db,
		Localizer: localization.NewLocalizer(db.GetConnection()),
	}
}

func (s *BotService) DetectLanguage(telegramLangCode string) string {
	switch telegramLangCode {
	case "ru", "ru-RU":
		return "ru"
	case "es", "es-ES", "es-MX":
		return "es"
	case "zh", "zh-CN", "zh-TW":
		return "zh"
	default:
		return "en"
	}
}

func (s *BotService) HandleUserRegistration(telegramID int64, username, firstName, telegramLangCode string) (*models.User, error) {
	user, err := s.DB.FindOrCreateUser(telegramID, username, firstName)
	if err != nil {
		return nil, err
	}

	detected := s.DetectLanguage(telegramLangCode)
	// Переопределяем язык на первом визите или если он пуст/дефолтен
	if user.Status == models.StatusNew || user.InterfaceLanguageCode == "" ||
		(user.InterfaceLanguageCode == "en" && detected != "" && detected != "en") {
		user.InterfaceLanguageCode = detected
		_ = s.DB.UpdateUserInterfaceLanguage(user.ID, detected)
	}
	return user, nil
}

func (s *BotService) GetWelcomeMessage(user *models.User) string {
	return s.Localizer.GetWithParams(user.InterfaceLanguageCode, "welcome_message", map[string]string{
		"name": user.FirstName,
	})
}

func (s *BotService) GetLanguagePrompt(user *models.User, promptType string) string {
	key := "choose_native_language"
	if promptType == "target" {
		key = "choose_target_language"
	}
	return s.Localizer.Get(user.InterfaceLanguageCode, key)
}

func (s *BotService) GetLocalizedLanguageName(langCode, interfaceLangCode string) string {
	return s.Localizer.GetLanguageName(langCode, interfaceLangCode)
}

func (s *BotService) GetLocalizedInterests(langCode string) (map[int]string, error) {
	return s.Localizer.GetInterests(langCode)
}
