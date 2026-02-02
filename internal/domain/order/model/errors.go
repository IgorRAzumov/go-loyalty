package model

import (
	"errors"
	"fmt"
	"time"
)

var (
	// ErrInvalidOrderNumber возвращается при невалидном номере заказа (например, не проходит алгоритм Луна).
	ErrInvalidOrderNumber = errors.New("invalid order number")
	// ErrOrderAlreadyUploaded возвращается, если номер заказа уже был загружен этим пользователем.
	ErrOrderAlreadyUploaded = errors.New("order already uploaded by this user")
	// ErrOrderAlreadyUploadedByAnother возвращается, если номер заказа уже был загружен другим пользователем.
	ErrOrderAlreadyUploadedByAnother = errors.New("order already uploaded by another user")

	// ErrAccrualNotRegistered возвращается, когда внешний сервис начислений не знает о заказе (HTTP 204).
	ErrAccrualNotRegistered = errors.New("accrual order not registered")
	// ErrAccrualRateLimited возвращается при превышении лимита запросов к сервису начислений (HTTP 429).
	ErrAccrualRateLimited = errors.New("accrual rate limited")
)

// RateLimitError описывает ситуацию, когда внешний сервис начислений ограничил запросы.
// RetryAfter указывает рекомендуемую паузу перед повтором.
type RateLimitError struct {
	RetryAfter time.Duration
}

// Error возвращает человекочитаемое описание ошибки.
func (rateLimitError RateLimitError) Error() string {
	if rateLimitError.RetryAfter > 0 {
		return fmt.Sprintf("%v (retry after %s)", ErrAccrualRateLimited, rateLimitError.RetryAfter)
	}
	return ErrAccrualRateLimited.Error()
}

// Unwrap позволяет проверять ошибку через errors.Is(err, ErrAccrualRateLimited).
func (rateLimitError RateLimitError) Unwrap() error { return ErrAccrualRateLimited }
