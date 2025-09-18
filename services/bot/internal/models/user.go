package models

import "time"

// Состояния пользователя.
const (
	StateNew                    = "new"
	StateWaitingLanguage        = "waiting_language"
	StateWaitingTargetLanguage  = "waiting_target_language"
	StateWaitingLanguageLevel   = "waiting_language_level"
	StateWaitingInterests       = "waiting_interests"
	StateWaitingTime            = "waiting_time"
	StateWaitingFeedback        = "waiting_feedback"
	StateWaitingFeedbackContact = "waiting_feedback_contact" // Для сбора контактной информации без username
	StateActive                 = "active"
)

// Статусы пользователя.
const (
	StatusNew     = "new"
	StatusFilling = "filling_profile"
	StatusActive  = "active"
	StatusPaused  = "paused"
)

type User struct {
	ID                     int       `db:"id"`
	TelegramID             int64     `db:"telegram_id"`
	Username               string    `db:"username"`
	FirstName              string    `db:"first_name"`
	NativeLanguageCode     string    `db:"native_language_code"`
	TargetLanguageCode     string    `db:"target_language_code"`
	TargetLanguageLevel    string    `db:"target_language_level"`
	InterfaceLanguageCode  string    `db:"interface_language_code"`
	State                  string    `db:"state"`
	Status                 string    `db:"status"`
	ProfileCompletionLevel int       `db:"profile_completion_level"`
	CreatedAt              time.Time `db:"created_at"`
	UpdatedAt              time.Time `db:"updated_at"`
	Interests              []int     `db:"-"` // Не храним в БД, загружаем отдельно
}
