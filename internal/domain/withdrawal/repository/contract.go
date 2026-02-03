package repository

import (
	"context"
	"time"

	"loyalty/internal/domain/withdrawal/model"

	"github.com/shopspring/decimal"
)

// WithdrawalsRepository — порт репозитория списаний.
type WithdrawalsRepository interface {
	// ListByUser возвращает списания пользователя.
	ListByUser(ctx context.Context, userID int64) ([]model.Withdrawal, error)
}

// AccountRepository — порт накопительного счёта для сценариев списаний.
type AccountRepository interface {
	// Withdraw списывает сумму с баланса (атомарно), создавая запись о списании.
	Withdraw(ctx context.Context, userID int64, orderNumber string, sum decimal.Decimal, now time.Time) error
}
