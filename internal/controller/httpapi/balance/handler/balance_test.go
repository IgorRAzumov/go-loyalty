package handler

import (
	"errors"
	"loyalty/internal/controller/httpapi/auth/authctx"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"go.uber.org/mock/gomock"

	balancemodel "loyalty/internal/domain/balance/model"
	"loyalty/internal/mocks"
)

func TestHandler_Get_200(t *testing.T) {
	gin.SetMode(gin.TestMode)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	uc := mocks.NewMockBalanceUsecase(ctrl)
	uc.EXPECT().
		GetBalance(gomock.Any(), int64(1)).
		Return(balancemodel.Balance{
			Current:   decimal.RequireFromString("10.5"),
			Withdrawn: decimal.RequireFromString("2"),
		}, nil)

	h := NewHandler(uc)

	r := gin.New()
	r.GET("/api/user/balance", h.Get)

	req := httptest.NewRequest(http.MethodGet, "/api/user/balance", nil)
	req = req.WithContext(authctx.WithUserID(req.Context(), 1))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("want %d, got %d", http.StatusOK, w.Code)
	}
	if got := w.Body.String(); got != "{\"current\":\"10.5\",\"withdrawn\":\"2\"}" {
		t.Fatalf("unexpected body: %s", got)
	}
}

func TestHandler_Get_500OnUnexpected(t *testing.T) {
	gin.SetMode(gin.TestMode)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	uc := mocks.NewMockBalanceUsecase(ctrl)
	uc.EXPECT().
		GetBalance(gomock.Any(), int64(1)).
		Return(balancemodel.Balance{}, errors.New("boom"))

	h := NewHandler(uc)

	r := gin.New()
	r.GET("/api/user/balance", h.Get)

	req := httptest.NewRequest(http.MethodGet, "/api/user/balance", nil)
	req = req.WithContext(authctx.WithUserID(req.Context(), 1))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("want %d, got %d", http.StatusInternalServerError, w.Code)
	}
}
