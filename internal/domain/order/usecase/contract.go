package usecase

import (
	"context"

	"loyalty/internal/domain/order/model"
)

// OrdersUsecase описывает сценарии загрузки и просмотра заказов пользователя.
type OrdersUsecase interface {
	// UploadOrder загружает номер заказа пользователя.
	UploadOrder(ctx context.Context, userID int64, number string) error

	// LoadOrders ListOrders возвращает список заказов пользователя (от новых к старым).
	LoadOrders(ctx context.Context, userID int64) ([]model.Order, error)
}
