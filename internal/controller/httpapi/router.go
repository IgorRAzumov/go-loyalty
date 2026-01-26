// Package httpapi реализует HTTP API сервиса (роутинг, wiring хендлеров и middleware).
package httpapi

import (
	"loyalty/internal/controller/httpapi/auth/handler"
	"loyalty/internal/controller/httpapi/auth/middleware"
	"loyalty/internal/domain/auth/service"
	"loyalty/internal/domain/auth/usecase"
	"net/http"

	"github.com/gin-gonic/gin"
)

// AuthDeps содержит зависимости HTTP-слоя авторизации, необходимые для регистрации маршрутов.
type AuthDeps struct {
	// AuthUsecase — usecase аутентификации (register/login), реализующий бизнес-сценарии.
	AuthUsecase usecase.AuthUsecase
	// TokenService — сервис токенов (инфраструктура), используемый для авторизации запросов.
	TokenService service.TokenService
}

// RegisterRoutes регистрирует все HTTP-маршруты сервиса на переданном gin.Engine.
func RegisterRoutes(routesEngine *gin.Engine, deps AuthDeps) {
	routesEngine.GET("/health", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "ok")
	})

	api := routesEngine.Group("/api")
	{
		registerAuthRoutes(api, deps.AuthUsecase)
		registerUserRoutes(api, deps.TokenService)
	}
}

func registerAuthRoutes(api *gin.RouterGroup, authUsecase usecase.AuthUsecase) {
	authHandler := handler.NewAuthHandler(authUsecase)
	api.POST("/user/register", authHandler.Register)
	api.POST("/user/login", authHandler.Login)
}

func registerUserRoutes(api *gin.RouterGroup, tokenService service.TokenService) {
	authed := api.Group("/user")
	authed.Use(middleware.NewAuthMiddleware(tokenService))

	// Минимальная реализация для теста авторизованной зоны
	authed.GET("/balance", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{
			"current":   0,
			"withdrawn": 0,
		})
	})
}
