package model

// User — доменная сущность пользователя для подсистемы аутентификации.
type User struct {
	ID           int64
	Login        string
	PasswordHash []byte
}
