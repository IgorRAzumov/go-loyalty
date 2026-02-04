package http

import (
	"context"
	"errors"
	"loyalty/internal/domain/accrual/model"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/sony/gobreaker"
)

func TestClient_GetOrderAccrual(t *testing.T) {
	tests := []struct {
		name           string
		serverResponse string
		serverStatus   int
		wantErr        bool
		wantNil        bool
		wantErrType    error
	}{
		{
			name:           "200 OK",
			serverResponse: `{"order":"123","status":"PROCESSED","accrual":500}`,
			serverStatus:   http.StatusOK,
			wantErr:        false,
			wantNil:        false,
		},
		{
			name:         "204 No Content",
			serverStatus: http.StatusNoContent,
			wantErr:      false,
			wantNil:      true,
		},
		{
			name:         "429 Too Many Requests",
			serverStatus: http.StatusTooManyRequests,
			wantErr:      true,
			wantNil:      true,
			wantErrType:  model.ErrTooManyRequests,
		},
		{
			name:           "500 Internal Server Error",
			serverResponse: "internal error",
			serverStatus:   http.StatusInternalServerError,
			wantErr:        true,
			wantNil:        true,
		},
		{
			name:           "invalid JSON",
			serverResponse: `{invalid`,
			serverStatus:   http.StatusOK,
			wantErr:        true,
			wantNil:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.serverStatus)
				if tt.serverResponse != "" {
					_, _ = w.Write([]byte(tt.serverResponse))
				}
			}))
			defer server.Close()

			c := NewClient(server.URL, 5*time.Second)
			resp, err := c.GetOrderAccrual(context.Background(), "123")

			if (err != nil) != tt.wantErr {
				t.Errorf("GetOrderAccrual() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErrType != nil && err != tt.wantErrType {
				t.Errorf("GetOrderAccrual() error type = %T, want %T", err, tt.wantErrType)
			}

			if (resp == nil) != tt.wantNil {
				t.Errorf("GetOrderAccrual() nil = %v, want %v", resp == nil, tt.wantNil)
			}
		})
	}
}

func TestClient_CircuitBreakerOpens(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("boom"))
	}))
	defer server.Close()

	c := NewClient(server.URL, 5*time.Second)
	// Делаем тест детерминированным: открываем breaker после 1 ошибки.
	c.breaker = gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name: "accrual-test",
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.ConsecutiveFailures >= 1
		},
		Timeout: 1 * time.Minute,
	})

	_, err := c.GetOrderAccrual(context.Background(), "123")
	if err == nil {
		t.Fatalf("expected error on first request")
	}

	_, err = c.GetOrderAccrual(context.Background(), "123")
	if err == nil {
		t.Fatalf("expected error when breaker is open")
	}
	if !errors.Is(err, model.ErrTemporarilyUnavailable) {
		t.Fatalf("expected ErrTemporarilyUnavailable, got %v", err)
	}
}

func TestGetRetryAfter(t *testing.T) {
	tests := []struct {
		name   string
		header string
		want   time.Duration
	}{
		{
			name:   "valid seconds",
			header: "60",
			want:   60 * time.Second,
		},
		{
			name:   "empty",
			header: "",
			want:   0,
		},
		{
			name:   "invalid",
			header: "abc",
			want:   0,
		},
		{
			name:   "zero",
			header: "0",
			want:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := &http.Response{
				Header: http.Header{},
			}
			if tt.header != "" {
				resp.Header.Set("Retry-After", tt.header)
			}
		})
	}
}
