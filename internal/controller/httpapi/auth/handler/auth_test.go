package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	networkmodel "loyalty/internal/controller/httpapi/auth/model"
	"loyalty/internal/controller/httpapi/common"
	"loyalty/internal/domain/auth/model"
	"loyalty/internal/domain/auth/usecase"
)

type mockUsecase struct {
	registerFn func(ctx context.Context, login, password string) (string, error)
	loginFn    func(ctx context.Context, login, password string) (string, error)
}

func (m *mockUsecase) Register(ctx context.Context, login, password string) (string, error) {
	return m.registerFn(ctx, login, password)
}
func (m *mockUsecase) Login(ctx context.Context, login, password string) (string, error) {
	return m.loginFn(ctx, login, password)
}

var _ usecase.AuthUsecase = (*mockUsecase)(nil)

func TestHandler_Register_SetsAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)

	uc := &mockUsecase{
		registerFn: func(context.Context, string, string) (string, error) { return "token123", nil },
		loginFn:    func(context.Context, string, string) (string, error) { panic("not used") },
	}
	h := NewAuthHandler(uc)

	r := gin.New()
	r.POST("/api/user/register", h.Register)

	body, _ := json.Marshal(networkmodel.LoginRequest{Login: "alice", Password: "longenough10"})
	req := httptest.NewRequest(http.MethodPost, "/api/user/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("want %d, got %d", http.StatusOK, w.Code)
	}
	if got := w.Header().Get("Authorization"); got == "" {
		t.Fatalf("expected Authorization header")
	}
	cookies := w.Result().Cookies()
	found := false
	for _, c := range cookies {
		if c.Name == "token" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected token cookie")
	}
}

func TestHandler_Login_UnauthorizedOnInvalidCreds(t *testing.T) {
	gin.SetMode(gin.TestMode)

	uc := &mockUsecase{
		registerFn: func(context.Context, string, string) (string, error) { panic("not used") },
		loginFn:    func(context.Context, string, string) (string, error) { return "", model.ErrInvalidCreds },
	}
	h := NewAuthHandler(uc)

	r := gin.New()
	r.POST("/api/user/login", h.Login)

	body, _ := json.Marshal(networkmodel.LoginRequest{Login: "alice", Password: "longenough10"})
	req := httptest.NewRequest(http.MethodPost, "/api/user/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("want %d, got %d", http.StatusUnauthorized, w.Code)
	}

	var resp map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	if got := resp[common.ErrKey]; got != common.CodeInvalidCreds {
		t.Fatalf("want %q, got %v", common.CodeInvalidCreds, got)
	}
}

func TestHandler_Register_ConflictOnLoginTaken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	uc := &mockUsecase{
		registerFn: func(context.Context, string, string) (string, error) { return "", model.ErrLoginTaken },
		loginFn:    func(context.Context, string, string) (string, error) { panic("not used") },
	}
	h := NewAuthHandler(uc)

	r := gin.New()
	r.POST("/api/user/register", h.Register)

	body, _ := json.Marshal(networkmodel.LoginRequest{Login: "alice", Password: "longenough10"})
	req := httptest.NewRequest(http.MethodPost, "/api/user/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Fatalf("want %d, got %d", http.StatusConflict, w.Code)
	}
}

func TestHandler_Register_500OnUnexpected(t *testing.T) {
	gin.SetMode(gin.TestMode)

	uc := &mockUsecase{
		registerFn: func(context.Context, string, string) (string, error) { return "", errors.New("boom") },
		loginFn:    func(context.Context, string, string) (string, error) { panic("not used") },
	}
	h := NewAuthHandler(uc)

	r := gin.New()
	r.POST("/api/user/register", h.Register)

	body, _ := json.Marshal(networkmodel.LoginRequest{Login: "alice", Password: "longenough10"})
	req := httptest.NewRequest(http.MethodPost, "/api/user/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("want %d, got %d", http.StatusInternalServerError, w.Code)
	}
}
