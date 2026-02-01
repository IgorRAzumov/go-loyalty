package gzip

import (
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestGzipMiddleware_StreamingMode(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(NewMiddleware(100, "/test"))

	router.GET("/test", func(ctx *gin.Context) {
		ctx.Writer.WriteString(strings.Repeat("a", 60))
		ctx.Writer.WriteString(strings.Repeat("b", 50))
		ctx.Writer.WriteString(strings.Repeat("c", 40))
		ctx.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", w.Code)
	}

	if w.Header().Get("Content-Encoding") != "gzip" {
		t.Fatalf("expected gzip encoding, got %q", w.Header().Get("Content-Encoding"))
	}

	gzReader, err := gzip.NewReader(w.Body)
	if err != nil {
		t.Fatalf("failed to create gzip reader: %v", err)
	}
	defer gzReader.Close()

	decompressed, err := io.ReadAll(gzReader)
	if err != nil {
		t.Fatalf("failed to decompress: %v", err)
	}

	expected := strings.Repeat("a", 60) + strings.Repeat("b", 50) + strings.Repeat("c", 40)
	if string(decompressed) != expected {
		t.Fatalf("want %q, got %q", expected, string(decompressed))
	}
}

func TestGzipMiddleware_SmallResponseUncompressed(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(NewMiddleware(100, "/test"))

	router.GET("/test", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "small") // 5 bytes < 100
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", w.Code)
	}

	if w.Header().Get("Content-Encoding") == "gzip" {
		t.Fatalf("small response should not be gzipped")
	}

	if w.Body.String() != "small" {
		t.Fatalf("want %q, got %q", "small", w.Body.String())
	}
}
