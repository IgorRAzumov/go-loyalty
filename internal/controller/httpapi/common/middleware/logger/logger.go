package logger

import (
	"loyalty/internal/controller/httpapi/common/middleware/routing"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

var httpLog = zerolog.New(os.Stderr)

// maxLoggedBodyBytes — сколько байт тела мы максимум буферизуем для логирования на ошибках.
// Значение < 0 означает "без лимита".
const maxLoggedBodyBytes = -1

// NewMiddleware возвращает middleware для логирования HTTP-запросов.
// Если enableBodyLogging = false, буферизация тел запросов/ответов отключена (production режим).
func NewMiddleware(enableBodyLogging bool, redactRoutes ...string) gin.HandlerFunc {
	rules := routing.ParsePaths(redactRoutes)

	if !enableBodyLogging {
		return func(ctx *gin.Context) {
			ctx.Next()

			status := ctx.Writer.Status()
			event := httpLog.Info()
			if status >= 400 {
				event = httpLog.Warn()
			}

			withCommonHTTPFields(event, ctx, status).
				Int("bytes", ctx.Writer.Size()).
				Msg("http")
		}
	}

	return withBodyLogging(rules)
}

func withBodyLogging(rules []string) func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		loggerWriter := newWriter(ctx.Writer, maxLoggedBodyBytes)
		ctx.Writer = loggerWriter

		loggerReader := newReader(ctx.Request.Body, maxLoggedBodyBytes)
		ctx.Request.Body = loggerReader

		ctx.Next()

		status := ctx.Writer.Status()
		event := httpLog.Info()
		if status >= 400 {
			event = httpLog.Warn()
		}

		withCommonHTTPFields(event, ctx, status).
			Int("bytes", ctx.Writer.Size()).
			Msg("http")

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

		withCommonHTTPFields(httpLog.Warn(), ctx, status).
			Str("request_body", string(requestBody)).
			Str("response_body", string(loggerWriter.bytes())).
			Msg("http body (error)")
	}
}

func withCommonHTTPFields(event *zerolog.Event, ctx *gin.Context, status int) *zerolog.Event {
	if ctx == nil || ctx.Request == nil || ctx.Request.URL == nil {
		return event.Int("status", status)
	}
	return event.
		Str("method", ctx.Request.Method).
		Str("path", ctx.Request.URL.Path).
		Str("query", ctx.Request.URL.RawQuery).
		Str("route", ctx.FullPath()).
		Int("status", status)
}
