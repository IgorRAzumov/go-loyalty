package model

import "github.com/shopspring/decimal"

// AccrualStatus представляет статус расчёта начисления в системе accrual.
type AccrualStatus string

const (
	// StatusRegistered — заказ зарегистрирован, но вознаграждение не рассчитано.
	StatusRegistered AccrualStatus = "REGISTERED"
	// StatusInvalid — заказ не принят к расчёту, и вознаграждение не будет начислено.
	StatusInvalid AccrualStatus = "INVALID"
	// StatusProcessing — расчёт начисления в процессе.
	StatusProcessing AccrualStatus = "PROCESSING"
	// StatusProcessed — расчёт начисления окончен.
	StatusProcessed AccrualStatus = "PROCESSED"
)

// Accrual представляет ответ от системы accrual о статусе начисления.
type Accrual struct {
	Order   string           `json:"order"`
	Status  AccrualStatus    `json:"status"`
	Accrual *decimal.Decimal `json:"accrual,omitempty"`
}
