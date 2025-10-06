package core

import (
	"context"
	"database/sql"
	"fmt"

	"language-exchange-bot/internal/config"
	errorsPkg "language-exchange-bot/internal/errors"
	"language-exchange-bot/internal/logging"
	"language-exchange-bot/internal/models"
)

// Константы для SQL запросов.
const (
	// countPrimaryInterestsQuery - запрос для подсчета основных интересов пользователя.
	countPrimaryInterestsQuery = `SELECT COUNT(*) FROM user_interest_selections WHERE user_id = $1 AND is_primary = true`

	// primaryInterestMultiplier - множитель для максимального балла основных интересов.
	primaryInterestMultiplier = 2
)

// InterestService handles user interest management and matching.
type InterestService struct {
	db           *sql.DB
	logger       *logging.DatabaseLogger
	errorHandler *errorsPkg.ErrorHandler
}

// NewInterestService creates a new InterestService instance.
func NewInterestService(db *sql.DB) *InterestService {
	return &InterestService{
		db:           db,
		logger:       logging.NewDatabaseLogger(),
		errorHandler: errorsPkg.NewErrorHandler(nil),
	}
}

// GetInterestCategories возвращает все категории интересов.
func (s *InterestService) GetInterestCategories() ([]models.InterestCategory, error) {
	query := `
		SELECT id, key_name, display_order, created_at 
		FROM interest_categories 
		ORDER BY display_order ASC
	`

	rows, err := s.db.QueryContext(context.Background(), query)
	if err != nil {
		return nil, fmt.Errorf("failed to query interests: %w", err)
	}

	if err := rows.Err(); err != nil {
		if closeErr := rows.Close(); closeErr != nil {
			return nil, fmt.Errorf("failed to close rows after error: %w (original error: %w)", closeErr, err)
		}

		return nil, fmt.Errorf("rows error: %w", err)
	}

	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			s.logger.ErrorWithContext(
				"Failed to close database rows",
				"", 0, 0, "GetInterestCategories",
				map[string]interface{}{
					"error": closeErr.Error(),
				},
			)
		}
	}()

	var categories []models.InterestCategory

	for rows.Next() {
		var category models.InterestCategory

		err := rows.Scan(&category.ID, &category.KeyName, &category.DisplayOrder, &category.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan interest category: %w", err)
		}

		categories = append(categories, category)
	}

	return categories, nil
}

// GetInterestsByCategory возвращает интересы по категории.
func (s *InterestService) GetInterestsByCategory(categoryID int) ([]models.Interest, error) {
	query := `
		SELECT id, key_name, category_id, display_order, type, created_at
		FROM interests 
		WHERE category_id = $1 
		ORDER BY display_order ASC, key_name ASC
	`

	rows, err := s.db.QueryContext(context.Background(), query, categoryID)
	if err != nil {
		return nil, fmt.Errorf("failed to query interests by category: %w", err)
	}

	return s.scanInterests(rows)
}

// GetUserInterestSelections возвращает выборы пользователя.
func (s *InterestService) GetUserInterestSelections(userID int) ([]models.InterestSelection, error) {
	query := `
		SELECT id, user_id, interest_id, is_primary, selection_order, created_at
		FROM user_interest_selections 
		WHERE user_id = $1 
		ORDER BY is_primary DESC, selection_order ASC
	`

	rows, err := s.db.QueryContext(context.Background(), query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query user interests: %w", err)
	}

	return s.scanInterestSelections(rows)
}

// AddUserInterestSelection добавляет выбор пользователя.
func (s *InterestService) AddUserInterestSelection(userID, interestID int, isPrimary bool) error {
	// Проверяем, не выбран ли уже этот интерес
	var exists bool

	checkQuery := `SELECT EXISTS(SELECT 1 FROM user_interest_selections WHERE user_id = $1 AND interest_id = $2)`

	err := s.db.QueryRowContext(context.Background(), checkQuery, userID, interestID).Scan(&exists)
	if err != nil {
		return fmt.Errorf("operation failed: %w", err)
	}

	if exists {
		return errorsPkg.ErrInterestAlreadySelected
	}

	// Получаем следующий порядок выбора
	var nextOrder int

	orderQuery := `SELECT COALESCE(MAX(selection_order), 0) + 1 FROM user_interest_selections WHERE user_id = $1`

	err = s.db.QueryRowContext(context.Background(), orderQuery, userID).Scan(&nextOrder)
	if err != nil {
		return fmt.Errorf("operation failed: %w", err)
	}

	// Если это основной интерес, проверяем лимиты
	if isPrimary {
		var primaryCount int

		countQuery := countPrimaryInterestsQuery

		err = s.db.QueryRowContext(context.Background(), countQuery, userID).Scan(&primaryCount)
		if err != nil {
			return fmt.Errorf("operation failed: %w", err)
		}

		// Получаем конфигурацию лимитов
		limits, limitsErr := s.GetInterestLimitsConfig()
		if limitsErr != nil {
			return limitsErr
		}

		if primaryCount >= limits.MaxPrimaryInterests {
			return errorsPkg.ErrMaxPrimaryInterestsReached
		}
	}

	// Добавляем выбор
	insertQuery := `
		INSERT INTO user_interest_selections (user_id, interest_id, is_primary, selection_order)
		VALUES ($1, $2, $3, $4)
	`

	_, err = s.db.ExecContext(context.Background(), insertQuery, userID, interestID, isPrimary, nextOrder)
	if err != nil {
		return fmt.Errorf("operation failed: %w", err)
	}

	return nil
}

// RemoveUserInterestSelection удаляет выбор пользователя.
func (s *InterestService) RemoveUserInterestSelection(userID, interestID int) error {
	query := `DELETE FROM user_interest_selections WHERE user_id = $1 AND interest_id = $2`

	_, err := s.db.ExecContext(context.Background(), query, userID, interestID)
	if err != nil {
		return fmt.Errorf("operation failed: %w", err)
	}

	return nil
}

// SetPrimaryInterest устанавливает интерес как основной.
func (s *InterestService) SetPrimaryInterest(userID, interestID int, isPrimary bool) error {
	// Проверяем лимиты основных интересов
	if isPrimary {
		var primaryCount int

		countQuery := countPrimaryInterestsQuery

		err := s.db.QueryRowContext(context.Background(), countQuery, userID).Scan(&primaryCount)
		if err != nil {
			return fmt.Errorf("operation failed: %w", err)
		}

		limits, err := s.GetInterestLimitsConfig()
		if err != nil {
			return fmt.Errorf("operation failed: %w", err)
		}

		if primaryCount >= limits.MaxPrimaryInterests {
			return errorsPkg.ErrMaxPrimaryInterestsReached
		}
	}

	query := `UPDATE user_interest_selections SET is_primary = $3 WHERE user_id = $1 AND interest_id = $2`

	_, err := s.db.ExecContext(context.Background(), query, userID, interestID, isPrimary)
	if err != nil {
		return fmt.Errorf("operation failed: %w", err)
	}

	return nil
}

// GetInterestLimitsConfig возвращает конфигурацию лимитов из файла.
func (s *InterestService) GetInterestLimitsConfig() (*config.InterestLimitsConfig, error) {
	interestsConfig, err := config.LoadInterestsConfig()
	if err != nil {
		return nil, fmt.Errorf("operation failed: %w", err)
	}

	return &interestsConfig.InterestLimits, nil
}

// GetMatchingConfig возвращает конфигурацию для алгоритма сопоставления из файла.
func (s *InterestService) GetMatchingConfig() (*config.MatchingConfig, error) {
	interestsConfig, err := config.LoadInterestsConfig()
	if err != nil {
		return nil, fmt.Errorf("operation failed: %w", err)
	}

	return &interestsConfig.Matching, nil
}

// CalculateCompatibilityScore вычисляет балл совместимости между пользователями.
func (s *InterestService) CalculateCompatibilityScore(user1ID, user2ID int) (int, error) {
	matchingConfig, err := s.GetMatchingConfig()
	if err != nil {
		return 0, err
	}

	user1Maps, err := s.buildUserInterestMaps(user1ID)
	if err != nil {
		return 0, err
	}

	user2Maps, err := s.buildUserInterestMaps(user2ID)
	if err != nil {
		return 0, err
	}

	score := s.calculateCompatibilityScore(user1Maps, user2Maps, matchingConfig)

	return score, nil
}

// UserInterestMaps содержит карты интересов пользователя.
type UserInterestMaps struct {
	AllInterests     map[int]bool
	PrimaryInterests map[int]bool
}

// GetUserInterestSummary возвращает сводку интересов пользователя.
func (s *InterestService) GetUserInterestSummary(userID int) (*models.UserInterestSummary, error) {
	selections, err := s.GetUserInterestSelections(userID)
	if err != nil {
		return nil, fmt.Errorf("operation failed: %w", err)
	}

	summary := &models.UserInterestSummary{
		UserID:              userID,
		TotalInterests:      len(selections),
		PrimaryInterests:    []models.InterestWithCategory{},
		AdditionalInterests: []models.InterestWithCategory{},
	}

	// Получаем детали интересов
	for _, selection := range selections {
		interest, err := s.GetInterestByID(selection.InterestID)
		if err != nil {
			continue
		}

		category, err := s.GetCategoryByID(interest.CategoryID)
		if err != nil {
			continue
		}

		interestWithCategory := models.InterestWithCategory{
			Interest:     *interest,
			CategoryName: category.KeyName,
			CategoryKey:  category.KeyName,
		}

		if selection.IsPrimary {
			summary.PrimaryInterests = append(summary.PrimaryInterests, interestWithCategory)
		} else {
			summary.AdditionalInterests = append(summary.AdditionalInterests, interestWithCategory)
		}
	}

	return summary, nil
}

// GetInterestByID возвращает интерес по ID.
func (s *InterestService) GetInterestByID(interestID int) (*models.Interest, error) {
	query := `SELECT id, key_name, category_id, display_order, type, created_at FROM interests WHERE id = $1`

	var interest models.Interest

	err := s.db.QueryRowContext(context.Background(), query, interestID).Scan(
		&interest.ID, &interest.KeyName, &interest.CategoryID,
		&interest.DisplayOrder, &interest.Type, &interest.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("operation failed: %w", err)
	}

	return &interest, nil
}

// GetCategoryByID возвращает категорию по ID.
func (s *InterestService) GetCategoryByID(categoryID int) (*models.InterestCategory, error) {
	query := `SELECT id, key_name, display_order, created_at FROM interest_categories WHERE id = $1`

	var category models.InterestCategory

	err := s.db.QueryRowContext(context.Background(), query, categoryID).Scan(
		&category.ID, &category.KeyName, &category.DisplayOrder, &category.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("operation failed: %w", err)
	}

	return &category, nil
}

// ValidateInterestSelection проверяет валидность выбора интересов.
func (s *InterestService) ValidateInterestSelection(userID, totalInterests int) error {
	limits, err := s.GetInterestLimitsConfig()
	if err != nil {
		return fmt.Errorf("operation failed: %w", err)
	}

	// // Вычисляем рекомендуемое количество основных интересов
	// recommendedPrimary := int(math.Ceil(float64(totalInterests) * limits.PrimaryPercentage))

	// // Ограничиваем минимумом и максимумом
	// if recommendedPrimary < limits.MinPrimaryInterests {
	// 	recommendedPrimary = limits.MinPrimaryInterests
	// }

	// if recommendedPrimary > limits.MaxPrimaryInterests {
	// 	recommendedPrimary = limits.MaxPrimaryInterests
	// }

	// Логируем рекомендацию для отладки
	s.logger.DebugWithContext(
		"Interest validation performed",
		"", 0, 0, "ValidateInterestSelection",
		map[string]interface{}{
			"user_id":         userID,
			"total_interests": totalInterests,
		},
	)

	// Получаем текущее количество основных интересов
	var currentPrimary int

	countQuery := `SELECT COUNT(*) FROM user_interest_selections WHERE user_id = $1 AND is_primary = true`

	err = s.db.QueryRowContext(context.Background(), countQuery, userID).Scan(&currentPrimary)
	if err != nil {
		return fmt.Errorf("operation failed: %w", err)
	}

	if currentPrimary < limits.MinPrimaryInterests {
		return errorsPkg.ErrMinPrimaryInterestsRequired
	}

	return nil
}

// GetInterestCategoryByID возвращает категорию интереса по ID.
func (s *InterestService) GetInterestCategoryByID(categoryID int) (*models.InterestCategory, error) {
	var category models.InterestCategory

	err := s.db.QueryRowContext(context.Background(), `
		SELECT id, key_name, display_order, created_at 
		FROM interest_categories 
		WHERE id = $1
	`, categoryID).Scan(
		&category.ID,
		&category.KeyName,
		&category.DisplayOrder,
		&category.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get interest category by ID %d: %w", categoryID, err)
	}

	return &category, nil
}

// buildUserInterestMaps создает карты интересов пользователя.
func (s *InterestService) buildUserInterestMaps(userID int) (*UserInterestMaps, error) {
	interests, err := s.GetUserInterestSelections(userID)
	if err != nil {
		return nil, fmt.Errorf("operation failed: %w", err)
	}

	allMap := make(map[int]bool)
	primaryMap := make(map[int]bool)

	for _, selection := range interests {
		allMap[selection.InterestID] = true
		if selection.IsPrimary {
			primaryMap[selection.InterestID] = true
		}
	}

	return &UserInterestMaps{
		AllInterests:     allMap,
		PrimaryInterests: primaryMap,
	}, nil
}

// calculateCompatibilityScore вычисляет балл совместимости.
func (s *InterestService) calculateCompatibilityScore(
	user1Maps, user2Maps *UserInterestMaps,
	config *config.MatchingConfig,
) int {
	score := 0

	for interestID := range user1Maps.AllInterests {
		if user2Maps.AllInterests[interestID] {
			score += s.calculateInterestScore(interestID, user1Maps, user2Maps, config)
		}
	}

	return score
}

// calculateInterestScore вычисляет балл за конкретный интерес.
func (s *InterestService) calculateInterestScore(
	interestID int,
	user1Maps, user2Maps *UserInterestMaps,
	config *config.MatchingConfig,
) int {
	switch {
	case user1Maps.PrimaryInterests[interestID] && user2Maps.PrimaryInterests[interestID]:
		// Оба пользователя считают этот интерес основным
		return config.PrimaryInterestScore * primaryInterestMultiplier
	case user1Maps.PrimaryInterests[interestID] || user2Maps.PrimaryInterests[interestID]:
		// Один из пользователей считает основным
		return config.PrimaryInterestScore + config.AdditionalInterestScore
	default:
		// Оба считают дополнительным
		return config.AdditionalInterestScore
	}
}

// scanInterests сканирует строки интересов.
func (s *InterestService) scanInterests(rows *sql.Rows) ([]models.Interest, error) {
	if err := rows.Err(); err != nil {
		if closeErr := rows.Close(); closeErr != nil {
			return nil, fmt.Errorf("failed to close rows after error: %w (original error: %w)", closeErr, err)
		}

		return nil, fmt.Errorf("rows error: %w", err)
	}

	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			s.logger.ErrorWithContext(
				"Failed to close database rows",
				"", 0, 0, "GetInterestCategories",
				map[string]interface{}{
					"error": closeErr.Error(),
				},
			)
		}
	}()

	var interests []models.Interest

	for rows.Next() {
		var interest models.Interest

		err := rows.Scan(&interest.ID, &interest.KeyName, &interest.CategoryID,
			&interest.DisplayOrder, &interest.Type, &interest.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan interest: %w", err)
		}

		interests = append(interests, interest)
	}

	return interests, nil
}

// scanInterestSelections сканирует строки выборов интересов.
func (s *InterestService) scanInterestSelections(rows *sql.Rows) ([]models.InterestSelection, error) {
	if err := rows.Err(); err != nil {
		if closeErr := rows.Close(); closeErr != nil {
			s.logger.ErrorWithContext(
				"Failed to close database rows",
				"", 0, 0, "GetInterestCategories",
				map[string]interface{}{
					"error": closeErr.Error(),
				},
			)
		}

		return nil, fmt.Errorf("rows error: %w", err)
	}

	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			s.logger.ErrorWithContext(
				"Failed to close database rows",
				"", 0, 0, "GetInterestCategories",
				map[string]interface{}{
					"error": closeErr.Error(),
				},
			)
		}
	}()

	var selections []models.InterestSelection

	for rows.Next() {
		var selection models.InterestSelection

		err := rows.Scan(&selection.ID, &selection.UserID, &selection.InterestID,
			&selection.IsPrimary, &selection.SelectionOrder, &selection.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user interest selection: %w", err)
		}

		selections = append(selections, selection)
	}

	return selections, nil
}

// GetAllInterests получает все интересы из системы.
func (s *InterestService) GetAllInterests() ([]models.Interest, error) {
	query := `
		SELECT id, key_name, category_id, display_order, type, created_at
		FROM interests 
		ORDER BY id
	`

	rows, err := s.db.QueryContext(context.Background(), query)
	if err != nil {
		return nil, fmt.Errorf("operation failed: %w", err)
	}
	defer rows.Close()

	var interests []models.Interest
	for rows.Next() {
		var interest models.Interest
		err := rows.Scan(
			&interest.ID,
			&interest.KeyName,
			&interest.CategoryID,
			&interest.DisplayOrder,
			&interest.Type,
			&interest.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("operation failed: %w", err)
		}
		interests = append(interests, interest)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("operation failed: %w", err)
	}

	return interests, nil
}

// GetInterestsByCategoryKey возвращает интересы по ключу категории
func (s *InterestService) GetInterestsByCategoryKey(categoryKey string) ([]models.Interest, error) {
	query := `
		SELECT i.id, i.key_name, i.category_id, i.display_order, i.type, i.created_at, ic.key_name
		FROM interests i
		JOIN interest_categories ic ON i.category_id = ic.id
		WHERE ic.key_name = $1 
		ORDER BY i.display_order ASC, i.key_name ASC
	`

	rows, err := s.db.QueryContext(context.Background(), query, categoryKey)
	if err != nil {
		return nil, fmt.Errorf("failed to query interests by category key: %w", err)
	}

	return s.scanInterestsWithCategory(rows)
}

// BatchUpdateUserInterests обновляет интересы пользователя батчем
func (s *InterestService) BatchUpdateUserInterests(userID int, selections []models.InterestSelection) error {
	// Начинаем транзакцию
	tx, err := s.db.BeginTx(context.Background(), nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	// Удаляем все текущие выборы пользователя
	_, err = tx.ExecContext(context.Background(), "DELETE FROM user_interest_selections WHERE user_id = $1", userID)
	if err != nil {
		return fmt.Errorf("failed to delete existing selections: %w", err)
	}

	// Вставляем новые выборы батчем
	if len(selections) > 0 {
		query := `
			INSERT INTO user_interest_selections (user_id, interest_id, is_primary, created_at)
			VALUES ($1, $2, $3, NOW())
		`

		stmt, err := tx.PrepareContext(context.Background(), query)
		if err != nil {
			return fmt.Errorf("failed to prepare statement: %w", err)
		}
		defer stmt.Close()

		for _, selection := range selections {
			_, err = stmt.ExecContext(context.Background(), userID, selection.InterestID, selection.IsPrimary)
			if err != nil {
				return fmt.Errorf("failed to insert selection: %w", err)
			}
		}
	}

	// Подтверждаем транзакцию
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// scanInterestsWithCategory сканирует строки интересов с категорией
func (s *InterestService) scanInterestsWithCategory(rows *sql.Rows) ([]models.Interest, error) {
	if err := rows.Err(); err != nil {
		if closeErr := rows.Close(); closeErr != nil {
			return nil, fmt.Errorf("failed to close rows after error: %w (original error: %w)", closeErr, err)
		}

		return nil, fmt.Errorf("rows error: %w", err)
	}

	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			s.logger.ErrorWithContext(
				"Failed to close database rows",
				"", 0, 0, "scanInterestsWithCategory",
				map[string]interface{}{
					"error": closeErr.Error(),
				},
			)
		}
	}()

	var interests []models.Interest

	for rows.Next() {
		var interest models.Interest

		err := rows.Scan(&interest.ID, &interest.KeyName, &interest.CategoryID,
			&interest.DisplayOrder, &interest.Type, &interest.CreatedAt, &interest.CategoryKey)
		if err != nil {
			return nil, fmt.Errorf("failed to scan interest with category: %w", err)
		}

		interests = append(interests, interest)
	}

	return interests, nil
}
