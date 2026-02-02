package balance

import (
	"context"
	"errors"
	"testing"

	balancemodel "loyalty/internal/domain/balance/model"

	"github.com/shopspring/decimal"
)

type mockBalanceService struct {
	balance balancemodel.Balance
	err     error
}

func (m *mockBalanceService) GetBalance(ctx context.Context, userID int64) (balancemodel.Balance, error) {
	if m.err != nil {
		return balancemodel.Balance{}, m.err
	}
	return m.balance, nil
}

func TestUsecase_GetBalance(t *testing.T) {
	tests := []struct {
		name    string
		userID  int64
		svc     *mockBalanceService
		wantErr bool
	}{
		{
			name:   "success",
			userID: 1,
			svc: &mockBalanceService{
				balance: balancemodel.Balance{
					Current:   decimal.NewFromFloat(100.5),
					Withdrawn: decimal.NewFromFloat(50),
				},
			},
			wantErr: false,
		},
		{
			name:   "service error",
			userID: 1,
			svc: &mockBalanceService{
				err: errors.New("db error"),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := NewUsecase(tt.svc)
			_, err := uc.GetBalance(context.Background(), tt.userID)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetBalance() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
