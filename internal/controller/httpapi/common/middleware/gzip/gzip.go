package gzip

import (
	"loyalty/internal/controller/httpapi/common/middleware/routing"
	"net/http"

	"strings"

	"github.com/gin-gonic/gin"
)

const (
	defaultStatus  = http.StatusOK
	acceptEncoding = "Accept-Encoding"
	gzipValue      = "gzip"
)

// NewMiddleware возвращает middleware, который сжимает HTTP-ответ gzip'ом
func NewMiddleware(minSizeBytes int, routes ...string) gin.HandlerFunc {
	rules := routing.ParsePaths(routes)
	return func(ctx *gin.Context) {
		if len(rules) > 0 && !routing.Allowed(ctx, rules) {
			ctx.Next()
			return
		}
		if !clientAcceptsGzip(ctx.Request) {
			ctx.Next()
			return
		}

		writer := newGzipMinWriter(ctx.Writer, minSizeBytes)
		ctx.Writer = writer

		ctx.Next()

		if writer.passthrough {
			return
		}
		writer.finalize(minSizeBytes)
	}
}

func clientAcceptsGzip(request *http.Request) bool {
	if request == nil {
		return false
	}
	enc := request.Header.Get(acceptEncoding)
	enc = strings.ToLower(enc)
	return strings.Contains(enc, gzipValue)
}
