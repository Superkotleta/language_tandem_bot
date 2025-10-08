package database

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite" // SQLite driver for testing
)

// TestNewDB_InvalidConnection tests database creation with invalid connection string.
// This test ensures that the database initialization properly handles
// connection failures and returns appropriate error messages.
func TestNewDB_InvalidConnection(t *testing.T) {
	// Attempt to connect to a non-existent database
	db, err := NewDB("postgres://invalid:invalid@invalid:5432/invalid?sslmode=disable")

	// Should return an error for invalid connection
	assert.Error(t, err)
	assert.Nil(t, db)
	assert.Contains(t, err.Error(), "failed to ping database")
}

// TestDB_GetConnection тестирует получение соединения
func TestDB_GetConnection(t *testing.T) {
	// Создаем mock DB для тестирования
	mockDB := &DB{
		conn: &sql.DB{}, // mock connection
	}

	conn := mockDB.GetConnection()
	assert.NotNil(t, conn)
}

// TestDB_Close тестирует закрытие соединения
func TestDB_Close(t *testing.T) {
	// Создаем тестовую базу данных в памяти
	db, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)

	// Создаем DB instance
	database := &DB{conn: db}

	// Close должен работать без ошибок
	err = database.Close()
	assert.NoError(t, err)
}

// TestDB_GetLanguages тестирует получение списка языков
func TestDB_GetLanguages(t *testing.T) {
	// Создаем тестовую базу данных в памяти
	db, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)
	defer func() {
		if closeErr := db.Close(); closeErr != nil {
			t.Logf("Failed to close database connection: %v", closeErr)
		}
	}()

	// Создаем таблицы и языки
	setupTestLanguagesTable(t, db)

	// Создаем DB instance
	database := &DB{conn: db}

	// Получаем языки
	languages, err := database.GetLanguages()
	require.NoError(t, err)
	assert.Len(t, languages, 2)

	assert.Equal(t, "en", languages[0].Code)
	assert.Equal(t, "English", languages[0].NameEn)
	assert.Equal(t, "ru", languages[1].Code)
	assert.Equal(t, "Russian", languages[1].NameEn)
}

// TestDB_GetLanguageByCode тестирует получение языка по коду
func TestDB_GetLanguageByCode(t *testing.T) {
	// Создаем тестовую базу данных в памяти
	db, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)
	defer func() {
		if closeErr := db.Close(); closeErr != nil {
			t.Logf("Failed to close database connection: %v", closeErr)
		}
	}()

	// Создаем таблицы и языки
	setupTestLanguagesTable(t, db)

	// Создаем DB instance
	database := &DB{conn: db}

	// Получаем язык по коду
	language, err := database.GetLanguageByCode("en")
	require.NoError(t, err)
	require.NotNil(t, language)

	assert.Equal(t, "en", language.Code)
	assert.Equal(t, "English", language.NameEn)
	assert.True(t, language.IsInterfaceLanguage)

	// Тестируем несуществующий код
	language, err = database.GetLanguageByCode("xx")
	assert.Error(t, err)
	assert.Nil(t, language)
}

// TestDB_GetInterests тестирует получение списка интересов
func TestDB_GetInterests(t *testing.T) {
	// Создаем тестовую базу данных в памяти
	db, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)
	defer func() {
		if closeErr := db.Close(); closeErr != nil {
			t.Logf("Failed to close database connection: %v", closeErr)
		}
	}()

	// Создаем таблицы и интересы
	setupTestInterestsTable(t, db)

	// Создаем DB instance
	database := &DB{conn: db}

	// Получаем интересы
	interests, err := database.GetInterests()
	require.NoError(t, err)
	assert.Len(t, interests, 3)

	assert.Equal(t, "Programming", interests[0].KeyName)
	assert.Equal(t, "Music", interests[1].KeyName)
	assert.Equal(t, "Sports", interests[2].KeyName)
}

// TestDB_GetUserByTelegramID тестирует получение пользователя по Telegram ID
func TestDB_GetUserByTelegramID(t *testing.T) {
	// Создаем тестовую базу данных в памяти
	db, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)
	defer func() {
		if closeErr := db.Close(); closeErr != nil {
			t.Logf("Failed to close database connection: %v", closeErr)
		}
	}()

	// Создаем таблицы и пользователя
	setupTestUsersTable(t, db)
	setupTestUser(t, db, 123456789, "testuser", "Test User")

	// Создаем DB instance
	database := &DB{conn: db}

	// Получаем пользователя
	user, err := database.GetUserByTelegramID(123456789)
	require.NoError(t, err)
	require.NotNil(t, user)

	assert.Equal(t, int64(123456789), user.TelegramID)
	assert.Equal(t, "testuser", user.Username)
	assert.Equal(t, "Test User", user.FirstName)
	assert.Equal(t, "en", user.InterfaceLanguageCode)

	// Тестируем несуществующего пользователя
	user, err = database.GetUserByTelegramID(999999999)
	assert.Error(t, err)
	assert.Nil(t, user)
}

// TestDB_SaveUserInterests тестирует сохранение интересов пользователя
func TestDB_SaveUserInterests(t *testing.T) {
	// Создаем тестовую базу данных в памяти
	db, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)
	defer func() {
		if closeErr := db.Close(); closeErr != nil {
			t.Logf("Failed to close database connection: %v", closeErr)
		}
	}()

	// Создаем таблицы
	setupTestUsersTable(t, db)
	setupTestUserInterestsTable(t, db)
	setupTestUser(t, db, 123456789, "testuser", "Test User")

	// Создаем DB instance
	database := &DB{conn: db}

	interestIDs := []int{1, 2, 3}

	// Сохраняем интересы (используем database ID = 1)
	err = database.SaveUserInterests(1, interestIDs)
	assert.NoError(t, err)

	// Проверяем, что интересы сохранены
	savedInterests, err := database.GetUserSelectedInterests(1)
	assert.NoError(t, err)
	assert.Equal(t, interestIDs, savedInterests)
}

// TestDB_FindOrCreateUser_NewUser тестирует создание нового пользователя
func TestDB_FindOrCreateUser_NewUser(t *testing.T) {
	// Создаем тестовую базу данных в памяти
	db, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)
	defer func() {
		if closeErr := db.Close(); closeErr != nil {
			t.Logf("Failed to close database connection: %v", closeErr)
		}
	}()

	// Создаем таблицы
	setupTestUsersTable(t, db)

	telegramID := int64(123456789)
	username := "testuser"
	firstName := "Test User"

	// Создаем DB instance
	database := &DB{conn: db}

	// Тестируем создание пользователя
	user, err := database.FindOrCreateUser(telegramID, username, firstName)
	require.NoError(t, err)
	require.NotNil(t, user)

	assert.Equal(t, telegramID, user.TelegramID)
	assert.Equal(t, username, user.Username)
	assert.Equal(t, firstName, user.FirstName)
	assert.Equal(t, "en", user.InterfaceLanguageCode)
	assert.Equal(t, "new", user.State)
	assert.Equal(t, "new", user.Status)
	assert.Equal(t, 0, user.ProfileCompletionLevel)
}

// TestDB_FindOrCreateUser_ExistingUser тестирует поиск существующего пользователя
func TestDB_FindOrCreateUser_ExistingUser(t *testing.T) {
	// Создаем тестовую базу данных в памяти
	db, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)
	defer func() {
		if closeErr := db.Close(); closeErr != nil {
			t.Logf("Failed to close database connection: %v", closeErr)
		}
	}()

	// Создаем таблицы и пользователя
	setupTestUsersTable(t, db)
	setupTestUser(t, db, 123456789, "testuser", "Test User")

	// Создаем DB instance
	database := &DB{conn: db}

	// Тестируем поиск существующего пользователя
	user, err := database.FindOrCreateUser(123456789, "updateduser", "Updated Name")
	require.NoError(t, err)
	require.NotNil(t, user)

	assert.Equal(t, int64(123456789), user.TelegramID)
	assert.Equal(t, "updateduser", user.Username)   // должно обновиться
	assert.Equal(t, "Updated Name", user.FirstName) // должно обновиться
}

// TestDB_UpdateUserState тестирует обновление состояния пользователя
func TestDB_UpdateUserState(t *testing.T) {
	// Создаем тестовую базу данных в памяти
	db, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)
	defer func() {
		if closeErr := db.Close(); closeErr != nil {
			t.Logf("Failed to close database connection: %v", closeErr)
		}
	}()

	// Создаем таблицы и пользователя
	setupTestUsersTable(t, db)
	setupTestUser(t, db, 123456789, "testuser", "Test User")

	// Создаем DB instance
	database := &DB{conn: db}

	// Обновляем состояние
	err = database.UpdateUserState(1, "waiting_language")
	require.NoError(t, err)

	// Проверяем обновление
	user, err := database.GetUserByTelegramID(123456789)
	require.NoError(t, err)
	assert.Equal(t, "waiting_language", user.State)
}

// TestDB_UpdateUserStatus тестирует обновление статуса пользователя
func TestDB_UpdateUserStatus(t *testing.T) {
	// Создаем тестовую базу данных в памяти
	db, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)
	defer func() {
		if closeErr := db.Close(); closeErr != nil {
			t.Logf("Failed to close database connection: %v", closeErr)
		}
	}()

	// Создаем таблицы и пользователя
	setupTestUsersTable(t, db)
	setupTestUser(t, db, 123456789, "testuser", "Test User")

	// Создаем DB instance
	database := &DB{conn: db}

	// Обновляем статус
	err = database.UpdateUserStatus(1, "active")
	require.NoError(t, err)

	// Проверяем обновление
	user, err := database.GetUserByTelegramID(123456789)
	require.NoError(t, err)
	assert.Equal(t, "active", user.Status)
}

// TestDB_UpdateUserInterfaceLanguage тестирует обновление языка интерфейса
func TestDB_UpdateUserInterfaceLanguage(t *testing.T) {
	// Создаем тестовую базу данных в памяти
	db, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)
	defer func() {
		if closeErr := db.Close(); closeErr != nil {
			t.Logf("Failed to close database connection: %v", closeErr)
		}
	}()

	// Создаем таблицы и пользователя
	setupTestUsersTable(t, db)
	setupTestUser(t, db, 123456789, "testuser", "Test User")

	// Создаем DB instance
	database := &DB{conn: db}

	// Обновляем язык интерфейса
	err = database.UpdateUserInterfaceLanguage(1, "ru")
	require.NoError(t, err)

	// Проверяем обновление
	user, err := database.GetUserByTelegramID(123456789)
	require.NoError(t, err)
	assert.Equal(t, "ru", user.InterfaceLanguageCode)
}

// TestDB_UpdateUserNativeLanguage тестирует обновление родного языка
func TestDB_UpdateUserNativeLanguage(t *testing.T) {
	// Создаем тестовую базу данных в памяти
	db, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	// Создаем таблицы и пользователя
	setupTestUsersTable(t, db)
	setupTestUser(t, db, 123456789, "testuser", "Test User")

	// Создаем DB instance
	database := &DB{conn: db}

	// Обновляем родной язык
	err = database.UpdateUserNativeLanguage(1, "ru")
	require.NoError(t, err)

	// Проверяем обновление
	user, err := database.GetUserByTelegramID(123456789)
	require.NoError(t, err)
	assert.Equal(t, "ru", user.NativeLanguageCode)
}

// TestDB_UpdateUserTargetLanguage тестирует обновление целевого языка
func TestDB_UpdateUserTargetLanguage(t *testing.T) {
	// Создаем тестовую базу данных в памяти
	db, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	// Создаем таблицы и пользователя
	setupTestUsersTable(t, db)
	setupTestUser(t, db, 123456789, "testuser", "Test User")

	// Создаем DB instance
	database := &DB{conn: db}

	// Обновляем целевой язык
	err = database.UpdateUserTargetLanguage(1, "en")
	require.NoError(t, err)

	// Проверяем обновление
	user, err := database.GetUserByTelegramID(123456789)
	require.NoError(t, err)
	assert.Equal(t, "en", user.TargetLanguageCode)
}

// TestDB_UpdateUserTargetLanguageLevel тестирует обновление уровня целевого языка
func TestDB_UpdateUserTargetLanguageLevel(t *testing.T) {
	// Создаем тестовую базу данных в памяти
	db, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	// Создаем таблицы и пользователя
	setupTestUsersTable(t, db)
	setupTestUser(t, db, 123456789, "testuser", "Test User")

	// Создаем DB instance
	database := &DB{conn: db}

	// Обновляем уровень языка
	err = database.UpdateUserTargetLanguageLevel(1, "intermediate")
	require.NoError(t, err)

	// Проверяем обновление
	user, err := database.GetUserByTelegramID(123456789)
	require.NoError(t, err)
	assert.Equal(t, "intermediate", user.TargetLanguageLevel)
}

// TestDB_RemoveUserInterest тестирует удаление интереса пользователя
func TestDB_RemoveUserInterest(t *testing.T) {
	// Создаем тестовую базу данных в памяти
	db, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	// Создаем таблицы
	setupTestUsersTable(t, db)
	setupTestUserInterestsTable(t, db)
	setupTestUser(t, db, 123456789, "testuser", "Test User")

	// Создаем DB instance
	database := &DB{conn: db}

	// Добавляем интерес
	err = database.SaveUserInterest(1, 1, true)
	require.NoError(t, err)

	// Проверяем, что интерес добавлен
	interests, err := database.GetUserSelectedInterests(1)
	require.NoError(t, err)
	assert.Len(t, interests, 1)
	assert.Equal(t, []int{1}, interests)

	// Удаляем интерес
	err = database.RemoveUserInterest(1, 1)
	require.NoError(t, err)

	// Проверяем, что интерес удален
	interests, err = database.GetUserSelectedInterests(1)
	require.NoError(t, err)
	assert.Len(t, interests, 0)
}

// TestDB_GetInterestByID тестирует получение интереса по ID
func TestDB_GetInterestByID(t *testing.T) {
	// Создаем тестовую базу данных в памяти
	db, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	// Создаем таблицы и интересы
	setupTestInterestsTable(t, db)

	// Создаем DB instance
	database := &DB{conn: db}

	// Получаем интерес по ID
	interest, err := database.GetInterestByID(1)
	require.NoError(t, err)
	require.NotNil(t, interest)

	assert.Equal(t, 1, interest.ID)
	assert.Equal(t, "Programming", interest.KeyName)
}

// TestDB_SaveNativeLanguage тестирует сохранение родного языка
func TestDB_SaveNativeLanguage(t *testing.T) {
	// Создаем тестовую базу данных в памяти
	db, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	// Создаем таблицы и пользователя
	setupTestUsersTable(t, db)
	setupTestUser(t, db, 123456789, "testuser", "Test User")

	// Создаем DB instance
	database := &DB{conn: db}

	// Сохраняем родной язык
	err = database.SaveNativeLanguage(1, "ru")
	require.NoError(t, err)

	// Проверяем сохранение
	user, err := database.GetUserByTelegramID(123456789)
	require.NoError(t, err)
	assert.Equal(t, "ru", user.NativeLanguageCode)
}

// TestDB_SaveTargetLanguage тестирует сохранение целевого языка
func TestDB_SaveTargetLanguage(t *testing.T) {
	// Создаем тестовую базу данных в памяти
	db, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	// Создаем таблицы и пользователя
	setupTestUsersTable(t, db)
	setupTestUser(t, db, 123456789, "testuser", "Test User")

	// Создаем DB instance
	database := &DB{conn: db}

	// Сохраняем целевой язык
	err = database.SaveTargetLanguage(1, "en")
	require.NoError(t, err)

	// Проверяем сохранение
	user, err := database.GetUserByTelegramID(123456789)
	require.NoError(t, err)
	assert.Equal(t, "en", user.TargetLanguageCode)
}

// TestDB_ResetUserProfile тестирует сброс профиля пользователя
func TestDB_ResetUserProfile(t *testing.T) {
	// Создаем тестовую базу данных в памяти
	db, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	// Создаем таблицы и пользователя с заполненным профилем
	setupTestUsersTable(t, db)
	setupTestUserInterestsTable(t, db)
	setupTestUser(t, db, 123456789, "testuser", "Test User")

	// Создаем DB instance и обновляем профиль
	database := &DB{conn: db}

	// Заполняем профиль
	err = database.UpdateUserState(1, "active")
	require.NoError(t, err)
	err = database.UpdateUserStatus(1, "active")
	require.NoError(t, err)
	err = database.UpdateUserInterfaceLanguage(1, "ru")
	require.NoError(t, err)
	err = database.SaveNativeLanguage(1, "ru")
	require.NoError(t, err)
	err = database.SaveTargetLanguage(1, "en")
	require.NoError(t, err)
	err = database.UpdateUserTargetLanguageLevel(1, "intermediate")
	require.NoError(t, err)

	// Проверяем, что профиль заполнен
	user, err := database.GetUserByTelegramID(123456789)
	require.NoError(t, err)
	assert.Equal(t, "active", user.State)
	assert.Equal(t, "active", user.Status)
	assert.Equal(t, "ru", user.InterfaceLanguageCode)
	assert.Equal(t, "ru", user.NativeLanguageCode)
	assert.Equal(t, "en", user.TargetLanguageCode)
	assert.Equal(t, "intermediate", user.TargetLanguageLevel)

	// Сбрасываем профиль
	err = database.ResetUserProfile(1)
	require.NoError(t, err)

	// Проверяем сброс
	user, err = database.GetUserByTelegramID(123456789)
	require.NoError(t, err)
	assert.Equal(t, "waiting_language", user.State)
	assert.Equal(t, "filling_profile", user.Status)
	assert.Equal(t, "", user.NativeLanguageCode)
	assert.Equal(t, "", user.TargetLanguageCode)
	assert.Equal(t, "", user.TargetLanguageLevel)
}

// TestDB_SaveUserFeedback тестирует сохранение отзыва пользователя
func TestDB_SaveUserFeedback(t *testing.T) {
	// Создаем тестовую базу данных в памяти
	db, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	// Создаем таблицы
	setupTestUsersTable(t, db)
	setupTestUserFeedbackTable(t, db)
	setupTestUser(t, db, 123456789, "testuser", "Test User")

	// Создаем DB instance
	database := &DB{conn: db}

	// Проверим, какой ID имеет пользователь
	user, err := database.GetUserByTelegramID(123456789)
	require.NoError(t, err)
	userID := user.ID

	// Сохраняем отзыв
	feedbackText := "This is a test feedback message"
	contactInfo := "test_contact"
	err = database.SaveUserFeedback(userID, feedbackText, &contactInfo)
	require.NoError(t, err)

	// Проверяем сохранение
	feedbacks, err := database.GetUserFeedbackByUserID(userID)
	require.NoError(t, err)
	require.Len(t, feedbacks, 1)

	assert.Equal(t, userID, feedbacks[0]["user_id"])
	assert.Equal(t, feedbackText, feedbacks[0]["feedback_text"])
	assert.Equal(t, contactInfo, feedbacks[0]["contact_info"])
	assert.Equal(t, false, feedbacks[0]["is_processed"])
}

// TestDB_GetUserFeedbackByUserID тестирует получение отзывов пользователя
func TestDB_GetUserFeedbackByUserID(t *testing.T) {
	// Создаем тестовую базу данных в памяти
	db, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	// Создаем таблицы
	setupTestUsersTable(t, db)
	setupTestUserFeedbackTable(t, db)
	setupTestUser(t, db, 123456789, "testuser", "Test User")

	// Создаем DB instance
	database := &DB{conn: db}

	// Получим userID
	user, err := database.GetUserByTelegramID(123456789)
	require.NoError(t, err)
	userID := user.ID

	// Сохраняем несколько отзывов
	contact1 := "contact1"
	contact2 := "contact2"
	err = database.SaveUserFeedback(userID, "First feedback", &contact1)
	require.NoError(t, err)
	err = database.SaveUserFeedback(userID, "Second feedback", &contact2)
	require.NoError(t, err)

	// Получаем отзывы
	feedbacks, err := database.GetUserFeedbackByUserID(userID)
	require.NoError(t, err)
	assert.Len(t, feedbacks, 2)

	assert.Equal(t, "First feedback", feedbacks[0]["feedback_text"])
	assert.Equal(t, "Second feedback", feedbacks[1]["feedback_text"])
}

// TestDB_GetUnprocessedFeedback тестирует получение необработанных отзывов
func TestDB_GetUnprocessedFeedback(t *testing.T) {
	// Создаем тестовую базу данных в памяти
	db, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	// Создаем таблицы
	setupTestUsersTable(t, db)
	setupTestUserFeedbackTable(t, db)
	setupTestUser(t, db, 123456789, "testuser", "Test User")
	setupTestUser(t, db, 987654321, "testuser2", "Test User 2")

	// Создаем DB instance
	database := &DB{conn: db}

	// Получим userIDs
	user1, err := database.GetUserByTelegramID(123456789)
	require.NoError(t, err)
	user2, err := database.GetUserByTelegramID(987654321)
	require.NoError(t, err)

	// Сохраняем отзывы
	contact1 := "contact1"
	contact2 := "contact2"
	err = database.SaveUserFeedback(user1.ID, "Unprocessed feedback 1", &contact1)
	require.NoError(t, err)
	err = database.SaveUserFeedback(user2.ID, "Unprocessed feedback 2", &contact2)
	require.NoError(t, err)

	// Получаем необработанные отзывы
	feedbacks, err := database.GetUnprocessedFeedback()
	require.NoError(t, err)
	assert.Len(t, feedbacks, 2)

	// Помечаем один отзыв как обработанный
	feedbackID := feedbacks[0]["id"].(int)
	err = database.MarkFeedbackProcessed(feedbackID, "Admin response")
	require.NoError(t, err)

	// Проверяем, что остался только один необработанный отзыв
	feedbacks, err = database.GetUnprocessedFeedback()
	require.NoError(t, err)
	assert.Len(t, feedbacks, 1)
	assert.Equal(t, "Unprocessed feedback 2", feedbacks[0]["feedback_text"])
}

// TestDB_MarkFeedbackProcessed тестирует пометку отзыва как обработанного
func TestDB_MarkFeedbackProcessed(t *testing.T) {
	// Создаем тестовую базу данных в памяти
	db, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	// Создаем таблицы
	setupTestUsersTable(t, db)
	setupTestUserFeedbackTable(t, db)
	setupTestUser(t, db, 123456789, "testuser", "Test User")

	// Создаем DB instance
	database := &DB{conn: db}

	// Получим userID
	user, err := database.GetUserByTelegramID(123456789)
	require.NoError(t, err)
	userID := user.ID

	// Сохраняем отзыв
	contact := "contact"
	err = database.SaveUserFeedback(userID, "Test feedback", &contact)
	require.NoError(t, err)

	// Получаем отзыв
	feedbacks, err := database.GetUserFeedbackByUserID(userID)
	require.NoError(t, err)
	require.Len(t, feedbacks, 1)

	feedbackID := feedbacks[0]["id"].(int)
	assert.Equal(t, false, feedbacks[0]["is_processed"])

	// Помечаем как обработанный
	err = database.MarkFeedbackProcessed(feedbackID, "Admin response")
	require.NoError(t, err)

	// Проверяем обновление
	feedbacks, err = database.GetUserFeedbackByUserID(userID)
	require.NoError(t, err)
	require.Len(t, feedbacks, 1)

	assert.Equal(t, true, feedbacks[0]["is_processed"])
	assert.Equal(t, "Admin response", feedbacks[0]["admin_response"])
}

// ===== Helper functions for test setup =====

// setupTestUserFeedbackTable creates user_feedback table for testing
func setupTestUserFeedbackTable(t *testing.T, db *sql.DB) {
	_, err := db.Exec(`
		CREATE TABLE user_feedback (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			feedback_text TEXT NOT NULL,
			contact_info TEXT,
			is_processed BOOLEAN DEFAULT 0,
			admin_response TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id)
		)
	`)
	require.NoError(t, err)
}

// setupTestLanguagesTable creates and populates languages table for testing
func setupTestLanguagesTable(t *testing.T, db *sql.DB) {
	// Create languages table
	_, err := db.Exec(`
		CREATE TABLE languages (
			id INTEGER PRIMARY KEY,
			code TEXT UNIQUE NOT NULL,
			name_native TEXT NOT NULL,
			name_en TEXT NOT NULL,
			is_interface_language BOOLEAN DEFAULT FALSE
		)
	`)
	require.NoError(t, err)

	// Insert test data
	_, err = db.Exec(`
		INSERT INTO languages (id, code, name_native, name_en, is_interface_language) VALUES
		(1, 'en', 'English', 'English', 1),
		(2, 'ru', 'Русский', 'Russian', 1)
	`)
	require.NoError(t, err)
}

// setupTestInterestsTable creates and populates interests table for testing
func setupTestInterestsTable(t *testing.T, db *sql.DB) {
	// Create interests table
	_, err := db.Exec(`
		CREATE TABLE interests (
			id INTEGER PRIMARY KEY,
			key_name TEXT UNIQUE NOT NULL,
			category_id INTEGER DEFAULT 1,
			display_order INTEGER DEFAULT 1,
			type TEXT DEFAULT 'primary',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	require.NoError(t, err)

	// Insert test data
	_, err = db.Exec(`
		INSERT INTO interests (id, key_name, category_id, display_order, type) VALUES
		(1, 'Programming', 1, 1, 'primary'),
		(2, 'Music', 2, 1, 'primary'),
		(3, 'Sports', 3, 1, 'primary')
	`)
	require.NoError(t, err)
}

// setupTestUsersTable creates users table for testing
func setupTestUsersTable(t *testing.T, db *sql.DB) {
	_, err := db.Exec(`
		CREATE TABLE users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			telegram_id INTEGER UNIQUE NOT NULL,
			username TEXT,
			first_name TEXT NOT NULL,
			native_language_code TEXT,
			target_language_code TEXT,
			target_language_level TEXT,
			interface_language_code TEXT DEFAULT 'en',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			state TEXT DEFAULT 'new',
			profile_completion_level INTEGER DEFAULT 0,
			status TEXT DEFAULT 'new'
		)
	`)
	require.NoError(t, err)
}

// setupTestUserInterestsTable creates user_interests table for testing
func setupTestUserInterestsTable(t *testing.T, db *sql.DB) {
	_, err := db.Exec(`
		CREATE TABLE user_interests (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			interest_id INTEGER NOT NULL,
			is_primary BOOLEAN DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id),
			UNIQUE(user_id, interest_id)
		)
	`)
	require.NoError(t, err)
}

// setupTestUser creates a test user in the database
func setupTestUser(t *testing.T, db *sql.DB, telegramID int64, username, firstName string) {
	_, err := db.Exec(`
		INSERT INTO users (telegram_id, username, first_name, interface_language_code, state, status)
		VALUES (?, ?, ?, 'en', 'new', 'new')
	`, telegramID, username, firstName)
	require.NoError(t, err)
}
