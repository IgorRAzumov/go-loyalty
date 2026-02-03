package service

import (
	"context"

	"loyalty/internal/domain/withdrawal/model"

	"github.com/shopspring/decimal"
)

// WithdrawalsService содержит прикладную логику работы со списаниями пользователя.
//
// Инкапсулирует работу с хранилищем (accountRepo + withdrawalsRepo) и используется
// usecase'ом как единый порт.
type WithdrawalsService interface {
	// Withdraw списывает баллы в счёт оплаты заказа.
	Withdraw(ctx context.Context, userID int64, orderNumber string, sum decimal.Decimal) error

	// ListWithdrawals возвращает список списаний пользователя (от новых к старым).
	ListWithdrawals(ctx context.Context, userID int64) ([]model.Withdrawal, error)
}
