package usecase

import (
	"context"

	"loyalty/internal/domain/withdrawal/model"

	"github.com/shopspring/decimal"
)

// WithdrawalsUsecase описывает сценарии списаний и их истории.
type WithdrawalsUsecase interface {
	// Withdraw списывает баллы в счёт оплаты заказа.
	Withdraw(ctx context.Context, userID int64, orderNumber string, sum decimal.Decimal) error

	// ListWithdrawals возвращает список списаний пользователя (от новых к старым).
	ListWithdrawals(ctx context.Context, userID int64) ([]model.Withdrawal, error)
}
