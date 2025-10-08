package handlers

import (
	"fmt"
	"log"

	"language-exchange-bot/internal/models"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// LanguageHandler интерфейс для обработки language операций.
type LanguageHandler interface {
	HandleLanguagesContinueFilling(callback *tgbotapi.CallbackQuery, user *models.User) error
	HandleLanguagesReselect(callback *tgbotapi.CallbackQuery, user *models.User) error
	HandleLanguageLevelSelection(callback *tgbotapi.CallbackQuery, user *models.User, levelCode string) error
	HandleBackToLanguageLevel(callback *tgbotapi.CallbackQuery, user *models.User) error
	HandleNativeLanguageCallback(callback *tgbotapi.CallbackQuery, user *models.User) error
	HandleTargetLanguageCallback(callback *tgbotapi.CallbackQuery, user *models.User) error
	HandleInterfaceLanguageSelection(callback *tgbotapi.CallbackQuery, user *models.User, langCode string) error
}

// LanguageHandlerImpl реализация обработчика language операций.
type LanguageHandlerImpl struct {
	base *BaseHandler
}

// NewLanguageHandler создает новый обработчик language операций.
func NewLanguageHandler(base *BaseHandler) *LanguageHandlerImpl {
	return &LanguageHandlerImpl{
		base: base,
	}
}

// HandleLanguagesContinueFilling продолжает заполнение профиля после выбора языков.
func (lh *LanguageHandlerImpl) HandleLanguagesContinueFilling(callback *tgbotapi.CallbackQuery, user *models.User) error {
	// Очищаем старые интересы при переходе к выбору интересов
	err := lh.base.service.DB.ClearUserInterests(user.ID)
	if err != nil {
		lh.base.service.LoggingService.Database().WarnWithContext(
			"Could not clear user interests",
			generateRequestID("HandleLanguagesContinueFilling"),
			int64(user.ID),
			callback.Message.Chat.ID,
			"HandleLanguagesContinueFilling",
			map[string]interface{}{"userID": user.ID, "error": err.Error()},
		)
	}

	// Предлагаем выбрать уровень владения языком
	langName := lh.base.service.Localizer.GetLanguageName(user.TargetLanguageCode, user.InterfaceLanguageCode)
	title := lh.base.service.Localizer.GetWithParams(user.InterfaceLanguageCode, "choose_level_title", map[string]string{
		"language": langName,
	})

	keyboard := lh.base.keyboardBuilder.CreateLanguageLevelKeyboardWithPrefix(user.InterfaceLanguageCode, user.TargetLanguageCode, "level_", true)
	editMsg := tgbotapi.NewEditMessageTextAndMarkup(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		title,
		keyboard,
	)
	_, err = lh.base.bot.Request(editMsg)

	return err
}

// HandleLanguagesReselect обрабатывает повторный выбор языков.
func (lh *LanguageHandlerImpl) HandleLanguagesReselect(callback *tgbotapi.CallbackQuery, user *models.User) error {
	// Сбрасываем выбор языков
	user.NativeLanguageCode = ""
	user.TargetLanguageCode = ""
	user.TargetLanguageLevel = ""

	// Обновляем статус пользователя
	if err := lh.base.service.DB.UpdateUserNativeLanguage(user.ID, ""); err != nil {
		log.Printf("Failed to reset native language for user %d: %v", user.ID, err)
	}
	if err := lh.base.service.DB.UpdateUserTargetLanguage(user.ID, ""); err != nil {
		log.Printf("Failed to reset target language for user %d: %v", user.ID, err)
	}
	if err := lh.base.service.DB.UpdateUserTargetLanguageLevel(user.ID, ""); err != nil {
		log.Printf("Failed to reset target language level for user %d: %v", user.ID, err)
	}

	// Предлагаем выбрать родной язык снова
	text := lh.base.service.Localizer.Get(user.InterfaceLanguageCode, "choose_native_language")
	keyboard := lh.base.keyboardBuilder.CreateLanguageKeyboard(user.InterfaceLanguageCode, "native", "", true)
	editMsg := tgbotapi.NewEditMessageTextAndMarkup(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		text,
		keyboard,
	)
	_, err := lh.base.bot.Request(editMsg)

	return err
}

// HandleLanguageLevelSelection обрабатывает выбор уровня владения языком.
func (lh *LanguageHandlerImpl) HandleLanguageLevelSelection(callback *tgbotapi.CallbackQuery, user *models.User, levelCode string) error {
	// Сохраняем уровень владения языком
	err := lh.base.service.DB.UpdateUserTargetLanguageLevel(user.ID, levelCode)
	if err != nil {
		return err
	}

	user.TargetLanguageLevel = levelCode

	// Переходим к новой системе интересов
	levelName := lh.base.service.Localizer.Get(user.InterfaceLanguageCode, "choose_level_"+levelCode)
	confirmMsg := "🎯 " + levelName + "\n\n" + lh.base.service.Localizer.Get(user.InterfaceLanguageCode, "interests_new_system")

	// Очищаем предыдущие выборы интересов пользователя
	err = lh.base.service.DB.ClearUserInterests(user.ID)
	if err != nil {
		lh.base.service.LoggingService.Database().WarnWithContext(
			"Could not clear user interests",
			generateRequestID("HandleLanguageLevelSelection"),
			int64(user.ID),
			callback.Message.Chat.ID,
			"HandleLanguageLevelSelection",
			map[string]interface{}{"userID": user.ID, "error": err.Error()},
		)
	}

	// Используем новую систему интересов
	keyboard := lh.base.keyboardBuilder.CreateInterestCategoriesKeyboard(user.InterfaceLanguageCode)
	editMsg := tgbotapi.NewEditMessageTextAndMarkup(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		confirmMsg,
		keyboard,
	)
	_, err = lh.base.bot.Request(editMsg)

	return err
}

// HandleNativeLanguageCallback обрабатывает выбор родного языка.
func (lh *LanguageHandlerImpl) HandleNativeLanguageCallback(callback *tgbotapi.CallbackQuery, user *models.User) error {
	langCode := callback.Data[len("lang_native_"):]

	// Сохраняем родной язык
	err := lh.base.service.DB.UpdateUserNativeLanguage(user.ID, langCode)
	if err != nil {
		return err
	}

	user.NativeLanguageCode = langCode

	// Обновляем статус пользователя
	err = lh.base.service.DB.UpdateUserState(user.ID, models.StateWaitingLanguage)
	if err != nil {
		return err
	}

	// Переход к следующему шагу онбординга
	return lh.proceedToNextOnboardingStep(callback, user, langCode)
}

// proceedToNextOnboardingStep переходит к следующему шагу онбординга.
func (lh *LanguageHandlerImpl) proceedToNextOnboardingStep(callback *tgbotapi.CallbackQuery, user *models.User, nativeLangCode string) error {
	if nativeLangCode == "ru" {
		return lh.handleRussianNativeLanguage(callback, user)
	}

	return lh.handleNonRussianNativeLanguage(callback, user, nativeLangCode)
}

// handleRussianNativeLanguage обрабатывает случай, когда русский выбран как родной язык.
func (lh *LanguageHandlerImpl) handleRussianNativeLanguage(callback *tgbotapi.CallbackQuery, user *models.User) error {
	// Если выбран русский как родной, предлагаем выбрать изучаемый язык
	text := lh.base.service.Localizer.Get(user.InterfaceLanguageCode, "choose_target_language")

	// Исключаем русский из списка изучаемых языков
	keyboard := lh.base.keyboardBuilder.CreateLanguageKeyboard(user.InterfaceLanguageCode, "target", "ru", true)
	editMsg := tgbotapi.NewEditMessageTextAndMarkup(callback.Message.Chat.ID, callback.Message.MessageID, text, keyboard)

	_, err := lh.base.bot.Request(editMsg)
	if err != nil {
		return err
	}

	// Обновляем статус для ожидания выбора изучаемого языка
	err = lh.base.service.DB.UpdateUserState(user.ID, models.StateWaitingTargetLanguage)
	if err != nil {
		return err
	}

	return nil
}

// handleNonRussianNativeLanguage обрабатывает случай, когда выбран не русский язык как родной.
func (lh *LanguageHandlerImpl) handleNonRussianNativeLanguage(callback *tgbotapi.CallbackQuery, user *models.User, nativeLangCode string) error {
	// Для всех других языков как родных автоматически устанавливаем русский как изучаемый
	err := lh.base.service.DB.UpdateUserTargetLanguage(user.ID, "ru")
	if err != nil {
		return err
	}

	user.TargetLanguageCode = "ru"

	// Получаем название выбранного языка для сообщения
	nativeLangName := lh.base.service.Localizer.GetLanguageName(nativeLangCode, user.InterfaceLanguageCode)

	// Показываем сообщение о том, что русский язык установлен автоматически
	targetExplanation := lh.base.service.Localizer.GetWithParams(
		user.InterfaceLanguageCode,
		"target_language_explanation",
		map[string]string{
			"native_lang": nativeLangName,
		},
	)

	// Предлагаем выбрать уровень владения русским языком
	langName := lh.base.service.Localizer.GetLanguageName(user.TargetLanguageCode, user.InterfaceLanguageCode)
	levelTitle := targetExplanation + "\n\n" + lh.base.service.Localizer.GetWithParams(user.InterfaceLanguageCode, "choose_level_title", map[string]string{
		"language": langName,
	})

	keyboard := lh.base.keyboardBuilder.CreateLanguageLevelKeyboardWithPrefix(user.InterfaceLanguageCode, user.TargetLanguageCode, "level_", true)
	editMsg := tgbotapi.NewEditMessageTextAndMarkup(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		levelTitle,
		keyboard,
	)

	_, err = lh.base.bot.Request(editMsg)
	if err != nil {
		return err
	}

	// Обновляем статус для ожидания выбора уровня
	err = lh.base.service.DB.UpdateUserState(user.ID, models.StateWaitingLanguageLevel)
	if err != nil {
		return err
	}

	return nil
}

// HandleTargetLanguageCallback обрабатывает выбор изучаемого языка.
func (lh *LanguageHandlerImpl) HandleTargetLanguageCallback(callback *tgbotapi.CallbackQuery, user *models.User) error {
	langCode := callback.Data[len("lang_target_"):]

	err := lh.base.service.DB.UpdateUserTargetLanguage(user.ID, langCode)
	if err != nil {
		return err
	}

	// ✅ ОЧИЩАЕМ СТАРЫЕ ИНТЕРЕСЫ при переходе к выбору интересов
	err = lh.base.service.DB.ClearUserInterests(user.ID)
	if err != nil {
		lh.base.service.LoggingService.Database().WarnWithContext(
			"Could not clear user interests",
			generateRequestID("HandleTargetLanguageCallback"),
			int64(user.ID),
			callback.Message.Chat.ID,
			"HandleTargetLanguageCallback",
			map[string]interface{}{"userID": user.ID, "error": err.Error()},
		)
	}

	user.TargetLanguageCode = langCode
	langName := lh.base.service.Localizer.GetLanguageName(langCode, user.InterfaceLanguageCode)

	// Предлагаем выбрать уровень владения языком
	title := lh.base.service.Localizer.GetWithParams(user.InterfaceLanguageCode, "choose_level_title", map[string]string{
		"language": langName,
	})

	keyboard := lh.base.keyboardBuilder.CreateLanguageLevelKeyboardWithPrefix(user.InterfaceLanguageCode, langCode, "level_", true)
	editMsg := tgbotapi.NewEditMessageTextAndMarkup(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		title,
		keyboard,
	)
	_, err = lh.base.bot.Request(editMsg)

	return err
}

// HandleInterfaceLanguageSelection обрабатывает выбор языка интерфейса.
func (lh *LanguageHandlerImpl) HandleInterfaceLanguageSelection(callback *tgbotapi.CallbackQuery, user *models.User, langCode string) error {
	if err := lh.base.service.DB.UpdateUserInterfaceLanguage(user.ID, langCode); err != nil {
		lh.base.service.LoggingService.Database().ErrorWithContext(
			"Error updating interface language",
			generateRequestID("HandleInterfaceLanguageSelection"),
			int64(user.ID),
			callback.Message.Chat.ID,
			"HandleInterfaceLanguageSelection",
			map[string]interface{}{"userID": user.ID, "langCode": langCode, "error": err.Error()},
		)

		return err
	}

	// Обновляем язык интерфейса пользователя и получаем новое сообщение
	user.InterfaceLanguageCode = langCode
	langName := lh.base.service.Localizer.GetLanguageName(langCode, langCode)
	text := fmt.Sprintf("%s\n\n%s: %s",
		lh.base.service.Localizer.Get(langCode, "choose_interface_language"),
		lh.base.service.Localizer.Get(langCode, "language_updated"),
		langName,
	)

	// Создаем клавиатуру с языками интерфейса (остальные кнопки остаются)
	keyboard := lh.base.keyboardBuilder.CreateLanguageKeyboard(langCode, "interface", "", true)

	// Редактируем сообщение, сохраняя клавиатуру
	editMsg := tgbotapi.NewEditMessageTextAndMarkup(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		text,
		keyboard,
	)
	_, err := lh.base.bot.Request(editMsg)

	return err
}

// HandleBackToLanguageLevel возвращает к выбору уровня языка.
func (lh *LanguageHandlerImpl) HandleBackToLanguageLevel(callback *tgbotapi.CallbackQuery, user *models.User) error {
	// Предлагаем выбрать уровень владения языком
	langName := lh.base.service.Localizer.GetLanguageName(user.TargetLanguageCode, user.InterfaceLanguageCode)
	title := lh.base.service.Localizer.GetWithParams(user.InterfaceLanguageCode, "choose_level_title", map[string]string{
		"language": langName,
	})

	keyboard := lh.base.keyboardBuilder.CreateLanguageLevelKeyboardWithPrefix(user.InterfaceLanguageCode, user.TargetLanguageCode, "level_", true)
	editMsg := tgbotapi.NewEditMessageTextAndMarkup(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		title,
		keyboard,
	)
	_, err := lh.base.bot.Request(editMsg)

	return err
}
