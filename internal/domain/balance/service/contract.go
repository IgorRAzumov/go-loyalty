package service

import (
	"context"

	"loyalty/internal/domain/balance/model"
)

// BalanceService содержит прикладную логику работы с балансом пользователя.
type BalanceService interface {
	// GetBalance возвращает баланс пользователя (current + withdrawn).
	GetBalance(ctx context.Context, userID int64) (model.Balance, error)
}
