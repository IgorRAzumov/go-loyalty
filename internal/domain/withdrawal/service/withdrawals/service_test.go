package withdrawals

import (
	"context"
	"errors"
	"testing"
	"time"

	"loyalty/internal/domain/withdrawal/model"
	"loyalty/internal/mocks"

	"github.com/shopspring/decimal"
	"go.uber.org/mock/gomock"
)

func TestService_Withdraw_CallsRepoWithValidSum(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	sum := decimal.NewFromFloat(123.45)
	accRepo := mocks.NewMockAccountRepository(ctrl)
	accRepo.EXPECT().
		Withdraw(gomock.Any(), int64(10), "79927398713", sum, gomock.Any()).
		DoAndReturn(func(ctx context.Context, userID int64, order string, s decimal.Decimal, now time.Time) error {
			if now.IsZero() {
				t.Fatal("expected time.Now to be passed")
			}
			return nil
		})

	wdRepo := mocks.NewMockWithdrawalsRepository(ctrl)
	wdRepo.EXPECT().ListByUser(gomock.Any(), gomock.Any()).MaxTimes(0)

	svc := NewService(accRepo, wdRepo)
	if err := svc.Withdraw(context.Background(), 10, "79927398713", sum); err != nil {
		t.Fatalf("unexpected err: %v", err)
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
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			accRepo := mocks.NewMockAccountRepository(ctrl)
			accRepo.EXPECT().Withdraw(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).MaxTimes(0)

			wdRepo := mocks.NewMockWithdrawalsRepository(ctrl)
			svc := NewService(accRepo, wdRepo)

			err := svc.Withdraw(context.Background(), 10, "79927398713", tt.sum)
			if err == nil || !errors.Is(err, model.ErrInvalidWithdrawSum) {
				t.Fatalf("want %v, got %v", model.ErrInvalidWithdrawSum, err)
			}
		})
	}
}

func TestService_Withdraw_PropagatesRepoError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repoErr := model.ErrInsufficientFunds
	accRepo := mocks.NewMockAccountRepository(ctrl)
	accRepo.EXPECT().
		Withdraw(gomock.Any(), int64(10), "79927398713", decimal.NewFromInt(100), gomock.Any()).
		Return(repoErr)

	wdRepo := mocks.NewMockWithdrawalsRepository(ctrl)
	svc := NewService(accRepo, wdRepo)

	err := svc.Withdraw(context.Background(), 10, "79927398713", decimal.NewFromInt(100))
	if !errors.Is(err, repoErr) {
		t.Fatalf("want %v, got %v", repoErr, err)
	}
}

func TestService_ListWithdrawals_DelegatesToRepo(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	accRepo := mocks.NewMockAccountRepository(ctrl)
	accRepo.EXPECT().Withdraw(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).MaxTimes(0)

	wdRepo := mocks.NewMockWithdrawalsRepository(ctrl)
	wdRepo.EXPECT().
		ListByUser(gomock.Any(), int64(10)).
		Return([]model.Withdrawal{}, nil)

	svc := NewService(accRepo, wdRepo)
	result, err := svc.ListWithdrawals(context.Background(), 10)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if result == nil {
		t.Fatalf("expected non-nil result")
	}
}
