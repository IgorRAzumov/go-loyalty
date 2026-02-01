package model

import "github.com/shopspring/decimal"

// Balance — состояние накопительного счёта пользователя (текущий баланс и сумма списаний).
type Balance struct {
	Current   decimal.Decimal
	Withdrawn decimal.Decimal
}
