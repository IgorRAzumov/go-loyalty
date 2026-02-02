package handler

import (
	"errors"
	"io"
	"loyalty/internal/controller/httpapi/auth/authctx"
	"loyalty/internal/controller/httpapi/order/model"
	"net/http"
	"strings"

	common "loyalty/internal/controller/httpapi/common/model"
	ordersmodel "loyalty/internal/domain/order/model"
	ordersusecase "loyalty/internal/domain/order/usecase"

	"github.com/gin-gonic/gin"
)

// Handler — HTTP-хендлеры сценариев заказов пользователя.
type Handler struct {
	usecase ordersusecase.OrdersUsecase
}

// NewHandler создаёт хендлеры заказов пользователя.
func NewHandler(usecase ordersusecase.OrdersUsecase) *Handler { return &Handler{usecase: usecase} }

// UploadOrder обрабатывает загрузку номера заказа пользователя.
func (handler *Handler) UploadOrder(ctx *gin.Context) {
	// ТЗ: строго Content-Type: text/plain
	contentType := ctx.GetHeader("Content-Type")
	if contentType != "text/plain" && contentType != "text/plain; charset=utf-8" {
		common.WriteError(ctx, http.StatusBadRequest, common.CodeBadRequest)
		return
	}

	body, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		common.WriteError(ctx, http.StatusBadRequest, common.CodeBadRequest)
		return
	}
	number := strings.TrimSpace(string(body))
	if number == "" {
		common.WriteError(ctx, http.StatusBadRequest, common.CodeBadRequest)
		return
	}

	userID, _ := authctx.UserID(ctx.Request.Context())
	err = handler.usecase.UploadOrder(ctx, userID, number)

	switch {
	case err == nil:
		ctx.Status(http.StatusAccepted)
	case errors.Is(err, ordersmodel.ErrOrderAlreadyUploaded):
		ctx.Status(http.StatusOK)
	case errors.Is(err, ordersmodel.ErrOrderAlreadyUploadedByAnother):
		ctx.Status(http.StatusConflict)
	case errors.Is(err, ordersmodel.ErrInvalidOrderNumber):
		ctx.Status(http.StatusUnprocessableEntity)
	default:
		status, code := common.MapError(err)
		common.WriteError(ctx, status, code)
	}
}

// ListOrders возвращает список загруженных заказов пользователя.
func (handler *Handler) ListOrders(ctx *gin.Context) {
	userID, _ := authctx.UserID(ctx.Request.Context())
	orders, err := handler.usecase.LoadOrders(ctx, userID)
	if err != nil {
		status, code := common.MapError(err)
		common.WriteError(ctx, status, code)
		return
	}
	if len(orders) == 0 {
		ctx.Status(http.StatusNoContent)
		return
	}

	resp := make([]model.OrderResponseItem, 0, len(orders))
	for _, o := range orders {
		resp = append(resp, model.OrderResponseItem{
			Number:     o.Number,
			Status:     string(o.Status),
			Accrual:    o.Accrual,
			UploadedAt: common.RFC3339Time{Time: o.UploadedAt},
		})
	}
	ctx.JSON(http.StatusOK, resp)
}
