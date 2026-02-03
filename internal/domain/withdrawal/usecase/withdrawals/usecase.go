package withdrawals

import (
	"context"

	ordersmodel "loyalty/internal/domain/order/model"
	orderssvc "loyalty/internal/domain/order/service"
	withdrawalsmodel "loyalty/internal/domain/withdrawal/model"
	withdrawalssvc "loyalty/internal/domain/withdrawal/service"
	"loyalty/internal/domain/withdrawal/usecase"

	"github.com/shopspring/decimal"
)

// Usecase — реализация usecase.WithdrawalsUsecase.
type Usecase struct {
	withdrawalsService   withdrawalssvc.WithdrawalsService
	orderNumberValidator orderssvc.OrderNumberValidator
}

// NewUsecase создаёт usecase списаний.
func NewUsecase(
	withdrawalsService withdrawalssvc.WithdrawalsService,
	orderNumberValidator orderssvc.OrderNumberValidator,
) *Usecase {
	return &Usecase{
		withdrawalsService:   withdrawalsService,
		orderNumberValidator: orderNumberValidator,
	}
}

// Withdraw списывает баллы в счёт оплаты заказа.
func (usecase *Usecase) Withdraw(ctx context.Context, userID int64, orderNumber string, sum decimal.Decimal) error {
	normalized, err := usecase.orderNumberValidator.ValidateNumber(orderNumber)
	if err != nil {
		return ordersmodel.ErrInvalidOrderNumber
	}
	return usecase.withdrawalsService.Withdraw(ctx, userID, normalized, sum)
}

// ListWithdrawals возвращает список списаний пользователя (от новых к старым).
func (usecase *Usecase) ListWithdrawals(ctx context.Context, userID int64) ([]withdrawalsmodel.Withdrawal, error) {
	return usecase.withdrawalsService.ListWithdrawals(ctx, userID)
}

var _ usecase.WithdrawalsUsecase = (*Usecase)(nil)
