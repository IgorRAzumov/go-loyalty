package orders

import (
	"context"
	"testing"

	"loyalty/internal/domain/order/model"

	"github.com/shopspring/decimal"
)

type mockRepo struct {
	created bool

	gotUserID int64
	gotNumber string
}

func (m *mockRepo) Create(ctx context.Context, userID int64, number string) error {
	m.created = true
	m.gotUserID = userID
	m.gotNumber = number
	return nil
}

func (m *mockRepo) ListByUser(context.Context, int64) ([]model.Order, error) { return nil, nil }
func (m *mockRepo) ListPending(context.Context) ([]model.Order, error)       { return nil, nil }
func (m *mockRepo) UpdateFromAccrual(context.Context, string, model.Status, *decimal.Decimal) error {
	return nil
}

type mockNumberService struct {
	normalized string
	err        error
}

func (m *mockNumberService) ValidateNumber(string) (string, error) { return m.normalized, m.err }

func TestService_UploadOrder_CallsRepoWithNormalizedNumber(t *testing.T) {
	repo := &mockRepo{}
	num := &mockNumberService{normalized: "79927398713"}
	svc := NewService(repo, num)

	if err := svc.UploadOrder(context.Background(), 10, " 79927398713 "); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if !repo.created {
		t.Fatalf("expected repo.Create to be called")
	}
	if repo.gotUserID != 10 {
		t.Fatalf("want userID %d, got %d", 10, repo.gotUserID)
	}
	if repo.gotNumber != "79927398713" {
		t.Fatalf("want number %q, got %q", "79927398713", repo.gotNumber)
	}
}

func TestService_UploadOrder_InvalidNumber_ReturnsDomainErrorAndDoesNotCreate(t *testing.T) {
	repo := &mockRepo{}
	num := &mockNumberService{err: model.ErrInvalidOrderNumber}
	svc := NewService(repo, num)

	err := svc.UploadOrder(context.Background(), 10, "bad")
	if err == nil || err != model.ErrInvalidOrderNumber {
		t.Fatalf("want %v, got %v", model.ErrInvalidOrderNumber, err)
	}
	if repo.created {
		t.Fatalf("did not expect repo.Create to be called")
	}
}
