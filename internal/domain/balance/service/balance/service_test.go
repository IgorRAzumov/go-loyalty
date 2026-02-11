package balance

import (
	"context"
	"errors"
	"testing"

	"loyalty/internal/domain/balance/model"
	"loyalty/internal/mocks"

	"github.com/shopspring/decimal"
	"go.uber.org/mock/gomock"
)

func TestService_GetBalance_DelegatesToRepo(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	expectedBalance := model.Balance{
		Current:   decimal.NewFromFloat(123.45),
		Withdrawn: decimal.NewFromFloat(67.89),
	}

	repo := mocks.NewMockBalanceRepository(ctrl)
	repo.EXPECT().
		GetBalance(gomock.Any(), int64(10)).
		Return(expectedBalance, nil)

	svc := NewService(repo)
	result, err := svc.GetBalance(context.Background(), 10)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if !result.Current.Equal(expectedBalance.Current) {
		t.Fatalf("want current %v, got %v", expectedBalance.Current, result.Current)
	}
	if !result.Withdrawn.Equal(expectedBalance.Withdrawn) {
		t.Fatalf("want withdrawn %v, got %v", expectedBalance.Withdrawn, result.Withdrawn)
	}
}

func TestService_GetBalance_PropagatesRepoError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repoErr := errors.New("database error")
	repo := mocks.NewMockBalanceRepository(ctrl)
	repo.EXPECT().
		GetBalance(gomock.Any(), int64(10)).
		Return(model.Balance{}, repoErr)

	svc := NewService(repo)
	_, err := svc.GetBalance(context.Background(), 10)
	if !errors.Is(err, repoErr) {
		t.Fatalf("want %v, got %v", repoErr, err)
	}
}
