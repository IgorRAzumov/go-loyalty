package accrual

import (
	"context"
	"errors"
	"loyalty/internal/domain/accrual/client"
	"loyalty/internal/domain/accrual/model"
	ordersmodel "loyalty/internal/domain/order/model"
	ordersrepo "loyalty/internal/domain/order/repository"
	orderssvc "loyalty/internal/domain/order/service"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

// Worker — фоновый воркер для обновления статусов заказов через систему accrual.
type Worker struct {
	ordersRepo     ordersrepo.OrdersRepository
	ordersService  orderssvc.OrdersService
	accrualClient  client.AccrualClient
	pollInterval   time.Duration
	maxConcurrency int
	queryTimeout   time.Duration
	requestDelay   time.Duration
	retryAfterMin  time.Duration
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
) *Worker {
	return &Worker{
		ordersRepo:     ordersRepo,
		ordersService:  ordersService,
		accrualClient:  accrualClient,
		pollInterval:   cfg.PollInterval,
		maxConcurrency: cfg.MaxConcurrency,
		queryTimeout:   cfg.QueryTimeout,
		requestDelay:   cfg.RequestDelay,
		retryAfterMin:  cfg.RetryAfterMin,
	}
}

// Start запускает воркер в фоне. Блокируется до отмены ctx.
func (worker *Worker) Start(ctx context.Context) {
	log.Info().
		Dur("poll_interval", worker.pollInterval).
		Int("max_concurrency", worker.maxConcurrency).
		Msg("accrual worker started")

	ticker := time.NewTicker(worker.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Info().Msg("accrual worker stopped")
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
		log.Error().Err(err).Msg("failed to list pending orders")
		return
	}

	if len(orders) == 0 {
		return
	}

	log.Debug().Int("count", len(orders)).Msg("processing pending orders")

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
	accrualResp, err := worker.accrualClient.GetOrderAccrual(ctx, order.Number)
	if err != nil {
		if errors.Is(err, model.ErrTooManyRequests) {
			log.Warn().
				Dur("retry_after", worker.retryAfterMin).
				Msg("accrual rate limit exceeded, pausing worker")
			time.Sleep(worker.retryAfterMin)
			return
		}

		log.Error().
			Err(err).
			Str("order", order.Number).
			Msg("failed to get accrual for order")
		return
	}

	if accrualResp == nil {
		log.Debug().Str("order", order.Number).Msg("order not registered in accrual system yet")
		return
	}

	updateCtx, cancel := context.WithTimeout(ctx, worker.queryTimeout)
	defer cancel()

	if err := worker.ordersService.UpdateFromAccrual(updateCtx, order.Number, accrualResp.Status, accrualResp.Accrual); err != nil {
		log.Error().
			Err(err).
			Str("order", order.Number).
			Str("accrual_status", string(accrualResp.Status)).
			Msg("failed to update order from accrual")
		return
	}

	log.Info().
		Str("order", order.Number).
		Str("old_status", string(order.Status)).
		Str("accrual_status", string(accrualResp.Status)).
		Interface("accrual", accrualResp.Accrual).
		Msg("order updated from accrual")
}
