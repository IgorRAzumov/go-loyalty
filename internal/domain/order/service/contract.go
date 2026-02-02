package service

import (
	"context"

	accrualmodel "loyalty/internal/domain/accrual/model"
	"loyalty/internal/domain/order/model"

	"github.com/shopspring/decimal"
)

// OrderNumberValidator валидирует и нормализует номера заказов.
type OrderNumberValidator interface {
	// ValidateNumber нормализует и валидирует номер заказа.
	ValidateNumber(number string) (normalized string, err error)
}

// OrdersService содержит прикладную логику работы с заказами пользователя.
//
// В отличие от OrderNumberValidator (чисто валидация/нормализация номера), этот сервис
// инкапсулирует работу с хранилищем (репозиторием) и используется usecase'ом как единый порт.
type OrdersService interface {
	// UploadOrder валидирует/нормализует номер заказа и сохраняет его в хранилище.
	UploadOrder(ctx context.Context, userID int64, number string) error

	// LoadOrders возвращает список заказов пользователя.
	LoadOrders(ctx context.Context, userID int64) ([]model.Order, error)

	// UpdateFromAccrual обновляет статус заказа по данным из системы accrual.
	// Инкапсулирует бизнес-логику маппинга статусов и правила обновления.
	UpdateFromAccrual(ctx context.Context, orderNumber string, accrualStatus accrualmodel.AccrualStatus, accrual *decimal.Decimal) error
}

// AccrualService — порт внешнего сервиса расчёта начислений.
type AccrualService interface {
	// GetAccrual возвращает статус расчёта и сумму начисления (если она есть).
	GetAccrual(ctx context.Context, orderNumber string) (status string, accrual *decimal.Decimal, err error)
}
