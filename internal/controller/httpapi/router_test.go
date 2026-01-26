package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	tokensvc "loyalty/internal/adapter/token/jwt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	networkmodel "loyalty/internal/controller/httpapi/auth/model"
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

func TestRegisterRoutes_Health(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	RegisterRoutes(r, AuthDeps{
		AuthUsecase: &mockAuthUsecase{
			registerFn: func(context.Context, string, string) (string, error) { return "", nil },
			loginFn:    func(context.Context, string, string) (string, error) { return "", nil },
		},
		TokenService: tokensvc.NewTokenService("secret", time.Hour),
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
	RegisterRoutes(r, AuthDeps{
		AuthUsecase: &mockAuthUsecase{
			registerFn: func(context.Context, string, string) (string, error) { return "", nil },
			loginFn:    func(context.Context, string, string) (string, error) { return "", nil },
		},
		TokenService: tokensvc.NewTokenService("secret", time.Hour),
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
	RegisterRoutes(r, AuthDeps{
		AuthUsecase: &mockAuthUsecase{
			registerFn: func(context.Context, string, string) (string, error) { return "", nil },
			loginFn:    func(context.Context, string, string) (string, error) { return "", nil },
		},
		TokenService: svc,
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
	RegisterRoutes(r, AuthDeps{
		AuthUsecase: &mockAuthUsecase{
			registerFn: func(context.Context, string, string) (string, error) {
				called = true
				return "tok", nil
			},
			loginFn: func(context.Context, string, string) (string, error) { return "", nil },
		},
		TokenService: tokensvc.NewTokenService("secret", time.Hour),
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
