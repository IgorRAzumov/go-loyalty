package order

import (
	"context"

	"loyalty/internal/domain/order/model"
	orderssvc "loyalty/internal/domain/order/service"
	"loyalty/internal/domain/order/usecase"
)

// Usecase — реализация usecase.OrdersUsecase.
type Usecase struct {
	ordersService orderssvc.OrdersService
}

// NewUsecase создаёт usecase заказов.
func NewUsecase(ordersService orderssvc.OrdersService) *Usecase {
	return &Usecase{ordersService: ordersService}
}

// UploadOrder загружает номер заказа пользователя.
func (usecase *Usecase) UploadOrder(ctx context.Context, userID int64, number string) error {
	return usecase.ordersService.UploadOrder(ctx, userID, number)
}

// LoadOrders ListOrders возвращает список заказов пользователя (от новых к старым).
func (usecase *Usecase) LoadOrders(ctx context.Context, userID int64) ([]model.Order, error) {
	return usecase.ordersService.LoadOrders(ctx, userID)
}

var _ usecase.OrdersUsecase = (*Usecase)(nil)
