package model

import "github.com/shopspring/decimal"

// WithdrawRequest — тело запроса на списание баллов.
type WithdrawRequest struct {
	Order string          `json:"order"`
	Sum   decimal.Decimal `json:"sum"`
}
