package domain

import "time"

// User status constants.
const (
	StatusFillingProfile = "filling_profile"
	StatusActive         = "active"
	StatusPaused         = "paused"
	StatusBanned         = "banned"
)

// User represents a user in the system.
type User struct {
	ID            string    `json:"id"`       // UUID
	SocialID      string    `json:"socialId"` // ID from platform
	Platform      string    `json:"platform"` // telegram, vk
	FirstName     string    `json:"firstName"`
	Username      string    `json:"username,omitempty"`
	NativeLang    string    `json:"nativeLang,omitempty"`
	TargetLang    string    `json:"targetLang,omitempty"`
	TargetLevel   string    `json:"targetLevel,omitempty"`
	InterfaceLang string    `json:"interfaceLang"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

// Constants for supported platforms.
const (
	PlatformTelegram = "telegram"
	PlatformVK       = "vk"
)
