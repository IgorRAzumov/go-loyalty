package handler

import (
	"context"
	"errors"
	"loyalty/internal/controller/httpapi/auth/authctx"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"

	balancemodel "loyalty/internal/domain/balance/model"
	balusecase "loyalty/internal/domain/balance/usecase"
)

type mockBalanceUsecase struct {
	getFn func(ctx context.Context, userID int64) (balancemodel.Balance, error)
}

func (m *mockBalanceUsecase) GetBalance(ctx context.Context, userID int64) (balancemodel.Balance, error) {
	return m.getFn(ctx, userID)
}

var _ balusecase.BalanceUsecase = (*mockBalanceUsecase)(nil)

func TestHandler_Get_200(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := NewHandler(&mockBalanceUsecase{
		getFn: func(context.Context, int64) (balancemodel.Balance, error) {
			return balancemodel.Balance{
				Current:   decimal.RequireFromString("10.5"),
				Withdrawn: decimal.RequireFromString("2"),
			}, nil
		},
	})

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

	h := NewHandler(&mockBalanceUsecase{
		getFn: func(context.Context, int64) (balancemodel.Balance, error) {
			return balancemodel.Balance{}, errors.New("boom")
		},
	})

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
