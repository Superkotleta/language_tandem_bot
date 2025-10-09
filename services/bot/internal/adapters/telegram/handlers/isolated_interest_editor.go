package handlers

import (
	"context"
	"fmt"
	"strings"
	"time"

	"language-exchange-bot/internal/cache"
	"language-exchange-bot/internal/core"
	"language-exchange-bot/internal/errors"
	"language-exchange-bot/internal/models"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// IsolatedInterestEditor управляет изолированной системой редактирования интересов.
type IsolatedInterestEditor struct {
	service         *core.BotService
	interestService *core.InterestService
	bot             *tgbotapi.BotAPI
	keyboardBuilder *KeyboardBuilder
	errorHandler    *errors.ErrorHandler
	cache           cache.ServiceInterface
}

// EditSession представляет сессию редактирования интересов.
type EditSession struct {
	UserID             int
	OriginalSelections []models.InterestSelection
	CurrentSelections  []models.InterestSelection
	Changes            []InterestChange
	CurrentCategory    string
	SessionStart       time.Time
	LastActivity       time.Time
}

// InterestChange представляет изменение в интересах.
type InterestChange struct {
	Action       string // "add", "remove", "set_primary", "unset_primary"
	InterestID   int
	InterestName string
	Category     string
	Timestamp    time.Time
}

// EditStats представляет статистику редактирования.
type EditStats struct {
	TotalSelected  int
	PrimaryCount   int
	CategoryCounts map[string]int
	ChangesCount   int
	LastUpdated    time.Time
}

// NewIsolatedInterestEditor создает новый редактор изолированных интересов.
func NewIsolatedInterestEditor(
	service *core.BotService,
	interestService *core.InterestService,
	bot *tgbotapi.BotAPI,
	keyboardBuilder *KeyboardBuilder,
	errorHandler *errors.ErrorHandler,
	cache cache.ServiceInterface,
) *IsolatedInterestEditor {
	return &IsolatedInterestEditor{
		service:         service,
		interestService: interestService,
		bot:             bot,
		keyboardBuilder: keyboardBuilder,
		errorHandler:    errorHandler,
		cache:           cache,
	}
}

// StartEditSession начинает новую сессию редактирования.
func (e *IsolatedInterestEditor) StartEditSession(callback *tgbotapi.CallbackQuery, user *models.User) error {
	e.service.LoggingService.Telegram().InfoWithContext(
		"Starting isolated edit session",
		generateRequestID("StartEditSession"),
		int64(user.ID),
		callback.Message.Chat.ID,
		"StartEditSession",
		map[string]interface{}{"userID": user.ID},
	)

	// Получаем оригинальные выборы пользователя
	originalSelections, err := e.interestService.GetUserInterestSelections(user.ID)
	if err != nil {
		return e.errorHandler.HandleTelegramError(err, callback.Message.Chat.ID, int64(user.ID), "GetUserInterestSelections")
	}

	// Создаем копию для редактирования
	currentSelections := make([]models.InterestSelection, len(originalSelections))
	copy(currentSelections, originalSelections)

	// Создаем сессию
	session := &EditSession{
		UserID:             user.ID,
		OriginalSelections: originalSelections,
		CurrentSelections:  currentSelections,
		Changes:            []InterestChange{},
		SessionStart:       time.Now(),
		LastActivity:       time.Now(),
	}

	// Сохраняем сессию в кеше
	cacheKey := fmt.Sprintf("edit_session_%d", user.ID)
	e.service.LoggingService.Cache().InfoWithContext(
		"Saving session to cache",
		generateRequestID("StartEditSession"),
		int64(user.ID),
		callback.Message.Chat.ID,
		"StartEditSession",
		map[string]interface{}{"userID": user.ID, "cacheKey": cacheKey},
	)

	err = e.cache.Set(context.Background(), cacheKey, session, 30*time.Minute)
	if err != nil {
		e.service.LoggingService.Cache().WarnWithContext(
			"Failed to cache edit session",
			generateRequestID("StartEditSession"),
			int64(user.ID),
			callback.Message.Chat.ID,
			"StartEditSession",
			map[string]interface{}{"userID": user.ID, "cacheKey": cacheKey, "error": err.Error()},
		)
	} else {
		e.service.LoggingService.Cache().InfoWithContext(
			"Successfully cached session",
			generateRequestID("StartEditSession"),
			int64(user.ID),
			callback.Message.Chat.ID,
			"StartEditSession",
			map[string]interface{}{"userID": user.ID, "cacheKey": cacheKey},
		)
	}

	// Показываем главное меню редактирования
	return e.ShowEditMainMenu(callback, user, session)
}

// showEditMainMenu показывает главное меню редактирования.
func (e *IsolatedInterestEditor) ShowEditMainMenu(callback *tgbotapi.CallbackQuery, user *models.User, session *EditSession) error {
	// Получаем статистику
	stats := e.calculateEditStats(session)

	// Создаем текст с хлебными крошками и статистикой
	breadcrumb := e.service.Localizer.Get(user.InterfaceLanguageCode, "edit_interests_breadcrumb")
	statsText := e.formatEditStats(stats, user.InterfaceLanguageCode)

	text := fmt.Sprintf("%s\n\n%s\n\n%s",
		breadcrumb,
		e.service.Localizer.Get(user.InterfaceLanguageCode, "edit_interests_main_menu"),
		statsText)

	// Создаем клавиатуру главного меню
	keyboard := e.createEditMainMenuKeyboard(user.InterfaceLanguageCode, stats)

	editMsg := tgbotapi.NewEditMessageTextAndMarkup(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		text,
		keyboard,
	)

	_, err := e.bot.Request(editMsg)

	return err
}

// showEditCategoriesMenu показывает меню категорий для редактирования.
func (e *IsolatedInterestEditor) ShowEditCategoriesMenu(callback *tgbotapi.CallbackQuery, user *models.User, session *EditSession) error {
	// Получаем категории с индикаторами прогресса
	categories, err := e.interestService.GetInterestCategories()
	if err != nil {
		return e.errorHandler.HandleTelegramError(err, callback.Message.Chat.ID, int64(user.ID), "GetInterestCategories")
	}

	// Создаем текст с хлебными крошками
	breadcrumb := e.service.Localizer.Get(user.InterfaceLanguageCode, "edit_interests_breadcrumb_categories")
	text := fmt.Sprintf("%s\n\n%s",
		breadcrumb,
		e.service.Localizer.Get(user.InterfaceLanguageCode, "edit_interests_choose_category"))

	// Создаем клавиатуру категорий с индикаторами
	keyboard := e.createEditCategoriesKeyboard(categories, session, user.InterfaceLanguageCode)

	editMsg := tgbotapi.NewEditMessageTextAndMarkup(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		text,
		keyboard,
	)

	_, err = e.bot.Request(editMsg)

	return err
}

// showEditCategoryInterests показывает интересы в категории для редактирования.
func (e *IsolatedInterestEditor) ShowEditCategoryInterests(callback *tgbotapi.CallbackQuery, user *models.User, session *EditSession, categoryKey string) error {
	// Обновляем текущую категорию в сессии
	session.CurrentCategory = categoryKey
	e.updateSession(session)

	// Получаем интересы в категории
	interests, err := e.interestService.GetInterestsByCategoryKey(categoryKey)
	if err != nil {
		return e.errorHandler.HandleTelegramError(err, callback.Message.Chat.ID, int64(user.ID), "GetInterestsByCategoryKey")
	}

	// Создаем карту выбранных интересов
	selectedMap := make(map[int]bool)
	for _, selection := range session.CurrentSelections {
		selectedMap[selection.InterestID] = true
	}

	// Создаем текст с хлебными крошками
	categoryName := e.service.Localizer.Get(user.InterfaceLanguageCode, "category_"+categoryKey)
	breadcrumb := fmt.Sprintf("%s > %s",
		e.service.Localizer.Get(user.InterfaceLanguageCode, "edit_interests_breadcrumb_categories"),
		categoryName)

	text := fmt.Sprintf("%s\n\n%s",
		breadcrumb,
		e.service.Localizer.Get(user.InterfaceLanguageCode, "edit_interests_in_category"))

	// Создаем клавиатуру интересов
	keyboard := e.createEditCategoryInterestsKeyboard(interests, selectedMap, categoryKey, user.InterfaceLanguageCode)

	editMsg := tgbotapi.NewEditMessageTextAndMarkup(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		text,
		keyboard,
	)

	_, err = e.bot.Request(editMsg)

	return err
}

// ShowEditPrimaryInterests показывает основные интересы для редактирования.
func (e *IsolatedInterestEditor) ShowEditPrimaryInterests(callback *tgbotapi.CallbackQuery, user *models.User, session *EditSession) error {
	e.service.LoggingService.Telegram().DebugWithContext(
		"ShowEditPrimaryInterests called",
		generateRequestID("ShowEditPrimaryInterests"),
		int64(user.ID),
		callback.Message.Chat.ID,
		"ShowEditPrimaryInterests",
		map[string]interface{}{"userID": user.ID},
	)

	// Получаем текущие выборы пользователя
	selections, err := e.interestService.GetUserInterestSelections(user.ID)
	if err != nil {
		e.service.LoggingService.Database().DebugWithContext(
			"ShowEditPrimaryInterests GetUserInterestSelections error",
			generateRequestID("ShowEditPrimaryInterests"),
			int64(user.ID),
			callback.Message.Chat.ID,
			"ShowEditPrimaryInterests",
			map[string]interface{}{"userID": user.ID, "error": err.Error()},
		)

		return e.errorHandler.HandleTelegramError(err, callback.Message.Chat.ID, int64(user.ID), "GetUserInterestSelections")
	}

	e.service.LoggingService.Database().DebugWithContext(
		"ShowEditPrimaryInterests found selections",
		generateRequestID("ShowEditPrimaryInterests"),
		int64(user.ID),
		callback.Message.Chat.ID,
		"ShowEditPrimaryInterests",
		map[string]interface{}{"userID": user.ID, "selectionsCount": len(selections)},
	)

	// Получаем общее количество интересов в системе
	allInterests, err := e.interestService.GetAllInterests()
	if err != nil {
		return e.errorHandler.HandleTelegramError(err, callback.Message.Chat.ID, int64(user.ID), "GetAllInterests")
	}

	// Вычисляем рекомендуемое количество основных интересов
	config := e.service.GetConfig()
	totalInterestsInSystem := len(allInterests)
	recommendedPrimary := int(float64(totalInterestsInSystem) * config.PrimaryPercentage)

	// Ограничиваем минимумом и максимумом
	if recommendedPrimary < config.MinPrimaryInterests {
		recommendedPrimary = config.MinPrimaryInterests
	}

	if recommendedPrimary > config.MaxPrimaryInterests {
		recommendedPrimary = config.MaxPrimaryInterests
	}

	// Подсчитываем уже выбранные основные интересы
	selectedPrimaryCount := 0

	for _, selection := range selections {
		if selection.IsPrimary {
			selectedPrimaryCount++
		}
	}

	// Создаем хлебные крошки
	breadcrumb := e.service.Localizer.Get(user.InterfaceLanguageCode, "edit_interests_breadcrumb_primary")

	// Создаем текст сообщения с счетчиком
	text := fmt.Sprintf("%s\n\n%s (%d из %d выбрано)",
		breadcrumb,
		e.service.Localizer.Get(user.InterfaceLanguageCode, "edit_interests_primary_description"),
		selectedPrimaryCount,
		recommendedPrimary)

	// Создаем клавиатуру для основных интересов
	keyboard := e.createEditPrimaryInterestsKeyboard(selections, user.InterfaceLanguageCode)

	editMsg := tgbotapi.NewEditMessageTextAndMarkup(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		text,
		keyboard,
	)

	_, err = e.bot.Request(editMsg)

	return err
}

// toggleInterestSelection переключает выбор интереса.
func (e *IsolatedInterestEditor) ToggleInterestSelection(callback *tgbotapi.CallbackQuery, user *models.User, interestID int) error {
	// Получаем сессию
	session, err := e.GetEditSession(user.ID)
	if err != nil {
		return e.errorHandler.HandleTelegramError(err, callback.Message.Chat.ID, int64(user.ID), "GetEditSession")
	}

	// Получаем информацию об интересе
	interest, err := e.interestService.GetInterestByID(interestID)
	if err != nil {
		return e.errorHandler.HandleTelegramError(err, callback.Message.Chat.ID, int64(user.ID), "GetInterestByID")
	}

	// Проверяем, выбран ли уже этот интерес
	isSelected := false

	for _, selection := range session.CurrentSelections {
		if selection.InterestID == interestID {
			isSelected = true

			break
		}
	}

	// Переключаем выбор
	if isSelected {
		// Удаляем выбор
		e.removeSelectionFromSession(session, interestID)
		e.addChange(session, InterestChange{
			Action:       "remove",
			InterestID:   interestID,
			InterestName: interest.KeyName,
			Category:     interest.CategoryKey,
			Timestamp:    time.Now(),
		})
	} else {
		// Добавляем выбор
		newSelection := models.InterestSelection{
			UserID:     user.ID,
			InterestID: interestID,
			IsPrimary:  false,
		}
		session.CurrentSelections = append(session.CurrentSelections, newSelection)
		e.addChange(session, InterestChange{
			Action:       "add",
			InterestID:   interestID,
			InterestName: interest.KeyName,
			Category:     interest.CategoryKey,
			Timestamp:    time.Now(),
		})
	}

	// Обновляем сессию
	e.updateSession(session)

	// Обновляем клавиатуру
	return e.ShowEditCategoryInterests(callback, user, session, session.CurrentCategory)
}

// showChangesPreview показывает предварительный просмотр изменений.
func (e *IsolatedInterestEditor) ShowChangesPreview(callback *tgbotapi.CallbackQuery, user *models.User, session *EditSession) error {
	// Создаем текст с изменениями
	text := e.formatChangesPreview(session, user.InterfaceLanguageCode)

	// Создаем клавиатуру для предварительного просмотра
	keyboard := e.createChangesPreviewKeyboard(user.InterfaceLanguageCode)

	editMsg := tgbotapi.NewEditMessageTextAndMarkup(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		text,
		keyboard,
	)

	_, err := e.bot.Request(editMsg)

	return err
}

// saveChanges сохраняет изменения.
func (e *IsolatedInterestEditor) SaveChanges(callback *tgbotapi.CallbackQuery, user *models.User, session *EditSession) error {
	e.service.LoggingService.Database().InfoWithContext(
		"Saving changes for user",
		generateRequestID("SaveChanges"),
		int64(user.ID),
		callback.Message.Chat.ID,
		"SaveChanges",
		map[string]interface{}{
			"userID":                 user.ID,
			"currentSelectionsCount": len(session.CurrentSelections),
			"changesCount":           len(session.Changes),
		},
	)

	// Валидируем выборы
	if err := e.validateSelections(session); err != nil {
		e.service.LoggingService.Database().ErrorWithContext(
			"Validation failed",
			generateRequestID("SaveChanges"),
			int64(user.ID),
			callback.Message.Chat.ID,
			"SaveChanges",
			map[string]interface{}{"userID": user.ID, "error": err.Error()},
		)

		return e.errorHandler.HandleTelegramError(err, callback.Message.Chat.ID, int64(user.ID), "ValidateSelections")
	}

	// Сохраняем изменения в базу данных
	e.service.LoggingService.Database().InfoWithContext(
		"Calling BatchUpdateUserInterests",
		generateRequestID("SaveChanges"),
		int64(user.ID),
		callback.Message.Chat.ID,
		"SaveChanges",
		map[string]interface{}{"userID": user.ID, "selectionsCount": len(session.CurrentSelections)},
	)

	err := e.interestService.BatchUpdateUserInterests(user.ID, session.CurrentSelections)
	if err != nil {
		e.service.LoggingService.Database().ErrorWithContext(
			"BatchUpdateUserInterests failed",
			generateRequestID("SaveChanges"),
			int64(user.ID),
			callback.Message.Chat.ID,
			"SaveChanges",
			map[string]interface{}{"userID": user.ID, "selectionsCount": len(session.CurrentSelections), "error": err.Error()},
		)

		return e.errorHandler.HandleTelegramError(err, callback.Message.Chat.ID, int64(user.ID), "BatchUpdateUserInterests")
	}

	// Очищаем сессию
	e.clearEditSession(user.ID)

	// Показываем уведомление об изменениях и возвращаемся к профилю
	changesCount := len(session.Changes)
	text := fmt.Sprintf("✅ Изменения успешно сохранены!\n\n📊 Внесено изменений: %d", changesCount)

	// Добавляем детальную информацию об изменениях
	if changesCount > 0 {
		text += "\n\n📝 Детали изменений:"
		text += e.formatChangesSummary(session, user.InterfaceLanguageCode)
	}

	// Создаем клавиатуру профиля
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("👤 Показать профиль", "profile_show"),
		),
	)

	editMsg := tgbotapi.NewEditMessageTextAndMarkup(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		text,
		keyboard,
	)

	_, err = e.bot.Request(editMsg)

	return err
}

// cancelEdit отменяет редактирование.
func (e *IsolatedInterestEditor) CancelEdit(callback *tgbotapi.CallbackQuery, user *models.User) error {
	e.service.LoggingService.Telegram().InfoWithContext(
		"Canceling edit for user",
		generateRequestID("CancelEdit"),
		int64(user.ID),
		callback.Message.Chat.ID,
		"CancelEdit",
		map[string]interface{}{"userID": user.ID},
	)

	// Получаем сессию для подсчета изменений
	session, err := e.GetEditSession(user.ID)

	changesCount := 0
	if err == nil {
		changesCount = len(session.Changes)
	}

	// Очищаем сессию
	e.clearEditSession(user.ID)

	// Показываем уведомление об отмене и возвращаемся к профилю
	text := fmt.Sprintf("❌ Редактирование отменено!\n\n📊 Отменено изменений: %d", changesCount)

	// Добавляем детальную информацию об отмененных изменениях
	if err == nil && changesCount > 0 {
		text += "\n\n📝 Отмененные изменения:"
		text += e.formatChangesSummary(session, user.InterfaceLanguageCode)
	}

	// Создаем клавиатуру профиля
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("👤 Показать профиль", "profile_show"),
		),
	)

	editMsg := tgbotapi.NewEditMessageTextAndMarkup(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		text,
		keyboard,
	)

	_, err = e.bot.Request(editMsg)

	return err
}

// Вспомогательные методы

func (e *IsolatedInterestEditor) GetEditSession(userID int) (*EditSession, error) {
	e.service.LoggingService.Cache().DebugWithContext(
		"Getting edit session for user",
		generateRequestID("GetEditSession"),
		int64(userID),
		0, // нет chatID в этой функции
		"GetEditSession",
		map[string]interface{}{"userID": userID},
	)

	var session EditSession

	cacheKey := fmt.Sprintf("edit_session_%d", userID)
	e.service.LoggingService.Cache().DebugWithContext(
		"Cache key generated",
		generateRequestID("GetEditSession"),
		int64(userID),
		0, // нет chatID в этой функции
		"GetEditSession",
		map[string]interface{}{"userID": userID, "cacheKey": cacheKey},
	)

	err := e.cache.Get(context.Background(), cacheKey, &session)
	if err != nil {
		e.service.LoggingService.Cache().ErrorWithContext(
			"Failed to get session from cache",
			generateRequestID("GetEditSession"),
			int64(userID),
			0, // нет chatID в этой функции
			"GetEditSession",
			map[string]interface{}{"userID": userID, "cacheKey": cacheKey, "error": err.Error()},
		)

		return nil, fmt.Errorf("session not found: %w", err)
	}

	e.service.LoggingService.Cache().DebugWithContext(
		"Successfully retrieved session",
		generateRequestID("GetEditSession"),
		int64(userID),
		0, // нет chatID в этой функции
		"GetEditSession",
		map[string]interface{}{"userID": userID, "selectionsCount": len(session.CurrentSelections)},
	)

	return &session, nil
}

func (e *IsolatedInterestEditor) updateSession(session *EditSession) {
	session.LastActivity = time.Now()
	_ = e.cache.Set(context.Background(), fmt.Sprintf("edit_session_%d", session.UserID), session, 30*time.Minute)
}

func (e *IsolatedInterestEditor) clearEditSession(userID int) {
	_ = e.cache.Delete(context.Background(), fmt.Sprintf("edit_session_%d", userID))
}

func (e *IsolatedInterestEditor) removeSelectionFromSession(session *EditSession, interestID int) {
	for i, selection := range session.CurrentSelections {
		if selection.InterestID == interestID {
			session.CurrentSelections = append(session.CurrentSelections[:i], session.CurrentSelections[i+1:]...)

			break
		}
	}
}

func (e *IsolatedInterestEditor) addChange(session *EditSession, change InterestChange) {
	session.Changes = append(session.Changes, change)
}

func (e *IsolatedInterestEditor) calculateEditStats(session *EditSession) EditStats {
	stats := EditStats{
		CategoryCounts: make(map[string]int),
		LastUpdated:    time.Now(),
	}

	for _, selection := range session.CurrentSelections {
		stats.TotalSelected++
		if selection.IsPrimary {
			stats.PrimaryCount++
		}
		// TODO: Добавить подсчет по категориям
	}

	stats.ChangesCount = len(session.Changes)

	return stats
}

func (e *IsolatedInterestEditor) formatEditStats(stats EditStats, lang string) string {
	return fmt.Sprintf("📊 %s: %d | ⭐ %s: %d | 🔄 %s: %d",
		e.service.Localizer.Get(lang, "total_interests"),
		stats.TotalSelected,
		e.service.Localizer.Get(lang, "primary_interests_label"),
		stats.PrimaryCount,
		e.service.Localizer.Get(lang, "changes_count"),
		stats.ChangesCount)
}

func (e *IsolatedInterestEditor) formatChangesPreview(session *EditSession, lang string) string {
	text := e.service.Localizer.Get(lang, "edit_interests_changes_preview") + "\n\n"

	if len(session.Changes) == 0 {
		text += e.service.Localizer.Get(lang, "no_changes_made")

		return text
	}

	// Группируем изменения
	added := []InterestChange{}
	removed := []InterestChange{}

	for _, change := range session.Changes {
		switch change.Action {
		case "add":
			added = append(added, change)
		case "remove":
			removed = append(removed, change)
		}
	}

	if len(added) > 0 {
		text += "✅ " + e.service.Localizer.Get(lang, "added_interests") + ":\n"
		for _, change := range added {
			text += fmt.Sprintf("• %s\n", change.InterestName)
		}

		text += "\n"
	}

	if len(removed) > 0 {
		text += "❌ " + e.service.Localizer.Get(lang, "removed_interests") + ":\n"
		for _, change := range removed {
			text += fmt.Sprintf("• %s\n", change.InterestName)
		}

		text += "\n"
	}

	return text
}

func (e *IsolatedInterestEditor) validateSelections(session *EditSession) error {
	// Разрешаем сохранение даже если нет выбранных интересов
	// Это позволяет пользователю очистить все свои интересы

	// Логируем только если service инициализирован (для совместимости с тестами)
	if e.service != nil && e.service.LoggingService != nil {
		e.service.LoggingService.Database().DebugWithContext(
			"Validating selections",
			generateRequestID("validateSelections"),
			int64(session.UserID),
			0, // нет chatID в этой функции
			"validateSelections",
			map[string]interface{}{"userID": session.UserID, "selectionsCount": len(session.CurrentSelections)},
		)
	}

	// TODO: Добавить дополнительные проверки валидации
	return nil
}

// formatChangesSummary форматирует краткое описание изменений.
func (e *IsolatedInterestEditor) formatChangesSummary(session *EditSession, lang string) string {
	if len(session.Changes) == 0 {
		return "\n• Изменений не было"
	}

	var (
		addedInterests   []string
		removedInterests []string
		primarySet       []string
		primaryUnset     []string
	)

	// Группируем изменения по типам

	for _, change := range session.Changes {
		interestName := e.service.Localizer.Get(lang, "interest_"+change.InterestName)

		switch change.Action {
		case "add":
			addedInterests = append(addedInterests, "• "+interestName)
		case "remove":
			removedInterests = append(removedInterests, "• "+interestName)
		case "set_primary":
			primarySet = append(primarySet, "• "+interestName)
		case "unset_primary":
			primaryUnset = append(primaryUnset, "• "+interestName)
		}
	}

	var summary strings.Builder

	if len(addedInterests) > 0 {
		summary.WriteString("\n\n➕ Добавлены:")

		for _, interest := range addedInterests {
			summary.WriteString("\n" + interest)
		}
	}

	if len(removedInterests) > 0 {
		summary.WriteString("\n\n➖ Удалены:")

		for _, interest := range removedInterests {
			summary.WriteString("\n" + interest)
		}
	}

	if len(primarySet) > 0 {
		summary.WriteString("\n\n⭐ Сделаны основными:")

		for _, interest := range primarySet {
			summary.WriteString("\n" + interest)
		}
	}

	if len(primaryUnset) > 0 {
		summary.WriteString("\n\n☐ Убраны из основных:")

		for _, interest := range primaryUnset {
			summary.WriteString("\n" + interest)
		}
	}

	return summary.String()
}

// togglePrimaryInterest переключает статус основного интереса.
func (e *IsolatedInterestEditor) TogglePrimaryInterest(callback *tgbotapi.CallbackQuery, user *models.User, interestID int) error {
	session, err := e.GetEditSession(user.ID)
	if err != nil {
		return e.errorHandler.HandleTelegramError(err, callback.Message.Chat.ID, int64(user.ID), "GetEditSession")
	}

	// Находим интерес в текущих выборах
	for i, selection := range session.CurrentSelections {
		if selection.InterestID == interestID {
			// Проверяем лимиты перед переключением
			currentPrimaryCount := 0

			for _, sel := range session.CurrentSelections {
				if sel.IsPrimary {
					currentPrimaryCount++
				}
			}

			// Если пытаемся сделать основным, проверяем максимум
			if !session.CurrentSelections[i].IsPrimary {
				// Получаем общее количество интересов в системе
				allInterests, err := e.interestService.GetAllInterests()
				if err != nil {
					return e.errorHandler.HandleTelegramError(err, callback.Message.Chat.ID, int64(user.ID), "GetAllInterests")
				}

				// Вычисляем рекомендуемое количество основных интересов
				config := e.service.GetConfig()
				totalInterestsInSystem := len(allInterests)
				recommendedPrimary := int(float64(totalInterestsInSystem) * config.PrimaryPercentage)

				// Ограничиваем минимумом и максимумом
				if recommendedPrimary < config.MinPrimaryInterests {
					recommendedPrimary = config.MinPrimaryInterests
				}

				if recommendedPrimary > config.MaxPrimaryInterests {
					recommendedPrimary = config.MaxPrimaryInterests
				}

				if currentPrimaryCount >= recommendedPrimary {
					// Показываем предупреждение о достижении максимума
					text := fmt.Sprintf("❌ Максимальное количество основных интересов (%d) уже выбрано!\n\n%s (%d из %d выбрано)",
						recommendedPrimary,
						e.service.Localizer.Get(user.InterfaceLanguageCode, "edit_interests_primary_description"),
						currentPrimaryCount,
						recommendedPrimary)

					keyboard := e.createEditPrimaryInterestsKeyboard(session.CurrentSelections, user.InterfaceLanguageCode)
					editMsg := tgbotapi.NewEditMessageTextAndMarkup(
						callback.Message.Chat.ID,
						callback.Message.MessageID,
						text,
						keyboard,
					)
					_, err = e.bot.Request(editMsg)

					return err
				}
			}

			// Переключаем статус основного
			session.CurrentSelections[i].IsPrimary = !session.CurrentSelections[i].IsPrimary

			// Добавляем изменение
			action := "unset_primary"
			if session.CurrentSelections[i].IsPrimary {
				action = "set_primary"
			}

			interest, err := e.interestService.GetInterestByID(interestID)
			if err == nil {
				e.addChange(session, InterestChange{
					Action:       action,
					InterestID:   interestID,
					InterestName: interest.KeyName,
					Category:     interest.CategoryKey,
					Timestamp:    time.Now(),
				})
			}

			break
		}
	}

	// Обновляем сессию
	e.updateSession(session)

	// Показываем обновленную клавиатуру основных интересов
	return e.showEditPrimaryInterests(callback, user, session)
}

// showEditPrimaryInterests показывает интерфейс редактирования основных интересов.
func (e *IsolatedInterestEditor) showEditPrimaryInterests(callback *tgbotapi.CallbackQuery, user *models.User, session *EditSession) error {
	// Создаем текст с хлебными крошками
	breadcrumb := e.service.Localizer.Get(user.InterfaceLanguageCode, "edit_interests_breadcrumb_primary")
	text := fmt.Sprintf("%s\n\n%s",
		breadcrumb,
		e.service.Localizer.Get(user.InterfaceLanguageCode, "edit_interests_primary_description"))

	// Создаем клавиатуру основных интересов
	keyboard := e.createEditPrimaryInterestsKeyboard(session.CurrentSelections, user.InterfaceLanguageCode)

	editMsg := tgbotapi.NewEditMessageTextAndMarkup(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		text,
		keyboard,
	)

	_, err := e.bot.Request(editMsg)

	return err
}

// massSelectCategory выбирает все интересы в категории.
func (e *IsolatedInterestEditor) MassSelectCategory(callback *tgbotapi.CallbackQuery, user *models.User, session *EditSession, categoryKey string) error {
	// Получаем все интересы в категории
	interests, err := e.interestService.GetInterestsByCategoryKey(categoryKey)
	if err != nil {
		return e.errorHandler.HandleTelegramError(err, callback.Message.Chat.ID, int64(user.ID), "GetInterestsByCategoryKey")
	}

	// Добавляем все интересы, которых еще нет в выборах
	existingIDs := make(map[int]bool)
	for _, selection := range session.CurrentSelections {
		existingIDs[selection.InterestID] = true
	}

	for _, interest := range interests {
		if !existingIDs[interest.ID] {
			newSelection := models.InterestSelection{
				UserID:     user.ID,
				InterestID: interest.ID,
				IsPrimary:  false,
			}
			session.CurrentSelections = append(session.CurrentSelections, newSelection)

			// Добавляем изменение
			e.addChange(session, InterestChange{
				Action:       "add",
				InterestID:   interest.ID,
				InterestName: interest.KeyName,
				Category:     categoryKey,
				Timestamp:    time.Now(),
			})
		}
	}

	// Обновляем сессию
	e.updateSession(session)

	// Показываем обновленную клавиатуру
	return e.ShowEditCategoryInterests(callback, user, session, categoryKey)
}

// massClearCategory очищает все интересы в категории.
func (e *IsolatedInterestEditor) MassClearCategory(callback *tgbotapi.CallbackQuery, user *models.User, session *EditSession, categoryKey string) error {
	// Получаем все интересы в категории
	interests, err := e.interestService.GetInterestsByCategoryKey(categoryKey)
	if err != nil {
		return e.errorHandler.HandleTelegramError(err, callback.Message.Chat.ID, int64(user.ID), "GetInterestsByCategoryKey")
	}

	// Создаем карту ID интересов в категории
	categoryInterestIDs := make(map[int]bool)
	for _, interest := range interests {
		categoryInterestIDs[interest.ID] = true
	}

	// Удаляем все интересы из этой категории
	var newSelections []models.InterestSelection

	for _, selection := range session.CurrentSelections {
		if !categoryInterestIDs[selection.InterestID] {
			newSelections = append(newSelections, selection)
		} else {
			// Добавляем изменение удаления
			interest, err := e.interestService.GetInterestByID(selection.InterestID)
			if err == nil {
				e.addChange(session, InterestChange{
					Action:       "remove",
					InterestID:   selection.InterestID,
					InterestName: interest.KeyName,
					Category:     categoryKey,
					Timestamp:    time.Now(),
				})
			}
		}
	}

	session.CurrentSelections = newSelections

	// Обновляем сессию
	e.updateSession(session)

	// Показываем обновленную клавиатуру
	return e.ShowEditCategoryInterests(callback, user, session, categoryKey)
}

// undoLastChange отменяет последнее изменение.
func (e *IsolatedInterestEditor) UndoLastChange(callback *tgbotapi.CallbackQuery, user *models.User, session *EditSession) error {
	if len(session.Changes) == 0 {
		// Нет изменений для отмены
		return e.ShowEditMainMenu(callback, user, session)
	}

	// Получаем последнее изменение
	lastChange := session.Changes[len(session.Changes)-1]

	// Отменяем изменение
	switch lastChange.Action {
	case "add":
		// Удаляем интерес
		e.removeSelectionFromSession(session, lastChange.InterestID)
	case "remove":
		// Добавляем интерес обратно
		newSelection := models.InterestSelection{
			UserID:     user.ID,
			InterestID: lastChange.InterestID,
			IsPrimary:  false,
		}
		session.CurrentSelections = append(session.CurrentSelections, newSelection)
	case "set_primary":
		// Убираем статус основного
		for i, selection := range session.CurrentSelections {
			if selection.InterestID == lastChange.InterestID {
				session.CurrentSelections[i].IsPrimary = false

				break
			}
		}
	case "unset_primary":
		// Устанавливаем статус основного
		for i, selection := range session.CurrentSelections {
			if selection.InterestID == lastChange.InterestID {
				session.CurrentSelections[i].IsPrimary = true

				break
			}
		}
	}

	// Удаляем последнее изменение из истории
	session.Changes = session.Changes[:len(session.Changes)-1]

	// Обновляем сессию
	e.updateSession(session)

	// Показываем главное меню
	return e.ShowEditMainMenu(callback, user, session)
}

// showEditStatistics показывает статистику редактирования.
func (e *IsolatedInterestEditor) ShowEditStatistics(callback *tgbotapi.CallbackQuery, user *models.User, session *EditSession) error {
	stats := e.calculateEditStats(session)

	// Создаем детальную статистику
	text := e.formatDetailedStatistics(stats, session, user.InterfaceLanguageCode)

	// Создаем клавиатуру для статистики
	keyboard := e.createStatisticsKeyboard(user.InterfaceLanguageCode)

	editMsg := tgbotapi.NewEditMessageTextAndMarkup(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		text,
		keyboard,
	)

	_, err := e.bot.Request(editMsg)

	return err
}

// formatDetailedStatistics форматирует детальную статистику.
func (e *IsolatedInterestEditor) formatDetailedStatistics(stats EditStats, session *EditSession, lang string) string {
	text := e.service.Localizer.Get(lang, "edit_interests_detailed_statistics") + "\n\n"

	// Общая статистика
	text += fmt.Sprintf("📊 %s: %d\n", e.service.Localizer.Get(lang, "total_interests"), stats.TotalSelected)
	text += fmt.Sprintf("⭐ %s: %d\n", e.service.Localizer.Get(lang, "primary_interests_label"), stats.PrimaryCount)
	text += fmt.Sprintf("🔄 %s: %d\n", e.service.Localizer.Get(lang, "changes_count"), stats.ChangesCount)
	text += fmt.Sprintf("⏱️ %s: %s\n\n", e.service.Localizer.Get(lang, "session_duration"),
		time.Since(session.SessionStart).Round(time.Minute))

	// Статистика по категориям
	if len(stats.CategoryCounts) > 0 {
		text += e.service.Localizer.Get(lang, "category_statistics") + ":\n"
		for category, count := range stats.CategoryCounts {
			categoryName := e.service.Localizer.Get(lang, "category_"+category)
			text += fmt.Sprintf("• %s: %d\n", categoryName, count)
		}
	}

	return text
}

// createStatisticsKeyboard создает клавиатуру для статистики.
func (e *IsolatedInterestEditor) createStatisticsKeyboard(interfaceLang string) tgbotapi.InlineKeyboardMarkup {
	var buttonRows [][]tgbotapi.InlineKeyboardButton

	// Навигация
	navRow := []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData(
			"🏠 "+e.service.Localizer.Get(interfaceLang, "back_to_main_menu"),
			"isolated_main_menu",
		),
		tgbotapi.NewInlineKeyboardButtonData(
			"👁️ "+e.service.Localizer.Get(interfaceLang, "preview_changes"),
			"isolated_preview_changes",
		),
	}
	buttonRows = append(buttonRows, navRow)

	return tgbotapi.NewInlineKeyboardMarkup(buttonRows...)
}
