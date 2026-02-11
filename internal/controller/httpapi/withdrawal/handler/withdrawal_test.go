package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"loyalty/internal/controller/httpapi/auth/authctx"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"go.uber.org/mock/gomock"

	common "loyalty/internal/controller/httpapi/common/model"
	ordersmodel "loyalty/internal/domain/order/model"
	withdrawalsmodel "loyalty/internal/domain/withdrawal/model"
	"loyalty/internal/mocks"
)

func TestHandler_Withdraw_402OnInsufficientFunds(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	uc := mocks.NewMockWithdrawalsUsecase(ctrl)
	uc.EXPECT().
		Withdraw(gomock.Any(), int64(1), "2377225624", gomock.Any()).
		Return(withdrawalsmodel.ErrInsufficientFunds)
	uc.EXPECT().ListWithdrawals(gomock.Any(), gomock.Any()).MaxTimes(0)

	h := NewHandler(uc)
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
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	uc := mocks.NewMockWithdrawalsUsecase(ctrl)
	uc.EXPECT().
		Withdraw(gomock.Any(), int64(1), "abc", gomock.Any()).
		Return(ordersmodel.ErrInvalidOrderNumber)
	uc.EXPECT().ListWithdrawals(gomock.Any(), gomock.Any()).MaxTimes(0)

	h := NewHandler(uc)
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
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	uc := mocks.NewMockWithdrawalsUsecase(ctrl)
	uc.EXPECT().
		Withdraw(gomock.Any(), int64(1), "2377225624", gomock.Any()).
		Return(withdrawalsmodel.ErrInvalidWithdrawSum)
	uc.EXPECT().ListWithdrawals(gomock.Any(), gomock.Any()).MaxTimes(0)

	h := NewHandler(uc)
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
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	uc := mocks.NewMockWithdrawalsUsecase(ctrl)
	uc.EXPECT().
		Withdraw(gomock.Any(), int64(1), "2377225624", gomock.Any()).
		Return(nil)
	uc.EXPECT().ListWithdrawals(gomock.Any(), gomock.Any()).MaxTimes(0)

	h := NewHandler(uc)
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
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	uc := mocks.NewMockWithdrawalsUsecase(ctrl)
	uc.EXPECT().Withdraw(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).MaxTimes(0)
	uc.EXPECT().ListWithdrawals(gomock.Any(), gomock.Any()).MaxTimes(0)

	h := NewHandler(uc)
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
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	uc := mocks.NewMockWithdrawalsUsecase(ctrl)
	uc.EXPECT().Withdraw(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).MaxTimes(0)
	uc.EXPECT().ListWithdrawals(gomock.Any(), int64(1)).Return(nil, nil)

	h := NewHandler(uc)
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
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	now := time.Date(2026, 1, 28, 12, 0, 0, 0, time.UTC)
	uc := mocks.NewMockWithdrawalsUsecase(ctrl)
	uc.EXPECT().Withdraw(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).MaxTimes(0)
	uc.EXPECT().ListWithdrawals(gomock.Any(), int64(1)).Return([]withdrawalsmodel.Withdrawal{
		{OrderNumber: "2377225624", Sum: decimal.RequireFromString("10.5"), ProcessedAt: now},
	}, nil)

	h := NewHandler(uc)
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
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	uc := mocks.NewMockWithdrawalsUsecase(ctrl)
	uc.EXPECT().Withdraw(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).MaxTimes(0)
	uc.EXPECT().ListWithdrawals(gomock.Any(), int64(1)).Return(nil, errors.New("boom"))

	h := NewHandler(uc)
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
