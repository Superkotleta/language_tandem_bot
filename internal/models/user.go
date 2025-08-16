package models

import "time"

type User struct {
	ID                     int       `json:"id" db:"id"`
	TelegramID             int64     `json:"telegram_id" db:"telegram_id"`
	Username               string    `json:"username" db:"username"`
	FirstName              string    `json:"first_name" db:"first_name"`
	CreatedAt              time.Time `json:"created_at" db:"created_at"`
	State                  string    `json:"state" db:"state"`
	ProfileCompletionLevel int       `json:"profile_completion_level" db:"profile_completion_level"`
	Status                 string    `json:"status" db:"status"`
}

// States для state machine
const (
	StateStart            = ""
	StateWaitingLanguage  = "waiting_language"
	StateWaitingInterests = "waiting_interests"
	StateWaitingTime      = "waiting_time"
	StateComplete         = "complete"
)

// User statuses
const (
	StatusNotStarted = "not_started"
	StatusFilling    = "filling"
	StatusReady      = "ready"
	StatusMatched    = "matched"
	StatusWaiting    = "waiting"
)
