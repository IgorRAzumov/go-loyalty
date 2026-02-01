package balance

import (
	"context"

	"loyalty/internal/domain/balance/model"
	balancesvc "loyalty/internal/domain/balance/service"
	"loyalty/internal/domain/balance/usecase"
)

// Usecase — реализация usecase.BalanceUsecase.
type Usecase struct {
	balanceService balancesvc.BalanceService
}

// NewUsecase создаёт usecase баланса.
func NewUsecase(balanceService balancesvc.BalanceService) *Usecase {
	return &Usecase{balanceService: balanceService}
}

// GetBalance возвращает баланс пользователя.
func (usecase *Usecase) GetBalance(ctx context.Context, userID int64) (model.Balance, error) {
	return usecase.balanceService.GetBalance(ctx, userID)
}

var _ usecase.BalanceUsecase = (*Usecase)(nil)
