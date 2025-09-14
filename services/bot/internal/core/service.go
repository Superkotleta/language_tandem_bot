package core

import (
	"fmt"
	"language-exchange-bot/internal/database"
	"language-exchange-bot/internal/localization"
	"language-exchange-bot/internal/models"
	"strings"
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
	// Определяем начальный язык интерфейса только для новых пользователей
	if user.Status == models.StatusNew || user.InterfaceLanguageCode == "" {
		// Для новых пользователей устанавливаем язык интерфейса по настройкам Telegram
		// Если язык не определен, используем русский как дефолт для проекта
		if detected == "" {
			user.InterfaceLanguageCode = "ru"
		} else {
			user.InterfaceLanguageCode = detected
		}
		_ = s.DB.UpdateUserInterfaceLanguage(user.ID, user.InterfaceLanguageCode)
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

// IsProfileCompleted проверяет наличие языков и хотя бы одного интереса.
func (s *BotService) IsProfileCompleted(user *models.User) (bool, error) {
	if user.NativeLanguageCode == "" || user.TargetLanguageCode == "" {
		return false, nil
	}
	ids, err := s.DB.GetUserSelectedInterests(user.ID)
	if err != nil {
		return false, err
	}
	return len(ids) > 0, nil
}

// BuildProfileSummary возвращает локализованное резюме профиля.
func (s *BotService) BuildProfileSummary(user *models.User) (string, error) {
	lang := user.InterfaceLanguageCode
	nativeName := s.Localizer.GetLanguageName(user.NativeLanguageCode, lang)
	targetName := s.Localizer.GetLanguageName(user.TargetLanguageCode, lang)

	// Определяем флаги языков
	var nativeFlag, targetFlag string
	switch user.NativeLanguageCode {
	case "ru":
		nativeFlag = "🇷🇺"
	case "en":
		nativeFlag = "🇺🇸"
	case "es":
		nativeFlag = "🇪🇸"
	case "zh":
		nativeFlag = "🇨🇳"
	default:
		nativeFlag = "🌍"
	}

	switch user.TargetLanguageCode {
	case "ru":
		targetFlag = "🇷🇺"
	case "en":
		targetFlag = "🇺🇸"
	case "es":
		targetFlag = "🇪🇸"
	case "zh":
		targetFlag = "🇨🇳"
	default:
		targetFlag = "🌍"
	}

	ids, err := s.DB.GetUserSelectedInterests(user.ID)
	if err != nil {
		ids = []int{}
	}
	allInterests, _ := s.Localizer.GetInterests(lang)

	var picked []string
	for _, id := range ids {
		if name, ok := allInterests[id]; ok {
			picked = append(picked, name)
		}
	}
	interestsLine := fmt.Sprintf("🎯 %s: %d", s.Localizer.Get(lang, "profile_field_interests"), len(picked))
	if len(picked) > 0 {
		interestsLine = fmt.Sprintf("🎯 %s: %d\n• %s", s.Localizer.Get(lang, "profile_field_interests"), len(picked), strings.Join(picked, ", "))
	}

	title := s.Localizer.Get(lang, "profile_summary_title")
	native := fmt.Sprintf("%s %s: %s", nativeFlag, s.Localizer.Get(lang, "profile_field_native"), nativeName)
	target := fmt.Sprintf("%s %s: %s", targetFlag, s.Localizer.Get(lang, "profile_field_target"), targetName)

	return fmt.Sprintf("%s\n\n%s\n%s\n%s", title, native, target, interestsLine), nil
}
