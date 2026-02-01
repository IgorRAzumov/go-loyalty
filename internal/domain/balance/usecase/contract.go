package usecase

import (
	"context"

	"loyalty/internal/domain/balance/model"
)

// BalanceUsecase описывает сценарии получения баланса пользователя.
type BalanceUsecase interface {
	GetBalance(ctx context.Context, userID int64) (model.Balance, error)
}
