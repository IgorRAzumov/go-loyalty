package retry

import (
	"context"
	"errors"
	"net"
	"time"
)

// ErrRetryable — маркер ошибки, при которой следует повторить попытку.
var ErrRetryable = errors.New("retryable")

// Config задаёт параметры повторных попыток.
type Config struct {
	MaxAttempts  int           // Максимум попыток (включая первую)
	InitialDelay time.Duration // Начальная задержка между попытками
	MaxDelay     time.Duration // Максимальная задержка
	Multiplier   float64       // Множитель для экспоненциального backoff
}

// DefaultConfig возвращает разумные значения по умолчанию.
func DefaultConfig() Config {
	return Config{
		MaxAttempts:  3,
		InitialDelay: 500 * time.Millisecond,
		MaxDelay:     5 * time.Second,
		Multiplier:   2.0,
	}
}

// Do выполняет fn с повторными попытками при retryable ошибках.
// Останавливается при успехе, при не-retryable ошибке или при отмене ctx.
func Do[Type any](ctx context.Context, cfg Config, request func() (Type, error), isRetryable func(err error) bool) (Type, error) {
	var lastError error
	var zero Type
	delay := cfg.InitialDelay

	for attempt := 0; attempt < cfg.MaxAttempts; attempt++ {
		result, err := request()
		if err == nil {
			return result, nil
		}

		lastError = err
		if !isRetryable(err) {
			return zero, err
		}
		if attempt == cfg.MaxAttempts-1 {
			return zero, err
		}

		select {
		case <-ctx.Done():
			return zero, lastError
		case <-time.After(delay):
			delay = time.Duration(float64(delay) * cfg.Multiplier)
			if delay > cfg.MaxDelay {
				delay = cfg.MaxDelay
			}
		}
	}

	return zero, lastError
}

// IsRetryableNetwork проверяет, является ли ошибка сетевой и подходящей для retry
// (таймаут, временная недоступность).
func IsRetryableNetwork(err error) bool {
	if errors.Is(err, ErrRetryable) {
		return true
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}

	var netErr net.Error
	if errors.As(err, &netErr) {
		return netErr.Temporary() || netErr.Timeout()
	}
	return false
}
