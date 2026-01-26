// Package user содержит доменный UserService (нормализация логина и работа с репозиторием).
package user

import (
	"context"
	"loyalty/internal/domain/auth/service"
	"strings"

	"loyalty/internal/domain/auth/model"
	"loyalty/internal/domain/auth/repository"
)

type userService struct {
	repo repository.UserRepository
}

const maxLoginLength = 64

// NewUserService создаёт доменный сервис пользователей (нормализация логина и доступ к репозиторию).
func NewUserService(repo repository.UserRepository) service.UserService {
	return &userService{repo: repo}
}

// CreateUser нормализует логин, проверяет ограничения и создаёт пользователя в репозитории.
func (service *userService) CreateUser(ctx context.Context, login string, passwordHash []byte) (model.User, error) {
	normalized, err := normalizeLogin(login)
	if err != nil {
		return model.User{}, err
	}
	return service.repo.Create(ctx, normalized, passwordHash)
}

// FindUserByLogin нормализует логин, проверяет ограничения и ищет пользователя в репозитории.
func (service *userService) FindUserByLogin(ctx context.Context, login string) (model.User, error) {
	normalized, err := normalizeLogin(login)
	if err != nil {
		return model.User{}, err
	}
	return service.repo.FindByLogin(ctx, normalized)
}

func normalizeLogin(login string) (string, error) {
	normalized := strings.TrimSpace(login)
	if normalized == "" {
		return "", model.ErrInvalidInput
	}
	if len(normalized) > maxLoginLength {
		return "", model.ErrInvalidInput
	}
	return normalized, nil
}

var _ service.UserService = (*userService)(nil)
