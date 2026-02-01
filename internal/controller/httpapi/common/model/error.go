package model

import (
	"errors"
	"loyalty/internal/domain/auth/model"
	ordersmodel "loyalty/internal/domain/order/model"
	withdrawalsmodel "loyalty/internal/domain/withdrawal/model"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ErrKey — имя поля в JSON-ответе, содержащее код ошибки (вариант 1: {"error":"<code>"}).
const ErrKey = "error"

// ErrorResponse — единый формат ответа об ошибке для API.
type ErrorResponse struct {
	Error string `json:"error"`
}

const (
	// CodeBadRequest — запрос не удалось распарсить/прочитать (например, невалидный JSON).
	CodeBadRequest = "bad_request"
	// CodeInvalidInput — нарушены бизнес-ограничения на входные данные.
	CodeInvalidInput = "invalid_input"
	// CodePasswordTooShort — пароль слишком короткий.
	CodePasswordTooShort = "password_too_short"
	// CodePasswordTooLong — пароль слишком длинный.
	CodePasswordTooLong = "password_too_long"
	// CodeLoginTaken — логин уже занят.
	CodeLoginTaken = "login_taken"
	// CodeInvalidCreds — неверная пара логин/пароль.
	CodeInvalidCreds = "invalid_credentials"
	// CodeUnauthorized — отсутствует/невалиден токен авторизации.
	CodeUnauthorized = "unauthorized"
	// CodeInvalidOrderNumber — неверный формат номера заказа / не проходит алгоритм Луна.
	CodeInvalidOrderNumber = "invalid_order_number"
	// CodeOrderAlreadyUploaded — номер заказа уже был загружен этим пользователем.
	CodeOrderAlreadyUploaded = "order_already_uploaded"
	// CodeOrderAlreadyUploadedByAnother — номер заказа уже был загружен другим пользователем.
	CodeOrderAlreadyUploadedByAnother = "order_already_uploaded_by_another"
	// CodeInsufficientFunds — на счету недостаточно средств.
	CodeInsufficientFunds = "insufficient_funds"
	// CodeInternal — внутренняя ошибка сервера (детали не раскрываются клиенту).
	CodeInternal = "internal"
)

// WriteError записывает ошибку в HTTP-ответ в едином формате.
// Если code пустой, функция отправляет только статус без тела.
func WriteError(ctx *gin.Context, status int, code string) {
	if code == "" {
		ctx.Status(status)
		return
	}
	ctx.JSON(status, ErrorResponse{Error: code})
}

// MapError сопоставляет доменную ошибку с HTTP-статусом и бизнес-кодом для JSON-ответа.
// Возвращаемый code должен быть стабильным для клиентов (Postman/фронт/автотесты).
func MapError(err error) (status int, code string) {
	switch {
	case err == nil:
		return http.StatusOK, ""

	case errors.Is(err, model.ErrInvalidInput):
		return http.StatusBadRequest, CodeInvalidInput
	case errors.Is(err, model.ErrPasswordTooShort):
		return http.StatusBadRequest, CodePasswordTooShort
	case errors.Is(err, model.ErrPasswordTooLong):
		return http.StatusBadRequest, CodePasswordTooLong
	case errors.Is(err, model.ErrLoginTaken):
		return http.StatusConflict, CodeLoginTaken

	case errors.Is(err, model.ErrInvalidCreds):
		return http.StatusUnauthorized, CodeInvalidCreds
	case errors.Is(err, model.ErrNotFound):
		// Do not reveal whether user exists.
		return http.StatusUnauthorized, CodeInvalidCreds
	case errors.Is(err, model.ErrInvalidToken):
		return http.StatusUnauthorized, CodeUnauthorized

	case errors.Is(err, ordersmodel.ErrInvalidOrderNumber):
		return http.StatusUnprocessableEntity, CodeInvalidOrderNumber
	case errors.Is(err, ordersmodel.ErrOrderAlreadyUploaded):
		return http.StatusOK, CodeOrderAlreadyUploaded
	case errors.Is(err, ordersmodel.ErrOrderAlreadyUploadedByAnother):
		return http.StatusConflict, CodeOrderAlreadyUploadedByAnother
	case errors.Is(err, withdrawalsmodel.ErrInsufficientFunds):
		return http.StatusPaymentRequired, CodeInsufficientFunds

	default:
		return http.StatusInternalServerError, CodeInternal
	}
}
