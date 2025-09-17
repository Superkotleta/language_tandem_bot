package mocks

import (
	"database/sql"
	"language-exchange-bot/internal/models"
	"time"
)

// DatabaseMock имитирует базу данных для тестов
type DatabaseMock struct {
	users     map[int64]*models.User
	languages map[string]*models.Language
	interests map[int]*models.Interest
	lastError error
}

// NewDatabaseMock создает новый мок базы данных
func NewDatabaseMock() *DatabaseMock {
	db := &DatabaseMock{
		users:     make(map[int64]*models.User),
		languages: make(map[string]*models.Language),
		interests: make(map[int]*models.Interest),
	}

	// Предзаполняем тестовыми языками
	db.seedLanguages()
	db.seedInterests()

	return db
}

// seedLanguages добавляет тестовые языки
func (db *DatabaseMock) seedLanguages() {
	languages := []*models.Language{
		{ID: 1, Code: "en", NameNative: "English", NameEn: "English", IsInterfaceLanguage: true},
		{ID: 2, Code: "ru", NameNative: "Русский", NameEn: "Russian", IsInterfaceLanguage: true},
		{ID: 3, Code: "es", NameNative: "Español", NameEn: "Spanish", IsInterfaceLanguage: true},
		{ID: 4, Code: "zh", NameNative: "中文", NameEn: "Chinese", IsInterfaceLanguage: true},
	}

	for _, lang := range languages {
		db.languages[lang.Code] = lang
	}
}

// seedInterests добавляет тестовые интересы
func (db *DatabaseMock) seedInterests() {
	interests := []*models.Interest{
		{ID: 1, Name: "movies", Type: "entertainment"},
		{ID: 2, Name: "music", Type: "entertainment"},
		{ID: 3, Name: "sports", Type: "activity"},
		{ID: 4, Name: "travel", Type: "activity"},
		{ID: 5, Name: "technology", Type: "knowledge"},
		{ID: 6, Name: "food", Type: "lifestyle"},
	}

	for _, interest := range interests {
		db.interests[interest.ID] = interest
	}
}

// GetUserByTelegramID находит пользователя по Telegram ID
func (db *DatabaseMock) GetUserByTelegramID(telegramID int64) (*models.User, error) {
	if db.lastError != nil {
		return nil, db.lastError
	}

	user, exists := db.users[telegramID]
	if !exists {
		return nil, sql.ErrNoRows
	}

	return user, nil
}

// CreateUser создает нового пользователя
func (db *DatabaseMock) CreateUser(telegramID int64, username, firstName, languageCode string) (*models.User, error) {
	if db.lastError != nil {
		return nil, db.lastError
	}

	user := &models.User{
		ID:                     len(db.users) + 1,
		TelegramID:             telegramID,
		Username:               username,
		FirstName:              firstName,
		NativeLanguageCode:     "",
		TargetLanguageCode:     "",
		InterfaceLanguageCode:  languageCode,
		CreatedAt:              time.Now(),
		UpdatedAt:              time.Now(),
		State:                  "new",
		ProfileCompletionLevel: 0,
		Status:                 "new",
	}

	db.users[telegramID] = user
	return user, nil
}

// FindOrCreateUser находит или создает пользователя (основной метод для BotService)
func (db *DatabaseMock) FindOrCreateUser(telegramID int64, username, firstName string) (*models.User, error) {
	if db.lastError != nil {
		return nil, db.lastError
	}

	// Сначала пытаемся найти существующего пользователя
	if user, exists := db.users[telegramID]; exists {
		// Обновляем информацию если она изменилась
		user.Username = username
		user.FirstName = firstName
		user.UpdatedAt = time.Now()
		return user, nil
	}

	// Если не найден, создаем нового
	return db.CreateUser(telegramID, username, firstName, "en")
}

// UpdateUser обновляет пользователя
func (db *DatabaseMock) UpdateUser(user *models.User) error {
	if db.lastError != nil {
		return db.lastError
	}

	user.UpdatedAt = time.Now()
	db.users[user.TelegramID] = user
	return nil
}

// GetLanguages возвращает все языки
func (db *DatabaseMock) GetLanguages() ([]*models.Language, error) {
	if db.lastError != nil {
		return nil, db.lastError
	}

	languages := make([]*models.Language, 0, len(db.languages))
	for _, lang := range db.languages {
		languages = append(languages, lang)
	}

	return languages, nil
}

// GetLanguageByCode возвращает язык по коду
func (db *DatabaseMock) GetLanguageByCode(code string) (*models.Language, error) {
	if db.lastError != nil {
		return nil, db.lastError
	}

	lang, exists := db.languages[code]
	if !exists {
		return nil, sql.ErrNoRows
	}

	return lang, nil
}

// GetInterests возвращает все интересы
func (db *DatabaseMock) GetInterests() ([]*models.Interest, error) {
	if db.lastError != nil {
		return nil, db.lastError
	}

	interests := make([]*models.Interest, 0, len(db.interests))
	for _, interest := range db.interests {
		interests = append(interests, interest)
	}

	return interests, nil
}

// SaveUserInterests сохраняет интересы пользователя
func (db *DatabaseMock) SaveUserInterests(userID int64, interestIDs []int) error {
	if db.lastError != nil {
		return db.lastError
	}

	// В реальной БД здесь была бы таблица user_interests
	// Для мока просто сохраняем в пользователе
	user, exists := db.users[userID]
	if exists {
		user.Interests = interestIDs
		user.UpdatedAt = time.Now()
	}

	return nil
}

// GetUserInterests возвращает интересы пользователя
func (db *DatabaseMock) GetUserInterests(userID int64) ([]int, error) {
	if db.lastError != nil {
		return nil, db.lastError
	}

	user, exists := db.users[userID]
	if !exists {
		return []int{}, nil
	}

	return user.Interests, nil
}

// GetUserSelectedInterests возвращает выбранные интересы пользователя (alias для GetUserInterests)
func (db *DatabaseMock) GetUserSelectedInterests(userID int) ([]int, error) {
	return db.GetUserInterests(int64(userID))
}

// UpdateUserInterfaceLanguage обновляет язык интерфейса пользователя
func (db *DatabaseMock) UpdateUserInterfaceLanguage(userID int, language string) error {
	if db.lastError != nil {
		return db.lastError
	}

	for _, user := range db.users {
		if user.ID == userID {
			user.InterfaceLanguageCode = language
			user.UpdatedAt = time.Now()
			break
		}
	}

	return nil
}

// UpdateUserState обновляет состояние пользователя
func (db *DatabaseMock) UpdateUserState(userID int, state string) error {
	if db.lastError != nil {
		return db.lastError
	}

	for _, user := range db.users {
		if user.ID == userID {
			user.State = state
			user.UpdatedAt = time.Now()
			break
		}
	}

	return nil
}

// UpdateUserStatus обновляет статус пользователя
func (db *DatabaseMock) UpdateUserStatus(userID int, status string) error {
	if db.lastError != nil {
		return db.lastError
	}

	for _, user := range db.users {
		if user.ID == userID {
			user.Status = status
			user.UpdatedAt = time.Now()
			break
		}
	}

	return nil
}

// SaveUserFeedback сохраняет отзыв пользователя (заглушка)
func (db *DatabaseMock) SaveUserFeedback(userID int, feedbackText string, contactInfo *string) error {
	if db.lastError != nil {
		return db.lastError
	}

	// Для тестов просто возвращаем успех
	return nil
}

// GetUnprocessedFeedback возвращает необработанные отзывы (заглушка)
func (db *DatabaseMock) GetUnprocessedFeedback() ([]map[string]interface{}, error) {
	if db.lastError != nil {
		return nil, db.lastError
	}

	// Для тестов возвращаем пустой список
	return []map[string]interface{}{}, nil
}

// MarkFeedbackProcessed помечает отзыв как обработанный (заглушка)
func (db *DatabaseMock) MarkFeedbackProcessed(feedbackID int, adminResponse string) error {
	if db.lastError != nil {
		return db.lastError
	}

	// Для тестов просто возвращаем успех
	return nil
}

// UpdateUserNativeLanguage обновляет родной язык пользователя
func (db *DatabaseMock) UpdateUserNativeLanguage(userID int, langCode string) error {
	if db.lastError != nil {
		return db.lastError
	}

	for _, user := range db.users {
		if user.ID == userID {
			user.NativeLanguageCode = langCode
			user.UpdatedAt = time.Now()
			break
		}
	}

	return nil
}

// UpdateUserTargetLanguage обновляет изучаемый язык пользователя
func (db *DatabaseMock) UpdateUserTargetLanguage(userID int, langCode string) error {
	if db.lastError != nil {
		return db.lastError
	}

	for _, user := range db.users {
		if user.ID == userID {
			user.TargetLanguageCode = langCode
			user.UpdatedAt = time.Now()
			break
		}
	}

	return nil
}

// UpdateUserTargetLanguageLevel обновляет уровень изучаемого языка
func (db *DatabaseMock) UpdateUserTargetLanguageLevel(userID int, level string) error {
	if db.lastError != nil {
		return db.lastError
	}

	for _, user := range db.users {
		if user.ID == userID {
			user.TargetLanguageLevel = level
			user.UpdatedAt = time.Now()
			break
		}
	}

	return nil
}

// ResetUserProfile сбрасывает профиль пользователя
func (db *DatabaseMock) ResetUserProfile(userID int) error {
	if db.lastError != nil {
		return db.lastError
	}

	for _, user := range db.users {
		if user.ID == userID {
			user.NativeLanguageCode = ""
			user.TargetLanguageCode = ""
			user.TargetLanguageLevel = ""
			user.State = "new"
			user.Status = "new"
			user.ProfileCompletionLevel = 0
			user.Interests = []int{}
			user.UpdatedAt = time.Now()
			break
		}
	}

	return nil
}

// SaveUserInterest сохраняет один интерес пользователя
func (db *DatabaseMock) SaveUserInterest(userID, interestID int, isPrimary bool) error {
	if db.lastError != nil {
		return db.lastError
	}

	// Находим пользователя по ID
	var targetUser *models.User
	for _, user := range db.users {
		if user.ID == userID {
			targetUser = user
			break
		}
	}

	if targetUser == nil {
		return nil // Пользователь не найден
	}

	// Проверяем, есть ли уже такой интерес
	for _, existingID := range targetUser.Interests {
		if existingID == interestID {
			return nil // Интерес уже есть
		}
	}

	// Добавляем новый интерес
	targetUser.Interests = append(targetUser.Interests, interestID)
	targetUser.UpdatedAt = time.Now()

	return nil
}

// RemoveUserInterest удаляет интерес пользователя
func (db *DatabaseMock) RemoveUserInterest(userID, interestID int) error {
	if db.lastError != nil {
		return db.lastError
	}

	// Находим пользователя по ID
	var targetUser *models.User
	for _, user := range db.users {
		if user.ID == userID {
			targetUser = user
			break
		}
	}

	if targetUser == nil {
		return nil // Пользователь не найден
	}

	// Ищем и удаляем интерес
	for i, existingID := range targetUser.Interests {
		if existingID == interestID {
			targetUser.Interests = append(targetUser.Interests[:i], targetUser.Interests[i+1:]...)
			targetUser.UpdatedAt = time.Now()
			break
		}
	}

	return nil
}

// ClearUserInterests очищает все интересы пользователя
func (db *DatabaseMock) ClearUserInterests(userID int) error {
	if db.lastError != nil {
		return db.lastError
	}

	// Находим пользователя по ID
	for _, user := range db.users {
		if user.ID == userID {
			user.Interests = []int{}
			user.UpdatedAt = time.Now()
			break
		}
	}

	return nil
}

// GetConnection возвращает соединение с БД (для мока возвращаем заглушку)
func (db *DatabaseMock) GetConnection() *sql.DB {
	// Возвращаем nil - локализатор должен справляться с этим
	return nil
}

// Close закрывает соединение с БД (для мока ничего не делает)
func (db *DatabaseMock) Close() error {
	return nil
}

// Вспомогательные методы для тестов

// SetError устанавливает ошибку, которую будут возвращать методы
func (db *DatabaseMock) SetError(err error) {
	db.lastError = err
}

// ClearError очищает установленную ошибку
func (db *DatabaseMock) ClearError() {
	db.lastError = nil
}

// GetUser возвращает пользователя по Telegram ID (для тестов)
func (db *DatabaseMock) GetUser(telegramID int64) *models.User {
	return db.users[telegramID]
}

// Reset очищает все данные в моке
func (db *DatabaseMock) Reset() {
	db.users = make(map[int64]*models.User)
	db.lastError = nil
	db.seedLanguages()
	db.seedInterests()
}
