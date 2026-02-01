package handler

import (
	networkmodel "loyalty/internal/controller/httpapi/auth/model"
	common "loyalty/internal/controller/httpapi/common/model"
	"loyalty/internal/domain/auth/usecase"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// Handler — HTTP-хендлеры аутентификации (register/login).
type Handler struct {
	authUsecase usecase.AuthUsecase
}

// NewAuthHandler создаёт хендлеры аутентификации (register/login).
func NewAuthHandler(authUsecase usecase.AuthUsecase) *Handler {
	return &Handler{authUsecase: authUsecase}
}

// Register обрабатывает регистрацию пользователя: валидирует запрос и возвращает токен.
func (handler *Handler) Register(ctx *gin.Context) {
	var request networkmodel.LoginRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		common.WriteError(ctx, http.StatusBadRequest, common.CodeBadRequest)
		return
	}

	token, err := handler.authUsecase.Register(ctx.Request.Context(), request.Login, request.Password)
	if err != nil {
		log.Error().Err(err).Str("login", request.Login).Msg("register failed")
		status, code := common.MapError(err)
		common.WriteError(ctx, status, code)
		return
	}
	writeAuth(ctx, token)
	ctx.Status(http.StatusOK)
}

// Login обрабатывает аутентификацию пользователя: валидирует запрос и возвращает токен. В реальной жизни пароль
// не должен передаваться в открытом виде от клиента
func (handler *Handler) Login(ctx *gin.Context) {
	var request networkmodel.LoginRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		common.WriteError(ctx, http.StatusBadRequest, common.CodeBadRequest)
		return
	}

	token, err := handler.authUsecase.Login(ctx.Request.Context(), request.Login, request.Password)
	if err != nil {
		log.Error().Err(err).Str("login", request.Login).Msg("login failed")
		status, code := common.MapError(err)
		common.WriteError(ctx, status, code)
		return
	}
	writeAuth(ctx, token)
	ctx.Status(http.StatusOK)
}

func writeAuth(ctx *gin.Context, token string) {
	ctx.Header("Authorization", "Bearer "+token)
	ctx.SetCookie("token", token, 0, "/", "", false, true)
}
