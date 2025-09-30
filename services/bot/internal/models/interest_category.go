package models

import "time"

// InterestCategory представляет категорию интересов.
type InterestCategory struct {
	ID           int       `db:"id"`
	KeyName      string    `db:"key_name"`
	DisplayOrder int       `db:"display_order"`
	CreatedAt    time.Time `db:"created_at"`
}

// InterestSelection представляет выбор пользователя.
type InterestSelection struct {
	ID             int       `db:"id"`
	UserID         int       `db:"user_id"`
	InterestID     int       `db:"interest_id"`
	IsPrimary      bool      `db:"is_primary"`
	SelectionOrder int       `db:"selection_order"`
	CreatedAt      time.Time `db:"created_at"`
}

// MatchingConfig представляет конфигурацию для алгоритма сопоставления.
type MatchingConfig struct {
	ID          int       `db:"id"`
	ConfigKey   string    `db:"config_key"`
	ConfigValue string    `db:"config_value"`
	Description string    `db:"description"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

// InterestLimitsConfig представляет конфигурацию лимитов основных интересов.
type InterestLimitsConfig struct {
	ID                  int       `db:"id"`
	MinPrimaryInterests int       `db:"min_primary_interests"`
	MaxPrimaryInterests int       `db:"max_primary_interests"`
	PrimaryPercentage   float64   `db:"primary_percentage"`
	CreatedAt           time.Time `db:"created_at"`
	UpdatedAt           time.Time `db:"updated_at"`
}

// InterestWithCategory представляет интерес с информацией о категории.
type InterestWithCategory struct {
	Interest

	CategoryName string `db:"category_name"`
	CategoryKey  string `db:"category_key"`
}

// UserInterestSummary представляет сводку интересов пользователя.
type UserInterestSummary struct {
	UserID              int
	TotalInterests      int
	PrimaryInterests    []InterestWithCategory
	AdditionalInterests []InterestWithCategory
}
