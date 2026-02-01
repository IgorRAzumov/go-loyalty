package middleware

import (
	tokensvc "loyalty/internal/adapter/token/jwt"
	"loyalty/internal/controller/httpapi/auth/authctx"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestAuthMiddleware_RejectsWithoutToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(NewAuthMiddleware(tokensvc.NewTokenService("secret", time.Hour)))
	r.GET("/x", func(c *gin.Context) { c.Status(http.StatusOK) })

	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("want %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestAuthMiddleware_AllowsWithBearerToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc := tokensvc.NewTokenService("secret", time.Hour)
	tok, err := svc.IssueToken(42, "alice", time.Now())
	if err != nil {
		t.Fatalf("issue token: %v", err)
	}

	r := gin.New()
	r.Use(NewAuthMiddleware(svc))
	r.GET("/x", func(c *gin.Context) {
		id, ok := authctx.UserID(c.Request.Context())
		if !ok || id != 42 {
			t.Fatalf("expected user id=42 in request context")
		}
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("want %d, got %d", http.StatusOK, w.Code)
	}
}

func TestAuthMiddleware_AllowsWithCookieToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc := tokensvc.NewTokenService("secret", time.Hour)
	tok, err := svc.IssueToken(42, "alice", time.Now())
	if err != nil {
		t.Fatalf("issue token: %v", err)
	}

	r := gin.New()
	r.Use(NewAuthMiddleware(svc))
	r.GET("/x", func(c *gin.Context) { c.Status(http.StatusOK) })

	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	req.AddCookie(&http.Cookie{Name: "token", Value: tok})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("want %d, got %d", http.StatusOK, w.Code)
	}
}
