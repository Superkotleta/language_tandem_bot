package models

import (
	"time"
)

// Language represents a language in the system.
type Language struct {
	ID         int64     `json:"id" db:"id"`
	Code       string    `json:"code" db:"code"`
	Name       string    `json:"name" db:"name"`
	NativeName string    `json:"native_name" db:"native_name"`
	IsActive   bool      `json:"is_active" db:"is_active"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}

// UserLanguage represents a user's language proficiency.
type UserLanguage struct {
	ID         int64     `json:"id" db:"id"`
	UserID     int64     `json:"user_id" db:"user_id"`
	LanguageID int64     `json:"language_id" db:"language_id"`
	Level      string    `json:"level" db:"level"`
	IsLearning bool      `json:"is_learning" db:"is_learning"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at"`
	Language   *Language `json:"language,omitempty"`
}

// CreateUserLanguageRequest represents the request to add a language to a user.
type CreateUserLanguageRequest struct {
	LanguageID int64  `json:"language_id" validate:"required,min=1"`
	Level      string `json:"level" validate:"required,oneof=beginner elementary intermediate upper_intermediate advanced native"`
	IsLearning bool   `json:"is_learning"`
}

// UpdateUserLanguageRequest represents the request to update a user's language.
type UpdateUserLanguageRequest struct {
	Level      string `json:"level" validate:"required,oneof=beginner elementary intermediate upper_intermediate advanced native"`
	IsLearning bool   `json:"is_learning"`
}

// UserLanguageResponse represents the response for user language data.
type UserLanguageResponse struct {
	ID         int64     `json:"id"`
	UserID     int64     `json:"user_id"`
	LanguageID int64     `json:"language_id"`
	Level      string    `json:"level"`
	IsLearning bool      `json:"is_learning"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	Language   *Language `json:"language,omitempty"`
}

// LanguageListResponse represents the response for a list of languages.
type LanguageListResponse struct {
	Languages  []Language `json:"languages"`
	Total      int64      `json:"total"`
	Page       int        `json:"page"`
	PerPage    int        `json:"per_page"`
	TotalPages int        `json:"total_pages"`
}

// UserLanguageListResponse represents the response for a list of user languages.
type UserLanguageListResponse struct {
	Languages  []UserLanguageResponse `json:"languages"`
	Total      int64                  `json:"total"`
	Page       int                    `json:"page"`
	PerPage    int                    `json:"per_page"`
	TotalPages int                    `json:"total_pages"`
}

// LanguageSearchRequest represents the request to search languages.
type LanguageSearchRequest struct {
	Query    string `json:"query,omitempty" form:"query"`
	IsActive *bool  `json:"is_active,omitempty" form:"is_active"`
	Page     int    `json:"page,omitempty" form:"page" validate:"min=1"`
	PerPage  int    `json:"per_page,omitempty" form:"per_page" validate:"min=1,max=100"`
}

// ToResponse converts a UserLanguage model to UserLanguageResponse.
func (ul *UserLanguage) ToResponse() UserLanguageResponse {
	return UserLanguageResponse{
		ID:         ul.ID,
		UserID:     ul.UserID,
		LanguageID: ul.LanguageID,
		Level:      ul.Level,
		IsLearning: ul.IsLearning,
		CreatedAt:  ul.CreatedAt,
		UpdatedAt:  ul.UpdatedAt,
		Language:   ul.Language,
	}
}

// Valid language levels.
var ValidLanguageLevels = []string{
	"beginner",
	"elementary",
	"intermediate",
	"upper_intermediate",
	"advanced",
	"native",
}

// IsValidLanguageLevel checks if the given level is valid.
func IsValidLanguageLevel(level string) bool {
	for _, validLevel := range ValidLanguageLevels {
		if level == validLevel {
			return true
		}
	}
	return false
}
