package model

import "github.com/shopspring/decimal"

// Response — ответ с балансом пользователя.
type Response struct {
	Current   decimal.Decimal `json:"current"`
	Withdrawn decimal.Decimal `json:"withdrawn"`
}
