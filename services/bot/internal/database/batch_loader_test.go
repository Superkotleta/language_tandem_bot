package database //nolint:testpackage

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBatchLoader_scanUserWithInterestRow(t *testing.T) {
	t.Parallel()
	// Создаем тестовую базу данных в памяти
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)

	defer func() { _ = db.Close() }()

	// Создаем таблицу для тестирования
	_, err = db.ExecContext(context.Background(), `
		CREATE TABLE test_users (
			id INTEGER PRIMARY KEY,
			telegram_id INTEGER,
			username TEXT,
			first_name TEXT,
			native_language_code TEXT,
			target_language_code TEXT,
			target_language_level TEXT,
			interface_language_code TEXT,
			created_at DATETIME,
			updated_at DATETIME,
			state TEXT,
			profile_completion_level INTEGER,
			status TEXT
		)
	`)
	require.NoError(t, err)

	// Создаем BatchLoader
	batchLoader := &BatchLoader{db: &DB{conn: db}}

	// Вставляем тестовые данные
	query := "\n\t\tINSERT INTO test_users (\n\t\t\tid, telegram_id, username, first_name,\n\t\t\tnative_language_code, target_language_code, target_language_level,\n\t\t\tinterface_language_code, created_at, updated_at,\n\t\t\tstate, profile_completion_level, status\n\t\t) VALUES (" +
		"1, 12345, 'testuser', 'Test', 'ru', 'en', 'intermediate', 'ru', \n\t\t\tdatetime('now'), 'active', 100, 'active')\n\t"
	_, err = db.ExecContext(context.Background(), query)
	require.NoError(t, err)

	// Выполняем запрос
	rows, err := db.QueryContext(context.Background(), `
		SELECT id, telegram_id, username, first_name,
			native_language_code, target_language_code, target_language_level,
			interface_language_code, created_at, updated_at,
			state, profile_completion_level, status, 5 as interest_id
		FROM test_users WHERE id = 1
	`)
	require.NoError(t, err)

	defer func() { _ = rows.Close() }()

	// Сканируем строку
	require.True(t, rows.Next())
	user, interestID, err := batchLoader.scanUserWithInterestRow(rows)

	// Проверяем результаты
	assert.NoError(t, err)
	assert.Equal(t, 1, user.ID)
	assert.Equal(t, int64(12345), user.TelegramID)
	assert.Equal(t, "testuser", user.Username)
	assert.Equal(t, "Test", user.FirstName)
	assert.True(t, interestID.Valid)
	assert.Equal(t, int64(5), interestID.Int64)
}

func TestBatchLoader_handleRowsError(t *testing.T) {
	t.Parallel()
	// Создаем тестовую базу данных в памяти
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)

	defer func() { _ = db.Close() }()

	// Создаем BatchLoader
	batchLoader := &BatchLoader{db: &DB{conn: db}}

	// Тест 1: Нет ошибки
	rows, err := db.QueryContext(context.Background(), "SELECT 1 as id")
	require.NoError(t, err)

	defer func() { _ = rows.Close() }()

	err = batchLoader.handleRowsError(rows, "TestOperation")
	assert.NoError(t, err)

	// Тест 2: Есть ошибка (закрываем rows чтобы вызвать ошибку)
	_ = rows.Close()
	err = batchLoader.handleRowsError(rows, "TestOperation")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "rows error in TestOperation")
}

func TestBatchLoader_BatchLoadUsersWithInterests_EmptyInput(t *testing.T) {
	t.Parallel()
	// Создаем тестовую базу данных в памяти
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)

	defer func() { _ = db.Close() }()

	// Создаем BatchLoader
	batchLoader := &BatchLoader{db: &DB{conn: db}}

	// Тестируем с пустым входом
	result, err := batchLoader.BatchLoadUsersWithInterests(context.Background(), []int64{})

	// Проверяем результаты
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Empty(t, result)
}

func TestBatchLoader_ContextTimeout(t *testing.T) {
	t.Parallel()
	// Создаем тестовую базу данных в памяти
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)

	defer func() { _ = db.Close() }()

	// Создаем BatchLoader
	batchLoader := &BatchLoader{db: &DB{conn: db}}

	// Создаем контекст с очень коротким таймаутом
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	// Ждем истечения таймаута
	time.Sleep(1 * time.Millisecond)

	// Выполняем тест
	result, err := batchLoader.BatchLoadUsersWithInterests(ctx, []int64{12345})

	// Проверяем результаты
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "context deadline exceeded")
}

func TestBatchLoader_Constants(t *testing.T) {
	t.Parallel()
	// Проверяем, что константы определены
	assert.Equal(t, 30*time.Second, DefaultQueryTimeout)
	assert.Equal(t, 1000, MaxBatchSize)
}

func TestBatchLoader_NewBatchLoader(t *testing.T) {
	t.Parallel()
	// Создаем тестовую базу данных в памяти
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)

	defer func() { _ = db.Close() }()

	// Создаем BatchLoader
	batchLoader := NewBatchLoader(&DB{conn: db})

	// Проверяем, что BatchLoader создан
	assert.NotNil(t, batchLoader)
	assert.NotNil(t, batchLoader.db)
}

func TestBatchLoader_HelperFunctions(t *testing.T) {
	t.Parallel()
	// Создаем тестовую базу данных в памяти
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)

	defer func() { _ = db.Close() }()

	// Создаем BatchLoader
	batchLoader := &BatchLoader{db: &DB{conn: db}}

	// Тестируем closeRowsSafely с nil rows
	batchLoader.closeRowsSafely(nil, "TestOperation")

	// Тестируем handleRowsError с nil rows
	err = batchLoader.handleRowsError(nil, "TestOperation")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "rows error in TestOperation")
}
