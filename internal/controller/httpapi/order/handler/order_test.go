package handler

import (
	"bytes"
	"context"
	"errors"
	"loyalty/internal/controller/httpapi/auth/authctx"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"

	ordersmodel "loyalty/internal/domain/order/model"
	ordersusecase "loyalty/internal/domain/order/usecase"
)

type mockOrdersUsecase struct {
	uploadFn func(ctx context.Context, userID int64, number string) error
	listFn   func(ctx context.Context, userID int64) ([]ordersmodel.Order, error)
}

func (m *mockOrdersUsecase) UploadOrder(ctx context.Context, userID int64, number string) error {
	return m.uploadFn(ctx, userID, number)
}
func (m *mockOrdersUsecase) LoadOrders(ctx context.Context, userID int64) ([]ordersmodel.Order, error) {
	return m.listFn(ctx, userID)
}

var _ ordersusecase.OrdersUsecase = (*mockOrdersUsecase)(nil)

func TestHandler_UploadOrder_202OnNew(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := NewHandler(&mockOrdersUsecase{
		uploadFn: func(context.Context, int64, string) error { return nil },
		listFn:   func(context.Context, int64) ([]ordersmodel.Order, error) { panic("not used") },
	})

	r := gin.New()
	r.POST("/api/user/orders", h.UploadOrder)

	req := httptest.NewRequest(http.MethodPost, "/api/user/orders", bytes.NewBufferString("79927398713"))
	req.Header.Set("Content-Type", "text/plain")
	req = req.WithContext(authctx.WithUserID(req.Context(), 1))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusAccepted {
		t.Fatalf("want %d, got %d", http.StatusAccepted, w.Code)
	}
}

func TestHandler_UploadOrder_200OnAlreadyUploaded(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := NewHandler(&mockOrdersUsecase{
		uploadFn: func(context.Context, int64, string) error { return ordersmodel.ErrOrderAlreadyUploaded },
		listFn:   func(context.Context, int64) ([]ordersmodel.Order, error) { panic("not used") },
	})

	r := gin.New()
	r.POST("/api/user/orders", h.UploadOrder)

	req := httptest.NewRequest(http.MethodPost, "/api/user/orders", bytes.NewBufferString("79927398713"))
	req.Header.Set("Content-Type", "text/plain")
	req = req.WithContext(authctx.WithUserID(req.Context(), 1))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("want %d, got %d", http.StatusOK, w.Code)
	}
}

func TestHandler_UploadOrder_409OnAlreadyUploadedByAnother(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := NewHandler(&mockOrdersUsecase{
		uploadFn: func(context.Context, int64, string) error { return ordersmodel.ErrOrderAlreadyUploadedByAnother },
		listFn:   func(context.Context, int64) ([]ordersmodel.Order, error) { panic("not used") },
	})

	r := gin.New()
	r.POST("/api/user/orders", h.UploadOrder)

	req := httptest.NewRequest(http.MethodPost, "/api/user/orders", bytes.NewBufferString("79927398713"))
	req.Header.Set("Content-Type", "text/plain")
	req = req.WithContext(authctx.WithUserID(req.Context(), 1))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Fatalf("want %d, got %d", http.StatusConflict, w.Code)
	}
}

func TestHandler_UploadOrder_422OnInvalidOrderNumber(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := NewHandler(&mockOrdersUsecase{
		uploadFn: func(context.Context, int64, string) error { return ordersmodel.ErrInvalidOrderNumber },
		listFn:   func(context.Context, int64) ([]ordersmodel.Order, error) { panic("not used") },
	})

	r := gin.New()
	r.POST("/api/user/orders", h.UploadOrder)

	req := httptest.NewRequest(http.MethodPost, "/api/user/orders", bytes.NewBufferString("abc"))
	req.Header.Set("Content-Type", "text/plain")
	req = req.WithContext(authctx.WithUserID(req.Context(), 1))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("want %d, got %d", http.StatusUnprocessableEntity, w.Code)
	}
}

func TestHandler_UploadOrder_400OnEmptyBody(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := NewHandler(&mockOrdersUsecase{
		uploadFn: func(context.Context, int64, string) error { panic("not used") },
		listFn:   func(context.Context, int64) ([]ordersmodel.Order, error) { panic("not used") },
	})

	r := gin.New()
	r.POST("/api/user/orders", h.UploadOrder)

	req := httptest.NewRequest(http.MethodPost, "/api/user/orders", bytes.NewBufferString("   "))
	req.Header.Set("Content-Type", "text/plain")
	req = req.WithContext(authctx.WithUserID(req.Context(), 1))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("want %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestHandler_ListOrders_204OnEmpty(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := NewHandler(&mockOrdersUsecase{
		uploadFn: func(context.Context, int64, string) error { panic("not used") },
		listFn:   func(context.Context, int64) ([]ordersmodel.Order, error) { return nil, nil },
	})

	r := gin.New()
	r.GET("/api/user/orders", h.ListOrders)

	req := httptest.NewRequest(http.MethodGet, "/api/user/orders", nil)
	req = req.WithContext(authctx.WithUserID(req.Context(), 1))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("want %d, got %d", http.StatusNoContent, w.Code)
	}
}

func TestHandler_ListOrders_200WithItems(t *testing.T) {
	gin.SetMode(gin.TestMode)

	now := time.Date(2026, 1, 28, 12, 0, 0, 0, time.UTC)
	accrual := decimal.RequireFromString("10.5")

	h := NewHandler(&mockOrdersUsecase{
		uploadFn: func(context.Context, int64, string) error { panic("not used") },
		listFn: func(context.Context, int64) ([]ordersmodel.Order, error) {
			return []ordersmodel.Order{
				{Number: "79927398713", Status: ordersmodel.StatusNew, Accrual: &accrual, UploadedAt: now},
			}, nil
		},
	})

	r := gin.New()
	r.GET("/api/user/orders", h.ListOrders)

	req := httptest.NewRequest(http.MethodGet, "/api/user/orders", nil)
	req = req.WithContext(authctx.WithUserID(req.Context(), 1))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("want %d, got %d", http.StatusOK, w.Code)
	}
	if !bytes.Contains(w.Body.Bytes(), []byte(`"number":"79927398713"`)) {
		t.Fatalf("unexpected body: %s", w.Body.String())
	}
	if !bytes.Contains(w.Body.Bytes(), []byte(`"status":"NEW"`)) {
		t.Fatalf("unexpected body: %s", w.Body.String())
	}
	if !bytes.Contains(w.Body.Bytes(), []byte(`"accrual":"10.5"`)) {
		t.Fatalf("unexpected body: %s", w.Body.String())
	}
	if !bytes.Contains(w.Body.Bytes(), []byte(`"uploaded_at":"`)) {
		t.Fatalf("unexpected body: %s", w.Body.String())
	}
}

func TestHandler_ListOrders_500OnUnexpectedError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := NewHandler(&mockOrdersUsecase{
		uploadFn: func(context.Context, int64, string) error { panic("not used") },
		listFn:   func(context.Context, int64) ([]ordersmodel.Order, error) { return nil, errors.New("boom") },
	})

	r := gin.New()
	r.GET("/api/user/orders", h.ListOrders)

	req := httptest.NewRequest(http.MethodGet, "/api/user/orders", nil)
	req = req.WithContext(authctx.WithUserID(req.Context(), 1))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("want %d, got %d", http.StatusInternalServerError, w.Code)
	}
}
