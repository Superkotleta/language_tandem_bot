package domain

import "time"

// User - основная сущность пользователя, не привязанная к конкретной платформе
type User struct {
	ID        int64     `json:"id"`
	SocialID  string    `json:"social_id"` // ID в соцсети (строка, т.к. в VK/Discord могут быть нюансы)
	Platform  string    `json:"platform"`  // telegram, vk, whatsapp
	FirstName string    `json:"first_name"`
	Username  string    `json:"username,omitempty"`
	Language  string    `json:"language"` // Код языка интерфейса (ru, en)
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Constants for supported platforms
const (
	PlatformTelegram = "telegram"
	PlatformVK       = "vk"
)


