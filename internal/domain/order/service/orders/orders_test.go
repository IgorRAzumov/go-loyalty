package orders

import (
	"context"
	"testing"

	"loyalty/internal/domain/order/model"
	"loyalty/internal/mocks"

	"go.uber.org/mock/gomock"
)

func TestService_UploadOrder_CallsRepoWithNormalizedNumber(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockOrdersRepository(ctrl)
	repo.EXPECT().
		Create(gomock.Any(), int64(10), "79927398713").
		Return(nil)

	num := mocks.NewMockOrderNumberValidator(ctrl)
	num.EXPECT().
		ValidateNumber(" 79927398713 ").
		Return("79927398713", nil)

	svc := NewService(repo, num)
	if err := svc.UploadOrder(context.Background(), 10, " 79927398713 "); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
}

func TestService_UploadOrder_InvalidNumber_ReturnsDomainErrorAndDoesNotCreate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockOrdersRepository(ctrl)
	repo.EXPECT().Create(gomock.Any(), gomock.Any(), gomock.Any()).MaxTimes(0)

	num := mocks.NewMockOrderNumberValidator(ctrl)
	num.EXPECT().
		ValidateNumber("bad").
		Return("", model.ErrInvalidOrderNumber)

	svc := NewService(repo, num)
	err := svc.UploadOrder(context.Background(), 10, "bad")
	if err == nil || err != model.ErrInvalidOrderNumber {
		t.Fatalf("want %v, got %v", model.ErrInvalidOrderNumber, err)
	}
}
