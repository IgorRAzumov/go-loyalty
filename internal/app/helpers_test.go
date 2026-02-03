package app

import (
	"loyalty/internal/config"
	"testing"
	"time"
)

// Тесты для helper функций, которые можно протестировать без побочных эффектов

func TestLoadDependencies(t *testing.T) {
	cfg := config.Config{
		JWTSecret:          "test-secret",
		JWTTTL:             time.Hour,
		DBQueryTimeout:     3 * time.Second,
		AuthRateLimitRPS:   100,
		AuthRateLimitBurst: 10,
	}

	// Mock DB (nil допустимо для теста конструкторов)
	deps, _ := loadDependencies(cfg, nil)

	if deps.AuthUsecase == nil {
		t.Error("loadDependencies() AuthUsecase is nil")
	}
	if deps.OrdersUsecase == nil {
		t.Error("loadDependencies() OrdersUsecase is nil")
	}
	if deps.BalanceUsecase == nil {
		t.Error("loadDependencies() BalanceUsecase is nil")
	}
	if deps.WithdrawalsUsecase == nil {
		t.Error("loadDependencies() WithdrawalsUsecase is nil")
	}
	if deps.TokenService == nil {
		t.Error("loadDependencies() TokenService is nil")
	}
}

func TestCreateAccrualClient(t *testing.T) {
	tests := []struct {
		name    string
		address string
		wantNil bool
	}{
		{
			name:    "with address",
			address: "http://localhost:8080",
			wantNil: false,
		},
		{
			name:    "empty address (mock)",
			address: "",
			wantNil: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.Config{AccrualSystemAddress: tt.address}
			client := createAccrualClient(cfg)
			if (client == nil) != tt.wantNil {
				t.Errorf("createAccrualClient() nil = %v, want %v", client == nil, tt.wantNil)
			}
		})
	}
}

// initLogger и loadConfig не тестируются напрямую,
// т.к. они вызывают os.Exit(2) при ошибках
