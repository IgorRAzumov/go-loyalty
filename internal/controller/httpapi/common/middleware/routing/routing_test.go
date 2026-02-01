package routing

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestParsePaths_TrimsAndSkipsEmpty(t *testing.T) {
	paths := ParsePaths([]string{"", "  ", "/api/user/orders", " /api/user/withdrawals "})
	if len(paths) != 2 {
		t.Fatalf("want %d paths, got %d", 2, len(paths))
	}
	if paths[0] != "/api/user/orders" {
		t.Fatalf("unexpected first path: %q", paths[0])
	}
	if paths[1] != "/api/user/withdrawals" {
		t.Fatalf("unexpected second path: %q", paths[1])
	}
}

func TestAllowed_UsesFullPathWhenMatched(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/api/user/orders", func(ctx *gin.Context) {
		ctx.Status(http.StatusOK)
	})

	paths := ParsePaths([]string{"/api/user/orders"})
	req := httptest.NewRequest(http.MethodGet, "/api/user/orders", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	called := false
	router2 := gin.New()
	router2.Use(func(ctx *gin.Context) {
		if !Allowed(ctx, paths) {
			t.Fatalf("expected route to be allowed")
		}
		called = true
	})
	router2.GET("/api/user/orders", func(ctx *gin.Context) { ctx.Status(http.StatusOK) })
	w2 := httptest.NewRecorder()
	router2.ServeHTTP(w2, req)
	if !called {
		t.Fatalf("expected middleware to be called")
	}
}

func TestAllowed_FallbacksToRequestPathWhenNoFullPath(t *testing.T) {
	gin.SetMode(gin.TestMode)
	paths := ParsePaths([]string{"/api/user/orders"})

	req := httptest.NewRequest(http.MethodGet, "/api/user/orders", nil)
	ctx := &gin.Context{Request: req}
	if !Allowed(ctx, paths) {
		t.Fatalf("expected route to be allowed by request path")
	}
}
