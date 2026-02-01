package repository

import (
	"context"

	"loyalty/internal/domain/balance/model"
)

// BalanceRepository — порт репозитория накопительных счетов (получение баланса).
type BalanceRepository interface {
	// GetBalance возвращает баланс пользователя.
	GetBalance(ctx context.Context, userID int64) (model.Balance, error)
}
