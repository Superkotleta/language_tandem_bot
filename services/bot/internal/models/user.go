// Package models defines the data structures used throughout the application.
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

// User представляет пользователя системы.
type User struct {
	ID                     int       `db:"id"                       json:"id"`
	TelegramID             int64     `db:"telegram_id"              json:"telegramId"`
	Username               string    `db:"username"                 json:"username"`
	FirstName              string    `db:"first_name"               json:"firstName"`
	NativeLanguageCode     string    `db:"native_language_code"     json:"nativeLanguageCode"`
	TargetLanguageCode     string    `db:"target_language_code"     json:"targetLanguageCode"`
	TargetLanguageLevel    string    `db:"target_language_level"    json:"targetLanguageLevel"`
	InterfaceLanguageCode  string    `db:"interface_language_code"  json:"interfaceLanguageCode"`
	State                  string    `db:"state"                    json:"state"`
	Status                 string    `db:"status"                   json:"status"`
	ProfileCompletionLevel int       `db:"profile_completion_level" json:"profileCompletionLevel"`
	CreatedAt              time.Time `db:"created_at"               json:"createdAt"`
	UpdatedAt              time.Time `db:"updated_at"               json:"updatedAt"`
	Interests              []int     `db:"-"                        json:"interests"` // Не храним в БД, загружаем отдельно

	// Дополнительные поля для расширенного профиля
	TimeAvailability      *TimeAvailability      `db:"-" json:"timeAvailability"`      // Временная доступность
	FriendshipPreferences *FriendshipPreferences `db:"-" json:"friendshipPreferences"` // Предпочтения общения
}

// TimeAvailability - временная доступность пользователя.
type TimeAvailability struct {
	DayType      string   `db:"day_type"      json:"dayType"`      // weekdays, weekends, any, specific
	SpecificDays []string `db:"specific_days" json:"specificDays"` // массив дней для specific
	TimeSlot     string   `db:"time_slot"     json:"timeSlot"`     // morning, day, evening, late
}

// FriendshipPreferences - предпочтения по общению.
type FriendshipPreferences struct {
	ActivityType       string `db:"activity_type"           json:"activityType"`           // movies, games, educational
	CommunicationStyle string `db:"communication_style"     json:"communicationStyle"`     // text, voice_msg, meet_person
	CommunicationFreq  string `db:"communication_frequency" json:"communicationFrequency"` // spontaneous, weekly, daily
}
