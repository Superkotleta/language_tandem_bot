package handlers

import (
	"fmt"
	"language-exchange-bot/internal/core"
	"language-exchange-bot/internal/models"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// LanguageHandler интерфейс для обработки language операций
type LanguageHandler interface {
	HandleLanguagesContinueFilling(callback *tgbotapi.CallbackQuery, user *models.User) error
	HandleLanguagesReselect(callback *tgbotapi.CallbackQuery, user *models.User) error
	HandleLanguageLevelSelection(callback *tgbotapi.CallbackQuery, user *models.User, levelCode string) error
	HandleNativeLanguageCallback(callback *tgbotapi.CallbackQuery, user *models.User) error
	HandleTargetLanguageCallback(callback *tgbotapi.CallbackQuery, user *models.User) error
	HandleInterfaceLanguageSelection(callback *tgbotapi.CallbackQuery, user *models.User, langCode string) error
}

// LanguageHandlerImpl реализация обработчика language операций
type LanguageHandlerImpl struct {
	service         *core.BotService
	bot             *tgbotapi.BotAPI
	keyboardBuilder *KeyboardBuilder
}

// NewLanguageHandler создает новый обработчик language операций
func NewLanguageHandler(service *core.BotService, bot *tgbotapi.BotAPI, keyboardBuilder *KeyboardBuilder) LanguageHandler {
	return &LanguageHandlerImpl{
		service:         service,
		bot:             bot,
		keyboardBuilder: keyboardBuilder,
	}
}

// sendMessage отправляет сообщение пользователю
func (lh *LanguageHandlerImpl) sendMessage(chatID int64, text string) error {
	msg := tgbotapi.NewMessage(chatID, text)
	_, err := lh.bot.Send(msg)
	return err
}

// HandleLanguagesContinueFilling продолжает заполнение профиля после выбора языков
func (lh *LanguageHandlerImpl) HandleLanguagesContinueFilling(callback *tgbotapi.CallbackQuery, user *models.User) error {
	// Очищаем старые интересы при переходе к выбору интересов
	err := lh.service.DB.ClearUserInterests(user.ID)
	if err != nil {
		log.Printf("Warning: could not clear user interests: %v", err)
	}

	// Предлагаем выбрать уровень владения языком
	langName := lh.service.Localizer.GetLanguageName(user.TargetLanguageCode, user.InterfaceLanguageCode)
	title := lh.service.Localizer.GetWithParams(user.InterfaceLanguageCode, "choose_level_title", map[string]string{
		"language": langName,
	})

	keyboard := lh.keyboardBuilder.CreateLanguageLevelKeyboardWithPrefix(user.InterfaceLanguageCode, user.TargetLanguageCode, "level_", true)
	editMsg := tgbotapi.NewEditMessageTextAndMarkup(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		title,
		keyboard,
	)
	_, err = lh.bot.Request(editMsg)
	return err
}

// HandleLanguagesReselect обрабатывает повторный выбор языков
func (lh *LanguageHandlerImpl) HandleLanguagesReselect(callback *tgbotapi.CallbackQuery, user *models.User) error {
	// Сбрасываем выбор языков
	user.NativeLanguageCode = ""
	user.TargetLanguageCode = ""
	user.TargetLanguageLevel = ""

	// Обновляем статус пользователя
	_ = lh.service.DB.UpdateUserNativeLanguage(user.ID, "")
	_ = lh.service.DB.UpdateUserTargetLanguage(user.ID, "")
	_ = lh.service.DB.UpdateUserTargetLanguageLevel(user.ID, "")

	// Предлагаем выбрать родной язык снова
	text := lh.service.Localizer.Get(user.InterfaceLanguageCode, "choose_native_language")
	keyboard := lh.keyboardBuilder.CreateLanguageKeyboard(user.InterfaceLanguageCode, "native", "", true)
	editMsg := tgbotapi.NewEditMessageTextAndMarkup(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		text,
		keyboard,
	)
	_, err := lh.bot.Request(editMsg)
	return err
}

// HandleLanguageLevelSelection обрабатывает выбор уровня владения языком
func (lh *LanguageHandlerImpl) HandleLanguageLevelSelection(callback *tgbotapi.CallbackQuery, user *models.User, levelCode string) error {
	// Сохраняем уровень владения языком
	err := lh.service.DB.UpdateUserTargetLanguageLevel(user.ID, levelCode)
	if err != nil {
		return err
	}
	user.TargetLanguageLevel = levelCode

	// Показываем подтверждение и переходим к выбору интересов
	levelName := lh.service.Localizer.Get(user.InterfaceLanguageCode, "choose_level_"+levelCode)
	confirmMsg := "🎯 " + levelName + "\n\n" + lh.service.Localizer.Get(user.InterfaceLanguageCode, "choose_interests")

	interests, err := lh.service.GetCachedInterests(user.InterfaceLanguageCode)
	if err != nil {
		return err
	}

	keyboard := lh.keyboardBuilder.CreateInterestsKeyboard(interests, []int{}, user.InterfaceLanguageCode)
	editMsg := tgbotapi.NewEditMessageTextAndMarkup(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		confirmMsg,
		keyboard,
	)
	_, err = lh.bot.Request(editMsg)
	return err
}

// HandleNativeLanguageCallback обрабатывает выбор родного языка
func (lh *LanguageHandlerImpl) HandleNativeLanguageCallback(callback *tgbotapi.CallbackQuery, user *models.User) error {
	langCode := callback.Data[len("lang_native_"):]

	// Сохраняем родной язык
	err := lh.service.DB.UpdateUserNativeLanguage(user.ID, langCode)
	if err != nil {
		return err
	}
	user.NativeLanguageCode = langCode

	// Обновляем статус пользователя
	lh.service.DB.UpdateUserState(user.ID, models.StateWaitingLanguage)

	// Переход к следующему шагу онбординга
	return lh.proceedToNextOnboardingStep(callback, user, langCode)
}

// proceedToNextOnboardingStep переходит к следующему шагу онбординга
func (lh *LanguageHandlerImpl) proceedToNextOnboardingStep(callback *tgbotapi.CallbackQuery, user *models.User, nativeLangCode string) error {
	if nativeLangCode == "ru" {
		// Если выбран русский как родной, предлагаем выбрать изучаемый язык
		text := lh.service.Localizer.Get(user.InterfaceLanguageCode, "choose_target_language")

		// Исключаем русский из списка изучаемых языков
		keyboard := lh.keyboardBuilder.CreateLanguageKeyboard(user.InterfaceLanguageCode, "target", "ru", true)
		editMsg := tgbotapi.NewEditMessageTextAndMarkup(callback.Message.Chat.ID, callback.Message.MessageID, text, keyboard)
		_, err := lh.bot.Request(editMsg)
		if err != nil {
			return err
		}

		// Обновляем статус для ожидания выбора изучаемого языка
		lh.service.DB.UpdateUserState(user.ID, models.StateWaitingTargetLanguage)
		return nil
	} else {
		// Для всех других языков как родных автоматически устанавливаем русский как изучаемый
		err := lh.service.DB.UpdateUserTargetLanguage(user.ID, "ru")
		if err != nil {
			return err
		}
		user.TargetLanguageCode = "ru"

		// Получаем название выбранного языка для сообщения
		nativeLangName := lh.service.Localizer.GetLanguageName(nativeLangCode, user.InterfaceLanguageCode)

		// Показываем сообщение о том, что русский язык установлен автоматически
		targetExplanation := lh.service.Localizer.GetWithParams(user.InterfaceLanguageCode, "target_language_explanation", map[string]string{
			"native_lang": nativeLangName,
		})

		// Предлагаем выбрать уровень владения русским языком
		langName := lh.service.Localizer.GetLanguageName(user.TargetLanguageCode, user.InterfaceLanguageCode)
		levelTitle := targetExplanation + "\n\n" + lh.service.Localizer.GetWithParams(user.InterfaceLanguageCode, "choose_level_title", map[string]string{
			"language": langName,
		})

		keyboard := lh.keyboardBuilder.CreateLanguageLevelKeyboardWithPrefix(user.InterfaceLanguageCode, user.TargetLanguageCode, "level_", true)
		editMsg := tgbotapi.NewEditMessageTextAndMarkup(
			callback.Message.Chat.ID,
			callback.Message.MessageID,
			levelTitle,
			keyboard,
		)
		_, err = lh.bot.Request(editMsg)
		if err != nil {
			return err
		}

		// Обновляем статус для ожидания выбора уровня
		lh.service.DB.UpdateUserState(user.ID, models.StateWaitingLanguageLevel)
		return nil
	}
}

// HandleTargetLanguageCallback обрабатывает выбор изучаемого языка
func (lh *LanguageHandlerImpl) HandleTargetLanguageCallback(callback *tgbotapi.CallbackQuery, user *models.User) error {
	langCode := callback.Data[len("lang_target_"):]
	err := lh.service.DB.UpdateUserTargetLanguage(user.ID, langCode)
	if err != nil {
		return err
	}

	// ✅ ОЧИЩАЕМ СТАРЫЕ ИНТЕРЕСЫ при переходе к выбору интересов
	err = lh.service.DB.ClearUserInterests(user.ID)
	if err != nil {
		log.Printf("Warning: could not clear user interests: %v", err)
	}

	user.TargetLanguageCode = langCode
	langName := lh.service.Localizer.GetLanguageName(langCode, user.InterfaceLanguageCode)

	// Предлагаем выбрать уровень владения языком
	title := lh.service.Localizer.GetWithParams(user.InterfaceLanguageCode, "choose_level_title", map[string]string{
		"language": langName,
	})

	keyboard := lh.keyboardBuilder.CreateLanguageLevelKeyboardWithPrefix(user.InterfaceLanguageCode, langCode, "level_", true)
	editMsg := tgbotapi.NewEditMessageTextAndMarkup(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		title,
		keyboard,
	)
	_, err = lh.bot.Request(editMsg)
	return err
}

// HandleInterfaceLanguageSelection обрабатывает выбор языка интерфейса
func (lh *LanguageHandlerImpl) HandleInterfaceLanguageSelection(callback *tgbotapi.CallbackQuery, user *models.User, langCode string) error {
	if err := lh.service.DB.UpdateUserInterfaceLanguage(user.ID, langCode); err != nil {
		log.Printf("Error updating interface language: %v", err)
		return err
	}

	// Обновляем язык интерфейса пользователя и получаем новое сообщение
	user.InterfaceLanguageCode = langCode
	langName := lh.service.Localizer.GetLanguageName(langCode, langCode)
	text := fmt.Sprintf("%s\n\n%s: %s",
		lh.service.Localizer.Get(langCode, "choose_interface_language"),
		lh.service.Localizer.Get(langCode, "language_updated"),
		langName,
	)

	// Создаем клавиатуру с языками интерфейса (остальные кнопки остаются)
	keyboard := lh.keyboardBuilder.CreateLanguageKeyboard(langCode, "interface", "", true)

	// Редактируем сообщение, сохраняя клавиатуру
	editMsg := tgbotapi.NewEditMessageTextAndMarkup(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		text,
		keyboard,
	)
	_, err := lh.bot.Request(editMsg)
	return err
}
