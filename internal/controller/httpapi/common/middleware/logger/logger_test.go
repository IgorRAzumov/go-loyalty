package logger

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestNewMiddleware_WithoutBodyLogging(t *testing.T) {
	gin.SetMode(gin.TestMode)

	middleware := NewMiddleware(false)

	router := gin.New()
	router.Use(middleware)
	router.GET("/test", func(c *gin.Context) {
		c.String(200, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("expected status 200, got %d", w.Code)
	}
	if w.Body.String() != "ok" {
		t.Errorf("expected body 'ok', got %s", w.Body.String())
	}
}

func TestNewMiddleware_WithBodyLogging(t *testing.T) {
	gin.SetMode(gin.TestMode)

	middleware := NewMiddleware(true)

	router := gin.New()
	router.Use(middleware)
	router.POST("/test", func(c *gin.Context) {
		body, _ := io.ReadAll(c.Request.Body)
		c.String(200, "received: "+string(body))
	})

	req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewBufferString("hello"))
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestNewMiddleware_ErrorStatus(t *testing.T) {
	gin.SetMode(gin.TestMode)

	middleware := NewMiddleware(true)

	router := gin.New()
	router.Use(middleware)
	router.GET("/error", func(c *gin.Context) {
		c.String(400, "bad request")
	})

	req := httptest.NewRequest(http.MethodGet, "/error", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != 400 {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestNewMiddleware_RedactRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	middleware := NewMiddleware(true, "/api/user/register", "/api/user/login")

	router := gin.New()
	router.Use(middleware)
	router.POST("/api/user/register", func(c *gin.Context) {
		c.String(400, "error")
	})

	req := httptest.NewRequest(http.MethodPost, "/api/user/register", bytes.NewBufferString("password"))
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != 400 {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestWithCommonHTTPFields(t *testing.T) {
	gin.SetMode(gin.TestMode)

	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	ctx.Request = httptest.NewRequest(http.MethodGet, "/test?query=1", nil)

	event := httpLog.Info()
	result := withCommonHTTPFields(event, ctx, 200)

	if result == nil {
		t.Error("expected non-nil event")
	}
}

func TestWithCommonHTTPFields_NilContext(t *testing.T) {
	event := httpLog.Info()
	result := withCommonHTTPFields(event, nil, 200)

	if result == nil {
		t.Error("expected non-nil event")
	}
}
