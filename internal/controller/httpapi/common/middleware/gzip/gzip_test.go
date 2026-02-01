package gzip

import (
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestGzipMiddleware_FastPathWithSmallContentLength(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(NewMiddleware(1024)) // minSize = 1024

	r.GET("/small", func(ctx *gin.Context) {
		body := []byte("small response")
		ctx.Header("Content-Length", "14") // 14 < 1024
		ctx.Data(http.StatusOK, "text/plain", body)
	})

	req := httptest.NewRequest(http.MethodGet, "/small", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Header().Get("Content-Encoding") != "" {
		t.Errorf("expected no Content-Encoding for small response with Content-Length, got %q", w.Header().Get("Content-Encoding"))
	}

	if w.Body.String() != "small response" {
		t.Errorf("expected body %q, got %q", "small response", w.Body.String())
	}
}

func TestGzipMiddleware_NoFastPathWithoutContentLength(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(NewMiddleware(1024))

	r.GET("/no-length", func(ctx *gin.Context) {
		body := []byte("small")
		ctx.Data(http.StatusOK, "text/plain", body)
	})

	req := httptest.NewRequest(http.MethodGet, "/no-length", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Body.String() != "small" {
		t.Errorf("expected body %q, got %q", "small", w.Body.String())
	}
}

func TestGzipMiddleware_NoFastPathWithLargeContentLength(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(NewMiddleware(10)) // minSize = 10

	r.GET("/large", func(ctx *gin.Context) {
		body := make([]byte, 2000)
		for i := range body {
			body[i] = 'x'
		}
		ctx.Header("Content-Length", "2000") // 2000 > 10
		ctx.Data(http.StatusOK, "text/plain", body)
	})

	req := httptest.NewRequest(http.MethodGet, "/large", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Header().Get("Content-Encoding") != "gzip" {
		t.Errorf("expected Content-Encoding=gzip for large response, got %q", w.Header().Get("Content-Encoding"))
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

	if len(decompressed) != 2000 {
		t.Errorf("expected decompressed size 2000, got %d", len(decompressed))
	}
}

func TestGzipMiddleware_FastPathPreservesHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(NewMiddleware(1024))

	r.GET("/headers", func(ctx *gin.Context) {
		ctx.Header("X-Custom-Header", "custom-value")
		ctx.Header("Content-Type", "application/json")
		ctx.Header("Content-Length", "50")
		ctx.Data(http.StatusOK, "application/json", []byte(`{"status":"ok"}`))
	})

	req := httptest.NewRequest(http.MethodGet, "/headers", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Header().Get("X-Custom-Header") != "custom-value" {
		t.Errorf("expected X-Custom-Header=custom-value, got %q", w.Header().Get("X-Custom-Header"))
	}

	if w.Header().Get("Content-Type") != "application/json" {
		t.Errorf("expected Content-Type=application/json, got %q", w.Header().Get("Content-Type"))
	}

	if w.Header().Get("Content-Encoding") != "" {
		t.Errorf("expected no Content-Encoding, got %q", w.Header().Get("Content-Encoding"))
	}
}
