package core

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"

	"language-exchange-bot/internal/config"
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
	db *sql.DB
}

// NewInterestService creates a new InterestService instance.
func NewInterestService(db *sql.DB) *InterestService {
	return &InterestService{db: db}
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

	defer func() {
		closeErr := rows.Close()
		if closeErr != nil {
			// Логируем ошибку закрытия, но не возвращаем её
			fmt.Printf("Warning: failed to close rows: %v\n", closeErr)
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

	defer func() {
		closeErr := rows.Close()
		if closeErr != nil {
			// Логируем ошибку закрытия, но не возвращаем её
			fmt.Printf("Warning: failed to close rows: %v\n", closeErr)
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

	defer func() {
		closeErr := rows.Close()
		if closeErr != nil {
			// Логируем ошибку закрытия, но не возвращаем её
			fmt.Printf("Warning: failed to close rows: %v\n", closeErr)
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
		return errors.New("интерес уже выбран")
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
			return fmt.Errorf("достигнут максимум основных интересов (%d)", limits.MaxPrimaryInterests)
		}
	}

	// Добавляем выбор
	insertQuery := `
		INSERT INTO user_interest_selections (user_id, interest_id, is_primary, selection_order)
		VALUES ($1, $2, $3, $4)
	`
	_, err = s.db.ExecContext(context.Background(), insertQuery, userID, interestID, isPrimary, nextOrder)

	return fmt.Errorf("operation failed: %w", err)
}

// RemoveUserInterestSelection удаляет выбор пользователя.
func (s *InterestService) RemoveUserInterestSelection(userID, interestID int) error {
	query := `DELETE FROM user_interest_selections WHERE user_id = $1 AND interest_id = $2`
	_, err := s.db.ExecContext(context.Background(), query, userID, interestID)

	return fmt.Errorf("operation failed: %w", err)
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
			return fmt.Errorf("достигнут максимум основных интересов (%d)", limits.MaxPrimaryInterests)
		}
	}

	query := `UPDATE user_interest_selections SET is_primary = $3 WHERE user_id = $1 AND interest_id = $2`
	_, err := s.db.ExecContext(context.Background(), query, userID, interestID, isPrimary)

	return fmt.Errorf("operation failed: %w", err)
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

	// Вычисляем рекомендуемое количество основных интересов
	recommendedPrimary := int(math.Ceil(float64(totalInterests) * limits.PrimaryPercentage))

	// Ограничиваем минимумом и максимумом
	if recommendedPrimary < limits.MinPrimaryInterests {
		recommendedPrimary = limits.MinPrimaryInterests
	}

	if recommendedPrimary > limits.MaxPrimaryInterests {
		recommendedPrimary = limits.MaxPrimaryInterests
	}

	// Логируем рекомендацию для отладки
	fmt.Printf("Рекомендуемое количество основных интересов: %d\n", recommendedPrimary)

	// Получаем текущее количество основных интересов
	var currentPrimary int

	countQuery := `SELECT COUNT(*) FROM user_interest_selections WHERE user_id = $1 AND is_primary = true`

	err = s.db.QueryRowContext(context.Background(), countQuery, userID).Scan(&currentPrimary)
	if err != nil {
		return fmt.Errorf("operation failed: %w", err)
	}

	if currentPrimary < limits.MinPrimaryInterests {
		return fmt.Errorf("необходимо выбрать минимум %d основных интересов", limits.MinPrimaryInterests)
	}

	return nil
}
