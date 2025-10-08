package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	errorsPkg "language-exchange-bot/internal/errors"
	"language-exchange-bot/internal/localization"
	"language-exchange-bot/internal/logging"
	"language-exchange-bot/internal/models"
)

// Константы для BatchLoader.

// BatchLoader оптимизирует загрузку данных для предотвращения N+1 запросов.
type BatchLoader struct {
	db           *DB
	logger       *logging.DatabaseLogger
	errorHandler *errorsPkg.ErrorHandler
}

// NewBatchLoader создает новый экземпляр BatchLoader.
func NewBatchLoader(db *DB) *BatchLoader {
	return &BatchLoader{
		db:           db,
		logger:       logging.NewDatabaseLogger(),
		errorHandler: errorsPkg.NewErrorHandler(nil),
	}
}

// handleRowsError обрабатывает ошибки rows с правильным закрытием.
func (bl *BatchLoader) handleRowsError(rows *sql.Rows, operation string) error {
	if rows == nil {
		return nil
	}

	if err := rows.Err(); err != nil {
		if closeErr := rows.Close(); closeErr != nil {
			return fmt.Errorf("failed to close rows after error in %s: %w (original error: %w)", operation, closeErr, err)
		}

		return fmt.Errorf("rows error in %s: %w", operation, err)
	}

	return nil
}

// closeRowsSafely безопасно закрывает rows с логированием ошибок.
func (bl *BatchLoader) closeRowsSafely(rows *sql.Rows, operation string) {
	if rows == nil {
		return
	}
	if closeErr := rows.Close(); closeErr != nil {
		log.Printf("Warning: failed to close rows in %s: %v", operation, closeErr)
	}
}

// UserWithInterests представляет пользователя с его интересами.
type UserWithInterests struct {
	*models.User

	Interests []int
}

// UserWithAllData представляет пользователя со всеми связанными данными.
type UserWithAllData struct {
	*models.User

	Interests    []int
	Translations map[int]string
	Languages    []*models.Language
}

// BatchLoadUsersWithInterests загружает пользователей с их интересами одним запросом.
func (bl *BatchLoader) BatchLoadUsersWithInterests(
	ctx context.Context,
	telegramIDs []int64,
) (map[int64]*UserWithInterests, error) {
	if len(telegramIDs) == 0 {
		return make(map[int64]*UserWithInterests), nil
	}

	query := getBatchLoadUsersWithInterestsQuery()

	// Создаем контекст с таймаутом если не передан
	if ctx == nil {
		ctx = context.Background()
	}

	ctx, cancel := context.WithTimeout(ctx, localization.DefaultQueryTimeout)
	defer cancel()

	rows, err := bl.db.conn.QueryContext(ctx, query, telegramIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to batch load users with interests: %w", err)
	}

	if err := bl.handleRowsError(rows, "BatchLoadUsersWithInterests"); err != nil {
		return nil, err
	}

	defer bl.closeRowsSafely(rows, "BatchLoadUsersWithInterests")

	return bl.processUsersWithInterestsRows(rows), nil
}

// getBatchLoadUsersWithInterestsQuery возвращает SQL запрос для загрузки пользователей с интересами.
func getBatchLoadUsersWithInterestsQuery() string {
	return `
		SELECT 
			u.id, u.telegram_id, u.username, u.first_name,
			COALESCE(u.native_language_code, '') as native_language_code,
			COALESCE(u.target_language_code, '') as target_language_code,
			COALESCE(u.target_language_level, '') as target_language_level,
			u.interface_language_code, u.created_at, u.updated_at, u.state,
			u.profile_completion_level, u.status,
			ui.interest_id
		FROM users u
		LEFT JOIN user_interests ui ON u.id = ui.user_id
		WHERE u.telegram_id = ANY($1)
		ORDER BY u.id, ui.interest_id
	`
}

// processUsersWithInterestsRows обрабатывает строки результата запроса пользователей с интересами.
func (bl *BatchLoader) processUsersWithInterestsRows(rows *sql.Rows) map[int64]*UserWithInterests {
	users := make(map[int64]*UserWithInterests)

	for rows.Next() {
		user, interestID, err := bl.scanUserWithInterestRow(rows)
		if err != nil {
			log.Printf("Error scanning user row: %v", err)

			continue
		}

		// Если пользователь еще не в мапе, создаем его
		if _, exists := users[user.TelegramID]; !exists {
			users[user.TelegramID] = &UserWithInterests{
				User:      &user,
				Interests: make([]int, 0),
			}
		}

		// Добавляем интерес, если он есть
		if interestID.Valid {
			users[user.TelegramID].Interests = append(users[user.TelegramID].Interests, int(interestID.Int64))
		}
	}

	return users
}

// scanUserWithInterestRow сканирует одну строку результата запроса пользователя с интересом.
func (bl *BatchLoader) scanUserWithInterestRow(rows *sql.Rows) (models.User, sql.NullInt64, error) {
	var (
		user       models.User
		interestID sql.NullInt64
	)

	err := rows.Scan(
		&user.ID, &user.TelegramID, &user.Username, &user.FirstName,
		&user.NativeLanguageCode, &user.TargetLanguageCode, &user.TargetLanguageLevel,
		&user.InterfaceLanguageCode, &user.CreatedAt, &user.UpdatedAt,
		&user.State, &user.ProfileCompletionLevel, &user.Status,
		&interestID,
	)
	if err != nil {
		return user, interestID, fmt.Errorf("failed to scan user row: %w", err)
	}

	return user, interestID, nil
}

// BatchLoadInterestsWithTranslations загружает интересы с переводами для нескольких языков.
func (bl *BatchLoader) BatchLoadInterestsWithTranslations(ctx context.Context, languages []string) (map[string]map[int]string, error) {
	if len(languages) == 0 {
		return make(map[string]map[int]string), nil
	}

	query := `
		SELECT 
			it.language_code,
			i.id,
			CASE
				WHEN it.name IS NOT NULL AND TRIM(it.name) != '' THEN it.name
				ELSE i.key_name
			END as name
		FROM interests i
		LEFT JOIN interest_translations it ON i.id = it.interest_id AND it.language_code = ANY($1)
		ORDER BY i.id, it.language_code
	`

	// Создаем контекст с таймаутом если не передан
	if ctx == nil {
		ctx = context.Background()
	}

	ctx, cancel := context.WithTimeout(ctx, localization.DefaultQueryTimeout)
	defer cancel()

	rows, err := bl.db.conn.QueryContext(ctx, query, languages)
	if err != nil {
		return nil, fmt.Errorf("failed to batch load interests with translations: %w", err)
	}

	if err := bl.handleRowsError(rows, "BatchLoadInterestsWithTranslations"); err != nil {
		return nil, err
	}

	defer func() { _ = rows.Close() }()
	defer bl.closeRowsSafely(rows, "BatchLoadInterestsWithTranslations")

	interests := make(map[string]map[int]string)

	for rows.Next() {
		var langCode string

		var interestID int

		var name string

		err := rows.Scan(&langCode, &interestID, &name)
		if err != nil {
			log.Printf("Error scanning interest row: %v", err)

			continue
		}

		if interests[langCode] == nil {
			interests[langCode] = make(map[int]string)
		}

		interests[langCode][interestID] = name
	}

	return interests, nil
}

// BatchLoadLanguagesWithTranslations загружает языки с переводами для нескольких языков.
func (bl *BatchLoader) BatchLoadLanguagesWithTranslations(ctx context.Context, languages []string) (map[string][]*models.Language, error) {
	if len(languages) == 0 {
		return make(map[string][]*models.Language), nil
	}

	query := `
		SELECT 
			lt.language_code,
			l.id, l.code, l.name_native, l.name_en
		FROM languages l
		LEFT JOIN language_translations lt ON l.id = lt.language_id AND lt.language_code = ANY($1)
		ORDER BY l.id, lt.language_code
	`

	// Создаем контекст с таймаутом если не передан
	if ctx == nil {
		ctx = context.Background()
	}

	ctx, cancel := context.WithTimeout(ctx, localization.DefaultQueryTimeout)
	defer cancel()

	rows, err := bl.db.conn.QueryContext(ctx, query, languages)
	if err != nil {
		return nil, fmt.Errorf("failed to batch load languages with translations: %w", err)
	}

	if err := bl.handleRowsError(rows, "BatchLoadLanguagesWithTranslations"); err != nil {
		return nil, err
	}

	defer func() { _ = rows.Close() }()
	defer bl.closeRowsSafely(rows, "BatchLoadLanguagesWithTranslations")

	langs := make(map[string][]*models.Language)

	for rows.Next() {
		var langCode string

		var lang models.Language

		err := rows.Scan(&langCode, &lang.ID, &lang.Code, &lang.NameNative, &lang.NameEn)
		if err != nil {
			log.Printf("Error scanning language row: %v", err)

			continue
		}

		if langs[langCode] == nil {
			langs[langCode] = make([]*models.Language, 0)
		}

		langs[langCode] = append(langs[langCode], &lang)
	}

	return langs, nil
}

// BatchLoadUserInterests загружает интересы для нескольких пользователей одним запросом.
func (bl *BatchLoader) BatchLoadUserInterests(ctx context.Context, userIDs []int) (map[int][]int, error) {
	if len(userIDs) == 0 {
		return make(map[int][]int), nil
	}

	// Создаем placeholders для SQLite: ?, ?, ?
	placeholders := make([]string, len(userIDs))
	args := make([]interface{}, len(userIDs))
	for i, id := range userIDs {
		placeholders[i] = "?"
		args[i] = id
	}

	query := fmt.Sprintf(`
		SELECT user_id, interest_id
		FROM user_interests
		WHERE user_id IN (%s)
		ORDER BY user_id, interest_id
	`, strings.Join(placeholders, ","))

	// Создаем контекст с таймаутом если не передан
	if ctx == nil {
		ctx = context.Background()
	}

	ctx, cancel := context.WithTimeout(ctx, localization.DefaultQueryTimeout)
	defer cancel()

	rows, err := bl.db.conn.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to batch load user interests: %w", err)
	}

	if err := bl.handleRowsError(rows, "BatchLoadUserInterests"); err != nil {
		return nil, err
	}

	defer func() { _ = rows.Close() }()
	defer bl.closeRowsSafely(rows, "BatchLoadUserInterests")

	interests := make(map[int][]int)

	for rows.Next() {
		var userID, interestID int

		err := rows.Scan(&userID, &interestID)
		if err != nil {
			log.Printf("Error scanning user interest row: %v", err)

			continue
		}

		if interests[userID] == nil {
			interests[userID] = make([]int, 0)
		}

		interests[userID] = append(interests[userID], interestID)
	}

	return interests, nil
}

// BatchLoadUsers загружает пользователей по Telegram ID одним запросом.
func (bl *BatchLoader) BatchLoadUsers(ctx context.Context, telegramIDs []int64) (map[int64]*models.User, error) {
	if len(telegramIDs) == 0 {
		return make(map[int64]*models.User), nil
	}

	query := `
		SELECT id, telegram_id, username, first_name,
		       COALESCE(native_language_code, '') as native_language_code,
		       COALESCE(target_language_code, '') as target_language_code,
		       COALESCE(target_language_level, '') as target_language_level,
		       interface_language_code, created_at, updated_at, state,
		       profile_completion_level, status
		FROM users
		WHERE telegram_id = ANY($1)
	`

	// Создаем контекст с таймаутом если не передан
	if ctx == nil {
		ctx = context.Background()
	}

	ctx, cancel := context.WithTimeout(ctx, localization.DefaultQueryTimeout)
	defer cancel()

	rows, err := bl.db.conn.QueryContext(ctx, query, telegramIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to batch load users: %w", err)
	}

	if err := bl.handleRowsError(rows, "BatchLoadUsers"); err != nil {
		return nil, err
	}

	defer func() { _ = rows.Close() }()
	defer bl.closeRowsSafely(rows, "BatchLoadUsers")

	users := make(map[int64]*models.User)

	for rows.Next() {
		var user models.User

		err := rows.Scan(
			&user.ID, &user.TelegramID, &user.Username, &user.FirstName,
			&user.NativeLanguageCode, &user.TargetLanguageCode, &user.TargetLanguageLevel,
			&user.InterfaceLanguageCode, &user.CreatedAt, &user.UpdatedAt,
			&user.State, &user.ProfileCompletionLevel, &user.Status,
		)
		if err != nil {
			log.Printf("Error scanning user row: %v", err)

			continue
		}

		users[user.TelegramID] = &user
	}

	return users, nil
}

// GetUserWithAllData загружает пользователя со всеми связанными данными одним запросом.
func (bl *BatchLoader) GetUserWithAllData(ctx context.Context, telegramID int64) (*UserWithAllData, error) {
	query := getUserWithAllDataQuery()

	// Создаем контекст с таймаутом если не передан
	if ctx == nil {
		ctx = context.Background()
	}

	ctx, cancel := context.WithTimeout(ctx, localization.DefaultQueryTimeout)
	defer cancel()

	rows, err := bl.db.conn.QueryContext(ctx, query, telegramID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user with all data: %w", err)
	}

	if err := bl.handleRowsError(rows, "GetUserWithAllData"); err != nil {
		return nil, err
	}

	defer bl.closeRowsSafely(rows, "GetUserWithAllData")

	userData, interests, translations, languages := bl.processUserDataRows(rows)

	if userData == nil {
		return nil, errorsPkg.ErrUserNotFound
	}

	userData.Interests = interests
	userData.Translations = translations
	userData.Languages = languages

	return userData, nil
}

// getUserWithAllDataQuery возвращает SQL запрос для получения пользователя со всеми данными.
func getUserWithAllDataQuery() string {
	return `
		SELECT 
			u.id, u.telegram_id, u.username, u.first_name,
			COALESCE(u.native_language_code, '') as native_language_code,
			COALESCE(u.target_language_code, '') as target_language_code,
			COALESCE(u.target_language_level, '') as target_language_level,
			u.interface_language_code, u.created_at, u.updated_at, u.state,
			u.profile_completion_level, u.status,
			ui.interest_id,
			it.name as interest_name,
			l.id as lang_id, l.code as lang_code, l.name_native, l.name_en
		FROM users u
		LEFT JOIN user_interests ui ON u.id = ui.user_id
		LEFT JOIN interest_translations it ON ui.interest_id = it.interest_id 
			AND it.language_code = u.interface_language_code
		LEFT JOIN languages l ON l.code = u.interface_language_code
		WHERE u.telegram_id = $1
		ORDER BY ui.interest_id
	`
}

// processUserDataRows обрабатывает строки результата запроса.
func (bl *BatchLoader) processUserDataRows(
	rows *sql.Rows,
) (*UserWithAllData, []int, map[int]string, []*models.Language) {
	var userData *UserWithAllData

	interests := make([]int, 0)
	translations := make(map[int]string)
	languages := make([]*models.Language, 0)

	for rows.Next() {
		user, interestID, interestName, langID, langCode, langNameNative, langNameEn, err := bl.scanUserDataRow(rows)
		if err != nil {
			log.Printf("Error scanning user data row: %v", err)

			continue
		}

		// Инициализируем userData только один раз
		if userData == nil {
			userData = &UserWithAllData{
				User:         &user,
				Interests:    make([]int, 0),
				Translations: make(map[int]string),
				Languages:    make([]*models.Language, 0),
			}
		}

		// Добавляем интерес, если он есть
		if interestID.Valid {
			interests = append(interests, int(interestID.Int64))
			if interestName.Valid {
				translations[int(interestID.Int64)] = interestName.String
			}
		}

		// Добавляем язык, если он есть
		if langID.Valid && langCode.Valid {
			lang := &models.Language{
				ID:                  int(langID.Int64),
				Code:                langCode.String,
				NameNative:          langNameNative.String,
				NameEn:              langNameEn.String,
				IsInterfaceLanguage: false,
				CreatedAt:           time.Now(),
			}
			languages = append(languages, lang)
		}
	}

	return userData, interests, translations, languages
}

// scanUserDataRow сканирует одну строку результата запроса.
func (bl *BatchLoader) scanUserDataRow(
	rows *sql.Rows,
) (models.User, sql.NullInt64, sql.NullString, sql.NullInt64, sql.NullString, sql.NullString, sql.NullString, error) {
	var (
		user                                 models.User
		interestID                           sql.NullInt64
		interestName                         sql.NullString
		langID                               sql.NullInt64
		langCode, langNameNative, langNameEn sql.NullString
	)

	err := rows.Scan(
		&user.ID, &user.TelegramID, &user.Username, &user.FirstName,
		&user.NativeLanguageCode, &user.TargetLanguageCode, &user.TargetLanguageLevel,
		&user.InterfaceLanguageCode, &user.CreatedAt, &user.UpdatedAt,
		&user.State, &user.ProfileCompletionLevel, &user.Status,
		&interestID, &interestName,
		&langID, &langCode, &langNameNative, &langNameEn,
	)

	return user, interestID, interestName, langID, langCode, langNameNative, langNameEn,
		fmt.Errorf("operation failed: %w", err)
}

// BatchLoadStats загружает статистику для нескольких типов одним запросом.
func (bl *BatchLoader) BatchLoadStats(statTypes []string) (map[string]map[string]interface{}, error) {
	if len(statTypes) == 0 {
		return make(map[string]map[string]interface{}), nil
	}

	stats := make(map[string]map[string]interface{})

	for _, statType := range statTypes {
		switch statType {
		case "users":
			stats[statType] = bl.loadUserStats()
		case "interests":
			stats[statType] = bl.loadInterestStats()
		case "feedbacks":
			stats[statType] = bl.loadFeedbackStats()
		}
	}

	return stats, nil
}

func (bl *BatchLoader) loadUserStats() map[string]interface{} {
	var totalUsers, activeUsers int

	err := bl.db.conn.QueryRowContext(context.Background(), "SELECT COUNT(*) FROM users").Scan(&totalUsers)
	if err != nil {
		bl.logger.ErrorWithContext(
			"Failed to get total users count",
			"", 0, 0, "LoadUserStats",
			map[string]interface{}{
				"error": err.Error(),
			},
		)
	}

	query := "SELECT COUNT(*) FROM users WHERE status = 'active'"

	err = bl.db.conn.QueryRowContext(context.Background(), query).Scan(&activeUsers)
	if err != nil {
		bl.logger.ErrorWithContext(
			"Failed to get active users count",
			"", 0, 0, "LoadUserStats",
			map[string]interface{}{
				"error": err.Error(),
			},
		)
	}

	return map[string]interface{}{
		"total_users":  totalUsers,
		"active_users": activeUsers,
		"timestamp":    time.Now(),
	}
}

func (bl *BatchLoader) loadInterestStats() map[string]interface{} {
	var totalInterests, popularInterests int

	err := bl.db.conn.QueryRowContext(context.Background(), "SELECT COUNT(*) FROM interests").Scan(&totalInterests)
	if err != nil {
		bl.logger.ErrorWithContext(
			"Failed to get total interests count",
			"", 0, 0, "LoadInterestStats",
			map[string]interface{}{
				"error": err.Error(),
			},
		)
	}

	err = bl.db.conn.QueryRowContext(context.Background(), `
		SELECT COUNT(*) FROM user_interests ui 
		JOIN interests i ON ui.interest_id = i.id 
		GROUP BY ui.interest_id 
		ORDER BY COUNT(*) DESC 
		LIMIT 1
	`).Scan(&popularInterests)
	if err != nil {
		bl.logger.ErrorWithContext(
			"Failed to get popular interests count",
			"", 0, 0, "LoadInterestStats",
			map[string]interface{}{
				"error": err.Error(),
			},
		)
	}

	return map[string]interface{}{
		"total_interests":   totalInterests,
		"popular_interests": popularInterests,
		"timestamp":         time.Now(),
	}
}

func (bl *BatchLoader) loadFeedbackStats() map[string]interface{} {
	var totalFeedbacks, processedFeedbacks int

	err := bl.db.conn.QueryRowContext(context.Background(), "SELECT COUNT(*) FROM user_feedback").Scan(&totalFeedbacks)
	if err != nil {
		bl.logger.ErrorWithContext(
			"Failed to get total feedbacks count",
			"", 0, 0, "LoadFeedbackStats",
			map[string]interface{}{
				"error": err.Error(),
			},
		)
	}

	query := "SELECT COUNT(*) FROM user_feedback WHERE is_processed = true"

	err = bl.db.conn.QueryRowContext(context.Background(), query).Scan(&processedFeedbacks)
	if err != nil {
		bl.logger.ErrorWithContext(
			"Failed to get processed feedbacks count",
			"", 0, 0, "LoadFeedbackStats",
			map[string]interface{}{
				"error": err.Error(),
			},
		)
	}

	return map[string]interface{}{
		"total_feedbacks":     totalFeedbacks,
		"processed_feedbacks": processedFeedbacks,
		"timestamp":           time.Now(),
	}
}

// ===== НОВЫЕ МЕТОДЫ БАТЧИНГА =====

// BatchUpdateUserInterests обновляет интересы пользователя батчем.
func (bl *BatchLoader) BatchUpdateUserInterests(ctx context.Context, userID int, interests []int, primaryInterests []int) error {
	if len(interests) == 0 {
		return nil
	}

	// Создаем контекст с таймаутом если не передан
	if ctx == nil {
		ctx = context.Background()
	}

	ctx, cancel := context.WithTimeout(ctx, localization.DefaultQueryTimeout)
	defer cancel()

	// Начинаем транзакцию
	tx, err := bl.db.conn.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	// Удаляем старые интересы
	_, err = tx.ExecContext(ctx, "DELETE FROM user_interest_selections WHERE user_id = $1", userID)
	if err != nil {
		return fmt.Errorf("failed to delete old interests: %w", err)
	}

	// Подготавливаем данные для вставки
	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO user_interest_selections (user_id, interest_id, is_primary, created_at)
		VALUES ($1, $2, $3, CURRENT_TIMESTAMP)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer func() {
		_ = stmt.Close()
	}()

	// Вставляем новые интересы
	for _, interestID := range interests {
		isPrimary := false
		for _, primaryID := range primaryInterests {
			if interestID == primaryID {
				isPrimary = true
				break
			}
		}

		_, err = stmt.ExecContext(ctx, userID, interestID, isPrimary)
		if err != nil {
			return fmt.Errorf("failed to insert interest %d: %w", interestID, err)
		}
	}

	// Коммитим транзакцию
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	bl.logger.InfoWithContext(
		"Batch updated user interests",
		"", int64(userID), 0, "BatchUpdateUserInterests",
		map[string]interface{}{
			"user_id":         userID,
			"interests_count": len(interests),
			"primary_count":   len(primaryInterests),
		},
	)

	return nil
}

// BatchLoadInterestCategories загружает категории интересов батчем.
func (bl *BatchLoader) BatchLoadInterestCategories(ctx context.Context, lang string) ([]*models.InterestCategory, error) {
	// Создаем контекст с таймаутом если не передан
	if ctx == nil {
		ctx = context.Background()
	}

	ctx, cancel := context.WithTimeout(ctx, localization.DefaultQueryTimeout)
	defer cancel()

	query := `
		SELECT ic.id, ic.key_name, ic.display_order,
		       COALESCE(ict.name, ic.key_name) as name,
		       COALESCE(ict.description, '') as description
		FROM interest_categories ic
		LEFT JOIN interest_category_translations ict ON ic.id = ict.category_id 
			AND ict.language_code = $1
		ORDER BY ic.display_order, ic.id
	`

	rows, err := bl.db.conn.QueryContext(ctx, query, lang)
	if err != nil {
		return nil, fmt.Errorf("failed to load interest categories: %w", err)
	}
	defer bl.closeRowsSafely(rows, "BatchLoadInterestCategories")

	var categories []*models.InterestCategory
	for rows.Next() {
		var category models.InterestCategory
		err := rows.Scan(
			&category.ID,
			&category.KeyName,
			&category.DisplayOrder,
			&category.Name,
			&category.Description,
		)
		if err != nil {
			log.Printf("Error scanning interest category row: %v", err)
			continue
		}

		categories = append(categories, &category)
	}

	return categories, nil
}

// BatchLoadUserStatistics загружает статистику пользователей батчем.
func (bl *BatchLoader) BatchLoadUserStatistics(ctx context.Context, userIDs []int64) (map[int64]map[string]interface{}, error) {
	if len(userIDs) == 0 {
		return make(map[int64]map[string]interface{}), nil
	}

	// Создаем контекст с таймаутом если не передан
	if ctx == nil {
		ctx = context.Background()
	}

	ctx, cancel := context.WithTimeout(ctx, localization.DefaultQueryTimeout)
	defer cancel()

	query := `
		SELECT 
			u.telegram_id,
			COUNT(DISTINCT ui.interest_id) as interests_count,
			COUNT(DISTINCT CASE WHEN ui.is_primary = true THEN ui.interest_id END) as primary_interests_count,
			u.profile_completion_level,
			u.status,
			u.created_at
		FROM users u
		LEFT JOIN user_interest_selections ui ON u.id = ui.user_id
		WHERE u.telegram_id = ANY($1)
		GROUP BY u.telegram_id, u.profile_completion_level, u.status, u.created_at
	`

	rows, err := bl.db.conn.QueryContext(ctx, query, userIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to load user statistics: %w", err)
	}
	defer bl.closeRowsSafely(rows, "BatchLoadUserStatistics")

	stats := make(map[int64]map[string]interface{})
	for rows.Next() {
		var telegramID int64
		var interestsCount, primaryInterestsCount int
		var profileCompletionLevel, status string
		var createdAt time.Time

		err := rows.Scan(
			&telegramID,
			&interestsCount,
			&primaryInterestsCount,
			&profileCompletionLevel,
			&status,
			&createdAt,
		)
		if err != nil {
			log.Printf("Error scanning user statistics row: %v", err)
			continue
		}

		stats[telegramID] = map[string]interface{}{
			"interests_count":          interestsCount,
			"primary_interests_count":  primaryInterestsCount,
			"profile_completion_level": profileCompletionLevel,
			"status":                   status,
			"created_at":               createdAt,
			"timestamp":                time.Now(),
		}
	}

	return stats, nil
}

// BatchLoadPopularInterests загружает популярные интересы батчем.
func (bl *BatchLoader) BatchLoadPopularInterests(ctx context.Context, limit int) ([]map[string]interface{}, error) {
	// Создаем контекст с таймаутом если не передан
	if ctx == nil {
		ctx = context.Background()
	}

	ctx, cancel := context.WithTimeout(ctx, localization.DefaultQueryTimeout)
	defer cancel()

	query := `
		SELECT 
			i.id, i.key_name,
			COUNT(ui.user_id) as user_count,
			COUNT(CASE WHEN ui.is_primary = true THEN 1 END) as primary_count
		FROM interests i
		LEFT JOIN user_interest_selections ui ON i.id = ui.interest_id
		GROUP BY i.id, i.key_name
		ORDER BY user_count DESC, primary_count DESC
		LIMIT $1
	`

	rows, err := bl.db.conn.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to load popular interests: %w", err)
	}
	defer bl.closeRowsSafely(rows, "BatchLoadPopularInterests")

	var interests []map[string]interface{}
	for rows.Next() {
		var id int
		var keyName string
		var userCount, primaryCount int

		err := rows.Scan(&id, &keyName, &userCount, &primaryCount)
		if err != nil {
			log.Printf("Error scanning popular interest row: %v", err)
			continue
		}

		interests = append(interests, map[string]interface{}{
			"id":            id,
			"key_name":      keyName,
			"user_count":    userCount,
			"primary_count": primaryCount,
		})
	}

	return interests, nil
}
