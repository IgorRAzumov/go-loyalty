package service

import (
	"context"
	"loyalty/internal/domain/auth/model"
	"time"
)

// AuthService содержит доменную логику аутентификации, не зависящую от инфраструктуры:
// валидация логина/пароля и работа с хешами паролей.
type AuthService interface {
	ValidateLogin(login string) (normalized string, err error)
	ValidatePassword(password string) error
	HashPassword(password string) ([]byte, error)
	ComparePassword(hash []byte, password string) error
}

// TokenService — инфраструктурный сервис токенов (выпуск и проверка access-token).
type TokenService interface {
	IssueToken(userID int64, login string, now time.Time) (string, error)
	ParseToken(token string) (*model.Claim, error)
}

// UserService инкапсулирует доступ к пользователям и их инварианты (например, нормализацию логина).
type UserService interface {
	CreateUser(ctx context.Context, login string, passwordHash []byte) (model.User, error)
	FindUserByLogin(ctx context.Context, login string) (model.User, error)
}
