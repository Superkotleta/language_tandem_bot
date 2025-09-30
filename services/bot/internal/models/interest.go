package models

import "time"

// Interest представляет интерес пользователя
type Interest struct {
	ID           int       `db:"id"`
	KeyName      string    `db:"key_name"`
	CategoryID   int       `db:"category_id"`
	DisplayOrder int       `db:"display_order"`
	Type         string    `db:"type"`
	CreatedAt    time.Time `db:"created_at"`
}
