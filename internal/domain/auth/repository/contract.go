// Package repository содержит контракты хранилища
package repository

import (
	"context"
	"loyalty/internal/domain/auth/model"
)

// UserRepository — конракт репозитория пользователей (создание и поиск по логину).
type UserRepository interface {
	Create(ctx context.Context, login string, passwordHash []byte) (model.User, error)
	FindByLogin(ctx context.Context, login string) (model.User, error)
}
