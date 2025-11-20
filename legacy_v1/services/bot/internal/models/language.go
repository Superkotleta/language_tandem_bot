package models

import "time"

// Language представляет язык в системе.
type Language struct {
	ID                  int       `db:"id"                    json:"id"`
	Code                string    `db:"code"                  json:"code"`
	NameNative          string    `db:"name_native"           json:"nameNative"`
	NameEn              string    `db:"name_en"               json:"nameEn"`
	IsInterfaceLanguage bool      `db:"is_interface_language" json:"isInterfaceLanguage"`
	CreatedAt           time.Time `db:"created_at"            json:"createdAt"`
}
