package middleware

import (
	"loyalty/internal/controller/httpapi/auth/authctx"
	common "loyalty/internal/controller/httpapi/common/model"
	"loyalty/internal/domain/auth/service"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	bearer = "bearer "
	token  = "token"
)

// NewAuthMiddleware создаёт middleware авторизации по JWT (Bearer или cookie).
func NewAuthMiddleware(tokenService service.TokenService) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		token := tokenFromRequest(ctx)
		if token == "" {
			common.WriteError(ctx, http.StatusUnauthorized, common.CodeUnauthorized)
			ctx.Abort()
			return
		}
		claims, err := tokenService.ParseToken(token)
		if err != nil || claims == nil || claims.UserID <= 0 {
			common.WriteError(ctx, http.StatusUnauthorized, common.CodeUnauthorized)
			ctx.Abort()
			return
		}
		ctx.Request = ctx.Request.WithContext(authctx.WithUserID(ctx.Request.Context(), claims.UserID))
		ctx.Next()
	}
}

func tokenFromRequest(ctx *gin.Context) string {
	if header := ctx.GetHeader("Authorization"); header != "" {
		if value := strings.TrimSpace(header); strings.HasPrefix(strings.ToLower(value), bearer) {
			return strings.TrimSpace(value[len(bearer):])
		}
	}
	if value, err := ctx.Cookie(token); err == nil {
		return strings.TrimSpace(value)
	}
	return ""
}
