package model

import (
	"time"

	"github.com/shopspring/decimal"
)

// Status — статус обработки заказа в системе лояльности.
type Status string

const (
	// StatusNew — заказ загружен в систему, но не попал в обработку.
	StatusNew Status = "NEW"
	// StatusProcessing — вознаграждение за заказ рассчитывается.
	StatusProcessing Status = "PROCESSING"
	// StatusInvalid — система расчёта вознаграждений отказала в расчёте.
	StatusInvalid Status = "INVALID"
	// StatusProcessed — данные по заказу проверены и информация о расчёте успешно получена.
	StatusProcessed Status = "PROCESSED"
)

// Order — доменная сущность заказа в системе лояльности.
type Order struct {
	Number     string
	UserID     int64
	Status     Status
	Accrual    *decimal.Decimal
	UploadedAt time.Time
}
