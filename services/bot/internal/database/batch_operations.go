package database

import (
	"context"
	"fmt"
	"language-exchange-bot/internal/models"
	"log"
	"strings"
	"time"
)

// BatchOperations предоставляет методы для массовых операций с базой данных.
type BatchOperations struct {
	db *DB
}

// NewBatchOperations создает новый экземпляр BatchOperations.
func NewBatchOperations(db *DB) *BatchOperations {
	return &BatchOperations{db: db}
}

// BatchInsertUsers выполняет массовую вставку пользователей.
func (bo *BatchOperations) BatchInsertUsers(ctx context.Context, users []*models.User) error {
	if len(users) == 0 {
		return nil
	}

	// Используем COPY для максимальной производительности
	query := `
		COPY users (telegram_id, username, first_name, native_language_code, 
		           target_language_code, target_language_level, interface_language_code, 
		           state, status, profile_completion_level, created_at, updated_at)
		FROM STDIN WITH (FORMAT csv, HEADER false)`

	tx, err := bo.db.conn.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			log.Printf("Failed to rollback transaction in BatchInsertUsers: %v", rollbackErr)
		}
	}()

	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}

	defer func() {
		if closeErr := stmt.Close(); closeErr != nil {
			log.Printf("Failed to close prepared statement in BatchInsertUsers: %v", closeErr)
		}
	}()

	// Подготавливаем данные для COPY
	var rows []string

	for _, user := range users {
		row := fmt.Sprintf("%d,%s,%s,%s,%s,%s,%s,%s,%s,%d,%s,%s",
			user.TelegramID,
			escapeCSV(user.Username),
			escapeCSV(user.FirstName),
			escapeCSV(user.NativeLanguageCode),
			escapeCSV(user.TargetLanguageCode),
			escapeCSV(user.TargetLanguageLevel),
			escapeCSV(user.InterfaceLanguageCode),
			escapeCSV(user.State),
			escapeCSV(user.Status),
			user.ProfileCompletionLevel,
			user.CreatedAt.Format(time.RFC3339),
			user.UpdatedAt.Format(time.RFC3339),
		)
		rows = append(rows, row)
	}

	// Выполняем COPY операцию
	_, err = stmt.ExecContext(ctx, strings.Join(rows, "\n"))
	if err != nil {
		return fmt.Errorf("failed to execute COPY: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	log.Printf("Batch inserted %d users", len(users))

	return nil
}

// BatchUpdateUsers выполняет массовое обновление пользователей.
func (bo *BatchOperations) BatchUpdateUsers(ctx context.Context, users []*models.User) error {
	if len(users) == 0 {
		return nil
	}

	// Используем VALUES для массового обновления
	query := `
		UPDATE users SET 
			username = data.username,
			first_name = data.first_name,
			native_language_code = data.native_language_code,
			target_language_code = data.target_language_code,
			target_language_level = data.target_language_level,
			interface_language_code = data.interface_language_code,
			state = data.state,
			status = data.status,
			profile_completion_level = data.profile_completion_level,
			updated_at = data.updated_at
		FROM (VALUES %s) AS data(id, username, first_name, native_language_code, 
		                        target_language_code, target_language_level, 
		                        interface_language_code, state, status, 
		                        profile_completion_level, updated_at)
		WHERE users.id = data.id`

	// Подготавливаем VALUES
	var values []string

	for _, user := range users {
		value := fmt.Sprintf("(%d, %s, %s, %s, %s, %s, %s, %s, %s, %d, %s)",
			user.ID,
			escapeSQL(user.Username),
			escapeSQL(user.FirstName),
			escapeSQL(user.NativeLanguageCode),
			escapeSQL(user.TargetLanguageCode),
			escapeSQL(user.TargetLanguageLevel),
			escapeSQL(user.InterfaceLanguageCode),
			escapeSQL(user.State),
			escapeSQL(user.Status),
			user.ProfileCompletionLevel,
			escapeSQL(user.UpdatedAt.Format(time.RFC3339)),
		)
		values = append(values, value)
	}

	// Выполняем обновление
	finalQuery := fmt.Sprintf(query, strings.Join(values, ","))

	_, err := bo.db.conn.ExecContext(ctx, finalQuery)
	if err != nil {
		return fmt.Errorf("failed to execute batch update: %w", err)
	}

	log.Printf("Batch updated %d users", len(users))

	return nil
}

// BatchInsertInterests выполняет массовую вставку интересов.
func (bo *BatchOperations) BatchInsertInterests(ctx context.Context, interests []*models.Interest) error {
	if len(interests) == 0 {
		return nil
	}

	query := `
		INSERT INTO interests (key_name, category_id, display_order, type, created_at)
		VALUES %s
		ON CONFLICT (key_name) DO NOTHING`

	// Подготавливаем VALUES
	var values []string

	for _, interest := range interests {
		value := fmt.Sprintf("(%s, %d, %d, %s, %s)",
			escapeSQL(interest.KeyName),
			interest.CategoryID,
			interest.DisplayOrder,
			escapeSQL(interest.Type),
			escapeSQL(interest.CreatedAt.Format(time.RFC3339)),
		)
		values = append(values, value)
	}

	finalQuery := fmt.Sprintf(query, strings.Join(values, ","))

	_, err := bo.db.conn.ExecContext(ctx, finalQuery)
	if err != nil {
		return fmt.Errorf("failed to execute batch insert interests: %w", err)
	}

	log.Printf("Batch inserted %d interests", len(interests))

	return nil
}

// BatchInsertUserInterests выполняет массовую вставку связей пользователь-интерес.
func (bo *BatchOperations) BatchInsertUserInterests(ctx context.Context, userID int, interestIDs []int) error {
	if len(interestIDs) == 0 {
		return nil
	}

	query := `
		INSERT INTO user_interests (user_id, interest_id, created_at)
		VALUES %s
		ON CONFLICT (user_id, interest_id) DO NOTHING`

	// Подготавливаем VALUES
	var values []string

	now := time.Now().Format(time.RFC3339)
	for _, interestID := range interestIDs {
		value := fmt.Sprintf("(%d, %d, %s)",
			userID,
			interestID,
			escapeSQL(now),
		)
		values = append(values, value)
	}

	finalQuery := fmt.Sprintf(query, strings.Join(values, ","))

	_, err := bo.db.conn.ExecContext(ctx, finalQuery)
	if err != nil {
		return fmt.Errorf("failed to execute batch insert user interests: %w", err)
	}

	log.Printf("Batch inserted %d user interests for user %d", len(interestIDs), userID)

	return nil
}

// BatchDeleteUserInterests выполняет массовое удаление связей пользователь-интерес.
func (bo *BatchOperations) BatchDeleteUserInterests(ctx context.Context, userID int, interestIDs []int) error {
	if len(interestIDs) == 0 {
		return nil
	}

	// Создаем плейсхолдеры для IN clause
	placeholders := make([]string, len(interestIDs))
	args := make([]interface{}, len(interestIDs)+1)
	args[0] = userID

	for i, interestID := range interestIDs {
		placeholders[i] = fmt.Sprintf("$%d", i+2)
		args[i+1] = interestID
	}

	query := fmt.Sprintf(`
		DELETE FROM user_interests 
		WHERE user_id = $1 AND interest_id IN (%s)`,
		strings.Join(placeholders, ","))

	_, err := bo.db.conn.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to execute batch delete user interests: %w", err)
	}

	log.Printf("Batch deleted %d user interests for user %d", len(interestIDs), userID)

	return nil
}

// BatchInsertLanguages выполняет массовую вставку языков.
func (bo *BatchOperations) BatchInsertLanguages(ctx context.Context, languages []*models.Language) error {
	if len(languages) == 0 {
		return nil
	}

	query := `
		INSERT INTO languages (code, name, native_name, created_at)
		VALUES %s
		ON CONFLICT (code) DO NOTHING`

	// Подготавливаем VALUES
	var values []string

	for _, language := range languages {
		value := fmt.Sprintf("(%s, %s, %s, %s)",
			escapeSQL(language.Code),
			escapeSQL(language.NameEn),
			escapeSQL(language.NameNative),
			escapeSQL(language.CreatedAt.Format(time.RFC3339)),
		)
		values = append(values, value)
	}

	finalQuery := fmt.Sprintf(query, strings.Join(values, ","))

	_, err := bo.db.conn.ExecContext(ctx, finalQuery)
	if err != nil {
		return fmt.Errorf("failed to execute batch insert languages: %w", err)
	}

	log.Printf("Batch inserted %d languages", len(languages))

	return nil
}

// BatchUpdateUserStats выполняет массовое обновление статистики пользователей.
func (bo *BatchOperations) BatchUpdateUserStats(ctx context.Context, userStats []UserStats) error {
	if len(userStats) == 0 {
		return nil
	}

	query := `
		UPDATE users SET 
			profile_completion_level = data.profile_completion_level,
			updated_at = data.updated_at
		FROM (VALUES %s) AS data(id, profile_completion_level, updated_at)
		WHERE users.id = data.id`

	// Подготавливаем VALUES
	var values []string

	for _, stats := range userStats {
		value := fmt.Sprintf("(%d, %d, %s)",
			stats.UserID,
			stats.ProfileCompletionLevel,
			escapeSQL(stats.UpdatedAt.Format(time.RFC3339)),
		)
		values = append(values, value)
	}

	finalQuery := fmt.Sprintf(query, strings.Join(values, ","))

	_, err := bo.db.conn.ExecContext(ctx, finalQuery)
	if err != nil {
		return fmt.Errorf("failed to execute batch update user stats: %w", err)
	}

	log.Printf("Batch updated stats for %d users", len(userStats))

	return nil
}

// BatchGetUsers выполняет массовое получение пользователей по ID.
func (bo *BatchOperations) BatchGetUsers(ctx context.Context, userIDs []int) ([]*models.User, error) {
	if len(userIDs) == 0 {
		return nil, nil
	}

	// Создаем плейсхолдеры для IN clause
	placeholders := make([]string, len(userIDs))
	args := make([]interface{}, len(userIDs))

	for i, userID := range userIDs {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = userID
	}

	query := fmt.Sprintf(`
		SELECT id, telegram_id, username, first_name, native_language_code,
		       target_language_code, target_language_level, interface_language_code,
		       state, status, profile_completion_level, created_at, updated_at
		FROM users 
		WHERE id IN (%s)`,
		strings.Join(placeholders, ","))

	rows, err := bo.db.conn.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute batch get users: %w", err)
	}

	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			log.Printf("Failed to close rows in BatchGetUsers: %v", closeErr)
		}
	}()

	var users []*models.User
	for rows.Next() {
		user := &models.User{}

		err := rows.Scan(
			&user.ID,
			&user.TelegramID,
			&user.Username,
			&user.FirstName,
			&user.NativeLanguageCode,
			&user.TargetLanguageCode,
			&user.TargetLanguageLevel,
			&user.InterfaceLanguageCode,
			&user.State,
			&user.Status,
			&user.ProfileCompletionLevel,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}

		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate users: %w", err)
	}

	log.Printf("Batch retrieved %d users", len(users))

	return users, nil
}

// UserStats представляет статистику пользователя для batch операций.
type UserStats struct {
	UserID                 int
	ProfileCompletionLevel int
	UpdatedAt              time.Time
}

// Helper functions

// escapeSQL экранирует строку для SQL запроса.
func escapeSQL(s string) string {
	if s == "" {
		return "NULL"
	}

	return fmt.Sprintf("'%s'", strings.ReplaceAll(s, "'", "''"))
}

// escapeCSV экранирует строку для CSV формата.
func escapeCSV(s string) string {
	if s == "" {
		return ""
	}
	// Экранируем кавычки и запятые
	s = strings.ReplaceAll(s, "\"", "\"\"")
	if strings.Contains(s, ",") || strings.Contains(s, "\"") || strings.Contains(s, "\n") {
		return fmt.Sprintf("\"%s\"", s)
	}

	return s
}
