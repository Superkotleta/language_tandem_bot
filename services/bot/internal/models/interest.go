package models

// Interest представляет интерес пользователя
type Interest struct {
	ID   int    `db:"id"`
	Name string `db:"name"`
	Type string `db:"type"`
}
