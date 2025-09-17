package models

import "time"

// Language представляет язык в системе
type Language struct {
	ID                  int       `db:"id"`
	Code                string    `db:"code"`
	NameNative          string    `db:"name_native"`
	NameEn              string    `db:"name_en"`
	IsInterfaceLanguage bool      `db:"is_interface_language"`
	CreatedAt           time.Time `db:"created_at"`
}
