package withdrawals

import (
	"context"
	"errors"
	"testing"

	ordersmodel "loyalty/internal/domain/order/model"
	withdrawalsmodel "loyalty/internal/domain/withdrawal/model"
	"loyalty/internal/mocks"

	"github.com/shopspring/decimal"
	"go.uber.org/mock/gomock"
)

func TestUsecase_Withdraw(t *testing.T) {
	tests := []struct {
		name          string
		validatorNorm string
		validatorErr  error
		withdrawErr   error
		wantErr       bool
	}{
		{
			name:          "success",
			validatorNorm: "12345678903",
			wantErr:       false,
		},
		{
			name:         "invalid order number",
			validatorErr: ordersmodel.ErrInvalidOrderNumber,
			wantErr:      true,
		},
		{
			name:          "service error",
			validatorNorm: "12345678903",
			withdrawErr:   errors.New("withdraw failed"),
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			svc := mocks.NewMockWithdrawalsService(ctrl)
			validator := mocks.NewMockOrderNumberValidator(ctrl)

			if tt.validatorErr != nil {
				validator.EXPECT().
					ValidateNumber("12345678903").
					Return("", tt.validatorErr)
				svc.EXPECT().Withdraw(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).MaxTimes(0)
			} else {
				validator.EXPECT().
					ValidateNumber("12345678903").
					Return(tt.validatorNorm, nil)
				svc.EXPECT().
					Withdraw(gomock.Any(), int64(1), tt.validatorNorm, decimal.NewFromFloat(100)).
					Return(tt.withdrawErr)
			}

			uc := NewUsecase(svc, validator)
			err := uc.Withdraw(context.Background(), 1, "12345678903", decimal.NewFromFloat(100))
			if (err != nil) != tt.wantErr {
				t.Errorf("Withdraw() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestUsecase_ListWithdrawals(t *testing.T) {
	tests := []struct {
		name        string
		withdrawals []withdrawalsmodel.Withdrawal
		listErr     error
		wantLen     int
		wantErr     bool
	}{
		{
			name: "success with items",
			withdrawals: []withdrawalsmodel.Withdrawal{
				{OrderNumber: "123"},
			},
			wantLen: 1,
			wantErr: false,
		},
		{
			name:    "empty list",
			wantLen: 0,
			wantErr: false,
		},
		{
			name:    "service error",
			listErr: errors.New("list failed"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			svc := mocks.NewMockWithdrawalsService(ctrl)
			svc.EXPECT().
				ListWithdrawals(gomock.Any(), int64(1)).
				Return(tt.withdrawals, tt.listErr)

			validator := mocks.NewMockOrderNumberValidator(ctrl)
			uc := NewUsecase(svc, validator)
			items, err := uc.ListWithdrawals(context.Background(), 1)
			if (err != nil) != tt.wantErr {
				t.Errorf("ListWithdrawals() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && len(items) != tt.wantLen {
				t.Errorf("ListWithdrawals() len = %v, want %v", len(items), tt.wantLen)
			}
		})
	}
}
