package handler

import (
	"bytes"
	"errors"
	"loyalty/internal/controller/httpapi/auth/authctx"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"go.uber.org/mock/gomock"

	ordersmodel "loyalty/internal/domain/order/model"
	"loyalty/internal/mocks"
)

func TestHandler_UploadOrder_202OnNew(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	uc := mocks.NewMockOrdersUsecase(ctrl)
	uc.EXPECT().UploadOrder(gomock.Any(), int64(1), "79927398713").Return(nil)
	uc.EXPECT().LoadOrders(gomock.Any(), gomock.Any()).MaxTimes(0)

	h := NewHandler(uc)
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
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	uc := mocks.NewMockOrdersUsecase(ctrl)
	uc.EXPECT().UploadOrder(gomock.Any(), int64(1), "79927398713").Return(ordersmodel.ErrOrderAlreadyUploaded)
	uc.EXPECT().LoadOrders(gomock.Any(), gomock.Any()).MaxTimes(0)

	h := NewHandler(uc)
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
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	uc := mocks.NewMockOrdersUsecase(ctrl)
	uc.EXPECT().UploadOrder(gomock.Any(), int64(1), "79927398713").Return(ordersmodel.ErrOrderAlreadyUploadedByAnother)
	uc.EXPECT().LoadOrders(gomock.Any(), gomock.Any()).MaxTimes(0)

	h := NewHandler(uc)
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
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	uc := mocks.NewMockOrdersUsecase(ctrl)
	uc.EXPECT().UploadOrder(gomock.Any(), int64(1), "abc").Return(ordersmodel.ErrInvalidOrderNumber)
	uc.EXPECT().LoadOrders(gomock.Any(), gomock.Any()).MaxTimes(0)

	h := NewHandler(uc)
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
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	uc := mocks.NewMockOrdersUsecase(ctrl)
	uc.EXPECT().UploadOrder(gomock.Any(), gomock.Any(), gomock.Any()).MaxTimes(0)
	uc.EXPECT().LoadOrders(gomock.Any(), gomock.Any()).MaxTimes(0)

	h := NewHandler(uc)
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
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	uc := mocks.NewMockOrdersUsecase(ctrl)
	uc.EXPECT().UploadOrder(gomock.Any(), gomock.Any(), gomock.Any()).MaxTimes(0)
	uc.EXPECT().LoadOrders(gomock.Any(), int64(1)).Return(nil, nil)

	h := NewHandler(uc)
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
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	now := time.Date(2026, 1, 28, 12, 0, 0, 0, time.UTC)
	accrual := decimal.RequireFromString("10.5")

	uc := mocks.NewMockOrdersUsecase(ctrl)
	uc.EXPECT().UploadOrder(gomock.Any(), gomock.Any(), gomock.Any()).MaxTimes(0)
	uc.EXPECT().LoadOrders(gomock.Any(), int64(1)).Return([]ordersmodel.Order{
		{Number: "79927398713", Status: ordersmodel.StatusNew, Accrual: &accrual, UploadedAt: now},
	}, nil)

	h := NewHandler(uc)
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
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	uc := mocks.NewMockOrdersUsecase(ctrl)
	uc.EXPECT().UploadOrder(gomock.Any(), gomock.Any(), gomock.Any()).MaxTimes(0)
	uc.EXPECT().LoadOrders(gomock.Any(), int64(1)).Return(nil, errors.New("boom"))

	h := NewHandler(uc)
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
