package orders

import (
	"context"
	"errors"
	"testing"

	accrualmodel "loyalty/internal/domain/accrual/model"
	"loyalty/internal/domain/order/model"

	"github.com/shopspring/decimal"
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
			repo := &mockRepoWithError{updateErr: tt.repoErr}
			svc := NewService(repo, &mockNumberValidator{})

			err := svc.UpdateFromAccrual(context.Background(), "123", tt.accrualStatus, tt.accrual)
			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateFromAccrual() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && !repo.updateCalled {
				t.Error("expected repo.UpdateFromAccrual to be called")
			}
		})
	}
}

func TestMapAccrualStatusToOrderStatus(t *testing.T) {
	tests := []struct {
		accrualStatus accrualmodel.AccrualStatus
		want          model.Status
	}{
		{accrualmodel.StatusRegistered, model.StatusProcessing},
		{accrualmodel.StatusProcessing, model.StatusProcessing},
		{accrualmodel.StatusInvalid, model.StatusInvalid},
		{accrualmodel.StatusProcessed, model.StatusProcessed},
		{"UNKNOWN", model.StatusProcessing},
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

type mockRepoWithError struct {
	mockRepo
	updateErr    error
	updateCalled bool
}

func (m *mockRepoWithError) UpdateFromAccrual(ctx context.Context, number string, status model.Status, accrual *decimal.Decimal) error {
	m.updateCalled = true
	return m.updateErr
}

type mockNumberValidator struct{}

func (m *mockNumberValidator) ValidateNumber(number string) (string, error) {
	return number, nil
}

func decimalPtr(v float64) *decimal.Decimal {
	d := decimal.NewFromFloat(v)
	return &d
}
