package withdrawals

import (
	"context"
	"time"

	"loyalty/internal/domain/withdrawal/model"
	withdrawalsrepo "loyalty/internal/domain/withdrawal/repository"
	withdrawalssvc "loyalty/internal/domain/withdrawal/service"

	"github.com/shopspring/decimal"
)

// Service — реализация withdrawalssvc.WithdrawalsService.
type Service struct {
	accountRepository     withdrawalsrepo.AccountRepository
	withdrawalsRepository withdrawalsrepo.WithdrawalsRepository
	now                   func() time.Time
}

// NewService создаёт прикладной сервис списаний.
func NewService(
	accountRepository withdrawalsrepo.AccountRepository,
	withdrawalsRepository withdrawalsrepo.WithdrawalsRepository,
) *Service {
	return &Service{
		accountRepository:     accountRepository,
		withdrawalsRepository: withdrawalsRepository,
		now:                   time.Now,
	}
}

// Withdraw списывает баллы в счёт оплаты заказа.
func (service *Service) Withdraw(ctx context.Context, userID int64, orderNumber string, sum decimal.Decimal) error {
	if sum.LessThanOrEqual(decimal.Zero) {
		return model.ErrInvalidWithdrawSum
	}
	return service.accountRepository.Withdraw(ctx, userID, orderNumber, sum, service.now())
}

// ListWithdrawals возвращает список списаний пользователя (от новых к старым).
func (service *Service) ListWithdrawals(ctx context.Context, userID int64) ([]model.Withdrawal, error) {
	return service.withdrawalsRepository.ListByUser(ctx, userID)
}

var _ withdrawalssvc.WithdrawalsService = (*Service)(nil)
