package models

import (
	"time"
)

// User represents a user in the system.
type User struct {
	ID                int64     `json:"id" db:"id"`
	TelegramID        *int64    `json:"telegram_id,omitempty" db:"telegram_id"`
	DiscordID         *int64    `json:"discord_id,omitempty" db:"discord_id"`
	Username          *string   `json:"username,omitempty" db:"username"`
	FirstName         *string   `json:"first_name,omitempty" db:"first_name"`
	LastName          *string   `json:"last_name,omitempty" db:"last_name"`
	Email             *string   `json:"email,omitempty" db:"email"`
	Phone             *string   `json:"phone,omitempty" db:"phone"`
	Bio               *string   `json:"bio,omitempty" db:"bio"`
	Age               *int      `json:"age,omitempty" db:"age"`
	Gender            *string   `json:"gender,omitempty" db:"gender"`
	Country           *string   `json:"country,omitempty" db:"country"`
	City              *string   `json:"city,omitempty" db:"city"`
	Timezone          *string   `json:"timezone,omitempty" db:"timezone"`
	ProfilePictureURL *string   `json:"profile_picture_url,omitempty" db:"profile_picture_url"`
	IsActive          bool      `json:"is_active" db:"is_active"`
	IsVerified        bool      `json:"is_verified" db:"is_verified"`
	LastSeen          time.Time `json:"last_seen" db:"last_seen"`
	CreatedAt         time.Time `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time `json:"updated_at" db:"updated_at"`
}

// CreateUserRequest represents the request to create a new user.
type CreateUserRequest struct {
	TelegramID        *int64  `json:"telegram_id,omitempty" validate:"omitempty,min=1"`
	DiscordID         *int64  `json:"discord_id,omitempty" validate:"omitempty,min=1"`
	Username          *string `json:"username,omitempty" validate:"omitempty,min=1,max=255"`
	FirstName         *string `json:"first_name,omitempty" validate:"omitempty,min=1,max=255"`
	LastName          *string `json:"last_name,omitempty" validate:"omitempty,min=1,max=255"`
	Email             *string `json:"email,omitempty" validate:"omitempty,email"`
	Phone             *string `json:"phone,omitempty" validate:"omitempty,min=1,max=20"`
	Bio               *string `json:"bio,omitempty" validate:"omitempty,max=1000"`
	Age               *int    `json:"age,omitempty" validate:"omitempty,min=13,max=120"`
	Gender            *string `json:"gender,omitempty" validate:"omitempty,oneof=male female other prefer_not_to_say"`
	Country           *string `json:"country,omitempty" validate:"omitempty,min=1,max=100"`
	City              *string `json:"city,omitempty" validate:"omitempty,min=1,max=100"`
	Timezone          *string `json:"timezone,omitempty" validate:"omitempty,min=1,max=50"`
	ProfilePictureURL *string `json:"profile_picture_url,omitempty" validate:"omitempty,url"`
}

// UpdateUserRequest represents the request to update a user.
type UpdateUserRequest struct {
	Username          *string `json:"username,omitempty" validate:"omitempty,min=1,max=255"`
	FirstName         *string `json:"first_name,omitempty" validate:"omitempty,min=1,max=255"`
	LastName          *string `json:"last_name,omitempty" validate:"omitempty,min=1,max=255"`
	Email             *string `json:"email,omitempty" validate:"omitempty,email"`
	Phone             *string `json:"phone,omitempty" validate:"omitempty,min=1,max=20"`
	Bio               *string `json:"bio,omitempty" validate:"omitempty,max=1000"`
	Age               *int    `json:"age,omitempty" validate:"omitempty,min=13,max=120"`
	Gender            *string `json:"gender,omitempty" validate:"omitempty,oneof=male female other prefer_not_to_say"`
	Country           *string `json:"country,omitempty" validate:"omitempty,min=1,max=100"`
	City              *string `json:"city,omitempty" validate:"omitempty,min=1,max=100"`
	Timezone          *string `json:"timezone,omitempty" validate:"omitempty,min=1,max=50"`
	ProfilePictureURL *string `json:"profile_picture_url,omitempty" validate:"omitempty,url"`
	IsActive          *bool   `json:"is_active,omitempty"`
}

// UserResponse represents the response for user data.
type UserResponse struct {
	ID                int64     `json:"id"`
	TelegramID        *int64    `json:"telegram_id,omitempty"`
	DiscordID         *int64    `json:"discord_id,omitempty"`
	Username          *string   `json:"username,omitempty"`
	FirstName         *string   `json:"first_name,omitempty"`
	LastName          *string   `json:"last_name,omitempty"`
	Email             *string   `json:"email,omitempty"`
	Phone             *string   `json:"phone,omitempty"`
	Bio               *string   `json:"bio,omitempty"`
	Age               *int      `json:"age,omitempty"`
	Gender            *string   `json:"gender,omitempty"`
	Country           *string   `json:"country,omitempty"`
	City              *string   `json:"city,omitempty"`
	Timezone          *string   `json:"timezone,omitempty"`
	ProfilePictureURL *string   `json:"profile_picture_url,omitempty"`
	IsActive          bool      `json:"is_active"`
	IsVerified        bool      `json:"is_verified"`
	LastSeen          time.Time `json:"last_seen"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
	CompletionScore   float64   `json:"completion_score,omitempty"`
}

// UserListResponse represents the response for a list of users.
type UserListResponse struct {
	Users      []UserResponse `json:"users"`
	Total      int64          `json:"total"`
	Page       int            `json:"page"`
	PerPage    int            `json:"per_page"`
	TotalPages int            `json:"total_pages"`
}

// UserSearchRequest represents the request to search users.
type UserSearchRequest struct {
	Query     string   `json:"query,omitempty" form:"query"`
	Country   string   `json:"country,omitempty" form:"country"`
	City      string   `json:"city,omitempty" form:"city"`
	Languages []string `json:"languages,omitempty" form:"languages"`
	Interests []string `json:"interests,omitempty" form:"interests"`
	MinAge    *int     `json:"min_age,omitempty" form:"min_age"`
	MaxAge    *int     `json:"max_age,omitempty" form:"max_age"`
	Gender    string   `json:"gender,omitempty" form:"gender"`
	Page      int      `json:"page,omitempty" form:"page" validate:"min=1"`
	PerPage   int      `json:"per_page,omitempty" form:"per_page" validate:"min=1,max=100"`
}

// ToResponse converts a User model to UserResponse.
func (u *User) ToResponse() UserResponse {
	return UserResponse{
		ID:                u.ID,
		TelegramID:        u.TelegramID,
		DiscordID:         u.DiscordID,
		Username:          u.Username,
		FirstName:         u.FirstName,
		LastName:          u.LastName,
		Email:             u.Email,
		Phone:             u.Phone,
		Bio:               u.Bio,
		Age:               u.Age,
		Gender:            u.Gender,
		Country:           u.Country,
		City:              u.City,
		Timezone:          u.Timezone,
		ProfilePictureURL: u.ProfilePictureURL,
		IsActive:          u.IsActive,
		IsVerified:        u.IsVerified,
		LastSeen:          u.LastSeen,
		CreatedAt:         u.CreatedAt,
		UpdatedAt:         u.UpdatedAt,
	}
}
