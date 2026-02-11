package accrual

import (
	"context"
	"errors"
	"testing"
	"time"

	accrualmodel "loyalty/internal/domain/accrual/model"
	ordersmodel "loyalty/internal/domain/order/model"
	"loyalty/internal/logger"
	"loyalty/internal/mocks"

	"github.com/shopspring/decimal"
	"go.uber.org/mock/gomock"
)

func TestWorker_processBatch(t *testing.T) {
	tests := []struct {
		name    string
		orders  []ordersmodel.Order
		listErr error
	}{
		{
			name:   "empty orders",
			orders: []ordersmodel.Order{},
		},
		{
			name:    "repo error",
			listErr: errors.New("db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := mocks.NewMockOrdersRepository(ctrl)
			repo.EXPECT().
				ListPending(gomock.Any()).
				Return(tt.orders, tt.listErr)

			svc := mocks.NewMockOrdersService(ctrl)
			client := mocks.NewMockAccrualClient(ctrl)

			w := NewWorker(repo, svc, client, DefaultConfig(), logger.NewNopLogger())
			w.processBatch(context.Background())
		})
	}
}

func TestWorker_processOrder(t *testing.T) {
	tests := []struct {
		name       string
		order      ordersmodel.Order
		accrual    *accrualmodel.Accrual
		accrualErr error
		updateErr  error
	}{
		{
			name:    "accrual found and updated",
			order:   ordersmodel.Order{Number: "123", Status: ordersmodel.StatusNew},
			accrual: &accrualmodel.Accrual{Order: "123", Status: accrualmodel.StatusProcessed, Accrual: decimalPtr(100)},
		},
		{
			name:    "accrual not found (204)",
			order:   ordersmodel.Order{Number: "123", Status: ordersmodel.StatusNew},
			accrual: nil,
		},
		{
			name:       "rate limit error",
			order:      ordersmodel.Order{Number: "123", Status: ordersmodel.StatusNew},
			accrualErr: accrualmodel.ErrTooManyRequests,
		},
		{
			name:       "network error",
			order:      ordersmodel.Order{Number: "123", Status: ordersmodel.StatusNew},
			accrualErr: errors.New("connection failed"),
		},
		{
			name:      "update error",
			order:     ordersmodel.Order{Number: "123", Status: ordersmodel.StatusNew},
			accrual:   &accrualmodel.Accrual{Order: "123", Status: accrualmodel.StatusProcessed, Accrual: decimalPtr(100)},
			updateErr: errors.New("update failed"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := mocks.NewMockOrdersRepository(ctrl)
			svc := mocks.NewMockOrdersService(ctrl)
			client := mocks.NewMockAccrualClient(ctrl)

			client.EXPECT().
				GetOrderAccrual(gomock.Any(), tt.order.Number).
				Return(tt.accrual, tt.accrualErr)

			if tt.accrual != nil && tt.accrualErr == nil {
				svc.EXPECT().
					UpdateFromAccrual(gomock.Any(), tt.order.Number, gomock.Any(), gomock.Any()).
					Return(tt.updateErr)
			}

			cfg := DefaultConfig()
			cfg.RequestDelay = 0
			cfg.RetryAfterMin = 10 * time.Millisecond

			w := NewWorker(repo, svc, client, cfg, logger.NewNopLogger())
			w.processOrder(context.Background(), tt.order)
		})
	}
}

func TestWorker_SleepIfPaused(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockOrdersRepository(ctrl)
	svc := mocks.NewMockOrdersService(ctrl)
	client := mocks.NewMockAccrualClient(ctrl)

	cfg := DefaultConfig()
	cfg.RequestDelay = 0
	cfg.RetryAfterMin = 0
	w := NewWorker(repo, svc, client, cfg, logger.NewNopLogger())

	pauseFor := 25 * time.Millisecond
	w.extendPauseUntil(time.Now().Add(pauseFor))

	start := time.Now()
	w.sleepIfPaused(context.Background())
	elapsed := time.Since(start)

	if elapsed < pauseFor-5*time.Millisecond {
		t.Fatalf("sleepIfPaused slept %s, want at least %s", elapsed, pauseFor)
	}
}

func decimalPtr(v float64) *decimal.Decimal {
	d := decimal.NewFromFloat(v)
	return &d
}
