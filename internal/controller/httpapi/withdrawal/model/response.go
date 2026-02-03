package model

import (
	common "loyalty/internal/controller/httpapi/common/model"

	"github.com/shopspring/decimal"
)

// WithdrawalResponseItem — элемент ответа списка списаний.
type WithdrawalResponseItem struct {
	Order       string             `json:"order"`
	Sum         decimal.Decimal    `json:"sum"`
	ProcessedAt common.RFC3339Time `json:"processed_at"`
}
