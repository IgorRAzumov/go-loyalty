package handler

import (
	"loyalty/internal/controller/httpapi/auth/authctx"
	"loyalty/internal/controller/httpapi/balance/model"
	"net/http"

	common "loyalty/internal/controller/httpapi/common/model"
	balusecase "loyalty/internal/domain/balance/usecase"

	"github.com/gin-gonic/gin"
)

// Handler — HTTP-хендлеры сценариев баланса пользователя.
type Handler struct {
	usecase balusecase.BalanceUsecase
}

// NewHandler создаёт хендлеры баланса пользователя.
func NewHandler(usecase balusecase.BalanceUsecase) *Handler { return &Handler{usecase: usecase} }

// Get возвращает текущий баланс пользователя.
func (handler *Handler) Get(ctx *gin.Context) {
	userID, ok := authctx.UserID(ctx.Request.Context())
	if !ok || userID <= 0 {
		common.WriteError(ctx, http.StatusBadRequest, common.CodeBadRequest)
		return
	}
	bal, err := handler.usecase.GetBalance(ctx, userID)
	if err != nil {
		status, code := common.MapError(err)
		common.WriteError(ctx, status, code)
		return
	}
	ctx.JSON(http.StatusOK, model.Response{
		Current:   bal.Current,
		Withdrawn: bal.Withdrawn,
	})
}
