package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"loyalty/internal/controller/httpapi/auth/authctx"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"

	common "loyalty/internal/controller/httpapi/common/model"
	ordersmodel "loyalty/internal/domain/order/model"
	withdrawalsmodel "loyalty/internal/domain/withdrawal/model"
	withdrawalsusecase "loyalty/internal/domain/withdrawal/usecase"
)

type mockWithdrawalsUsecase struct {
	withdrawFn func(ctx context.Context, userID int64, orderNumber string, sum decimal.Decimal) error
	listFn     func(ctx context.Context, userID int64) ([]withdrawalsmodel.Withdrawal, error)
}

func (m *mockWithdrawalsUsecase) Withdraw(ctx context.Context, userID int64, orderNumber string, sum decimal.Decimal) error {
	return m.withdrawFn(ctx, userID, orderNumber, sum)
}
func (m *mockWithdrawalsUsecase) ListWithdrawals(ctx context.Context, userID int64) ([]withdrawalsmodel.Withdrawal, error) {
	return m.listFn(ctx, userID)
}

var _ withdrawalsusecase.WithdrawalsUsecase = (*mockWithdrawalsUsecase)(nil)

func TestHandler_Withdraw_402OnInsufficientFunds(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := NewHandler(&mockWithdrawalsUsecase{
		withdrawFn: func(context.Context, int64, string, decimal.Decimal) error {
			return withdrawalsmodel.ErrInsufficientFunds
		},
		listFn: func(context.Context, int64) ([]withdrawalsmodel.Withdrawal, error) { panic("not used") },
	})

	r := gin.New()
	r.POST("/api/user/balance/withdraw", h.Withdraw)

	body, _ := json.Marshal(map[string]any{"order": "2377225624", "sum": 10})
	req := httptest.NewRequest(http.MethodPost, "/api/user/balance/withdraw", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(authctx.WithUserID(req.Context(), 1))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusPaymentRequired {
		t.Fatalf("want %d, got %d", http.StatusPaymentRequired, w.Code)
	}
	if got := w.Body.String(); got != `{"error":"insufficient_funds"}` {
		t.Fatalf("unexpected body: %s", got)
	}
}

func TestHandler_Withdraw_422OnInvalidOrderNumber(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := NewHandler(&mockWithdrawalsUsecase{
		withdrawFn: func(context.Context, int64, string, decimal.Decimal) error { return ordersmodel.ErrInvalidOrderNumber },
		listFn:     func(context.Context, int64) ([]withdrawalsmodel.Withdrawal, error) { panic("not used") },
	})

	r := gin.New()
	r.POST("/api/user/balance/withdraw", h.Withdraw)

	body, _ := json.Marshal(map[string]any{"order": "abc", "sum": 10})
	req := httptest.NewRequest(http.MethodPost, "/api/user/balance/withdraw", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(authctx.WithUserID(req.Context(), 1))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("want %d, got %d", http.StatusUnprocessableEntity, w.Code)
	}
	if got := w.Body.String(); got != `{"error":"invalid_order_number"}` {
		t.Fatalf("unexpected body: %s", got)
	}
}

func TestHandler_Withdraw_400OnInvalidSum(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := NewHandler(&mockWithdrawalsUsecase{
		withdrawFn: func(context.Context, int64, string, decimal.Decimal) error {
			return withdrawalsmodel.ErrInvalidWithdrawSum
		},
		listFn: func(context.Context, int64) ([]withdrawalsmodel.Withdrawal, error) { panic("not used") },
	})

	r := gin.New()
	r.POST("/api/user/balance/withdraw", h.Withdraw)

	body, _ := json.Marshal(map[string]any{"order": "2377225624", "sum": 0})
	req := httptest.NewRequest(http.MethodPost, "/api/user/balance/withdraw", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(authctx.WithUserID(req.Context(), 1))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("want %d, got %d", http.StatusBadRequest, w.Code)
	}
	if got := w.Body.String(); got != `{"error":"bad_request"}` {
		t.Fatalf("unexpected body: %s", got)
	}
}

func TestHandler_Withdraw_200OnOK(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := NewHandler(&mockWithdrawalsUsecase{
		withdrawFn: func(context.Context, int64, string, decimal.Decimal) error { return nil },
		listFn:     func(context.Context, int64) ([]withdrawalsmodel.Withdrawal, error) { panic("not used") },
	})

	r := gin.New()
	r.POST("/api/user/balance/withdraw", h.Withdraw)

	body, _ := json.Marshal(map[string]any{"order": "2377225624", "sum": 10})
	req := httptest.NewRequest(http.MethodPost, "/api/user/balance/withdraw", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(authctx.WithUserID(req.Context(), 1))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("want %d, got %d", http.StatusOK, w.Code)
	}
}

func TestHandler_Withdraw_400OnBadJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := NewHandler(&mockWithdrawalsUsecase{
		withdrawFn: func(context.Context, int64, string, decimal.Decimal) error { panic("not used") },
		listFn:     func(context.Context, int64) ([]withdrawalsmodel.Withdrawal, error) { panic("not used") },
	})

	r := gin.New()
	r.POST("/api/user/balance/withdraw", h.Withdraw)

	req := httptest.NewRequest(http.MethodPost, "/api/user/balance/withdraw", bytes.NewBufferString("{"))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(authctx.WithUserID(req.Context(), 1))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("want %d, got %d", http.StatusBadRequest, w.Code)
	}
	if got := w.Body.String(); got != `{"error":"bad_request"}` {
		t.Fatalf("unexpected body: %s", got)
	}
}

func TestHandler_List_204OnEmpty(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := NewHandler(&mockWithdrawalsUsecase{
		withdrawFn: func(context.Context, int64, string, decimal.Decimal) error { panic("not used") },
		listFn:     func(context.Context, int64) ([]withdrawalsmodel.Withdrawal, error) { return nil, nil },
	})

	r := gin.New()
	r.GET("/api/user/withdrawals", h.List)

	req := httptest.NewRequest(http.MethodGet, "/api/user/withdrawals", nil)
	req = req.WithContext(authctx.WithUserID(req.Context(), 1))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("want %d, got %d", http.StatusNoContent, w.Code)
	}
}

func TestHandler_List_200WithItems(t *testing.T) {
	gin.SetMode(gin.TestMode)

	now := time.Date(2026, 1, 28, 12, 0, 0, 0, time.UTC)
	h := NewHandler(&mockWithdrawalsUsecase{
		withdrawFn: func(context.Context, int64, string, decimal.Decimal) error { panic("not used") },
		listFn: func(context.Context, int64) ([]withdrawalsmodel.Withdrawal, error) {
			return []withdrawalsmodel.Withdrawal{
				{OrderNumber: "2377225624", Sum: decimal.RequireFromString("10.5"), ProcessedAt: now},
			}, nil
		},
	})

	r := gin.New()
	r.GET("/api/user/withdrawals", h.List)

	req := httptest.NewRequest(http.MethodGet, "/api/user/withdrawals", nil)
	req = req.WithContext(authctx.WithUserID(req.Context(), 1))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("want %d, got %d", http.StatusOK, w.Code)
	}
	if !bytes.Contains(w.Body.Bytes(), []byte(`"order":"2377225624"`)) {
		t.Fatalf("unexpected body: %s", w.Body.String())
	}
	if !bytes.Contains(w.Body.Bytes(), []byte(`"sum":"10.5"`)) {
		t.Fatalf("unexpected body: %s", w.Body.String())
	}
	if !bytes.Contains(w.Body.Bytes(), []byte(`"processed_at":"`)) {
		t.Fatalf("unexpected body: %s", w.Body.String())
	}
}

func TestHandler_List_500OnUnexpectedError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := NewHandler(&mockWithdrawalsUsecase{
		withdrawFn: func(context.Context, int64, string, decimal.Decimal) error { panic("not used") },
		listFn:     func(context.Context, int64) ([]withdrawalsmodel.Withdrawal, error) { return nil, errors.New("boom") },
	})

	r := gin.New()
	r.GET("/api/user/withdrawals", h.List)

	req := httptest.NewRequest(http.MethodGet, "/api/user/withdrawals", nil)
	req = req.WithContext(authctx.WithUserID(req.Context(), 1))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("want %d, got %d", http.StatusInternalServerError, w.Code)
	}

	var resp map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	if resp[common.ErrKey] != common.CodeInternal {
		t.Fatalf("want %q, got %v", common.CodeInternal, resp[common.ErrKey])
	}
}
