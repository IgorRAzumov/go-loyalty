package util

import (
	"context"
	"time"
)

// queryTimeout — дефолтный таймаут для SQL запросов (защита от зависших запросов).
var queryTimeout = 3 * time.Second

// SetQueryTimeout устанавливает глобальный таймаут для всех SQL запросов.
// Должен вызываться при инициализации приложения (до создания репозиториев).
func SetQueryTimeout(timeout time.Duration) {
	if timeout > 0 {
		queryTimeout = timeout
	}
}

// WithQueryTimeout создаёт контекст с deadline для выполнения SQL запроса.
// Если родительский контекст уже имеет deadline, выбирается более ранний.
func WithQueryTimeout(parent context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(parent, queryTimeout)
}
