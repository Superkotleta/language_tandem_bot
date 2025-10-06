package database

import (
	"testing"

	"language-exchange-bot/internal/models"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpdateUser(t *testing.T) {
	// Создаем мок базы данных
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	// Создаем тестового пользователя
	user := &models.User{
		ID:                     1,
		TelegramID:             12345,
		Username:               "testuser",
		FirstName:              "Test",
		NativeLanguageCode:     "ru",
		TargetLanguageCode:     "en",
		TargetLanguageLevel:    "intermediate",
		InterfaceLanguageCode:  "ru",
		State:                  "idle",
		Status:                 "active",
		ProfileCompletionLevel: 80,
	}

	// Ожидаем выполнения UPDATE запроса
	mock.ExpectExec(`UPDATE users SET username = \$1, first_name = \$2, native_language_code = \$3,
		    target_language_code = \$4, target_language_level = \$5,
		    interface_language_code = \$6, state = \$7, status = \$8,
		    profile_completion_level = \$9, updated_at = NOW\(\)
		WHERE id = \$10`).
		WithArgs(user.Username, user.FirstName, user.NativeLanguageCode,
			user.TargetLanguageCode, user.TargetLanguageLevel,
			user.InterfaceLanguageCode, user.State, user.Status,
			user.ProfileCompletionLevel, user.ID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Создаем экземпляр DB с моком
	testDB := &DB{
		conn: db,
	}

	// Выполняем тест
	err = testDB.UpdateUser(user)

	// Проверяем результат
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRemoveUserInterest(t *testing.T) {
	// Создаем мок базы данных
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	userID := 1
	interestID := 5

	// Ожидаем выполнения DELETE запроса
	mock.ExpectExec(`DELETE FROM user_interests
        WHERE user_id = \$1 AND interest_id = \$2`).
		WithArgs(userID, interestID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Создаем экземпляр DB с моком
	testDB := &DB{
		conn: db,
	}

	// Выполняем тест
	err = testDB.RemoveUserInterest(userID, interestID)

	// Проверяем результат
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestMarkFeedbackProcessed(t *testing.T) {
	// Создаем мок базы данных
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	feedbackID := 1
	adminResponse := "Спасибо за отзыв!"

	// Ожидаем выполнения UPDATE запроса
	mock.ExpectExec(`UPDATE user_feedback
        SET is_processed = true, admin_response = \$1, updated_at = NOW\(\)
        WHERE id = \$2`).
		WithArgs(adminResponse, feedbackID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Создаем экземпляр DB с моком
	testDB := &DB{
		conn: db,
	}

	// Выполняем тест
	err = testDB.MarkFeedbackProcessed(feedbackID, adminResponse)

	// Проверяем результат
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateUserState(t *testing.T) {
	// Создаем мок базы данных
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	userID := 1
	newState := "waiting_feedback"

	// Ожидаем выполнения UPDATE запроса
	mock.ExpectExec(`UPDATE users SET state = \$1, updated_at = NOW\(\) WHERE id = \$2`).
		WithArgs(newState, userID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Создаем экземпляр DB с моком
	testDB := &DB{
		conn: db,
	}

	// Выполняем тест
	err = testDB.UpdateUserState(userID, newState)

	// Проверяем результат
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestClearUserInterests(t *testing.T) {
	// Создаем мок базы данных
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	userID := 1

	// Ожидаем выполнения DELETE запроса
	mock.ExpectExec(`DELETE FROM user_interest_selections WHERE user_id = \$1`).
		WithArgs(userID).
		WillReturnResult(sqlmock.NewResult(0, 5)) // Удалено 5 записей

	// Создаем экземпляр DB с моком
	testDB := &DB{
		conn: db,
	}

	// Выполняем тест
	err = testDB.ClearUserInterests(userID)

	// Проверяем результат
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
