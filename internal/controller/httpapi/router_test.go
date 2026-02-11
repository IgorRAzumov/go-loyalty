package httpapi

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	tokensvc "loyalty/internal/adapter/token/jwt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"go.uber.org/mock/gomock"

	networkmodel "loyalty/internal/controller/httpapi/auth/model"
	balancemodel "loyalty/internal/domain/balance/model"
	ordersmodel "loyalty/internal/domain/order/model"
	withdrawalsmodel "loyalty/internal/domain/withdrawal/model"
	"loyalty/internal/logger"
	"loyalty/internal/mocks"
)

func mustIssueToken(t *testing.T) (*tokensvc.Service, string) {
	t.Helper()
	svc := tokensvc.NewTokenService("secret", time.Hour)
	tok, err := svc.IssueToken(1, "alice", time.Now())
	if err != nil {
		t.Fatalf("issue token: %v", err)
	}
	return svc, tok
}

func mustGunzip(t *testing.T, b []byte) []byte {
	t.Helper()
	r, err := gzip.NewReader(bytes.NewReader(b))
	if err != nil {
		t.Fatalf("gzip reader: %v", err)
	}
	defer func() { _ = r.Close() }()
	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("gzip read: %v", err)
	}
	return out
}

func defaultDeps(t *testing.T, ctrl *gomock.Controller) Deps {
	auth := mocks.NewMockAuthUsecase(ctrl)
	auth.EXPECT().Register(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return("", nil)
	auth.EXPECT().Login(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return("", nil)

	orders := mocks.NewMockOrdersUsecase(ctrl)
	orders.EXPECT().UploadOrder(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(nil)
	orders.EXPECT().LoadOrders(gomock.Any(), gomock.Any()).AnyTimes().Return(nil, nil)

	balance := mocks.NewMockBalanceUsecase(ctrl)
	balance.EXPECT().GetBalance(gomock.Any(), gomock.Any()).AnyTimes().Return(balancemodel.Balance{Current: decimal.Zero, Withdrawn: decimal.Zero}, nil)

	withdrawals := mocks.NewMockWithdrawalsUsecase(ctrl)
	withdrawals.EXPECT().Withdraw(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(nil)
	withdrawals.EXPECT().ListWithdrawals(gomock.Any(), gomock.Any()).AnyTimes().Return(nil, nil)

	return Deps{
		AuthUsecase:           auth,
		OrdersUsecase:         orders,
		BalanceUsecase:        balance,
		WithdrawalsUsecase:    withdrawals,
		TokenService:          tokensvc.NewTokenService("secret", time.Hour),
		Logger:                logger.NewNopLogger(),
		EnableHTTPBodyLogging: false,
		AuthRateLimitRPS:      100,
		AuthRateLimitBurst:    20,
	}
}

func TestRegisterRoutes_Health(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	r := gin.New()
	RegisterRoutes(r, defaultDeps(t, ctrl))

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("want %d, got %d", http.StatusOK, w.Code)
	}
}

func TestRegisterRoutes_UserBalance_UnauthorizedWithoutToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	r := gin.New()
	RegisterRoutes(r, defaultDeps(t, ctrl))

	req := httptest.NewRequest(http.MethodGet, "/api/user/balance", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("want %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestRegisterRoutes_UserBalance_OKWithToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc, tok := mustIssueToken(t)
	deps := defaultDeps(t, ctrl)
	deps.TokenService = svc

	r := gin.New()
	RegisterRoutes(r, deps)

	req := httptest.NewRequest(http.MethodGet, "/api/user/balance", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("want %d, got %d", http.StatusOK, w.Code)
	}
	var resp map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["current"] == nil || resp["withdrawn"] == nil {
		t.Fatalf("expected current/withdrawn fields, got: %s", w.Body.String())
	}
}

func TestRegisterRoutes_Register_UsesUsecase(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	auth := mocks.NewMockAuthUsecase(ctrl)
	auth.EXPECT().Register(gomock.Any(), "alice", "longenough10").Return("tok", nil)
	auth.EXPECT().Login(gomock.Any(), gomock.Any(), gomock.Any()).MaxTimes(0)

	orders := mocks.NewMockOrdersUsecase(ctrl)
	orders.EXPECT().UploadOrder(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(nil)
	orders.EXPECT().LoadOrders(gomock.Any(), gomock.Any()).AnyTimes().Return(nil, nil)

	balance := mocks.NewMockBalanceUsecase(ctrl)
	balance.EXPECT().GetBalance(gomock.Any(), gomock.Any()).AnyTimes().Return(balancemodel.Balance{}, nil)

	withdrawals := mocks.NewMockWithdrawalsUsecase(ctrl)
	withdrawals.EXPECT().Withdraw(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(nil)
	withdrawals.EXPECT().ListWithdrawals(gomock.Any(), gomock.Any()).AnyTimes().Return(nil, nil)

	r := gin.New()
	RegisterRoutes(r, Deps{
		AuthUsecase:           auth,
		OrdersUsecase:         orders,
		BalanceUsecase:        balance,
		WithdrawalsUsecase:    withdrawals,
		TokenService:          tokensvc.NewTokenService("secret", time.Hour),
		Logger:                logger.NewNopLogger(),
		EnableHTTPBodyLogging: false,
		AuthRateLimitRPS:      100,
		AuthRateLimitBurst:    20,
	})

	body, _ := json.Marshal(networkmodel.LoginRequest{Login: "alice", Password: "longenough10"})
	req := httptest.NewRequest(http.MethodPost, "/api/user/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("want %d, got %d", http.StatusOK, w.Code)
	}
}

func TestRegisterRoutes_UserOrders_UnauthorizedWithoutToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	r := gin.New()
	RegisterRoutes(r, defaultDeps(t, ctrl))

	req := httptest.NewRequest(http.MethodPost, "/api/user/orders", bytes.NewBufferString("79927398713"))
	req.Header.Set("Content-Type", "text/plain")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("want %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestRegisterRoutes_UserWithdrawals_UnauthorizedWithoutToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	r := gin.New()
	RegisterRoutes(r, defaultDeps(t, ctrl))

	req := httptest.NewRequest(http.MethodGet, "/api/user/withdrawals", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("want %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestRegisterRoutes_UserOrders_GzipWhenAcceptedAndLarge(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc, tok := mustIssueToken(t)

	orders := make([]ordersmodel.Order, 0, 200)
	now := time.Now()
	for i := 0; i < 200; i++ {
		acc := decimal.NewFromInt(int64(i))
		orders = append(orders, ordersmodel.Order{
			Number:     fmt.Sprintf("%013d%013d", i, i),
			UserID:     1,
			Status:     ordersmodel.StatusProcessed,
			Accrual:    &acc,
			UploadedAt: now.Add(-time.Duration(i) * time.Second),
		})
	}

	ordersUC := mocks.NewMockOrdersUsecase(ctrl)
	ordersUC.EXPECT().UploadOrder(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(nil)
	ordersUC.EXPECT().LoadOrders(gomock.Any(), int64(1)).Return(orders, nil)

	deps := defaultDeps(t, ctrl)
	deps.OrdersUsecase = ordersUC
	deps.TokenService = svc

	r := gin.New()
	RegisterRoutes(r, deps)

	req := httptest.NewRequest(http.MethodGet, "/api/user/orders", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	req.Header.Set("Accept-Encoding", "gzip")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("want %d, got %d", http.StatusOK, w.Code)
	}
	if got := w.Header().Get("Content-Encoding"); got != "gzip" {
		t.Fatalf("expected gzip response, got Content-Encoding=%q", got)
	}
	if vary := w.Header().Values("Vary"); len(vary) == 0 {
		t.Fatalf("expected Vary header to be set")
	}

	raw := mustGunzip(t, w.Body.Bytes())
	var resp []map[string]any
	if err := json.Unmarshal(raw, &resp); err != nil {
		t.Fatalf("unmarshal: %v; body=%s", err, string(raw))
	}
	if len(resp) != 200 {
		t.Fatalf("want %d items, got %d", 200, len(resp))
	}
}

func TestRegisterRoutes_UserOrders_NoGzipWhenSmall(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc, tok := mustIssueToken(t)

	acc := decimal.NewFromInt(1)
	orders := []ordersmodel.Order{
		{Number: "79927398713", UserID: 1, Status: ordersmodel.StatusNew, Accrual: &acc, UploadedAt: time.Now()},
	}

	ordersUC := mocks.NewMockOrdersUsecase(ctrl)
	ordersUC.EXPECT().UploadOrder(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(nil)
	ordersUC.EXPECT().LoadOrders(gomock.Any(), int64(1)).Return(orders, nil)

	deps := defaultDeps(t, ctrl)
	deps.OrdersUsecase = ordersUC
	deps.TokenService = svc

	r := gin.New()
	RegisterRoutes(r, deps)

	req := httptest.NewRequest(http.MethodGet, "/api/user/orders", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	req.Header.Set("Accept-Encoding", "gzip")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("want %d, got %d", http.StatusOK, w.Code)
	}
	if got := w.Header().Get("Content-Encoding"); got != "" {
		t.Fatalf("expected uncompressed response, got Content-Encoding=%q", got)
	}
	var resp []map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v; body=%s", err, w.Body.String())
	}
	if len(resp) != 1 {
		t.Fatalf("want %d items, got %d", 1, len(resp))
	}
}

func TestRegisterRoutes_UserWithdrawals_GzipWhenAcceptedAndLarge(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc, tok := mustIssueToken(t)

	items := make([]withdrawalsmodel.Withdrawal, 0, 200)
	now := time.Now()
	for i := 0; i < 200; i++ {
		items = append(items, withdrawalsmodel.Withdrawal{
			UserID:      1,
			OrderNumber: fmt.Sprintf("%013d%013d", i, i),
			Sum:         decimal.NewFromInt(int64(i + 1)),
			ProcessedAt: now.Add(-time.Duration(i) * time.Second),
		})
	}

	withdrawalsUC := mocks.NewMockWithdrawalsUsecase(ctrl)
	withdrawalsUC.EXPECT().Withdraw(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(nil)
	withdrawalsUC.EXPECT().ListWithdrawals(gomock.Any(), int64(1)).Return(items, nil)

	deps := defaultDeps(t, ctrl)
	deps.WithdrawalsUsecase = withdrawalsUC
	deps.TokenService = svc

	r := gin.New()
	RegisterRoutes(r, deps)

	req := httptest.NewRequest(http.MethodGet, "/api/user/withdrawals", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	req.Header.Set("Accept-Encoding", "gzip")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("want %d, got %d", http.StatusOK, w.Code)
	}
	if got := w.Header().Get("Content-Encoding"); got != "gzip" {
		t.Fatalf("expected gzip response, got Content-Encoding=%q", got)
	}

	raw := mustGunzip(t, w.Body.Bytes())
	var resp []map[string]any
	if err := json.Unmarshal(raw, &resp); err != nil {
		t.Fatalf("unmarshal: %v; body=%s", err, string(raw))
	}
	if len(resp) != 200 {
		t.Fatalf("want %d items, got %d", 200, len(resp))
	}
}
