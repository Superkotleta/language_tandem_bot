package models

import (
	"time"
)

// UserPreference represents a user's preferences for matching.
type UserPreference struct {
	ID                 int64     `json:"id" db:"id"`
	UserID             int64     `json:"user_id" db:"user_id"`
	MinAge             *int      `json:"min_age,omitempty" db:"min_age"`
	MaxAge             *int      `json:"max_age,omitempty" db:"max_age"`
	PreferredGender    *string   `json:"preferred_gender,omitempty" db:"preferred_gender"`
	PreferredCountries []string  `json:"preferred_countries,omitempty" db:"preferred_countries"`
	PreferredLanguages []string  `json:"preferred_languages,omitempty" db:"preferred_languages"`
	MaxDistance        *int      `json:"max_distance,omitempty" db:"max_distance"`
	TimezoneOffset     *int      `json:"timezone_offset,omitempty" db:"timezone_offset"`
	AvailabilityStart  *string   `json:"availability_start,omitempty" db:"availability_start"`
	AvailabilityEnd    *string   `json:"availability_end,omitempty" db:"availability_end"`
	IsOnlineOnly       bool      `json:"is_online_only" db:"is_online_only"`
	CreatedAt          time.Time `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time `json:"updated_at" db:"updated_at"`
}

// CreateUserPreferenceRequest represents the request to create user preferences.
type CreateUserPreferenceRequest struct {
	MinAge             *int     `json:"min_age,omitempty" validate:"omitempty,min=13,max=120"`
	MaxAge             *int     `json:"max_age,omitempty" validate:"omitempty,min=13,max=120"`
	PreferredGender    *string  `json:"preferred_gender,omitempty" validate:"omitempty,oneof=male female other any"`
	PreferredCountries []string `json:"preferred_countries,omitempty"`
	PreferredLanguages []string `json:"preferred_languages,omitempty"`
	MaxDistance        *int     `json:"max_distance,omitempty" validate:"omitempty,min=1,max=10000"`
	TimezoneOffset     *int     `json:"timezone_offset,omitempty" validate:"omitempty,min=-720,max=720"`
	AvailabilityStart  *string  `json:"availability_start,omitempty" validate:"omitempty"`
	AvailabilityEnd    *string  `json:"availability_end,omitempty" validate:"omitempty"`
	IsOnlineOnly       bool     `json:"is_online_only"`
}

// UpdateUserPreferenceRequest represents the request to update user preferences.
type UpdateUserPreferenceRequest struct {
	MinAge             *int     `json:"min_age,omitempty" validate:"omitempty,min=13,max=120"`
	MaxAge             *int     `json:"max_age,omitempty" validate:"omitempty,min=13,max=120"`
	PreferredGender    *string  `json:"preferred_gender,omitempty" validate:"omitempty,oneof=male female other any"`
	PreferredCountries []string `json:"preferred_countries,omitempty"`
	PreferredLanguages []string `json:"preferred_languages,omitempty"`
	MaxDistance        *int     `json:"max_distance,omitempty" validate:"omitempty,min=1,max=10000"`
	TimezoneOffset     *int     `json:"timezone_offset,omitempty" validate:"omitempty,min=-720,max=720"`
	AvailabilityStart  *string  `json:"availability_start,omitempty" validate:"omitempty"`
	AvailabilityEnd    *string  `json:"availability_end,omitempty" validate:"omitempty"`
	IsOnlineOnly       *bool    `json:"is_online_only,omitempty"`
}

// UserPreferenceResponse represents the response for user preference data.
type UserPreferenceResponse struct {
	ID                 int64     `json:"id"`
	UserID             int64     `json:"user_id"`
	MinAge             *int      `json:"min_age,omitempty"`
	MaxAge             *int      `json:"max_age,omitempty"`
	PreferredGender    *string   `json:"preferred_gender,omitempty"`
	PreferredCountries []string  `json:"preferred_countries,omitempty"`
	PreferredLanguages []string  `json:"preferred_languages,omitempty"`
	MaxDistance        *int      `json:"max_distance,omitempty"`
	TimezoneOffset     *int      `json:"timezone_offset,omitempty"`
	AvailabilityStart  *string   `json:"availability_start,omitempty"`
	AvailabilityEnd    *string   `json:"availability_end,omitempty"`
	IsOnlineOnly       bool      `json:"is_online_only"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

// ToResponse converts a UserPreference model to UserPreferenceResponse.
func (up *UserPreference) ToResponse() UserPreferenceResponse {
	return UserPreferenceResponse{
		ID:                 up.ID,
		UserID:             up.UserID,
		MinAge:             up.MinAge,
		MaxAge:             up.MaxAge,
		PreferredGender:    up.PreferredGender,
		PreferredCountries: up.PreferredCountries,
		PreferredLanguages: up.PreferredLanguages,
		MaxDistance:        up.MaxDistance,
		TimezoneOffset:     up.TimezoneOffset,
		AvailabilityStart:  up.AvailabilityStart,
		AvailabilityEnd:    up.AvailabilityEnd,
		IsOnlineOnly:       up.IsOnlineOnly,
		CreatedAt:          up.CreatedAt,
		UpdatedAt:          up.UpdatedAt,
	}
}

// Valid gender preferences.
var ValidGenderPreferences = []string{
	"male",
	"female",
	"other",
	"any",
}

// IsValidGenderPreference checks if the given gender preference is valid.
func IsValidGenderPreference(gender string) bool {
	for _, validGender := range ValidGenderPreferences {
		if gender == validGender {
			return true
		}
	}
	return false
}
