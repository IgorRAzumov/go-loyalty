package ratelimit

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// NewMiddleware создаёт middleware для ограничения частоты запросов (rate limiting).
// Использует token bucket алгоритм: rps запросов/сек + burst для всплесков.
func NewMiddleware(rps int, burst int) gin.HandlerFunc {
	limiter := rate.NewLimiter(rate.Limit(rps), burst)

	return func(ctx *gin.Context) {
		if !limiter.Allow() {
			ctx.Header("Retry-After", "1")
			ctx.Status(http.StatusTooManyRequests)
			ctx.Abort()
			return
		}
		ctx.Next()
	}
}
