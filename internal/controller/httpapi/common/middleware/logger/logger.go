package logger

import (
	"loyalty/internal/controller/httpapi/common/middleware/routing"
	applogger "loyalty/internal/logger"

	"github.com/gin-gonic/gin"
)

// maxLoggedBodyBytes — сколько байт тела мы максимум буферизуем для логирования на ошибках.
// Значение < 0 означает "без лимита".
const maxLoggedBodyBytes = -1

// NewMiddleware возвращает middleware для логирования HTTP-запросов.
// Если enableBodyLogging = false, буферизация тел запросов/ответов отключена (production режим).
func NewMiddleware(log applogger.Logger, enableBodyLogging bool, redactRoutes ...string) gin.HandlerFunc {
	rules := routing.ParsePaths(redactRoutes)

	if !enableBodyLogging {
		return func(ctx *gin.Context) {
			ctx.Next()

			status := ctx.Writer.Status()
			event := log.Info()
			if status >= 400 {
				event = log.Warn()
			}

			withCommonHTTPFields(event, ctx, status).
				Int("bytes", ctx.Writer.Size()).
				Message("http")
		}
	}

	return withBodyLogging(log, rules)
}

func withBodyLogging(log applogger.Logger, rules []string) func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		loggerWriter := newWriter(ctx.Writer, maxLoggedBodyBytes)
		ctx.Writer = loggerWriter

		loggerReader := newReader(ctx.Request.Body, maxLoggedBodyBytes)
		ctx.Request.Body = loggerReader

		ctx.Next()

		status := ctx.Writer.Status()
		event := log.Info()
		if status >= 400 {
			event = log.Warn()
		}

		withCommonHTTPFields(event, ctx, status).
			Int("bytes", ctx.Writer.Size()).
			Message("http")

		if status < 400 {
			return
		}

		route := ctx.FullPath()
		if route == "" {
			route = ctx.Request.URL.Path
		}
		requestBody := loggerReader.bytes()
		if routing.Allowed(ctx, rules) {
			requestBody = []byte("<<redacted>>")
		}

		withCommonHTTPFields(log.Warn(), ctx, status).
			String("request_body", string(requestBody)).
			String("response_body", string(loggerWriter.bytes())).
			Message("http body (error)")
	}
}

func withCommonHTTPFields(event applogger.Event, ctx *gin.Context, status int) applogger.Event {
	if ctx == nil || ctx.Request == nil || ctx.Request.URL == nil {
		return event.Int("status", status)
	}
	return event.
		String("method", ctx.Request.Method).
		String("path", ctx.Request.URL.Path).
		String("query", ctx.Request.URL.RawQuery).
		String("route", ctx.FullPath()).
		Int("status", status)
}
