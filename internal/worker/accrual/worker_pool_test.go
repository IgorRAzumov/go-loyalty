package accrual

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	accrualmodel "loyalty/internal/domain/accrual/model"
	ordersmodel "loyalty/internal/domain/order/model"
	"loyalty/internal/mocks"

	"github.com/shopspring/decimal"
	"go.uber.org/mock/gomock"
)

func TestWorkerPool_ParallelProcessing(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	orders := []ordersmodel.Order{
		{Number: "1", Status: ordersmodel.StatusNew},
		{Number: "2", Status: ordersmodel.StatusNew},
		{Number: "3", Status: ordersmodel.StatusNew},
		{Number: "4", Status: ordersmodel.StatusNew},
		{Number: "5", Status: ordersmodel.StatusNew},
	}

	repo := mocks.NewMockOrdersRepository(ctrl)
	repo.EXPECT().ListPending(gomock.Any()).Return(orders, nil)

	svc := mocks.NewMockOrdersService(ctrl)
	svc.EXPECT().
		UpdateFromAccrual(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil).Times(5)

	client := mocks.NewMockAccrualClient(ctrl)
	for _, o := range orders {
		client.EXPECT().
			GetOrderAccrual(gomock.Any(), o.Number).
			Return(&accrualmodel.Accrual{Order: o.Number, Status: accrualmodel.StatusProcessed, Accrual: decimalPtr(100)}, nil)
	}

	cfg := DefaultConfig()
	cfg.MaxConcurrency = 3
	cfg.RequestDelay = 10 * time.Millisecond

	w := NewWorker(repo, svc, client, cfg)

	start := time.Now()
	w.processBatch(context.Background())
	elapsed := time.Since(start)

	if elapsed > 40*time.Millisecond {
		t.Errorf("processBatch took %v, expected <40ms with concurrency", elapsed)
	}
}

func TestWorkerPool_GracefulShutdown(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	orders := make([]ordersmodel.Order, 100)
	for i := range orders {
		orders[i] = ordersmodel.Order{Number: string(rune(i)), Status: ordersmodel.StatusNew}
	}

	repo := mocks.NewMockOrdersRepository(ctrl)
	repo.EXPECT().ListPending(gomock.Any()).Return(orders, nil)

	var processedCount int32
	svc := mocks.NewMockOrdersService(ctrl)
	svc.EXPECT().
		UpdateFromAccrual(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(context.Context, string, accrualmodel.AccrualStatus, *decimal.Decimal) error {
			atomic.AddInt32(&processedCount, 1)
			return nil
		}).AnyTimes()

	client := mocks.NewMockAccrualClient(ctrl)
	client.EXPECT().
		GetOrderAccrual(gomock.Any(), gomock.Any()).
		Return(&accrualmodel.Accrual{Status: accrualmodel.StatusProcessed, Accrual: decimalPtr(100)}, nil).AnyTimes()

	cfg := DefaultConfig()
	cfg.MaxConcurrency = 5
	cfg.RequestDelay = 1 * time.Millisecond

	w := NewWorker(repo, svc, client, cfg)

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(10 * time.Millisecond)
		cancel()
	}()

	w.processBatch(ctx)

	processed := atomic.LoadInt32(&processedCount)
	if processed == 0 {
		t.Error("no orders processed before cancellation")
	}
	if processed >= 100 {
		t.Error("all orders processed despite cancellation")
	}
	t.Logf("processed %d orders before cancellation", processed)
}
