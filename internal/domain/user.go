package domain

import "time"

// User status constants
const (
	StatusFillingProfile = "filling_profile"
	StatusActive         = "active"
	StatusPaused         = "paused"
	StatusBanned         = "banned"
)

// User represents a user in the system
type User struct {
	ID            string    `json:"id"`        // UUID
	SocialID      string    `json:"social_id"` // ID from platform
	Platform      string    `json:"platform"`  // telegram, vk
	FirstName     string    `json:"first_name"`
	Username      string    `json:"username,omitempty"`
	NativeLang    string    `json:"native_lang,omitempty"`
	TargetLang    string    `json:"target_lang,omitempty"`
	TargetLevel   string    `json:"target_level,omitempty"`
	InterfaceLang string    `json:"interface_lang"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// Constants for supported platforms
const (
	PlatformTelegram = "telegram"
	PlatformVK       = "vk"
)
