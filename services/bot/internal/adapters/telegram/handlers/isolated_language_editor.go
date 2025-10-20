package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"language-exchange-bot/internal/models"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// IsolatedLanguageEditor управляет изолированным редактированием языковых настроек
type IsolatedLanguageEditor struct {
	baseHandler *BaseHandler
}

// LanguageEditSession представляет сессию редактирования языков
type LanguageEditSession struct {
	UserID               int              `json:"user_id"`
	OriginalNativeLang   string           `json:"original_native_lang"`
	OriginalTargetLang   string           `json:"original_target_lang"`
	OriginalTargetLevel  string           `json:"original_target_level"`
	CurrentNativeLang    string           `json:"current_native_lang"`
	CurrentTargetLang    string           `json:"current_target_lang"`
	CurrentTargetLevel   string           `json:"current_target_level"`
	Changes              []LanguageChange `json:"changes"`
	CurrentStep          string           `json:"current_step"` // "native", "target", "level", "preview"
	SessionStart         time.Time        `json:"session_start"`
	LastActivity         time.Time        `json:"last_activity"`
	InterfaceLanguage    string           `json:"interface_language"`
}

// LanguageChange представляет изменение языковых настроек
type LanguageChange struct {
	Field     string      `json:"field"` // "native_language", "target_language", "target_level"
	OldValue  interface{} `json:"old_value"`
	NewValue  interface{} `json:"new_value"`
	Timestamp time.Time   `json:"timestamp"`
}

// NewIsolatedLanguageEditor создает новый изолированный редактор языков
func NewIsolatedLanguageEditor(baseHandler *BaseHandler) *IsolatedLanguageEditor {
	return &IsolatedLanguageEditor{
		baseHandler: baseHandler,
	}
}

// =============================================================================
// ОСНОВНЫЕ МЕТОДЫ УПРАВЛЕНИЯ СЕССИЯМИ
// =============================================================================

// StartEditSession начинает сессию редактирования языков
func (e *IsolatedLanguageEditor) StartEditSession(callback *tgbotapi.CallbackQuery, user *models.User) error {
	loggingService := e.baseHandler.service.LoggingService
	requestID := generateRequestID("StartLanguageEditSession")

	loggingService.LogRequestStart(requestID, int64(user.ID), callback.Message.Chat.ID, "StartLanguageEditSession")
	loggingService.Telegram().InfoWithContext(
		"Starting language edit session",
		requestID,
		int64(user.ID),
		callback.Message.Chat.ID,
		"StartLanguageEditSession",
		map[string]interface{}{
			"user_id":            user.ID,
			"native_lang":        user.NativeLanguageCode,
			"target_lang":        user.TargetLanguageCode,
			"target_level":       user.TargetLanguageLevel,
			"interface_language": user.InterfaceLanguageCode,
		},
	)

	// Создаем сессию редактирования
	session := &LanguageEditSession{
		UserID:               user.ID,
		OriginalNativeLang:   user.NativeLanguageCode,
		OriginalTargetLang:   user.TargetLanguageCode,
		OriginalTargetLevel:  user.TargetLanguageLevel,
		CurrentNativeLang:    user.NativeLanguageCode,
		CurrentTargetLang:    user.TargetLanguageCode,
		CurrentTargetLevel:   user.TargetLanguageLevel,
		Changes:              []LanguageChange{},
		CurrentStep:          "main_menu",
		SessionStart:         time.Now(),
		LastActivity:         time.Now(),
		InterfaceLanguage:    user.InterfaceLanguageCode,
	}

	// Сохраняем сессию в Redis
	if err := e.saveSession(user.ID, session); err != nil {
		loggingService.Telegram().ErrorWithContext(
			"Failed to save language edit session",
			requestID,
			int64(user.ID),
			callback.Message.Chat.ID,
			"StartLanguageEditSession",
			map[string]interface{}{
				"user_id": user.ID,
				"error":   err.Error(),
			},
		)
		return fmt.Errorf("failed to save session: %w", err)
	}

	loggingService.Telegram().InfoWithContext(
		"Language edit session started successfully",
		requestID,
		int64(user.ID),
		callback.Message.Chat.ID,
		"StartLanguageEditSession",
		map[string]interface{}{
			"user_id":      user.ID,
			"session_step": session.CurrentStep,
		},
	)

	// Показываем главное меню редактирования
	return e.showMainEditMenu(callback, user, session)
}

// showMainEditMenu показывает главное меню редактирования языков
func (e *IsolatedLanguageEditor) showMainEditMenu(callback *tgbotapi.CallbackQuery, user *models.User, session *LanguageEditSession) error {
	text := e.buildLanguageSettingsText(user, session)
	keyboard := e.createEditMainMenuKeyboard(user.InterfaceLanguageCode, session)

	return e.baseHandler.messageFactory.EditWithKeyboard(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		text,
		&keyboard,
	)
}

// buildLanguageSettingsText формирует текст с текущими настройками языков
func (e *IsolatedLanguageEditor) buildLanguageSettingsText(user *models.User, session *LanguageEditSession) string {
	localizer := e.baseHandler.service.Localizer
	lang := user.InterfaceLanguageCode

	// Заголовок
	text := localizer.Get(lang, "profile_edit_languages") + "\n\n"

	// Родной язык
	nativeLangName := localizer.GetLanguageName(session.CurrentNativeLang, lang)
	text += "🏠 " + localizer.Get(lang, "native_language") + ": " + nativeLangName

	if session.CurrentNativeLang != session.OriginalNativeLang {
		text += " ✏️"
	}
	text += "\n"

	// Изучаемый язык
	targetLangName := localizer.GetLanguageName(session.CurrentTargetLang, lang)
	text += "📚 " + localizer.Get(lang, "target_language") + ": " + targetLangName

	if session.CurrentTargetLang != session.OriginalTargetLang {
		text += " ✏️"
	}
	text += "\n"

	// Уровень владения
	levelName := localizer.Get(lang, "level_"+session.CurrentTargetLevel)
	text += "📊 " + localizer.Get(lang, "language_level") + ": " + levelName

	if session.CurrentTargetLevel != session.OriginalTargetLevel {
		text += " ✏️"
	}
	text += "\n"

	// Количество изменений
	if len(session.Changes) > 0 {
		text += "\n✨ " + localizer.Get(lang, "changes_made") + ": " + fmt.Sprintf("%d", len(session.Changes))
	}

	return text
}

// createEditMainMenuKeyboard создает клавиатуру главного меню редактирования
func (e *IsolatedLanguageEditor) createEditMainMenuKeyboard(interfaceLang string, session *LanguageEditSession) tgbotapi.InlineKeyboardMarkup {
	localizer := e.baseHandler.service.Localizer

	buttonRows := [][]tgbotapi.InlineKeyboardButton{
		// Редактирование родного языка
		{
			tgbotapi.NewInlineKeyboardButtonData(
				"🏠 "+localizer.Get(interfaceLang, "edit_native_language"),
				"isolated_lang_edit_native",
			),
		},
		// Редактирование изучаемого языка
		{
			tgbotapi.NewInlineKeyboardButtonData(
				"📚 "+localizer.Get(interfaceLang, "edit_target_language"),
				"isolated_lang_edit_target",
			),
		},
		// Редактирование уровня
		{
			tgbotapi.NewInlineKeyboardButtonData(
				"📊 "+localizer.Get(interfaceLang, "edit_language_level"),
				"isolated_lang_edit_level",
			),
		},
	}

	// Если есть изменения, показываем кнопку предпросмотра
	if len(session.Changes) > 0 {
		buttonRows = append(buttonRows, []tgbotapi.InlineKeyboardButton{
			tgbotapi.NewInlineKeyboardButtonData(
				"👁️ "+localizer.Get(interfaceLang, "preview_changes"),
				"isolated_lang_preview",
			),
		})
	}

	// Кнопки управления - используем специфичные для языкового редактора
	controlRow := []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData(
			"💾 "+localizer.Get(interfaceLang, "save_changes"),
			"isolated_lang_save",
		),
		tgbotapi.NewInlineKeyboardButtonData(
			"❌ "+localizer.Get(interfaceLang, "cancel_edit"),
			"isolated_lang_cancel",
		),
	}
	buttonRows = append(buttonRows, controlRow)

	return tgbotapi.NewInlineKeyboardMarkup(buttonRows...)
}

// =============================================================================
// РЕДАКТИРОВАНИЕ РОДНОГО ЯЗЫКА
// =============================================================================

// HandleEditNativeLanguage обрабатывает начало редактирования родного языка
func (e *IsolatedLanguageEditor) HandleEditNativeLanguage(callback *tgbotapi.CallbackQuery, user *models.User) error {
	session, err := e.getSession(user.ID)
	if err != nil {
		return e.StartEditSession(callback, user)
	}

	session.CurrentStep = "native"
	session.LastActivity = time.Now()

	if err := e.saveSession(user.ID, session); err != nil {
		return err
	}

	text := e.baseHandler.service.Localizer.Get(user.InterfaceLanguageCode, "choose_native_language")
	keyboard := e.createNativeLanguageKeyboard(user.InterfaceLanguageCode, session)

	return e.baseHandler.messageFactory.EditWithKeyboard(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		text,
		&keyboard,
	)
}

// createNativeLanguageKeyboard создает клавиатуру выбора родного языка
func (e *IsolatedLanguageEditor) createNativeLanguageKeyboard(interfaceLang string, session *LanguageEditSession) tgbotapi.InlineKeyboardMarkup {
	kb := e.baseHandler.keyboardBuilder

	// Получаем базовую клавиатуру с языками
	keyboard := kb.CreateLanguageKeyboard(interfaceLang, "isolated_native", "", false)

	// Добавляем кнопки навигации
	backRow := []tgbotapi.InlineKeyboardButton{
		kb.CreateBackButton(interfaceLang, "isolated_lang_back_to_menu"),
	}

	keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, backRow)

	return keyboard
}

// HandleNativeLanguageSelection обрабатывает выбор родного языка
func (e *IsolatedLanguageEditor) HandleNativeLanguageSelection(callback *tgbotapi.CallbackQuery, user *models.User, langCode string) error {
	session, err := e.getSession(user.ID)
	if err != nil {
		return e.StartEditSession(callback, user)
	}

	// Записываем изменение
	if session.CurrentNativeLang != langCode {
		change := LanguageChange{
			Field:     "native_language",
			OldValue:  session.CurrentNativeLang,
			NewValue:  langCode,
			Timestamp: time.Now(),
		}
		session.Changes = append(session.Changes, change)
		session.CurrentNativeLang = langCode

		// Если выбран не русский, автоматически устанавливаем русский как изучаемый
		if langCode != "ru" && session.CurrentTargetLang != "ru" {
			targetChange := LanguageChange{
				Field:     "target_language",
				OldValue:  session.CurrentTargetLang,
				NewValue:  "ru",
				Timestamp: time.Now(),
			}
			session.Changes = append(session.Changes, targetChange)
			session.CurrentTargetLang = "ru"
		}

		// Если выбран русский и текущий изучаемый - тоже русский, сбрасываем изучаемый
		if langCode == "ru" && session.CurrentTargetLang == "ru" {
			targetChange := LanguageChange{
				Field:     "target_language",
				OldValue:  session.CurrentTargetLang,
				NewValue:  "",
				Timestamp: time.Now(),
			}
			session.Changes = append(session.Changes, targetChange)
			session.CurrentTargetLang = ""
		}
	}

	session.CurrentStep = "main_menu"
	session.LastActivity = time.Now()

	if err := e.saveSession(user.ID, session); err != nil {
		return err
	}

	// Возвращаемся в главное меню
	return e.showMainEditMenu(callback, user, session)
}

// =============================================================================
// РЕДАКТИРОВАНИЕ ИЗУЧАЕМОГО ЯЗЫКА
// =============================================================================

// HandleEditTargetLanguage обрабатывает начало редактирования изучаемого языка
func (e *IsolatedLanguageEditor) HandleEditTargetLanguage(callback *tgbotapi.CallbackQuery, user *models.User) error {
	session, err := e.getSession(user.ID)
	if err != nil {
		return e.StartEditSession(callback, user)
	}

	// Проверяем, что родной язык - русский
	if session.CurrentNativeLang != "ru" {
		text := e.baseHandler.service.Localizer.Get(user.InterfaceLanguageCode, "target_language_locked")
		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			[]tgbotapi.InlineKeyboardButton{
				e.baseHandler.keyboardBuilder.CreateBackButton(user.InterfaceLanguageCode, "isolated_lang_back_to_menu"),
			},
		)
		return e.baseHandler.messageFactory.EditWithKeyboard(
			callback.Message.Chat.ID,
			callback.Message.MessageID,
			text,
			&keyboard,
		)
	}

	session.CurrentStep = "target"
	session.LastActivity = time.Now()

	if err := e.saveSession(user.ID, session); err != nil {
		return err
	}

	text := e.baseHandler.service.Localizer.Get(user.InterfaceLanguageCode, "choose_target_language")
	keyboard := e.createTargetLanguageKeyboard(user.InterfaceLanguageCode, session)

	return e.baseHandler.messageFactory.EditWithKeyboard(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		text,
		&keyboard,
	)
}

// createTargetLanguageKeyboard создает клавиатуру выбора изучаемого языка
func (e *IsolatedLanguageEditor) createTargetLanguageKeyboard(interfaceLang string, session *LanguageEditSession) tgbotapi.InlineKeyboardMarkup {
	kb := e.baseHandler.keyboardBuilder

	// Исключаем родной язык из списка
	keyboard := kb.CreateLanguageKeyboard(interfaceLang, "isolated_target", session.CurrentNativeLang, false)

	// Добавляем кнопки навигации
	backRow := []tgbotapi.InlineKeyboardButton{
		kb.CreateBackButton(interfaceLang, "isolated_lang_back_to_menu"),
	}

	keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, backRow)

	return keyboard
}

// HandleTargetLanguageSelection обрабатывает выбор изучаемого языка
func (e *IsolatedLanguageEditor) HandleTargetLanguageSelection(callback *tgbotapi.CallbackQuery, user *models.User, langCode string) error {
	session, err := e.getSession(user.ID)
	if err != nil {
		return e.StartEditSession(callback, user)
	}

	// Записываем изменение
	if session.CurrentTargetLang != langCode {
		change := LanguageChange{
			Field:     "target_language",
			OldValue:  session.CurrentTargetLang,
			NewValue:  langCode,
			Timestamp: time.Now(),
		}
		session.Changes = append(session.Changes, change)
		session.CurrentTargetLang = langCode
	}

	session.CurrentStep = "level"
	session.LastActivity = time.Now()

	if err := e.saveSession(user.ID, session); err != nil {
		return err
	}

	// Автоматически переходим к выбору уровня
	return e.showLevelSelection(callback, user, session)
}

// =============================================================================
// РЕДАКТИРОВАНИЕ УРОВНЯ ВЛАДЕНИЯ
// =============================================================================

// HandleEditLanguageLevel обрабатывает начало редактирования уровня
func (e *IsolatedLanguageEditor) HandleEditLanguageLevel(callback *tgbotapi.CallbackQuery, user *models.User) error {
	session, err := e.getSession(user.ID)
	if err != nil {
		return e.StartEditSession(callback, user)
	}

	session.CurrentStep = "level"
	session.LastActivity = time.Now()

	if err := e.saveSession(user.ID, session); err != nil {
		return err
	}

	return e.showLevelSelection(callback, user, session)
}

// showLevelSelection показывает экран выбора уровня владения языком
func (e *IsolatedLanguageEditor) showLevelSelection(callback *tgbotapi.CallbackQuery, user *models.User, session *LanguageEditSession) error {
	localizer := e.baseHandler.service.Localizer
	langName := localizer.GetLanguageName(session.CurrentTargetLang, user.InterfaceLanguageCode)

	text := localizer.GetWithParams(user.InterfaceLanguageCode, "choose_level_title", map[string]string{
		"language": langName,
	})

	keyboard := e.createLevelKeyboard(user.InterfaceLanguageCode, session)

	return e.baseHandler.messageFactory.EditWithKeyboard(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		text,
		&keyboard,
	)
}

// createLevelKeyboard создает клавиатуру выбора уровня владения
func (e *IsolatedLanguageEditor) createLevelKeyboard(interfaceLang string, session *LanguageEditSession) tgbotapi.InlineKeyboardMarkup {
	kb := e.baseHandler.keyboardBuilder

	keyboard := kb.CreateLanguageLevelKeyboardWithPrefix(interfaceLang, session.CurrentTargetLang, "isolated_level_", false)

	// Добавляем кнопки навигации
	backRow := []tgbotapi.InlineKeyboardButton{
		kb.CreateBackButton(interfaceLang, "isolated_lang_back_to_menu"),
	}

	keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, backRow)

	return keyboard
}

// HandleLanguageLevelSelection обрабатывает выбор уровня владения языком
func (e *IsolatedLanguageEditor) HandleLanguageLevelSelection(callback *tgbotapi.CallbackQuery, user *models.User, levelCode string) error {
	session, err := e.getSession(user.ID)
	if err != nil {
		return e.StartEditSession(callback, user)
	}

	// Записываем изменение
	if session.CurrentTargetLevel != levelCode {
		change := LanguageChange{
			Field:     "target_level",
			OldValue:  session.CurrentTargetLevel,
			NewValue:  levelCode,
			Timestamp: time.Now(),
		}
		session.Changes = append(session.Changes, change)
		session.CurrentTargetLevel = levelCode
	}

	session.CurrentStep = "main_menu"
	session.LastActivity = time.Now()

	if err := e.saveSession(user.ID, session); err != nil {
		return err
	}

	// Возвращаемся в главное меню
	return e.showMainEditMenu(callback, user, session)
}

// =============================================================================
// ПРЕДПРОСМОТР ИЗМЕНЕНИЙ
// =============================================================================

// HandlePreviewChanges показывает предпросмотр всех изменений
func (e *IsolatedLanguageEditor) HandlePreviewChanges(callback *tgbotapi.CallbackQuery, user *models.User) error {
	session, err := e.getSession(user.ID)
	if err != nil {
		return e.StartEditSession(callback, user)
	}

	session.CurrentStep = "preview"
	session.LastActivity = time.Now()

	if err := e.saveSession(user.ID, session); err != nil {
		return err
	}

	text := e.buildChangesPreviewText(user, session)
	keyboard := e.createChangesPreviewKeyboard(user.InterfaceLanguageCode, session)

	return e.baseHandler.messageFactory.EditWithKeyboard(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		text,
		&keyboard,
	)
}

// buildChangesPreviewText формирует текст предпросмотра изменений
func (e *IsolatedLanguageEditor) buildChangesPreviewText(user *models.User, session *LanguageEditSession) string {
	localizer := e.baseHandler.service.Localizer
	lang := user.InterfaceLanguageCode

	text := "📋 " + localizer.Get(lang, "preview_changes") + "\n\n"

	// Показываем все изменения
	for i, change := range session.Changes {
		text += fmt.Sprintf("%d. ", i+1)

		switch change.Field {
		case "native_language":
			oldLang := localizer.GetLanguageName(change.OldValue.(string), lang)
			newLang := localizer.GetLanguageName(change.NewValue.(string), lang)
			text += localizer.Get(lang, "native_language") + ": " + oldLang + " → " + newLang
		case "target_language":
			oldLang := localizer.GetLanguageName(change.OldValue.(string), lang)
			newLang := localizer.GetLanguageName(change.NewValue.(string), lang)
			text += localizer.Get(lang, "target_language") + ": " + oldLang + " → " + newLang
		case "target_level":
			oldLevel := localizer.Get(lang, "level_"+change.OldValue.(string))
			newLevel := localizer.Get(lang, "level_"+change.NewValue.(string))
			text += localizer.Get(lang, "language_level") + ": " + oldLevel + " → " + newLevel
		}

		text += "\n"
	}

	text += "\n" + localizer.Get(lang, "confirm_save_changes")

	return text
}

// createChangesPreviewKeyboard создает клавиатуру предпросмотра изменений
func (e *IsolatedLanguageEditor) createChangesPreviewKeyboard(interfaceLang string, session *LanguageEditSession) tgbotapi.InlineKeyboardMarkup {
	localizer := e.baseHandler.service.Localizer
	kb := e.baseHandler.keyboardBuilder

	mainRow := []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData(
			"💾 "+localizer.Get(interfaceLang, "save_changes"),
			"isolated_lang_save",
		),
		tgbotapi.NewInlineKeyboardButtonData(
			localizer.Get(interfaceLang, "undo_last_change"),
			"isolated_lang_undo_last",
		),
	}

	// Навигация
	backRow := []tgbotapi.InlineKeyboardButton{
		kb.CreateBackButton(interfaceLang, "isolated_lang_back_to_menu"),
		tgbotapi.NewInlineKeyboardButtonData(
			"❌ "+localizer.Get(interfaceLang, "cancel_edit"),
			"isolated_lang_cancel",
		),
	}

	return tgbotapi.NewInlineKeyboardMarkup(mainRow, backRow)
}

// =============================================================================
// ОТМЕНА ПОСЛЕДНЕГО ИЗМЕНЕНИЯ
// =============================================================================

// HandleUndoLastChange отменяет последнее изменение
func (e *IsolatedLanguageEditor) HandleUndoLastChange(callback *tgbotapi.CallbackQuery, user *models.User) error {
	session, err := e.getSession(user.ID)
	if err != nil {
		return e.StartEditSession(callback, user)
	}

	if len(session.Changes) == 0 {
		// Нет изменений для отмены
		text := e.baseHandler.service.Localizer.Get(user.InterfaceLanguageCode, "no_changes_to_undo")
		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			[]tgbotapi.InlineKeyboardButton{
				e.baseHandler.keyboardBuilder.CreateBackButton(user.InterfaceLanguageCode, "isolated_lang_back_to_menu"),
			},
		)
		return e.baseHandler.messageFactory.EditWithKeyboard(
			callback.Message.Chat.ID,
			callback.Message.MessageID,
			text,
			&keyboard,
		)
	}

	// Удаляем последнее изменение
	lastChange := session.Changes[len(session.Changes)-1]
	session.Changes = session.Changes[:len(session.Changes)-1]

	// Откатываем значение
	switch lastChange.Field {
	case "native_language":
		session.CurrentNativeLang = lastChange.OldValue.(string)
	case "target_language":
		session.CurrentTargetLang = lastChange.OldValue.(string)
	case "target_level":
		session.CurrentTargetLevel = lastChange.OldValue.(string)
	}

	session.LastActivity = time.Now()

	if err := e.saveSession(user.ID, session); err != nil {
		return err
	}

	// Если мы были в предпросмотре, показываем обновленный предпросмотр
	if session.CurrentStep == "preview" {
		if len(session.Changes) == 0 {
			// Если больше нет изменений, возвращаемся в главное меню
			session.CurrentStep = "main_menu"
			if err := e.saveSession(user.ID, session); err != nil {
				return err
			}
			return e.showMainEditMenu(callback, user, session)
		}
		return e.HandlePreviewChanges(callback, user)
	}

	// Иначе возвращаемся в главное меню
	session.CurrentStep = "main_menu"
	if err := e.saveSession(user.ID, session); err != nil {
		return err
	}
	return e.showMainEditMenu(callback, user, session)
}

// =============================================================================
// СОХРАНЕНИЕ И ОТМЕНА
// =============================================================================

// HandleSaveChanges сохраняет все изменения
func (e *IsolatedLanguageEditor) HandleSaveChanges(callback *tgbotapi.CallbackQuery, user *models.User) error {
	session, err := e.getSession(user.ID)
	if err != nil {
		return e.StartEditSession(callback, user)
	}

	loggingService := e.baseHandler.service.LoggingService
	requestID := generateRequestID("SaveLanguageChanges")

	loggingService.Telegram().InfoWithContext(
		"Saving language changes",
		requestID,
		int64(user.ID),
		callback.Message.Chat.ID,
		"SaveLanguageChanges",
		map[string]interface{}{
			"user_id":       user.ID,
			"changes_count": len(session.Changes),
		},
	)

	// Применяем все изменения
	for _, change := range session.Changes {
		switch change.Field {
		case "native_language":
			if err := e.baseHandler.service.DB.UpdateUserNativeLanguage(user.ID, session.CurrentNativeLang); err != nil {
				loggingService.Telegram().ErrorWithContext(
					"Failed to update native language",
					requestID,
					int64(user.ID),
					callback.Message.Chat.ID,
					"SaveLanguageChanges",
					map[string]interface{}{
						"user_id": user.ID,
						"error":   err.Error(),
					},
				)
				return err
			}
			user.NativeLanguageCode = session.CurrentNativeLang

		case "target_language":
			if err := e.baseHandler.service.DB.UpdateUserTargetLanguage(user.ID, session.CurrentTargetLang); err != nil {
				loggingService.Telegram().ErrorWithContext(
					"Failed to update target language",
					requestID,
					int64(user.ID),
					callback.Message.Chat.ID,
					"SaveLanguageChanges",
					map[string]interface{}{
						"user_id": user.ID,
						"error":   err.Error(),
					},
				)
				return err
			}
			user.TargetLanguageCode = session.CurrentTargetLang

		case "target_level":
			if err := e.baseHandler.service.DB.UpdateUserTargetLanguageLevel(user.ID, session.CurrentTargetLevel); err != nil {
				loggingService.Telegram().ErrorWithContext(
					"Failed to update target language level",
					requestID,
					int64(user.ID),
					callback.Message.Chat.ID,
					"SaveLanguageChanges",
					map[string]interface{}{
						"user_id": user.ID,
						"error":   err.Error(),
					},
				)
				return err
			}
			user.TargetLanguageLevel = session.CurrentTargetLevel
		}
	}

	// Удаляем сессию
	if err := e.deleteSession(user.ID); err != nil {
		loggingService.Telegram().ErrorWithContext(
			"Failed to delete language edit session",
			requestID,
			int64(user.ID),
			callback.Message.Chat.ID,
			"SaveLanguageChanges",
			map[string]interface{}{
				"user_id": user.ID,
				"error":   err.Error(),
			},
		)
	}

	loggingService.Telegram().InfoWithContext(
		"Language changes saved successfully",
		requestID,
		int64(user.ID),
		callback.Message.Chat.ID,
		"SaveLanguageChanges",
		map[string]interface{}{
			"user_id":       user.ID,
			"changes_count": len(session.Changes),
		},
	)

	// Показываем сообщение об успешном сохранении и профиль
	text := e.baseHandler.service.Localizer.Get(user.InterfaceLanguageCode, "changes_saved_successfully") + "\n\n"

	summary, err := e.baseHandler.service.BuildProfileSummary(user)
	if err == nil {
		text += summary + "\n\n"
	}

	text += e.baseHandler.service.Localizer.Get(user.InterfaceLanguageCode, "profile_actions")

	keyboard := e.baseHandler.keyboardBuilder.CreateProfileMenuKeyboard(user.InterfaceLanguageCode)

	return e.baseHandler.messageFactory.EditWithKeyboard(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		text,
		&keyboard,
	)
}

// HandleCancelEdit отменяет редактирование без сохранения
func (e *IsolatedLanguageEditor) HandleCancelEdit(callback *tgbotapi.CallbackQuery, user *models.User) error {
	session, err := e.getSession(user.ID)
	if err != nil {
		// Сессия не найдена, просто возвращаемся к профилю
		return e.showProfileAfterCancel(callback, user)
	}

	loggingService := e.baseHandler.service.LoggingService
	requestID := generateRequestID("CancelLanguageEdit")

	loggingService.Telegram().InfoWithContext(
		"Cancelling language edit session",
		requestID,
		int64(user.ID),
		callback.Message.Chat.ID,
		"CancelLanguageEdit",
		map[string]interface{}{
			"user_id":       user.ID,
			"changes_count": len(session.Changes),
		},
	)

	// Удаляем сессию
	if err := e.deleteSession(user.ID); err != nil {
		loggingService.Telegram().ErrorWithContext(
			"Failed to delete language edit session",
			requestID,
			int64(user.ID),
			callback.Message.Chat.ID,
			"CancelLanguageEdit",
			map[string]interface{}{
				"user_id": user.ID,
				"error":   err.Error(),
			},
		)
	}

	return e.showProfileAfterCancel(callback, user)
}

// showProfileAfterCancel показывает профиль после отмены редактирования
func (e *IsolatedLanguageEditor) showProfileAfterCancel(callback *tgbotapi.CallbackQuery, user *models.User) error {
	summary, err := e.baseHandler.service.BuildProfileSummary(user)
	if err != nil {
		return err
	}

	text := summary + "\n\n" + e.baseHandler.service.Localizer.Get(user.InterfaceLanguageCode, "profile_actions")
	keyboard := e.baseHandler.keyboardBuilder.CreateProfileMenuKeyboard(user.InterfaceLanguageCode)

	return e.baseHandler.messageFactory.EditWithKeyboard(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		text,
		&keyboard,
	)
}

// HandleBackToMenu возвращает в главное меню редактирования
func (e *IsolatedLanguageEditor) HandleBackToMenu(callback *tgbotapi.CallbackQuery, user *models.User) error {
	session, err := e.getSession(user.ID)
	if err != nil {
		return e.StartEditSession(callback, user)
	}

	session.CurrentStep = "main_menu"
	session.LastActivity = time.Now()

	if err := e.saveSession(user.ID, session); err != nil {
		return err
	}

	return e.showMainEditMenu(callback, user, session)
}

// =============================================================================
// ВСПОМОГАТЕЛЬНЫЕ МЕТОДЫ ДЛЯ РАБОТЫ С REDIS
// =============================================================================

// getSessionKey возвращает ключ для хранения сессии в Redis
func (e *IsolatedLanguageEditor) getSessionKey(userID int) string {
	return fmt.Sprintf("language_edit_session:%d", userID)
}

// saveSession сохраняет сессию в кеш
func (e *IsolatedLanguageEditor) saveSession(userID int, session *LanguageEditSession) error {
	key := e.getSessionKey(userID)
	return e.baseHandler.service.Cache.Set(context.Background(), key, session, time.Hour)
}

// getSession получает сессию из кеша
func (e *IsolatedLanguageEditor) getSession(userID int) (*LanguageEditSession, error) {
	key := e.getSessionKey(userID)

	var data string
	if err := e.baseHandler.service.Cache.Get(context.Background(), key, &data); err != nil {
		return nil, fmt.Errorf("edit session not found")
	}

	var session LanguageEditSession
	if err := json.Unmarshal([]byte(data), &session); err != nil {
		return nil, fmt.Errorf("failed to unmarshal edit session: %w", err)
	}

	return &session, nil
}

// deleteSession удаляет сессию из кеша
func (e *IsolatedLanguageEditor) deleteSession(userID int) error {
	key := e.getSessionKey(userID)
	return e.baseHandler.service.Cache.Delete(context.Background(), key)
}
