package accrual

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"loyalty/internal/domain/accrual/model"
	ordersmodel "loyalty/internal/domain/order/model"
)

// TestWorkerPool_ParallelProcessing проверяет что заказы обрабатываются параллельно
func TestWorkerPool_ParallelProcessing(t *testing.T) {
	var processedCount int32
	repo := &mockOrdersRepo{
		orders: []ordersmodel.Order{
			{Number: "1", Status: ordersmodel.StatusNew},
			{Number: "2", Status: ordersmodel.StatusNew},
			{Number: "3", Status: ordersmodel.StatusNew},
			{Number: "4", Status: ordersmodel.StatusNew},
			{Number: "5", Status: ordersmodel.StatusNew},
		},
	}

	client := &countingAccrualClient{
		counter: &processedCount,
	}

	cfg := DefaultConfig()
	cfg.MaxConcurrency = 3
	cfg.RequestDelay = 10 * time.Millisecond

	w := NewWorker(repo, &mockOrdersService{}, client, cfg)

	start := time.Now()
	w.processBatch(context.Background())
	elapsed := time.Since(start)

	// С параллельной обработкой (3 воркера) это должно занять ~20ms
	// (5 заказов / 3 воркера = 2 раунда * 10ms)
	// Без параллелизации: 5 * 10ms = 50ms
	if elapsed > 40*time.Millisecond {
		t.Errorf("processBatch took %v, expected <40ms with concurrency", elapsed)
	}

	if atomic.LoadInt32(&processedCount) != 5 {
		t.Errorf("processed %d orders, want 5", atomic.LoadInt32(&processedCount))
	}
}

func TestWorkerPool_GracefulShutdown(t *testing.T) {
	repo := &mockOrdersRepo{
		orders: make([]ordersmodel.Order, 100), // Много заказов
	}
	for i := range repo.orders {
		repo.orders[i] = ordersmodel.Order{Number: string(rune(i)), Status: ordersmodel.StatusNew}
	}

	var processedCount int32
	client := &countingAccrualClient{counter: &processedCount}

	cfg := DefaultConfig()
	cfg.MaxConcurrency = 5
	cfg.RequestDelay = 1 * time.Millisecond

	w := NewWorker(repo, &mockOrdersService{}, client, cfg)

	ctx, cancel := context.WithCancel(context.Background())

	// Отменяем контекст через 10ms
	go func() {
		time.Sleep(10 * time.Millisecond)
		cancel()
	}()

	w.processBatch(ctx)

	// Должны обработать хотя бы несколько заказов, но не все 100
	processed := atomic.LoadInt32(&processedCount)
	if processed == 0 {
		t.Error("no orders processed before cancellation")
	}
	if processed >= 100 {
		t.Error("all orders processed despite cancellation")
	}
	t.Logf("processed %d orders before cancellation", processed)
}

type countingAccrualClient struct {
	counter *int32
	mu      sync.Mutex
}

func (c *countingAccrualClient) GetOrderAccrual(ctx context.Context, orderNumber string) (*model.Accrual, error) {
	atomic.AddInt32(c.counter, 1)
	return &model.Accrual{
		Order:   orderNumber,
		Status:  model.StatusProcessed,
		Accrual: decimalPtr(100),
	}, nil
}
