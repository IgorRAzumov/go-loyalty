package mock

import (
	"context"
	"loyalty/internal/domain/accrual/model"
	"math/rand"
	"time"

	"github.com/shopspring/decimal"
)

var rng = rand.New(rand.NewSource(time.Now().UnixNano()))

// Client — простой mock для accrual client (для тестов или когда accrual система недоступна).
type Client struct {
}

// NewClient создаёт mock клиент с пустыми ответами.
func NewClient() *Client {
	return &Client{}
}

// NewClientWithDefaults создаёт mock клиент, который возвращает случайные значения от 0 до 100.
func NewClientWithDefaults() *Client {
	return NewClient()
}

// GetOrderAccrual возвращает случайный ответ с начислением от 0 до 100 или предопределённый (если задан).
func (c *Client) GetOrderAccrual(ctx context.Context, orderNumber string) (*model.Accrual, error) {
	statuses := []model.AccrualStatus{
		model.StatusRegistered,
		model.StatusProcessing,
		model.StatusProcessed,
		model.StatusInvalid,
	}
	status := statuses[rng.Intn(len(statuses))]

	var accrualPtr *decimal.Decimal
	if status == model.StatusProcessed {
		accrualPtr = decimalPtr(rng.Float64() * 100)
	}

	return &model.Accrual{
		Order:   orderNumber,
		Status:  status,
		Accrual: accrualPtr,
	}, nil
}

func decimalPtr(v float64) *decimal.Decimal {
	value := decimal.NewFromFloat(v)
	return &value
}
