package model

import (
	common "loyalty/internal/controller/httpapi/common/model"

	"github.com/shopspring/decimal"
)

// OrderResponseItem — элемент ответа списка заказов пользователя.
type OrderResponseItem struct {
	Number     string             `json:"number"`
	Status     string             `json:"status"`
	Accrual    *decimal.Decimal   `json:"accrual,omitempty"`
	UploadedAt common.RFC3339Time `json:"uploaded_at"`
}
