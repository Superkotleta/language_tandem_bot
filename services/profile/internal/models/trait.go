package models

import (
	"time"
)

// UserTrait represents a user's trait (personality, learning style, etc.)
type UserTrait struct {
	ID         int64     `json:"id" db:"id"`
	UserID     int64     `json:"user_id" db:"user_id"`
	TraitType  string    `json:"trait_type" db:"trait_type"`
	TraitValue string    `json:"trait_value" db:"trait_value"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at"`
}

// CreateUserTraitRequest represents the request to create a user trait.
type CreateUserTraitRequest struct {
	TraitType  string `json:"trait_type" validate:"required,min=1,max=50"`
	TraitValue string `json:"trait_value" validate:"required,min=1,max=100"`
}

// UpdateUserTraitRequest represents the request to update a user trait.
type UpdateUserTraitRequest struct {
	TraitValue string `json:"trait_value" validate:"required,min=1,max=100"`
}

// UserTraitResponse represents the response for user trait data.
type UserTraitResponse struct {
	ID         int64     `json:"id"`
	UserID     int64     `json:"user_id"`
	TraitType  string    `json:"trait_type"`
	TraitValue string    `json:"trait_value"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// UserTraitListResponse represents the response for a list of user traits.
type UserTraitListResponse struct {
	Traits     []UserTraitResponse `json:"traits"`
	Total      int64               `json:"total"`
	Page       int                 `json:"page"`
	PerPage    int                 `json:"per_page"`
	TotalPages int                 `json:"total_pages"`
}

// ToResponse converts a UserTrait model to UserTraitResponse.
func (ut *UserTrait) ToResponse() UserTraitResponse {
	return UserTraitResponse{
		ID:         ut.ID,
		UserID:     ut.UserID,
		TraitType:  ut.TraitType,
		TraitValue: ut.TraitValue,
		CreatedAt:  ut.CreatedAt,
		UpdatedAt:  ut.UpdatedAt,
	}
}

// Valid trait types.
var ValidTraitTypes = []string{
	"personality",
	"learning_style",
	"communication_style",
	"interests",
	"goals",
	"experience_level",
	"availability",
	"preferred_activities",
	"hobbies",
	"lifestyle",
	"values",
	"motivation",
}

// IsValidTraitType checks if the given trait type is valid.
func IsValidTraitType(traitType string) bool {
	for _, validType := range ValidTraitTypes {
		if traitType == validType {
			return true
		}
	}
	return false
}
