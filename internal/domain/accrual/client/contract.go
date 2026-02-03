package client

import (
	"context"
	"loyalty/internal/domain/accrual/model"
)

// AccrualClient представляет клиент для взаимодействия с системой расчёта начислений.
type AccrualClient interface {
	// GetOrderAccrual получает информацию о начислении для заказа.
	GetOrderAccrual(ctx context.Context, orderNumber string) (*model.Accrual, error)
}
