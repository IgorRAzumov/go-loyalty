package model

import (
	"errors"
	"net/http"
	"testing"

	authmodel "loyalty/internal/domain/auth/model"
	ordersmodel "loyalty/internal/domain/order/model"
	withdrawalsmodel "loyalty/internal/domain/withdrawal/model"
)

func TestMapError(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantStatus int
		wantCode   string
	}{
		{
			name:       "nil error",
			err:        nil,
			wantStatus: http.StatusOK,
			wantCode:   "",
		},
		{
			name:       "invalid input",
			err:        authmodel.ErrInvalidInput,
			wantStatus: http.StatusBadRequest,
			wantCode:   CodeInvalidInput,
		},
		{
			name:       "password too short",
			err:        authmodel.ErrPasswordTooShort,
			wantStatus: http.StatusBadRequest,
			wantCode:   CodePasswordTooShort,
		},
		{
			name:       "password too long",
			err:        authmodel.ErrPasswordTooLong,
			wantStatus: http.StatusBadRequest,
			wantCode:   CodePasswordTooLong,
		},
		{
			name:       "login taken",
			err:        authmodel.ErrLoginTaken,
			wantStatus: http.StatusConflict,
			wantCode:   CodeLoginTaken,
		},
		{
			name:       "invalid credentials",
			err:        authmodel.ErrInvalidCreds,
			wantStatus: http.StatusUnauthorized,
			wantCode:   CodeInvalidCreds,
		},
		{
			name:       "not found",
			err:        authmodel.ErrNotFound,
			wantStatus: http.StatusUnauthorized,
			wantCode:   CodeInvalidCreds,
		},
		{
			name:       "invalid token",
			err:        authmodel.ErrInvalidToken,
			wantStatus: http.StatusUnauthorized,
			wantCode:   CodeUnauthorized,
		},
		{
			name:       "invalid order number",
			err:        ordersmodel.ErrInvalidOrderNumber,
			wantStatus: http.StatusUnprocessableEntity,
			wantCode:   CodeInvalidOrderNumber,
		},
		{
			name:       "order already uploaded",
			err:        ordersmodel.ErrOrderAlreadyUploaded,
			wantStatus: http.StatusOK,
			wantCode:   CodeOrderAlreadyUploaded,
		},
		{
			name:       "order uploaded by another",
			err:        ordersmodel.ErrOrderAlreadyUploadedByAnother,
			wantStatus: http.StatusConflict,
			wantCode:   CodeOrderAlreadyUploadedByAnother,
		},
		{
			name:       "insufficient funds",
			err:        withdrawalsmodel.ErrInsufficientFunds,
			wantStatus: http.StatusPaymentRequired,
			wantCode:   CodeInsufficientFunds,
		},
		{
			name:       "unknown error",
			err:        errors.New("unknown"),
			wantStatus: http.StatusInternalServerError,
			wantCode:   CodeInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, code := MapError(tt.err)
			if status != tt.wantStatus {
				t.Errorf("MapError() status = %v, want %v", status, tt.wantStatus)
			}
			if code != tt.wantCode {
				t.Errorf("MapError() code = %v, want %v", code, tt.wantCode)
			}
		})
	}
}
