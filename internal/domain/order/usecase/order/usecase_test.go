package order

import (
	"context"
	"errors"
	"testing"

	ordersmodel "loyalty/internal/domain/order/model"
	"loyalty/internal/mocks"

	"go.uber.org/mock/gomock"
)

func TestUsecase_UploadOrder(t *testing.T) {
	tests := []struct {
		name      string
		uploadErr error
		wantErr   bool
	}{
		{
			name:    "success",
			wantErr: false,
		},
		{
			name:      "service error",
			uploadErr: errors.New("upload failed"),
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			svc := mocks.NewMockOrdersService(ctrl)
			svc.EXPECT().
				UploadOrder(gomock.Any(), int64(1), "12345678903").
				Return(tt.uploadErr)

			uc := NewUsecase(svc)
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
		orders  []ordersmodel.Order
		loadErr error
		wantLen int
		wantErr bool
	}{
		{
			name: "success with orders",
			orders: []ordersmodel.Order{
				{Number: "123"},
			},
			wantLen: 1,
			wantErr: false,
		},
		{
			name:    "empty orders",
			orders:  []ordersmodel.Order{},
			wantLen: 0,
			wantErr: false,
		},
		{
			name:    "service error",
			loadErr: errors.New("load failed"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			svc := mocks.NewMockOrdersService(ctrl)
			svc.EXPECT().
				LoadOrders(gomock.Any(), int64(1)).
				Return(tt.orders, tt.loadErr)

			uc := NewUsecase(svc)
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
