package mock

import (
	"context"
	"testing"

	"loyalty/internal/domain/accrual/model"
)

func TestClient_GetOrderAccrual(t *testing.T) {
	tests := []struct {
		name        string
		client      *Client
		orderNumber string
		wantNil     bool
		checkRandom bool
	}{
		{
			name:        "predefined response",
			client:      NewClient(),
			orderNumber: "12345678903",
			wantNil:     false,
			checkRandom: false,
		},
		{
			name:        "random response",
			client:      NewClientWithDefaults(),
			orderNumber: "unknown",
			wantNil:     false,
			checkRandom: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := tt.client.GetOrderAccrual(context.Background(), tt.orderNumber)
			if err != nil {
				t.Errorf("GetOrderAccrual() error = %v", err)
				return
			}
			if (resp == nil) != tt.wantNil {
				t.Errorf("GetOrderAccrual() nil = %v, want %v", resp == nil, tt.wantNil)
				return
			}

			if tt.checkRandom && resp != nil {
				// Проверяем, что начисление в диапазоне 0-100
				if resp.Accrual != nil {
					accrualFloat, _ := resp.Accrual.Float64()
					if accrualFloat < 0 || accrualFloat > 100 {
						t.Errorf("GetOrderAccrual() accrual = %v, want 0-100", accrualFloat)
					}
				}
				// Проверяем, что статус валидный
				validStatuses := map[model.AccrualStatus]bool{
					model.StatusRegistered: true,
					model.StatusProcessing: true,
					model.StatusProcessed:  true,
					model.StatusInvalid:    true,
				}
				if !validStatuses[resp.Status] {
					t.Errorf("GetOrderAccrual() status = %v, want valid status", resp.Status)
				}
			}
		})
	}
}
