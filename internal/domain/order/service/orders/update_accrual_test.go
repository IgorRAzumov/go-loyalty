package orders

import (
	"context"
	"errors"
	"testing"

	accrualmodel "loyalty/internal/domain/accrual/model"
	ordersmodel "loyalty/internal/domain/order/model"
	"loyalty/internal/mocks"

	"github.com/shopspring/decimal"
	"go.uber.org/mock/gomock"
)

func TestService_UpdateFromAccrual(t *testing.T) {
	tests := []struct {
		name          string
		accrualStatus accrualmodel.AccrualStatus
		accrual       *decimal.Decimal
		repoErr       error
		wantErr       bool
	}{
		{
			name:          "PROCESSED with accrual",
			accrualStatus: accrualmodel.StatusProcessed,
			accrual:       decimalPtr(100),
			wantErr:       false,
		},
		{
			name:          "INVALID",
			accrualStatus: accrualmodel.StatusInvalid,
			accrual:       nil,
			wantErr:       false,
		},
		{
			name:          "PROCESSING",
			accrualStatus: accrualmodel.StatusProcessing,
			accrual:       nil,
			wantErr:       false,
		},
		{
			name:          "REGISTERED",
			accrualStatus: accrualmodel.StatusRegistered,
			accrual:       nil,
			wantErr:       false,
		},
		{
			name:          "repo error",
			accrualStatus: accrualmodel.StatusProcessed,
			accrual:       decimalPtr(100),
			repoErr:       errors.New("db error"),
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			expectedStatus := mapAccrualStatusToOrderStatus(tt.accrualStatus)
			repo := mocks.NewMockOrdersRepository(ctrl)
			repo.EXPECT().
				UpdateFromAccrual(gomock.Any(), "123", expectedStatus, tt.accrual).
				Return(tt.repoErr)

			num := mocks.NewMockOrderNumberValidator(ctrl)
			svc := NewService(repo, num)

			err := svc.UpdateFromAccrual(context.Background(), "123", tt.accrualStatus, tt.accrual)
			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateFromAccrual() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMapAccrualStatusToOrderStatus(t *testing.T) {
	tests := []struct {
		accrualStatus accrualmodel.AccrualStatus
		want          ordersmodel.Status
	}{
		{accrualmodel.StatusRegistered, ordersmodel.StatusProcessing},
		{accrualmodel.StatusProcessing, ordersmodel.StatusProcessing},
		{accrualmodel.StatusInvalid, ordersmodel.StatusInvalid},
		{accrualmodel.StatusProcessed, ordersmodel.StatusProcessed},
		{"UNKNOWN", ordersmodel.StatusProcessing},
	}

	for _, tt := range tests {
		t.Run(string(tt.accrualStatus), func(t *testing.T) {
			got := mapAccrualStatusToOrderStatus(tt.accrualStatus)
			if got != tt.want {
				t.Errorf("mapAccrualStatusToOrderStatus() = %v, want %v", got, tt.want)
			}
		})
	}
}

func decimalPtr(v float64) *decimal.Decimal {
	d := decimal.NewFromFloat(v)
	return &d
}
