// Package memory содержит in-memory реализации адаптеров (для локального запуска/тестов).
package memory

import (
	"context"
	"loyalty/internal/domain/auth/model"
	"loyalty/internal/domain/auth/repository"
	"sync"
)

// UserRepository — простая in-memory реализация repository.UserRepository (хранит пользователей по логину).
type UserRepository struct {
	mu      sync.RWMutex
	nextID  int64
	byLogin map[string]model.User
}

// NewRepository создаёт новый in-memory репозиторий пользователей.
func NewRepository() *UserRepository {
	return &UserRepository{
		nextID:  1,
		byLogin: make(map[string]model.User),
	}
}

// Create создаёт пользователя с уникальным логином.
func (repository *UserRepository) Create(_ context.Context, login string, passwordHash []byte) (model.User, error) {
	repository.mu.Lock()
	defer repository.mu.Unlock()

	if _, ok := repository.byLogin[login]; ok {
		return model.User{}, model.ErrLoginTaken
	}
	user := model.User{
		ID:           repository.nextID,
		Login:        login,
		PasswordHash: append([]byte(nil), passwordHash...),
	}
	repository.nextID++
	repository.byLogin[login] = user
	return user, nil
}

// FindByLogin возвращает пользователя по логину или ErrNotFound.
func (repository *UserRepository) FindByLogin(_ context.Context, login string) (model.User, error) {
	repository.mu.RLock()
	defer repository.mu.RUnlock()

	user, ok := repository.byLogin[login]
	if !ok {
		return model.User{}, model.ErrNotFound
	}
	user.PasswordHash = append([]byte(nil), user.PasswordHash...)
	return user, nil
}

var _ repository.UserRepository = (*UserRepository)(nil)
