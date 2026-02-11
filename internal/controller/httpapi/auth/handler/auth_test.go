package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"go.uber.org/mock/gomock"

	networkmodel "loyalty/internal/controller/httpapi/auth/model"
	common "loyalty/internal/controller/httpapi/common/model"
	"loyalty/internal/domain/auth/model"
	"loyalty/internal/mocks"
)

func TestHandler_Register_SetsAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	uc := mocks.NewMockAuthUsecase(ctrl)
	uc.EXPECT().
		Register(gomock.Any(), "alice", "longenough10").
		Return("token123", nil)
	uc.EXPECT().Login(gomock.Any(), gomock.Any(), gomock.Any()).MaxTimes(0)

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

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	uc := mocks.NewMockAuthUsecase(ctrl)
	uc.EXPECT().Register(gomock.Any(), gomock.Any(), gomock.Any()).MaxTimes(0)
	uc.EXPECT().
		Login(gomock.Any(), "alice", "longenough10").
		Return("", model.ErrInvalidCreds)

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

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	uc := mocks.NewMockAuthUsecase(ctrl)
	uc.EXPECT().
		Register(gomock.Any(), "alice", "longenough10").
		Return("", model.ErrLoginTaken)
	uc.EXPECT().Login(gomock.Any(), gomock.Any(), gomock.Any()).MaxTimes(0)

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

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	uc := mocks.NewMockAuthUsecase(ctrl)
	uc.EXPECT().
		Register(gomock.Any(), "alice", "longenough10").
		Return("", errors.New("boom"))
	uc.EXPECT().Login(gomock.Any(), gomock.Any(), gomock.Any()).MaxTimes(0)

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
