package model

import "errors"

// ErrTooManyRequests возвращается при превышении rate limit (429 Too Many Requests).
var ErrTooManyRequests = errors.New("accrual system rate limit exceeded")

// ErrTemporarilyUnavailable возвращается, когда запросы в accrual временно прекращены
// (например, из-за открытого circuit breaker после серии ошибок).
var ErrTemporarilyUnavailable = errors.New("accrual system temporarily unavailable")
