package handler

import (
	"errors"
	"loyalty/internal/controller/httpapi/auth/authctx"
	"loyalty/internal/controller/httpapi/withdrawal/model"
	"net/http"

	common "loyalty/internal/controller/httpapi/common/model"
	ordersmodel "loyalty/internal/domain/order/model"
	withdrawalsmodel "loyalty/internal/domain/withdrawal/model"
	withdrawalsusecase "loyalty/internal/domain/withdrawal/usecase"

	"github.com/gin-gonic/gin"
)

// Handler — HTTP-хендлеры сценариев списаний пользователя.
type Handler struct {
	usecase withdrawalsusecase.WithdrawalsUsecase
}

// NewHandler создаёт хендлеры списаний пользователя.
func NewHandler(usecase withdrawalsusecase.WithdrawalsUsecase) *Handler {
	return &Handler{usecase: usecase}
}

// Withdraw обрабатывает запрос на списание баллов.
func (handler *Handler) Withdraw(ctx *gin.Context) {
	var req model.WithdrawRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		common.WriteError(ctx, http.StatusBadRequest, common.CodeBadRequest)
		return
	}
	userID, _ := authctx.UserID(ctx.Request.Context())

	if err := handler.usecase.Withdraw(ctx, userID, req.Order, req.Sum); err != nil {
		switch {
		case errors.Is(err, withdrawalsmodel.ErrInsufficientFunds):
			common.WriteError(ctx, http.StatusPaymentRequired, common.CodeInsufficientFunds)
		case errors.Is(err, ordersmodel.ErrInvalidOrderNumber):
			common.WriteError(ctx, http.StatusUnprocessableEntity, common.CodeInvalidOrderNumber)
		case errors.Is(err, withdrawalsmodel.ErrInvalidWithdrawSum):
			common.WriteError(ctx, http.StatusBadRequest, common.CodeBadRequest)
		default:
			status, code := common.MapError(err)
			common.WriteError(ctx, status, code)
		}
		return
	}
	ctx.Status(http.StatusOK)
}

// List возвращает список списаний пользователя.
// по-хорошему нужна пагинация - по ТЗ не было
func (handler *Handler) List(ctx *gin.Context) {
	userID, _ := authctx.UserID(ctx.Request.Context())
	items, err := handler.usecase.ListWithdrawals(ctx, userID)
	if err != nil {
		status, code := common.MapError(err)
		common.WriteError(ctx, status, code)
		return
	}
	if len(items) == 0 {
		ctx.Status(http.StatusNoContent)
		return
	}

	result := make([]model.WithdrawalResponseItem, 0, len(items))
	for _, w := range items {
		result = append(result, model.WithdrawalResponseItem{
			Order:       w.OrderNumber,
			Sum:         w.Sum,
			ProcessedAt: common.RFC3339Time{Time: w.ProcessedAt},
		})
	}
	ctx.JSON(http.StatusOK, result)
}
