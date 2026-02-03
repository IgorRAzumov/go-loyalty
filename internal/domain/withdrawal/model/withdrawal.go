package model

import (
	"errors"
	"time"

	"github.com/shopspring/decimal"
)

var (
	// ErrInvalidWithdrawSum возвращается, если сумма списания некорректна (<= 0).
	ErrInvalidWithdrawSum = errors.New("invalid withdraw sum")
	// ErrInsufficientFunds возвращается, если на счету недостаточно средств для списания.
	ErrInsufficientFunds = errors.New("insufficient funds")
)

// Withdrawal — доменная сущность списания баллов пользователем.
type Withdrawal struct {
	UserID      int64
	OrderNumber string
	Sum         decimal.Decimal
	ProcessedAt time.Time
}
