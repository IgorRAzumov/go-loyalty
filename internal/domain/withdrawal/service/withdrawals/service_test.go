package withdrawals

import (
	"context"
	"errors"
	"testing"
	"time"

	"loyalty/internal/domain/withdrawal/model"

	"github.com/shopspring/decimal"
)

type mockAccountRepo struct {
	withdrawn bool
	gotUserID int64
	gotOrder  string
	gotSum    decimal.Decimal
	gotTime   time.Time
	err       error
}

func (m *mockAccountRepo) GetBalance(context.Context, int64) (decimal.Decimal, error) {
	return decimal.Zero, nil
}

func (m *mockAccountRepo) Withdraw(ctx context.Context, userID int64, orderNumber string, sum decimal.Decimal, processedAt time.Time) error {
	m.withdrawn = true
	m.gotUserID = userID
	m.gotOrder = orderNumber
	m.gotSum = sum
	m.gotTime = processedAt
	return m.err
}

type mockWithdrawalsRepo struct {
	called bool
}

func (m *mockWithdrawalsRepo) ListByUser(context.Context, int64) ([]model.Withdrawal, error) {
	m.called = true
	return []model.Withdrawal{}, nil
}

func TestService_Withdraw_CallsRepoWithValidSum(t *testing.T) {
	accRepo := &mockAccountRepo{}
	wdRepo := &mockWithdrawalsRepo{}
	svc := NewService(accRepo, wdRepo)

	sum := decimal.NewFromFloat(123.45)
	if err := svc.Withdraw(context.Background(), 10, "79927398713", sum); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if !accRepo.withdrawn {
		t.Fatalf("expected accountRepository.Withdraw to be called")
	}
	if accRepo.gotUserID != 10 {
		t.Fatalf("want userID %d, got %d", 10, accRepo.gotUserID)
	}
	if accRepo.gotOrder != "79927398713" {
		t.Fatalf("want order %q, got %q", "79927398713", accRepo.gotOrder)
	}
	if !accRepo.gotSum.Equal(sum) {
		t.Fatalf("want sum %v, got %v", sum, accRepo.gotSum)
	}
	if accRepo.gotTime.IsZero() {
		t.Fatalf("expected time.Now to be passed")
	}
}

func TestService_Withdraw_InvalidSum_ReturnsErrInvalidWithdrawSum(t *testing.T) {
	tests := []struct {
		name string
		sum  decimal.Decimal
	}{
		{"zero", decimal.Zero},
		{"negative", decimal.NewFromFloat(-10)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			accRepo := &mockAccountRepo{}
			wdRepo := &mockWithdrawalsRepo{}
			svc := NewService(accRepo, wdRepo)

			err := svc.Withdraw(context.Background(), 10, "79927398713", tt.sum)
			if err == nil || !errors.Is(err, model.ErrInvalidWithdrawSum) {
				t.Fatalf("want %v, got %v", model.ErrInvalidWithdrawSum, err)
			}
			if accRepo.withdrawn {
				t.Fatalf("did not expect accountRepository.Withdraw to be called")
			}
		})
	}
}

func TestService_Withdraw_PropagatesRepoError(t *testing.T) {
	repoErr := model.ErrInsufficientFunds
	accRepo := &mockAccountRepo{err: repoErr}
	wdRepo := &mockWithdrawalsRepo{}
	svc := NewService(accRepo, wdRepo)

	err := svc.Withdraw(context.Background(), 10, "79927398713", decimal.NewFromInt(100))
	if !errors.Is(err, repoErr) {
		t.Fatalf("want %v, got %v", repoErr, err)
	}
}

func TestService_ListWithdrawals_DelegatesToRepo(t *testing.T) {
	accRepo := &mockAccountRepo{}
	wdRepo := &mockWithdrawalsRepo{}
	svc := NewService(accRepo, wdRepo)

	result, err := svc.ListWithdrawals(context.Background(), 10)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if !wdRepo.called {
		t.Fatalf("expected withdrawalsRepository.ListByUser to be called")
	}
	if result == nil {
		t.Fatalf("expected non-nil result")
	}
}
