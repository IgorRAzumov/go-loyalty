package httpapi

import (
	"loyalty/internal/controller/httpapi/auth/handler"
	"loyalty/internal/controller/httpapi/auth/middleware"
	userbalance "loyalty/internal/controller/httpapi/balance/handler"
	"loyalty/internal/controller/httpapi/common/middleware/gzip"
	"loyalty/internal/controller/httpapi/common/middleware/logger"
	"loyalty/internal/controller/httpapi/common/middleware/ratelimit"
	userorders "loyalty/internal/controller/httpapi/order/handler"
	userwithdrawals "loyalty/internal/controller/httpapi/withdrawal/handler"
	"loyalty/internal/domain/auth/service"
	authusecase "loyalty/internal/domain/auth/usecase"
	balanceusecase "loyalty/internal/domain/balance/usecase"
	ordersusecase "loyalty/internal/domain/order/usecase"
	withdrawalsusecase "loyalty/internal/domain/withdrawal/usecase"

	"github.com/gin-gonic/gin"
)

// Deps содержит зависимости HTTP-слоя, необходимые для регистрации маршрутов.
type Deps struct {
	AuthUsecase        authusecase.AuthUsecase
	OrdersUsecase      ordersusecase.OrdersUsecase
	BalanceUsecase     balanceusecase.BalanceUsecase
	WithdrawalsUsecase withdrawalsusecase.WithdrawalsUsecase
	TokenService       service.TokenService

	EnableHTTPBodyLogging bool

	AuthRateLimitRPS   int
	AuthRateLimitBurst int
}

func RegisterRoutes(router *gin.Engine, deps Deps) {
	registerRoutes(router, deps)
}

func InitRouter(deps Deps) *gin.Engine {
	router := gin.New()
	router.Use(logger.NewMiddleware(deps.EnableHTTPBodyLogging, "/api/user/register", "/api/user/login"))
	router.Use(gin.Recovery())
	registerRoutes(router, deps)
	return router
}

func registerRoutes(routesEngine *gin.Engine, deps Deps) {
	routesEngine.GET("/health", func(ctx *gin.Context) {
		ctx.String(200, "ok")
	})

	api := routesEngine.Group("/api")
	registerAuthRoutes(api, deps)

	authed := api.Group("/user")
	authed.Use(middleware.NewAuthMiddleware(deps.TokenService))
	authed.Use(gzip.NewMiddleware(1024, "/api/user/orders", "/api/user/withdrawals"))

	registerOrdersRoutes(authed, deps.OrdersUsecase)
	registerBalanceRoutes(authed, deps.BalanceUsecase)
	registerWithdrawalsRoutes(authed, deps.WithdrawalsUsecase)
}

func registerAuthRoutes(api *gin.RouterGroup, deps Deps) {
	authHandler := handler.NewAuthHandler(deps.AuthUsecase)
	rateLimiter := ratelimit.NewMiddleware(deps.AuthRateLimitRPS, deps.AuthRateLimitBurst)
	api.POST("/user/register", rateLimiter, authHandler.Register)
	api.POST("/user/login", rateLimiter, authHandler.Login)
}

func registerOrdersRoutes(authed *gin.RouterGroup, ordersUsecase ordersusecase.OrdersUsecase) {
	ordersHandler := userorders.NewHandler(ordersUsecase)
	authed.POST("/orders", ordersHandler.UploadOrder)
	authed.GET("/orders", ordersHandler.ListOrders)
}

func registerBalanceRoutes(authed *gin.RouterGroup, balanceUsecase balanceusecase.BalanceUsecase) {
	balanceHandler := userbalance.NewHandler(balanceUsecase)
	authed.GET("/balance", balanceHandler.Get)
}

func registerWithdrawalsRoutes(authed *gin.RouterGroup, withdrawalsUsecase withdrawalsusecase.WithdrawalsUsecase) {
	withdrawalsHandler := userwithdrawals.NewHandler(withdrawalsUsecase)
	authed.POST("/balance/withdraw", withdrawalsHandler.Withdraw)
	authed.GET("/withdrawals", withdrawalsHandler.List)
}
