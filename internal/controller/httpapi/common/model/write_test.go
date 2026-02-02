package model

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestWriteError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		status     int
		code       string
		wantStatus int
		wantBody   string
	}{
		{
			name:       "with code",
			status:     http.StatusBadRequest,
			code:       CodeBadRequest,
			wantStatus: http.StatusBadRequest,
			wantBody:   `{"error":"bad_request"}`,
		},
		{
			name:       "empty code",
			status:     http.StatusOK,
			code:       "",
			wantStatus: http.StatusOK,
			wantBody:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			WriteError(c, tt.status, tt.code)

			if w.Code != tt.wantStatus {
				t.Errorf("Status = %v, want %v", w.Code, tt.wantStatus)
			}

			if tt.wantBody != "" {
				got := w.Body.String()
				// Gin добавляет newline к JSON
				if got != tt.wantBody && got != tt.wantBody+"\n" {
					t.Errorf("Body = %v, want %v", got, tt.wantBody)
				}
			}
		})
	}
}
