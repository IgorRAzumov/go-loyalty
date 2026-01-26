// Package auth содержит реализацию доменного AuthService (валидация и пароли).
package auth

import (
	"loyalty/internal/domain/auth/model"
	"loyalty/internal/domain/auth/service"
	authutil "loyalty/internal/util/auth"
	"strings"
)

// Service — реализация доменного сервиса аутентификации (валидация и пароли).
type Service struct {
}

// bcrypt в Go учитывает только первые 72 байта пароля, поэтому мы явно ограничиваем длину.
const maxPasswordLength = 72

// NewAuthService создаёт доменный сервис аутентификации (валидация и пароли).
func NewAuthService() *Service {
	return &Service{}
}

// ValidateLogin нормализует логин (trim) и проверяет базовые ограничения.
func (service *Service) ValidateLogin(login string) (string, error) {
	normalized := strings.TrimSpace(login)
	if normalized == "" {
		return "", model.ErrInvalidInput
	}
	return normalized, nil
}

// ValidatePassword проверяет пароль на минимальные требования (в т.ч. длину).
func (service *Service) ValidatePassword(password string) error {
	if password == "" {
		return model.ErrInvalidInput
	}
	if len(password) < 10 {
		return model.ErrPasswordTooShort
	}
	if len(password) > maxPasswordLength {
		return model.ErrPasswordTooLong
	}
	return nil
}

// HashPassword валидирует и хеширует пароль для безопасного хранения.
func (service *Service) HashPassword(password string) ([]byte, error) {
	if err := service.ValidatePassword(password); err != nil {
		return nil, err
	}
	return authutil.HashPassword(password)
}

// ComparePassword валидирует пароль и сравнивает его с хешем.
func (service *Service) ComparePassword(hash []byte, password string) error {
	if err := service.ValidatePassword(password); err != nil {
		return err
	}
	return authutil.ComparePassword(hash, password)
}

var _ service.AuthService = (*Service)(nil)
