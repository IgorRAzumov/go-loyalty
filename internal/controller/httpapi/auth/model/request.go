package model

// LoginRequest — тело запроса для регистрации/логина пользователя.
type LoginRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}
