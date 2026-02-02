package orders

import (
	"context"
	"fmt"

	accrualmodel "loyalty/internal/domain/accrual/model"
	"loyalty/internal/domain/order/model"
	ordersrepo "loyalty/internal/domain/order/repository"
	orderssvc "loyalty/internal/domain/order/service"

	"github.com/shopspring/decimal"
)

// Service — реализация orderssvc.OrdersService.
type Service struct {
	repo            ordersrepo.OrdersRepository
	numberValidator orderssvc.OrderNumberValidator
}

// NewService создаёт прикладной сервис заказов.
func NewService(repo ordersrepo.OrdersRepository, numberValidator orderssvc.OrderNumberValidator) *Service {
	return &Service{repo: repo, numberValidator: numberValidator}
}

// UploadOrder валидирует/нормализует номер заказа и сохраняет его.
func (service *Service) UploadOrder(ctx context.Context, userID int64, number string) error {
	normalized, err := service.numberValidator.ValidateNumber(number)
	if err != nil {
		return model.ErrInvalidOrderNumber
	}
	return service.repo.Create(ctx, userID, normalized)
}

// LoadOrders возвращает список заказов пользователя.
func (service *Service) LoadOrders(ctx context.Context, userID int64) ([]model.Order, error) {
	return service.repo.ListByUser(ctx, userID)
}

// UpdateFromAccrual обновляет статус заказа по данным из системы accrual.
// Инкапсулирует бизнес-логику маппинга статусов и правила обновления.
func (service *Service) UpdateFromAccrual(
	ctx context.Context,
	orderNumber string,
	accrualStatus accrualmodel.AccrualStatus,
	accrual *decimal.Decimal,
) error {
	// Маппим статус из accrual в доменный статус заказа
	orderStatus := mapAccrualStatusToOrderStatus(accrualStatus)

	// Обновляем заказ в репозитории
	if err := service.repo.UpdateFromAccrual(ctx, orderNumber, orderStatus, accrual); err != nil {
		return fmt.Errorf("update order from accrual: %w", err)
	}

	return nil
}

// mapAccrualStatusToOrderStatus маппит статус из системы accrual в статус заказа.
func mapAccrualStatusToOrderStatus(accrualStatus accrualmodel.AccrualStatus) model.Status {
	switch accrualStatus {
	case accrualmodel.StatusRegistered, accrualmodel.StatusProcessing:
		return model.StatusProcessing
	case accrualmodel.StatusInvalid:
		return model.StatusInvalid
	case accrualmodel.StatusProcessed:
		return model.StatusProcessed
	default:
		// Неизвестный статус - оставляем в PROCESSING для повторной проверки
		return model.StatusProcessing
	}
}

var _ orderssvc.OrdersService = (*Service)(nil)
