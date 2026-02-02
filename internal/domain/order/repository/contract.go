package repository

import (
	"context"

	"loyalty/internal/domain/order/model"

	"github.com/shopspring/decimal"
)

// OrdersRepository — порт репозитория заказов (загрузка, выдача и обновление статусов/начислений).
type OrdersRepository interface {
	// Create создаёт заказ со статусом NEW для пользователя.
	Create(ctx context.Context, userID int64, number string) error

	// ListByUser возвращает список заказов пользователя
	ListByUser(ctx context.Context, userID int64) ([]model.Order, error)

	// ListPending возвращает все заказы, которые нужно проверить/обновить через accrual-сервис.
	ListPending(ctx context.Context) ([]model.Order, error)

	// UpdateFromAccrual обновляет статус/начисление заказа по данным внешнего accrual-сервиса.
	UpdateFromAccrual(ctx context.Context, number string, status model.Status, accrual *decimal.Decimal) error
}
