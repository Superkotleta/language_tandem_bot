package database

import (
	"context"
	"database/sql"
	"fmt"
	"language-exchange-bot/internal/models"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// OptimizedDB оптимизированная версия для работы с PostgreSQL.
type OptimizedDB struct {
	pool   *pgxpool.Pool
	ctx    context.Context
	cancel context.CancelFunc
}

// NewOptimizedDB создает оптимизированное соединение с БД.
func NewOptimizedDB(databaseURL string) (*OptimizedDB, error) {
	ctx, cancel := context.WithCancel(context.Background())

	// Настройки connection pool
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to parse database URL: %w", err)
	}

	// Оптимизация настроек пула соединений
	config.MaxConns = 25                      // Максимум соединений
	config.MinConns = 5                       // Минимум соединений
	config.MaxConnLifetime = time.Hour        // Время жизни соединения
	config.MaxConnIdleTime = time.Minute * 30 // Время простоя соединения
	config.HealthCheckPeriod = time.Minute    // Период проверки здоровья

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Проверяем соединение
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		cancel()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &OptimizedDB{
		pool:   pool,
		ctx:    ctx,
		cancel: cancel,
	}, nil
}

// Close закрывает соединение с БД.
func (db *OptimizedDB) Close() error {
	db.cancel()
	db.pool.Close()
	return nil
}

// GetConnection возвращает соединение из пула.
func (db *OptimizedDB) GetConnection() *sql.DB {
	// Для совместимости с существующим кодом
	// В реальном проекте лучше перейти на pgx полностью
	return nil
}

// GetPool возвращает пул соединений pgx.
func (db *OptimizedDB) GetPool() *pgxpool.Pool {
	return db.pool
}

// FindOrCreateUser оптимизированное создание/поиск пользователя.
func (db *OptimizedDB) FindOrCreateUser(telegramID int64, username, firstName string) (*models.User, error) {
	ctx, cancel := context.WithTimeout(db.ctx, 5*time.Second)
	defer cancel()

	query := `
        INSERT INTO users (telegram_id, username, first_name, interface_language_code)
        VALUES ($1, $2, $3, 'en')
        ON CONFLICT (telegram_id) DO UPDATE SET
            username = EXCLUDED.username,
            first_name = EXCLUDED.first_name,
            updated_at = NOW()
        RETURNING id, telegram_id, username, first_name,
        COALESCE(native_language_code, '') as native_language_code,
        COALESCE(target_language_code, '') as target_language_code,
        COALESCE(target_language_level, '') as target_language_level,
        interface_language_code, created_at, updated_at, state,
        profile_completion_level, status
    `

	user := &models.User{}
	err := db.pool.QueryRow(ctx, query, telegramID, username, firstName).Scan(
		&user.ID, &user.TelegramID, &user.Username, &user.FirstName,
		&user.NativeLanguageCode, &user.TargetLanguageCode, &user.TargetLanguageLevel,
		&user.InterfaceLanguageCode, &user.CreatedAt, &user.UpdatedAt,
		&user.State, &user.ProfileCompletionLevel, &user.Status,
	)

	return user, err
}

// UpdateUserProfileBatch обновляет профиль пользователя в одной транзакции.
func (db *OptimizedDB) UpdateUserProfileBatch(userID int, updates map[string]interface{}) error {
	ctx, cancel := context.WithTimeout(db.ctx, 10*time.Second)
	defer cancel()

	tx, err := db.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Обновляем основные поля пользователя
	if len(updates) > 0 {
		setParts := []string{}
		args := []interface{}{}
		argIndex := 1

		for field, value := range updates {
			if field == "interests" {
				continue // Обрабатываем отдельно
			}
			setParts = append(setParts, fmt.Sprintf("%s = $%d", field, argIndex))
			args = append(args, value)
			argIndex++
		}

		if len(setParts) > 0 {
			setParts = append(setParts, "updated_at = NOW()")
			query := fmt.Sprintf("UPDATE users SET %s WHERE id = $%d",
				setParts[0], argIndex)

			for i := 1; i < len(setParts); i++ {
				query += fmt.Sprintf(", %s", setParts[i])
			}
			query += fmt.Sprintf(" WHERE id = $%d", argIndex)
			args = append(args, userID)

			_, err = tx.Exec(ctx, query, args...)
			if err != nil {
				return fmt.Errorf("failed to update user: %w", err)
			}
		}
	}

	// Обрабатываем интересы если есть
	if interests, ok := updates["interests"].([]int); ok {
		// Удаляем старые интересы
		_, err = tx.Exec(ctx, "DELETE FROM user_interests WHERE user_id = $1", userID)
		if err != nil {
			return fmt.Errorf("failed to clear user interests: %w", err)
		}

		// Добавляем новые интересы batch операцией
		if len(interests) > 0 {
			batch := &pgx.Batch{}
			for _, interestID := range interests {
				batch.Queue("INSERT INTO user_interests (user_id, interest_id, is_primary, created_at) VALUES ($1, $2, $3, NOW())",
					userID, interestID, false)
			}

			results := tx.SendBatch(ctx, batch)
			defer results.Close()

			for i := 0; i < len(interests); i++ {
				_, err = results.Exec()
				if err != nil {
					return fmt.Errorf("failed to insert user interest: %w", err)
				}
			}
		}
	}

	return tx.Commit(ctx)
}

// GetUserSelectedInterests оптимизированное получение интересов пользователя.
func (db *OptimizedDB) GetUserSelectedInterests(userID int) ([]int, error) {
	ctx, cancel := context.WithTimeout(db.ctx, 3*time.Second)
	defer cancel()

	query := `SELECT interest_id FROM user_interests WHERE user_id = $1 ORDER BY interest_id`
	rows, err := db.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var interests []int
	for rows.Next() {
		var interestID int
		if err := rows.Scan(&interestID); err != nil {
			return nil, err
		}
		interests = append(interests, interestID)
	}

	return interests, nil
}

// SaveUserInterestsBatch сохраняет интересы пользователя batch операцией.
func (db *OptimizedDB) SaveUserInterestsBatch(userID int, interestIDs []int) error {
	ctx, cancel := context.WithTimeout(db.ctx, 5*time.Second)
	defer cancel()

	tx, err := db.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Удаляем старые интересы
	_, err = tx.Exec(ctx, "DELETE FROM user_interests WHERE user_id = $1", userID)
	if err != nil {
		return fmt.Errorf("failed to clear user interests: %w", err)
	}

	// Добавляем новые интересы batch операцией
	if len(interestIDs) > 0 {
		batch := &pgx.Batch{}
		for _, interestID := range interestIDs {
			batch.Queue("INSERT INTO user_interests (user_id, interest_id, is_primary, created_at) VALUES ($1, $2, $3, NOW())",
				userID, interestID, false)
		}

		results := tx.SendBatch(ctx, batch)
		defer results.Close()

		for i := 0; i < len(interestIDs); i++ {
			_, err = results.Exec()
			if err != nil {
				return fmt.Errorf("failed to insert user interest: %w", err)
			}
		}
	}

	return tx.Commit(ctx)
}

// GetUnprocessedFeedbackBatch получает необработанные отзывы с пагинацией.
func (db *OptimizedDB) GetUnprocessedFeedbackBatch(limit, offset int) ([]map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(db.ctx, 5*time.Second)
	defer cancel()

	query := `
        SELECT uf.id, uf.feedback_text, uf.contact_info, uf.created_at,
               u.username, u.telegram_id, u.first_name
        FROM user_feedback uf
        JOIN users u ON uf.user_id = u.id
        WHERE uf.is_processed = false
        ORDER BY uf.created_at ASC
        LIMIT $1 OFFSET $2
    `

	rows, err := db.pool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var feedbacks []map[string]interface{}
	for rows.Next() {
		var (
			id           int
			feedbackText string
			contactInfo  *string
			createdAt    time.Time
			username     *string
			telegramID   int64
			firstName    string
		)

		err := rows.Scan(&id, &feedbackText, &contactInfo, &createdAt, &username, &telegramID, &firstName)
		if err != nil {
			continue
		}

		feedback := map[string]interface{}{
			"id":            id,
			"feedback_text": feedbackText,
			"created_at":    createdAt,
			"telegram_id":   telegramID,
			"first_name":    firstName,
		}

		if username != nil {
			feedback["username"] = *username
		} else {
			feedback["username"] = nil
		}

		if contactInfo != nil {
			feedback["contact_info"] = *contactInfo
		} else {
			feedback["contact_info"] = nil
		}

		feedbacks = append(feedbacks, feedback)
	}

	return feedbacks, nil
}

// GetAllFeedback получает все отзывы для администрирования.
func (db *OptimizedDB) GetAllFeedback() ([]map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(db.ctx, 5*time.Second)
	defer cancel()

	query := `
        SELECT uf.id, uf.feedback_text, uf.contact_info, uf.created_at, uf.is_processed, uf.admin_response,
               u.username, u.telegram_id, u.first_name
        FROM user_feedback uf
        JOIN users u ON uf.user_id = u.id
        ORDER BY uf.created_at DESC
    `

	rows, err := db.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var feedbacks []map[string]interface{}
	for rows.Next() {
		var (
			id            int
			feedbackText  string
			contactInfo   sql.NullString
			createdAt     sql.NullTime
			isProcessed   bool
			adminResponse sql.NullString
			username      sql.NullString
			telegramID    int64
			firstName     string
		)

		err := rows.Scan(&id, &feedbackText, &contactInfo, &createdAt, &isProcessed, &adminResponse, &username, &telegramID, &firstName)
		if err != nil {
			continue // Пропускаем ошибочные записи
		}

		feedback := map[string]interface{}{
			"id":            id,
			"feedback_text": feedbackText,
			"created_at":    createdAt.Time,
			"is_processed":  isProcessed,
			"telegram_id":   telegramID,
			"first_name":    firstName,
		}

		if username.Valid {
			feedback["username"] = username.String
		} else {
			feedback["username"] = nil
		}

		if contactInfo.Valid {
			feedback["contact_info"] = contactInfo.String
		} else {
			feedback["contact_info"] = nil
		}

		if adminResponse.Valid {
			feedback["admin_response"] = adminResponse.String
		} else {
			feedback["admin_response"] = nil
		}

		feedbacks = append(feedbacks, feedback)
	}

	return feedbacks, nil
}

// MarkFeedbackProcessedBatch помечает несколько отзывов как обработанные.
func (db *OptimizedDB) MarkFeedbackProcessedBatch(feedbackIDs []int, adminResponse string) error {
	ctx, cancel := context.WithTimeout(db.ctx, 5*time.Second)
	defer cancel()

	if len(feedbackIDs) == 0 {
		return nil
	}

	// Создаем плейсхолдеры для IN clause
	placeholders := make([]string, len(feedbackIDs))
	args := make([]interface{}, len(feedbackIDs)+1)

	for i, id := range feedbackIDs {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = id
	}
	args[len(feedbackIDs)] = adminResponse

	query := fmt.Sprintf(`
        UPDATE user_feedback
        SET is_processed = true, admin_response = $%d, updated_at = NOW()
        WHERE id IN (%s)
    `, len(feedbackIDs)+1, placeholders[0])

	for i := 1; i < len(placeholders); i++ {
		query += fmt.Sprintf(", %s", placeholders[i])
	}

	_, err := db.pool.Exec(ctx, query, args...)
	return err
}

// GetUserDataForFeedback получает данные пользователя для формирования уведомления о новом отзыве.
func (db *OptimizedDB) GetUserDataForFeedback(userID int) (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(db.ctx, 3*time.Second)
	defer cancel()

	query := `
        SELECT telegram_id, username, first_name
        FROM users
        WHERE id = $1
    `

	var (
		telegramID int64
		username   sql.NullString
		firstName  sql.NullString
	)

	err := db.pool.QueryRow(ctx, query, userID).Scan(&telegramID, &username, &firstName)
	if err != nil {
		return nil, err
	}

	userData := map[string]interface{}{
		"telegram_id": telegramID,
		"first_name":  firstName.String,
	}

	if username.Valid {
		userData["username"] = &username.String
	} else {
		userData["username"] = nil
	}

	return userData, nil
}

// DeleteFeedback удаляет отзыв из базы данных.
func (db *OptimizedDB) DeleteFeedback(feedbackID int) error {
	ctx, cancel := context.WithTimeout(db.ctx, 3*time.Second)
	defer cancel()

	query := `DELETE FROM user_feedback WHERE id = $1`
	_, err := db.pool.Exec(ctx, query, feedbackID)
	return err
}

// ArchiveFeedback архивирует отзыв (помечает как обработанный).
func (db *OptimizedDB) ArchiveFeedback(feedbackID int) error {
	ctx, cancel := context.WithTimeout(db.ctx, 3*time.Second)
	defer cancel()

	query := `UPDATE user_feedback SET is_processed = true WHERE id = $1`
	_, err := db.pool.Exec(ctx, query, feedbackID)
	return err
}

// UnarchiveFeedback разархивирует отзыв (помечает как необработанный).
func (db *OptimizedDB) UnarchiveFeedback(feedbackID int) error {
	ctx, cancel := context.WithTimeout(db.ctx, 3*time.Second)
	defer cancel()

	query := `UPDATE user_feedback SET is_processed = false WHERE id = $1`
	_, err := db.pool.Exec(ctx, query, feedbackID)
	return err
}

// UpdateFeedbackStatus обновляет статус отзыва.
func (db *OptimizedDB) UpdateFeedbackStatus(feedbackID int, isProcessed bool) error {
	ctx, cancel := context.WithTimeout(db.ctx, 3*time.Second)
	defer cancel()

	query := `UPDATE user_feedback SET is_processed = $1 WHERE id = $2`
	_, err := db.pool.Exec(ctx, query, isProcessed, feedbackID)
	return err
}

// DeleteAllProcessedFeedbacks удаляет все обработанные отзывы.
func (db *OptimizedDB) DeleteAllProcessedFeedbacks() (int, error) {
	ctx, cancel := context.WithTimeout(db.ctx, 5*time.Second)
	defer cancel()

	query := `DELETE FROM user_feedback WHERE is_processed = true`
	result, err := db.pool.Exec(ctx, query)
	if err != nil {
		return 0, err
	}

	rowsAffected := result.RowsAffected()
	return int(rowsAffected), nil
}

// HealthCheck проверяет здоровье соединения с БД.
func (db *OptimizedDB) HealthCheck() error {
	ctx, cancel := context.WithTimeout(db.ctx, 2*time.Second)
	defer cancel()

	return db.pool.Ping(ctx)
}

// GetStats возвращает статистику пула соединений.
func (db *OptimizedDB) GetStats() map[string]interface{} {
	stats := db.pool.Stat()
	return map[string]interface{}{
		"max_conns":           stats.MaxConns(),
		"acquired_conns":      stats.AcquiredConns(),
		"constructing_conns":  stats.ConstructingConns(),
		"idle_conns":          stats.IdleConns(),
		"total_conns":         stats.TotalConns(),
		"new_conns_count":     stats.NewConnsCount(),
		"acquire_duration":    stats.AcquireDuration(),
		"acquire_count":       stats.AcquireCount(),
		"empty_acquire_count": stats.EmptyAcquireCount(),
	}
}
