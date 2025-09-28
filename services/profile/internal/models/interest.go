package models

import (
	"time"
)

// Interest represents an interest in the system.
type Interest struct {
	ID          int64     `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Category    string    `json:"category" db:"category"`
	Description *string   `json:"description,omitempty" db:"description"`
	IsActive    bool      `json:"is_active" db:"is_active"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// UserInterest represents a user's interest.
type UserInterest struct {
	ID         int64     `json:"id" db:"id"`
	UserID     int64     `json:"user_id" db:"user_id"`
	InterestID int64     `json:"interest_id" db:"interest_id"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	Interest   *Interest `json:"interest,omitempty"`
}

// CreateUserInterestRequest represents the request to add an interest to a user.
type CreateUserInterestRequest struct {
	InterestID int64 `json:"interest_id" validate:"required,min=1"`
}

// InterestListResponse represents the response for a list of interests.
type InterestListResponse struct {
	Interests  []Interest `json:"interests"`
	Total      int64      `json:"total"`
	Page       int        `json:"page"`
	PerPage    int        `json:"per_page"`
	TotalPages int        `json:"total_pages"`
}

// UserInterestListResponse represents the response for a list of user interests.
type UserInterestListResponse struct {
	Interests  []UserInterestResponse `json:"interests"`
	Total      int64                  `json:"total"`
	Page       int                    `json:"page"`
	PerPage    int                    `json:"per_page"`
	TotalPages int                    `json:"total_pages"`
}

// UserInterestResponse represents the response for user interest data.
type UserInterestResponse struct {
	ID         int64     `json:"id"`
	UserID     int64     `json:"user_id"`
	InterestID int64     `json:"interest_id"`
	CreatedAt  time.Time `json:"created_at"`
	Interest   *Interest `json:"interest,omitempty"`
}

// InterestSearchRequest represents the request to search interests.
type InterestSearchRequest struct {
	Query    string `json:"query,omitempty" form:"query"`
	Category string `json:"category,omitempty" form:"category"`
	IsActive *bool  `json:"is_active,omitempty" form:"is_active"`
	Page     int    `json:"page,omitempty" form:"page" validate:"min=1"`
	PerPage  int    `json:"per_page,omitempty" form:"per_page" validate:"min=1,max=100"`
}

// ToResponse converts a UserInterest model to UserInterestResponse.
func (ui *UserInterest) ToResponse() UserInterestResponse {
	return UserInterestResponse{
		ID:         ui.ID,
		UserID:     ui.UserID,
		InterestID: ui.InterestID,
		CreatedAt:  ui.CreatedAt,
		Interest:   ui.Interest,
	}
}

// Valid interest categories.
var ValidInterestCategories = []string{
	"technology",
	"science",
	"arts",
	"sports",
	"music",
	"travel",
	"food",
	"books",
	"movies",
	"gaming",
	"photography",
	"fitness",
	"nature",
	"culture",
	"education",
	"business",
	"health",
	"fashion",
	"automotive",
	"other",
}

// IsValidInterestCategory checks if the given category is valid.
func IsValidInterestCategory(category string) bool {
	for _, validCategory := range ValidInterestCategories {
		if category == validCategory {
			return true
		}
	}
	return false
}
