package routing

import (
	"strings"

	"github.com/gin-gonic/gin"
)

// ParsePaths нормализует список маршрутов, отбрасывая пустые элементы.
// Формат: только путь (например, "/api/user/orders").
func ParsePaths(routes []string) []string {
	if len(routes) == 0 {
		return nil
	}

	paths := make([]string, 0, len(routes))
	for _, route := range routes {
		path := strings.TrimSpace(route)
		if path == "" {
			continue
		}
		paths = append(paths, path)
	}
	return paths
}

// Allowed возвращает true, если текущий маршрут входит в список paths.
func Allowed(ctx *gin.Context, paths []string) bool {
	if ctx == nil {
		return false
	}
	path := ctx.FullPath()

	if path == "" && ctx.Request != nil && ctx.Request.URL != nil {
		path = ctx.Request.URL.Path
	}

	if path == "" {
		return false
	}

	for _, rule := range paths {
		if rule == path {
			return true
		}
	}
	return false
}
