package accrual

import (
	"context"
	"errors"
	"loyalty/internal/domain/accrual/client"
	"loyalty/internal/domain/accrual/model"
	ordersmodel "loyalty/internal/domain/order/model"
	ordersrepo "loyalty/internal/domain/order/repository"
	orderssvc "loyalty/internal/domain/order/service"
	"loyalty/internal/logger"
	"sync"
	"sync/atomic"
	"time"
)

// Worker — фоновый воркер для обновления статусов заказов через систему accrual.
type Worker struct {
	ordersRepo         ordersrepo.OrdersRepository
	ordersService      orderssvc.OrdersService
	accrualClient      client.AccrualClient
	log                logger.Logger
	pollInterval       time.Duration
	maxConcurrency     int
	queryTimeout       time.Duration
	requestDelay       time.Duration
	retryAfterMin      time.Duration
	pauseUntilUnixNano atomic.Int64
}

// Config содержит параметры воркера.
type Config struct {
	PollInterval   time.Duration // Интервал опроса БД (по умолчанию 5s)
	MaxConcurrency int           // Количество параллельных воркеров (по умолчанию 5)
	QueryTimeout   time.Duration // Таймаут для БД операций (по умолчанию 3s)
	RequestDelay   time.Duration // Задержка между запросами (по умолчанию 100ms)
	RetryAfterMin  time.Duration // Минимальная пауза при 429 (по умолчанию 60s)
}

// DefaultConfig возвращает дефолтную конфигурацию воркера.
func DefaultConfig() Config {
	return Config{
		PollInterval:   5 * time.Second,
		MaxConcurrency: 5,
		QueryTimeout:   3 * time.Second,
		RequestDelay:   100 * time.Millisecond,
		RetryAfterMin:  60 * time.Second,
	}
}

// NewWorker создаёт воркер для обновления заказов через accrual.
func NewWorker(
	ordersRepo ordersrepo.OrdersRepository,
	ordersService orderssvc.OrdersService,
	accrualClient client.AccrualClient,
	cfg Config,
	log logger.Logger,
) *Worker {
	return &Worker{
		ordersRepo:     ordersRepo,
		ordersService:  ordersService,
		accrualClient:  accrualClient,
		log:            log,
		pollInterval:   cfg.PollInterval,
		maxConcurrency: cfg.MaxConcurrency,
		queryTimeout:   cfg.QueryTimeout,
		requestDelay:   cfg.RequestDelay,
		retryAfterMin:  cfg.RetryAfterMin,
	}
}

// Start запускает воркер в фоне. Блокируется до отмены ctx.
func (worker *Worker) Start(ctx context.Context) {
	worker.log.Info().
		Duration("poll_interval", worker.pollInterval).
		Int("max_concurrency", worker.maxConcurrency).
		Message("accrual worker started")

	ticker := time.NewTicker(worker.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			worker.log.Info().Message("accrual worker stopped")
			return
		case <-ticker.C:
			worker.processBatch(ctx)
		}
	}
}

func (worker *Worker) processBatch(ctx context.Context) {
	queryCtx, cancel := context.WithTimeout(ctx, worker.queryTimeout)
	defer cancel()

	orders, err := worker.ordersRepo.ListPending(queryCtx)
	if err != nil {
		worker.log.Error().Error(err).Message("failed to list pending orders")
		return
	}

	if len(orders) == 0 {
		return
	}

	worker.log.Debug().Int("count", len(orders)).Message("processing pending orders")

	ordersChan := make(chan ordersmodel.Order, len(orders))

	var wg sync.WaitGroup
	for i := 0; i < worker.maxConcurrency; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			for order := range ordersChan {
				if ctx.Err() != nil {
					return
				}

				worker.sleepIfPaused(ctx)
				if ctx.Err() != nil {
					return
				}
				time.Sleep(worker.requestDelay)
				worker.processOrder(ctx, order)
			}
		}(i)
	}

	for _, order := range orders {
		ordersChan <- order
	}
	close(ordersChan)

	wg.Wait()
}

func (worker *Worker) processOrder(ctx context.Context, order ordersmodel.Order) {
	worker.sleepIfPaused(ctx)
	if ctx.Err() != nil {
		return
	}

	accrualResp, err := worker.accrualClient.GetOrderAccrual(ctx, order.Number)
	if err != nil {
		if errors.Is(err, model.ErrTooManyRequests) {
			pauseUntil := time.Now().Add(worker.retryAfterMin)
			worker.extendPauseUntil(pauseUntil)
			worker.log.Warn().
				Duration("retry_after", worker.retryAfterMin).
				Time("pause_until", pauseUntil).
				Message("accrual rate limit exceeded, pausing workers")
			return
		}
		if errors.Is(err, model.ErrTemporarilyUnavailable) {
			pauseUntil := time.Now().Add(worker.retryAfterMin)
			worker.extendPauseUntil(pauseUntil)
			worker.log.Warn().
				Duration("retry_after", worker.retryAfterMin).
				Time("pause_until", pauseUntil).
				Message("accrual temporarily unavailable, pausing workers")
			return
		}

		worker.log.Error().
			Error(err).
			String("order", order.Number).
			Message("failed to get accrual for order")
		return
	}

	if accrualResp == nil {
		worker.log.Debug().String("order", order.Number).Message("order not registered in accrual system yet")
		return
	}

	updateCtx, cancel := context.WithTimeout(ctx, worker.queryTimeout)
	defer cancel()

	if err := worker.ordersService.UpdateFromAccrual(updateCtx, order.Number, accrualResp.Status, accrualResp.Accrual); err != nil {
		worker.log.Error().
			Error(err).
			String("order", order.Number).
			String("accrual_status", string(accrualResp.Status)).
			Message("failed to update order from accrual")
		return
	}

	worker.log.Info().
		String("order", order.Number).
		String("old_status", string(order.Status)).
		String("accrual_status", string(accrualResp.Status)).
		Interface("accrual", accrualResp.Accrual).
		Message("order updated from accrual")
}

func (worker *Worker) extendPauseUntil(until time.Time) {
	newVal := until.UnixNano()
	for {
		old := worker.pauseUntilUnixNano.Load()
		if old >= newVal {
			return
		}
		if worker.pauseUntilUnixNano.CompareAndSwap(old, newVal) {
			return
		}
	}
}

func (worker *Worker) sleepIfPaused(ctx context.Context) {
	until := time.Unix(0, worker.pauseUntilUnixNano.Load())
	if !until.After(time.Now()) {
		return
	}

	timer := time.NewTimer(time.Until(until))
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return
	case <-timer.C:
		return
	}
}
