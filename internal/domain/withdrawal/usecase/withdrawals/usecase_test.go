package withdrawals

import (
	"context"
	"errors"
	"testing"

	ordersmodel "loyalty/internal/domain/order/model"
	withdrawalsmodel "loyalty/internal/domain/withdrawal/model"

	"github.com/shopspring/decimal"
)

type mockWithdrawalsService struct {
	withdrawErr error
	withdrawals []withdrawalsmodel.Withdrawal
	listErr     error
}

func (m *mockWithdrawalsService) Withdraw(ctx context.Context, userID int64, orderNumber string, sum decimal.Decimal) error {
	return m.withdrawErr
}

func (m *mockWithdrawalsService) ListWithdrawals(ctx context.Context, userID int64) ([]withdrawalsmodel.Withdrawal, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	return m.withdrawals, nil
}

type mockOrderNumberValidator struct {
	normalized string
	err        error
}

func (m *mockOrderNumberValidator) ValidateNumber(number string) (string, error) {
	if m.err != nil {
		return "", m.err
	}
	return m.normalized, nil
}

func TestUsecase_Withdraw(t *testing.T) {
	tests := []struct {
		name      string
		svc       *mockWithdrawalsService
		validator *mockOrderNumberValidator
		wantErr   bool
	}{
		{
			name:      "success",
			svc:       &mockWithdrawalsService{},
			validator: &mockOrderNumberValidator{normalized: "12345678903"},
			wantErr:   false,
		},
		{
			name:      "invalid order number",
			svc:       &mockWithdrawalsService{},
			validator: &mockOrderNumberValidator{err: ordersmodel.ErrInvalidOrderNumber},
			wantErr:   true,
		},
		{
			name: "service error",
			svc: &mockWithdrawalsService{
				withdrawErr: errors.New("withdraw failed"),
			},
			validator: &mockOrderNumberValidator{normalized: "12345678903"},
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := NewUsecase(tt.svc, tt.validator)
			err := uc.Withdraw(context.Background(), 1, "12345678903", decimal.NewFromFloat(100))
			if (err != nil) != tt.wantErr {
				t.Errorf("Withdraw() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestUsecase_ListWithdrawals(t *testing.T) {
	tests := []struct {
		name    string
		svc     *mockWithdrawalsService
		wantLen int
		wantErr bool
	}{
		{
			name: "success with items",
			svc: &mockWithdrawalsService{
				withdrawals: []withdrawalsmodel.Withdrawal{
					{OrderNumber: "123"},
				},
			},
			wantLen: 1,
			wantErr: false,
		},
		{
			name:    "empty list",
			svc:     &mockWithdrawalsService{withdrawals: []withdrawalsmodel.Withdrawal{}},
			wantLen: 0,
			wantErr: false,
		},
		{
			name: "service error",
			svc: &mockWithdrawalsService{
				listErr: errors.New("list failed"),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := NewUsecase(tt.svc, &mockOrderNumberValidator{normalized: "123"})
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
