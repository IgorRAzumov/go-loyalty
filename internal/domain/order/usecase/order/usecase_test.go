package order

import (
	"context"
	"errors"
	"testing"

	accrualmodel "loyalty/internal/domain/accrual/model"
	ordersmodel "loyalty/internal/domain/order/model"

	"github.com/shopspring/decimal"
)

type mockOrdersService struct {
	uploadErr error
	orders    []ordersmodel.Order
	loadErr   error
}

func (m *mockOrdersService) UploadOrder(ctx context.Context, userID int64, number string) error {
	return m.uploadErr
}

func (m *mockOrdersService) LoadOrders(ctx context.Context, userID int64) ([]ordersmodel.Order, error) {
	if m.loadErr != nil {
		return nil, m.loadErr
	}
	return m.orders, nil
}

func (m *mockOrdersService) UpdateFromAccrual(ctx context.Context, orderNumber string, accrualStatus accrualmodel.AccrualStatus, accrual *decimal.Decimal) error {
	return nil
}

func TestUsecase_UploadOrder(t *testing.T) {
	tests := []struct {
		name    string
		svc     *mockOrdersService
		wantErr bool
	}{
		{
			name:    "success",
			svc:     &mockOrdersService{},
			wantErr: false,
		},
		{
			name: "service error",
			svc: &mockOrdersService{
				uploadErr: errors.New("upload failed"),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := NewUsecase(tt.svc)
			err := uc.UploadOrder(context.Background(), 1, "12345678903")
			if (err != nil) != tt.wantErr {
				t.Errorf("UploadOrder() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestUsecase_LoadOrders(t *testing.T) {
	tests := []struct {
		name    string
		svc     *mockOrdersService
		wantLen int
		wantErr bool
	}{
		{
			name: "success with orders",
			svc: &mockOrdersService{
				orders: []ordersmodel.Order{
					{Number: "123"},
				},
			},
			wantLen: 1,
			wantErr: false,
		},
		{
			name:    "empty orders",
			svc:     &mockOrdersService{orders: []ordersmodel.Order{}},
			wantLen: 0,
			wantErr: false,
		},
		{
			name: "service error",
			svc: &mockOrdersService{
				loadErr: errors.New("load failed"),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := NewUsecase(tt.svc)
			orders, err := uc.LoadOrders(context.Background(), 1)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadOrders() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && len(orders) != tt.wantLen {
				t.Errorf("LoadOrders() len = %v, want %v", len(orders), tt.wantLen)
			}
		})
	}
}
