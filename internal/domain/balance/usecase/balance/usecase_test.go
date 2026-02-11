package balance

import (
	"context"
	"errors"
	"testing"

	balancemodel "loyalty/internal/domain/balance/model"
	"loyalty/internal/mocks"

	"github.com/shopspring/decimal"
	"go.uber.org/mock/gomock"
)

func TestUsecase_GetBalance(t *testing.T) {
	tests := []struct {
		name    string
		userID  int64
		balance balancemodel.Balance
		err     error
		wantErr bool
	}{
		{
			name:   "success",
			userID: 1,
			balance: balancemodel.Balance{
				Current:   decimal.NewFromFloat(100.5),
				Withdrawn: decimal.NewFromFloat(50),
			},
			wantErr: false,
		},
		{
			name:    "service error",
			userID:  1,
			err:     errors.New("db error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			svc := mocks.NewMockBalanceService(ctrl)
			svc.EXPECT().
				GetBalance(gomock.Any(), tt.userID).
				Return(tt.balance, tt.err)

			uc := NewUsecase(svc)
			_, err := uc.GetBalance(context.Background(), tt.userID)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetBalance() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
