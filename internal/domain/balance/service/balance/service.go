package balance

import (
	"context"

	"loyalty/internal/domain/balance/model"
	balancerepo "loyalty/internal/domain/balance/repository"
	balancesvc "loyalty/internal/domain/balance/service"
)

// Service — реализация balancesvc.BalanceService.
type Service struct {
	repo balancerepo.BalanceRepository
}

// NewService создаёт прикладной сервис баланса.
func NewService(repo balancerepo.BalanceRepository) *Service {
	return &Service{repo: repo}
}

// GetBalance возвращает баланс пользователя.
func (service *Service) GetBalance(ctx context.Context, userID int64) (model.Balance, error) {
	return service.repo.GetBalance(ctx, userID)
}

var _ balancesvc.BalanceService = (*Service)(nil)
