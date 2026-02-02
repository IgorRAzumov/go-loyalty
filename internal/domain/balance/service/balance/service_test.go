package balance

import (
	"context"
	"errors"
	"testing"

	"loyalty/internal/domain/balance/model"

	"github.com/shopspring/decimal"
)

type mockBalanceRepo struct {
	balance model.Balance
	err     error
	called  bool
}

func (m *mockBalanceRepo) GetBalance(ctx context.Context, userID int64) (model.Balance, error) {
	m.called = true
	return m.balance, m.err
}

func TestService_GetBalance_DelegatesToRepo(t *testing.T) {
	expectedBalance := model.Balance{
		Current:   decimal.NewFromFloat(123.45),
		Withdrawn: decimal.NewFromFloat(67.89),
	}
	repo := &mockBalanceRepo{balance: expectedBalance}
	svc := NewService(repo)

	result, err := svc.GetBalance(context.Background(), 10)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if !repo.called {
		t.Fatalf("expected repo.GetBalance to be called")
	}
	if !result.Current.Equal(expectedBalance.Current) {
		t.Fatalf("want current %v, got %v", expectedBalance.Current, result.Current)
	}
	if !result.Withdrawn.Equal(expectedBalance.Withdrawn) {
		t.Fatalf("want withdrawn %v, got %v", expectedBalance.Withdrawn, result.Withdrawn)
	}
}

func TestService_GetBalance_PropagatesRepoError(t *testing.T) {
	repoErr := errors.New("database error")
	repo := &mockBalanceRepo{err: repoErr}
	svc := NewService(repo)

	_, err := svc.GetBalance(context.Background(), 10)
	if !errors.Is(err, repoErr) {
		t.Fatalf("want %v, got %v", repoErr, err)
	}
}
