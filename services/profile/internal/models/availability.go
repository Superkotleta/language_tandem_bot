package models

import (
	"time"
)

// UserTimeAvailability represents a user's time availability.
type UserTimeAvailability struct {
	ID        int64     `json:"id" db:"id"`
	UserID    int64     `json:"user_id" db:"user_id"`
	DayOfWeek int       `json:"day_of_week" db:"day_of_week"`
	StartTime string    `json:"start_time" db:"start_time"`
	EndTime   string    `json:"end_time" db:"end_time"`
	Timezone  string    `json:"timezone" db:"timezone"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// CreateUserTimeAvailabilityRequest represents the request to create user time availability.
type CreateUserTimeAvailabilityRequest struct {
	DayOfWeek int    `json:"day_of_week" validate:"required,min=0,max=6"`
	StartTime string `json:"start_time" validate:"required"`
	EndTime   string `json:"end_time" validate:"required"`
	Timezone  string `json:"timezone" validate:"required,min=1,max=50"`
}

// UpdateUserTimeAvailabilityRequest represents the request to update user time availability.
type UpdateUserTimeAvailabilityRequest struct {
	StartTime string `json:"start_time" validate:"required"`
	EndTime   string `json:"end_time" validate:"required"`
	Timezone  string `json:"timezone" validate:"required,min=1,max=50"`
}

// UserTimeAvailabilityResponse represents the response for user time availability data.
type UserTimeAvailabilityResponse struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	DayOfWeek int       `json:"day_of_week"`
	StartTime string    `json:"start_time"`
	EndTime   string    `json:"end_time"`
	Timezone  string    `json:"timezone"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// UserTimeAvailabilityListResponse represents the response for a list of user time availability.
type UserTimeAvailabilityListResponse struct {
	Availability []UserTimeAvailabilityResponse `json:"availability"`
	Total        int64                          `json:"total"`
	Page         int                            `json:"page"`
	PerPage      int                            `json:"per_page"`
	TotalPages   int                            `json:"total_pages"`
}

// ToResponse converts a UserTimeAvailability model to UserTimeAvailabilityResponse.
func (uta *UserTimeAvailability) ToResponse() UserTimeAvailabilityResponse {
	return UserTimeAvailabilityResponse{
		ID:        uta.ID,
		UserID:    uta.UserID,
		DayOfWeek: uta.DayOfWeek,
		StartTime: uta.StartTime,
		EndTime:   uta.EndTime,
		Timezone:  uta.Timezone,
		CreatedAt: uta.CreatedAt,
		UpdatedAt: uta.UpdatedAt,
	}
}

// Valid day of week values (0 = Sunday, 6 = Saturday).
var ValidDaysOfWeek = []int{0, 1, 2, 3, 4, 5, 6}

// DayNames maps day numbers to names.
var DayNames = map[int]string{
	0: "Sunday",
	1: "Monday",
	2: "Tuesday",
	3: "Wednesday",
	4: "Thursday",
	5: "Friday",
	6: "Saturday",
}

// IsValidDayOfWeek checks if the given day of week is valid.
func IsValidDayOfWeek(day int) bool {
	for _, validDay := range ValidDaysOfWeek {
		if day == validDay {
			return true
		}
	}
	return false
}

// GetDayName returns the name of the day for the given day number.
func GetDayName(day int) string {
	if name, exists := DayNames[day]; exists {
		return name
	}
	return "Unknown"
}
