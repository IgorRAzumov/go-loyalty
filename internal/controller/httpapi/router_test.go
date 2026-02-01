package httpapi

import (
	"bytes"
	"compress/gzip"
	"context"
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

	networkmodel "loyalty/internal/controller/httpapi/auth/model"
	balancemodel "loyalty/internal/domain/balance/model"
	ordersmodel "loyalty/internal/domain/order/model"
	withdrawalsmodel "loyalty/internal/domain/withdrawal/model"
)

type mockAuthUsecase struct {
	registerFn func(ctx context.Context, login, password string) (string, error)
	loginFn    func(ctx context.Context, login, password string) (string, error)
}

func (m *mockAuthUsecase) Register(ctx context.Context, login, password string) (string, error) {
	return m.registerFn(ctx, login, password)
}
func (m *mockAuthUsecase) Login(ctx context.Context, login, password string) (string, error) {
	return m.loginFn(ctx, login, password)
}

type mockOrdersUsecase struct{}

func (m *mockOrdersUsecase) UploadOrder(context.Context, int64, string) error { return nil }
func (m *mockOrdersUsecase) LoadOrders(context.Context, int64) ([]ordersmodel.Order, error) {
	return nil, nil
}

type mockOrdersUsecaseWithOrders struct {
	orders []ordersmodel.Order
}

func (m *mockOrdersUsecaseWithOrders) UploadOrder(context.Context, int64, string) error { return nil }
func (m *mockOrdersUsecaseWithOrders) LoadOrders(context.Context, int64) ([]ordersmodel.Order, error) {
	return m.orders, nil
}

type mockBalanceUsecase struct{}

func (m *mockBalanceUsecase) GetBalance(context.Context, int64) (balancemodel.Balance, error) {
	return balancemodel.Balance{Current: decimal.Zero, Withdrawn: decimal.Zero}, nil
}

type mockWithdrawalsUsecase struct{}

func (m *mockWithdrawalsUsecase) Withdraw(context.Context, int64, string, decimal.Decimal) error {
	return nil
}
func (m *mockWithdrawalsUsecase) ListWithdrawals(context.Context, int64) ([]withdrawalsmodel.Withdrawal, error) {
	return nil, nil
}

type mockWithdrawalsUsecaseWithItems struct {
	items []withdrawalsmodel.Withdrawal
}

func (m *mockWithdrawalsUsecaseWithItems) Withdraw(context.Context, int64, string, decimal.Decimal) error {
	return nil
}
func (m *mockWithdrawalsUsecaseWithItems) ListWithdrawals(context.Context, int64) ([]withdrawalsmodel.Withdrawal, error) {
	return m.items, nil
}

func mustIssueToken(t *testing.T) (svc *tokensvc.Service, token string) {
	t.Helper()

	svc = tokensvc.NewTokenService("secret", time.Hour)
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

func TestRegisterRoutes_Health(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	RegisterRoutes(r, Deps{
		AuthUsecase: &mockAuthUsecase{
			registerFn: func(context.Context, string, string) (string, error) { return "", nil },
			loginFn:    func(context.Context, string, string) (string, error) { return "", nil },
		},
		OrdersUsecase:         &mockOrdersUsecase{},
		BalanceUsecase:        &mockBalanceUsecase{},
		WithdrawalsUsecase:    &mockWithdrawalsUsecase{},
		TokenService:          tokensvc.NewTokenService("secret", time.Hour),
		EnableHTTPBodyLogging: false,
		AuthRateLimitRPS:      100,
		AuthRateLimitBurst:    20,
	})

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("want %d, got %d", http.StatusOK, w.Code)
	}
}

func TestRegisterRoutes_UserBalance_UnauthorizedWithoutToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	RegisterRoutes(r, Deps{
		AuthUsecase: &mockAuthUsecase{
			registerFn: func(context.Context, string, string) (string, error) { return "", nil },
			loginFn:    func(context.Context, string, string) (string, error) { return "", nil },
		},
		OrdersUsecase:         &mockOrdersUsecase{},
		BalanceUsecase:        &mockBalanceUsecase{},
		WithdrawalsUsecase:    &mockWithdrawalsUsecase{},
		TokenService:          tokensvc.NewTokenService("secret", time.Hour),
		EnableHTTPBodyLogging: false,
		AuthRateLimitRPS:      100,
		AuthRateLimitBurst:    20,
	})

	req := httptest.NewRequest(http.MethodGet, "/api/user/balance", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("want %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestRegisterRoutes_UserBalance_OKWithToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := tokensvc.NewTokenService("secret", time.Hour)
	tok, err := svc.IssueToken(1, "alice", time.Now())
	if err != nil {
		t.Fatalf("issue token: %v", err)
	}

	r := gin.New()
	RegisterRoutes(r, Deps{
		AuthUsecase: &mockAuthUsecase{
			registerFn: func(context.Context, string, string) (string, error) { return "", nil },
			loginFn:    func(context.Context, string, string) (string, error) { return "", nil },
		},
		OrdersUsecase:         &mockOrdersUsecase{},
		BalanceUsecase:        &mockBalanceUsecase{},
		WithdrawalsUsecase:    &mockWithdrawalsUsecase{},
		TokenService:          svc,
		EnableHTTPBodyLogging: false,
		AuthRateLimitRPS:      100,
		AuthRateLimitBurst:    20,
	})

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
	called := false
	r := gin.New()
	RegisterRoutes(r, Deps{
		AuthUsecase: &mockAuthUsecase{
			registerFn: func(context.Context, string, string) (string, error) {
				called = true
				return "tok", nil
			},
			loginFn: func(context.Context, string, string) (string, error) { return "", nil },
		},
		OrdersUsecase:         &mockOrdersUsecase{},
		BalanceUsecase:        &mockBalanceUsecase{},
		WithdrawalsUsecase:    &mockWithdrawalsUsecase{},
		TokenService:          tokensvc.NewTokenService("secret", time.Hour),
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
	if !called {
		t.Fatalf("expected usecase to be called")
	}
}

func TestRegisterRoutes_UserOrders_UnauthorizedWithoutToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	RegisterRoutes(r, Deps{
		AuthUsecase: &mockAuthUsecase{
			registerFn: func(context.Context, string, string) (string, error) { return "", nil },
			loginFn:    func(context.Context, string, string) (string, error) { return "", nil },
		},
		OrdersUsecase:         &mockOrdersUsecase{},
		BalanceUsecase:        &mockBalanceUsecase{},
		WithdrawalsUsecase:    &mockWithdrawalsUsecase{},
		TokenService:          tokensvc.NewTokenService("secret", time.Hour),
		EnableHTTPBodyLogging: false,
		AuthRateLimitRPS:      100,
		AuthRateLimitBurst:    20,
	})

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
	r := gin.New()
	RegisterRoutes(r, Deps{
		AuthUsecase: &mockAuthUsecase{
			registerFn: func(context.Context, string, string) (string, error) { return "", nil },
			loginFn:    func(context.Context, string, string) (string, error) { return "", nil },
		},
		OrdersUsecase:         &mockOrdersUsecase{},
		BalanceUsecase:        &mockBalanceUsecase{},
		WithdrawalsUsecase:    &mockWithdrawalsUsecase{},
		TokenService:          tokensvc.NewTokenService("secret", time.Hour),
		EnableHTTPBodyLogging: false,
		AuthRateLimitRPS:      100,
		AuthRateLimitBurst:    20,
	})

	req := httptest.NewRequest(http.MethodGet, "/api/user/withdrawals", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("want %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestRegisterRoutes_UserOrders_GzipWhenAcceptedAndLarge(t *testing.T) {
	gin.SetMode(gin.TestMode)
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

	r := gin.New()
	RegisterRoutes(r, Deps{
		AuthUsecase: &mockAuthUsecase{
			registerFn: func(context.Context, string, string) (string, error) { return "", nil },
			loginFn:    func(context.Context, string, string) (string, error) { return "", nil },
		},
		OrdersUsecase:         &mockOrdersUsecaseWithOrders{orders: orders},
		BalanceUsecase:        &mockBalanceUsecase{},
		WithdrawalsUsecase:    &mockWithdrawalsUsecase{},
		TokenService:          svc,
		EnableHTTPBodyLogging: false,
		AuthRateLimitRPS:      100,
		AuthRateLimitBurst:    20,
	})

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
	svc, tok := mustIssueToken(t)

	acc := decimal.NewFromInt(1)
	orders := []ordersmodel.Order{
		{
			Number:     "79927398713",
			UserID:     1,
			Status:     ordersmodel.StatusNew,
			Accrual:    &acc,
			UploadedAt: time.Now(),
		},
	}

	r := gin.New()
	RegisterRoutes(r, Deps{
		AuthUsecase: &mockAuthUsecase{
			registerFn: func(context.Context, string, string) (string, error) { return "", nil },
			loginFn:    func(context.Context, string, string) (string, error) { return "", nil },
		},
		OrdersUsecase:         &mockOrdersUsecaseWithOrders{orders: orders},
		BalanceUsecase:        &mockBalanceUsecase{},
		WithdrawalsUsecase:    &mockWithdrawalsUsecase{},
		TokenService:          svc,
		EnableHTTPBodyLogging: false,
		AuthRateLimitRPS:      100,
		AuthRateLimitBurst:    20,
	})

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

	r := gin.New()
	RegisterRoutes(r, Deps{
		AuthUsecase: &mockAuthUsecase{
			registerFn: func(context.Context, string, string) (string, error) { return "", nil },
			loginFn:    func(context.Context, string, string) (string, error) { return "", nil },
		},
		OrdersUsecase:         &mockOrdersUsecase{},
		BalanceUsecase:        &mockBalanceUsecase{},
		WithdrawalsUsecase:    &mockWithdrawalsUsecaseWithItems{items: items},
		TokenService:          svc,
		EnableHTTPBodyLogging: false,
		AuthRateLimitRPS:      100,
		AuthRateLimitBurst:    20,
	})

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
