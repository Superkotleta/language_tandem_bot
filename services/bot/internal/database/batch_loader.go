package database

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"language-exchange-bot/internal/models"

	"github.com/lib/pq"
)

// BatchLoader оптимизирует загрузку данных для предотвращения N+1 запросов
type BatchLoader struct {
	db *DB
}

// NewBatchLoader создает новый экземпляр BatchLoader
func NewBatchLoader(db *DB) *BatchLoader {
	return &BatchLoader{db: db}
}

// UserWithInterests представляет пользователя с его интересами
type UserWithInterests struct {
	*models.User
	Interests []int
}

// UserWithAllData представляет пользователя со всеми связанными данными
type UserWithAllData struct {
	*models.User
	Interests    []int
	Translations map[int]string
	Languages    []*models.Language
}

// BatchLoadUsersWithInterests загружает пользователей с их интересами одним запросом
func (bl *BatchLoader) BatchLoadUsersWithInterests(telegramIDs []int64) (map[int64]*UserWithInterests, error) {
	if len(telegramIDs) == 0 {
		return make(map[int64]*UserWithInterests), nil
	}

	query := `
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

	rows, err := bl.db.conn.Query(query, telegramIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to batch load users with interests: %w", err)
	}
	defer rows.Close()

	users := make(map[int64]*UserWithInterests)

	for rows.Next() {
		var user models.User

		var interestID sql.NullInt64

		err := rows.Scan(
			&user.ID, &user.TelegramID, &user.Username, &user.FirstName,
			&user.NativeLanguageCode, &user.TargetLanguageCode, &user.TargetLanguageLevel,
			&user.InterfaceLanguageCode, &user.CreatedAt, &user.UpdatedAt,
			&user.State, &user.ProfileCompletionLevel, &user.Status,
			&interestID,
		)
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

	return users, nil
}

// BatchLoadInterestsWithTranslations загружает интересы с переводами для нескольких языков
func (bl *BatchLoader) BatchLoadInterestsWithTranslations(languages []string) (map[string]map[int]string, error) {
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

	rows, err := bl.db.conn.Query(query, languages)
	if err != nil {
		return nil, fmt.Errorf("failed to batch load interests with translations: %w", err)
	}
	defer rows.Close()

	interests := make(map[string]map[int]string)

	for rows.Next() {
		var langCode string

		var id int

		var name string

		err := rows.Scan(&langCode, &id, &name)
		if err != nil {
			log.Printf("Error scanning interest row: %v", err)
			continue
		}

		if interests[langCode] == nil {
			interests[langCode] = make(map[int]string)
		}

		interests[langCode][id] = name
	}

	return interests, nil
}

// BatchLoadLanguagesWithTranslations загружает языки с переводами для нескольких языков
func (bl *BatchLoader) BatchLoadLanguagesWithTranslations(languages []string) (map[string][]*models.Language, error) {
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

	rows, err := bl.db.conn.Query(query, languages)
	if err != nil {
		return nil, fmt.Errorf("failed to batch load languages with translations: %w", err)
	}
	defer rows.Close()

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

// BatchLoadUserInterests загружает интересы для нескольких пользователей одним запросом
func (bl *BatchLoader) BatchLoadUserInterests(userIDs []int) (map[int][]int, error) {
	if len(userIDs) == 0 {
		return make(map[int][]int), nil
	}

	query := `
		SELECT user_id, interest_id
		FROM user_interests
		WHERE user_id = ANY($1)
		ORDER BY user_id, interest_id
	`

	rows, err := bl.db.conn.Query(query, pq.Array(userIDs))
	if err != nil {
		return nil, fmt.Errorf("failed to batch load user interests: %w", err)
	}
	defer rows.Close()

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

// BatchLoadUsers загружает пользователей по Telegram ID одним запросом
func (bl *BatchLoader) BatchLoadUsers(telegramIDs []int64) (map[int64]*models.User, error) {
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

	rows, err := bl.db.conn.Query(query, telegramIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to batch load users: %w", err)
	}
	defer rows.Close()

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

// GetUserWithAllData загружает пользователя со всеми связанными данными одним запросом
func (bl *BatchLoader) GetUserWithAllData(telegramID int64) (*UserWithAllData, error) {
	query := `
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

	rows, err := bl.db.conn.Query(query, telegramID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user with all data: %w", err)
	}
	defer rows.Close()

	var userData *UserWithAllData

	interests := make([]int, 0)

	translations := make(map[int]string)

	languages := make([]*models.Language, 0)

	for rows.Next() {
		var user models.User

		var interestID sql.NullInt64

		var interestName sql.NullString

		var langID sql.NullInt64

		var langCode, langNameNative, langNameEn sql.NullString

		err := rows.Scan(
			&user.ID, &user.TelegramID, &user.Username, &user.FirstName,
			&user.NativeLanguageCode, &user.TargetLanguageCode, &user.TargetLanguageLevel,
			&user.InterfaceLanguageCode, &user.CreatedAt, &user.UpdatedAt,
			&user.State, &user.ProfileCompletionLevel, &user.Status,
			&interestID, &interestName,
			&langID, &langCode, &langNameNative, &langNameEn,
		)
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
				ID:         int(langID.Int64),
				Code:       langCode.String,
				NameNative: langNameNative.String,
				NameEn:     langNameEn.String,
			}
			languages = append(languages, lang)
		}
	}

	if userData == nil {
		return nil, fmt.Errorf("user not found: %d", telegramID)
	}

	userData.Interests = interests
	userData.Translations = translations
	userData.Languages = languages

	return userData, nil
}

// BatchLoadStats загружает статистику для нескольких типов одним запросом
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

	if err := bl.db.conn.QueryRow("SELECT COUNT(*) FROM users").Scan(&totalUsers); err != nil {
		fmt.Printf("Error getting total users count: %v\n", err)
	}

	if err := bl.db.conn.QueryRow("SELECT COUNT(*) FROM users WHERE status = 'active'").Scan(&activeUsers); err != nil {
		fmt.Printf("Error getting active users count: %v\n", err)
	}

	return map[string]interface{}{
		"total_users":  totalUsers,
		"active_users": activeUsers,
		"timestamp":    time.Now(),
	}
}

func (bl *BatchLoader) loadInterestStats() map[string]interface{} {
	var totalInterests, popularInterests int

	if err := bl.db.conn.QueryRow("SELECT COUNT(*) FROM interests").Scan(&totalInterests); err != nil {
		fmt.Printf("Error getting total interests count: %v\n", err)
	}

	if err := bl.db.conn.QueryRow(`
		SELECT COUNT(*) FROM user_interests ui 
		JOIN interests i ON ui.interest_id = i.id 
		GROUP BY ui.interest_id 
		ORDER BY COUNT(*) DESC 
		LIMIT 1
	`).Scan(&popularInterests); err != nil {
		fmt.Printf("Error getting popular interests count: %v\n", err)
	}

	return map[string]interface{}{
		"total_interests":   totalInterests,
		"popular_interests": popularInterests,
		"timestamp":         time.Now(),
	}
}

func (bl *BatchLoader) loadFeedbackStats() map[string]interface{} {
	var totalFeedbacks, processedFeedbacks int

	if err := bl.db.conn.QueryRow("SELECT COUNT(*) FROM user_feedback").Scan(&totalFeedbacks); err != nil {
		fmt.Printf("Error getting total feedbacks count: %v\n", err)
	}

	if err := bl.db.conn.QueryRow("SELECT COUNT(*) FROM user_feedback WHERE is_processed = true").Scan(&processedFeedbacks); err != nil {
		fmt.Printf("Error getting processed feedbacks count: %v\n", err)
	}

	return map[string]interface{}{
		"total_feedbacks":     totalFeedbacks,
		"processed_feedbacks": processedFeedbacks,
		"timestamp":           time.Now(),
	}
}
