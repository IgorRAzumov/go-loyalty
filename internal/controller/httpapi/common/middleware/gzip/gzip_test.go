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

func TestGzipWriter_Status(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)

	middleware := NewMiddleware(0, "/test")

	router := gin.New()
	router.Use(middleware)
	router.GET("/test", func(c *gin.Context) {
		// Проверяем что Status() работает
		status := c.Writer.Status()
		if status != http.StatusOK && status != 0 {
			t.Errorf("unexpected status: %d", status)
		}
		c.String(200, "ok")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	router.ServeHTTP(ctx.Writer, req)
}

func TestGzipWriter_Size(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)

	middleware := NewMiddleware(0, "/test")

	router := gin.New()
	router.Use(middleware)
	router.GET("/test", func(c *gin.Context) {
		// Проверяем что Size() работает
		size := c.Writer.Size()
		_ = size // используем
		c.String(200, "ok")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	router.ServeHTTP(ctx.Writer, req)
}

func TestGzipWriter_Flush(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()

	middleware := NewMiddleware(0, "/test")

	router := gin.New()
	router.Use(middleware)
	router.GET("/test", func(c *gin.Context) {
		c.Writer.WriteString("test")
		c.Writer.(gin.ResponseWriter).Flush()
		c.Writer.WriteString("flush")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	router.ServeHTTP(w, req)
}

func TestClientAcceptsGzip_Nil(t *testing.T) {
	if clientAcceptsGzip(nil) {
		t.Error("expected false for nil request")
	}
}

func TestClientAcceptsGzip_NoHeader(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	if clientAcceptsGzip(req) {
		t.Error("expected false without Accept-Encoding")
	}
}

func TestClientAcceptsGzip_Deflate(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Accept-Encoding", "deflate")
	if clientAcceptsGzip(req) {
		t.Error("expected false for deflate")
	}
}

func TestClientAcceptsGzip_Mixed(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	if !clientAcceptsGzip(req) {
		t.Error("expected true for mixed with gzip")
	}
}

func TestClientAcceptsGzip_Uppercase(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Accept-Encoding", "GZIP")
	if !clientAcceptsGzip(req) {
		t.Error("expected true for uppercase GZIP")
	}
}

func TestNewMiddleware_NoAcceptEncoding(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(NewMiddleware(100))

	r.GET("/test", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "no compression")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	// Не указываем Accept-Encoding
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Header().Get("Content-Encoding") != "" {
		t.Errorf("expected no compression without Accept-Encoding")
	}
}

func TestNewMiddleware_AlreadyCompressed(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(NewMiddleware(100))

	r.GET("/test", func(ctx *gin.Context) {
		ctx.Header("Content-Encoding", "br")
		ctx.String(http.StatusOK, "already compressed")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Header().Get("Content-Encoding") != "br" {
		t.Errorf("expected Content-Encoding=br")
	}
}

func TestNewMiddleware_LargeResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(NewMiddleware(100))

	r.GET("/test", func(ctx *gin.Context) {
		// Большой ответ >100 байт
		ctx.String(http.StatusOK, strings.Repeat("x", 200))
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Header().Get("Content-Encoding") != "gzip" {
		t.Errorf("expected gzip compression for large response")
	}

	// Проверяем, что данные действительно сжаты
	gr, err := gzip.NewReader(w.Body)
	if err != nil {
		t.Fatalf("failed to create gzip reader: %v", err)
	}
	defer gr.Close()

	decompressed, err := io.ReadAll(gr)
	if err != nil {
		t.Fatalf("failed to decompress: %v", err)
	}

	if string(decompressed) != strings.Repeat("x", 200) {
		t.Errorf("decompressed content mismatch")
	}
}

func TestNewMiddleware_ConditionalRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(NewMiddleware(100, "/api/user/orders"))

	r.GET("/api/user/orders", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, strings.Repeat("order ", 50))
	})

	r.GET("/other", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, strings.Repeat("other ", 50))
	})

	// Проверяем /api/user/orders - должен сжиматься
	req1 := httptest.NewRequest(http.MethodGet, "/api/user/orders", nil)
	req1.Header.Set("Accept-Encoding", "gzip")
	w1 := httptest.NewRecorder()
	r.ServeHTTP(w1, req1)

	if w1.Header().Get("Content-Encoding") != "gzip" {
		t.Errorf("expected gzip for /api/user/orders")
	}

	// Проверяем /other - не должен сжиматься (не в списке)
	req2 := httptest.NewRequest(http.MethodGet, "/other", nil)
	req2.Header.Set("Accept-Encoding", "gzip")
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)

	if w2.Header().Get("Content-Encoding") == "gzip" {
		t.Errorf("expected no gzip for /other")
	}
}
