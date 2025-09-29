package models

import "time"

// Состояния пользователя
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

// Статусы пользователя
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

	// Дополнительные поля для расширенного профиля
	TimeAvailability      *TimeAvailability      `db:"-"` // Временная доступность
	FriendshipPreferences *FriendshipPreferences `db:"-"` // Предпочтения общения
}

// TimeAvailability - временная доступность пользователя
type TimeAvailability struct {
	DayType      string   `db:"day_type"`      // weekdays, weekends, any, specific
	SpecificDays []string `db:"specific_days"` // массив дней для specific
	TimeSlot     string   `db:"time_slot"`     // morning, day, evening, late
}

// FriendshipPreferences - предпочтения по общению
type FriendshipPreferences struct {
	ActivityType       string `db:"activity_type"`           // movies, games, casual_chat, creative, active, educational
	CommunicationStyle string `db:"communication_style"`     // text, voice_msg, audio_call, video_call, meet_person
	CommunicationFreq  string `db:"communication_frequency"` // spontaneous, weekly, daily, intensive
}
