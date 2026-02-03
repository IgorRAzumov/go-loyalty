package accrual

import (
	"context"
	"errors"
	"testing"
	"time"

	accrualmodel "loyalty/internal/domain/accrual/model"
	ordersmodel "loyalty/internal/domain/order/model"

	"github.com/shopspring/decimal"
)

type mockOrdersRepo struct {
	orders      []ordersmodel.Order
	listErr     error
	updateErr   error
	updateCalls int
}

func (m *mockOrdersRepo) Create(ctx context.Context, userID int64, number string) error {
	return nil
}

func (m *mockOrdersRepo) ListByUser(ctx context.Context, userID int64) ([]ordersmodel.Order, error) {
	return nil, nil
}

func (m *mockOrdersRepo) ListPending(ctx context.Context) ([]ordersmodel.Order, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	return m.orders, nil
}

func (m *mockOrdersRepo) UpdateFromAccrual(ctx context.Context, number string, status ordersmodel.Status, accrual *decimal.Decimal) error {
	m.updateCalls++
	return m.updateErr
}

type mockOrdersService struct {
	updateErr error
}

func (m *mockOrdersService) UploadOrder(ctx context.Context, userID int64, orderNumber string) error {
	return nil
}

func (m *mockOrdersService) LoadOrders(ctx context.Context, userID int64) ([]ordersmodel.Order, error) {
	return nil, nil
}

func (m *mockOrdersService) UpdateFromAccrual(ctx context.Context, orderNumber string, accrualStatus accrualmodel.AccrualStatus, accrual *decimal.Decimal) error {
	return m.updateErr
}

type mockAccrualClient struct {
	response *accrualmodel.Accrual
	err      error
}

func (m *mockAccrualClient) GetOrderAccrual(ctx context.Context, orderNumber string) (*accrualmodel.Accrual, error) {
	return m.response, m.err
}

func TestWorker_processBatch(t *testing.T) {
	tests := []struct {
		name string
		repo *mockOrdersRepo
	}{
		{
			name: "empty orders",
			repo: &mockOrdersRepo{orders: []ordersmodel.Order{}},
		},
		{
			name: "repo error",
			repo: &mockOrdersRepo{listErr: errors.New("db error")},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := NewWorker(tt.repo, &mockOrdersService{}, &mockAccrualClient{}, DefaultConfig())
			w.processBatch(context.Background())
		})
	}
}

func TestWorker_processOrder(t *testing.T) {
	tests := []struct {
		name    string
		repo    *mockOrdersRepo
		service *mockOrdersService
		client  *mockAccrualClient
		order   ordersmodel.Order
	}{
		{
			name:    "accrual found and updated",
			repo:    &mockOrdersRepo{},
			service: &mockOrdersService{},
			client: &mockAccrualClient{
				response: &accrualmodel.Accrual{
					Order:   "123",
					Status:  accrualmodel.StatusProcessed,
					Accrual: decimalPtr(100),
				},
			},
			order: ordersmodel.Order{Number: "123", Status: ordersmodel.StatusNew},
		},
		{
			name:    "accrual not found (204)",
			repo:    &mockOrdersRepo{},
			service: &mockOrdersService{},
			client:  &mockAccrualClient{response: nil},
			order:   ordersmodel.Order{Number: "123", Status: ordersmodel.StatusNew},
		},
		{
			name:    "rate limit error",
			repo:    &mockOrdersRepo{},
			service: &mockOrdersService{},
			client:  &mockAccrualClient{err: accrualmodel.ErrTooManyRequests},
			order:   ordersmodel.Order{Number: "123", Status: ordersmodel.StatusNew},
		},
		{
			name:    "network error",
			repo:    &mockOrdersRepo{},
			service: &mockOrdersService{},
			client:  &mockAccrualClient{err: errors.New("connection failed")},
			order:   ordersmodel.Order{Number: "123", Status: ordersmodel.StatusNew},
		},
		{
			name:    "update error",
			repo:    &mockOrdersRepo{},
			service: &mockOrdersService{updateErr: errors.New("update failed")},
			client: &mockAccrualClient{
				response: &accrualmodel.Accrual{
					Order:   "123",
					Status:  accrualmodel.StatusProcessed,
					Accrual: decimalPtr(100),
				},
			},
			order: ordersmodel.Order{Number: "123", Status: ordersmodel.StatusNew},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := DefaultConfig()
			cfg.RequestDelay = 0
			cfg.RetryAfterMin = 10 * time.Millisecond
			w := NewWorker(tt.repo, tt.service, tt.client, cfg)
			w.processOrder(context.Background(), tt.order)
		})
	}
}

func decimalPtr(v float64) *decimal.Decimal {
	d := decimal.NewFromFloat(v)
	return &d
}
