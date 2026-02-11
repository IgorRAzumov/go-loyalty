package http

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"loyalty/internal/domain/accrual/model"
	"loyalty/internal/util/retry"
	"net/http"
	"strings"
	"time"

	"github.com/sony/gobreaker"
)

// Client реализует accrual.AccrualClient через HTTP.
type Client struct {
	baseURL    string
	httpClient *http.Client
	breaker    *gobreaker.CircuitBreaker
}

// NewClient создаёт HTTP-клиент для системы accrual.
func NewClient(baseURL string, timeout time.Duration) *Client {
	cb := initBreaker()

	return &Client{
		baseURL:    strings.TrimSuffix(baseURL, "/"),
		httpClient: &http.Client{Timeout: timeout},
		breaker:    cb,
	}
}

// GetOrderAccrual получает информацию о начислении для заказа.
func (client *Client) GetOrderAccrual(ctx context.Context, orderNumber string) (*model.Accrual, error) {
	res, err := retry.Do(ctx, retry.DefaultConfig(), func() (*model.Accrual, error) {
		raw, execErr := client.breaker.Execute(func() (any, error) {
			return client.doRequest(ctx, orderNumber)
		})
		if execErr != nil {
			if errors.Is(execErr, gobreaker.ErrOpenState) || errors.Is(execErr, gobreaker.ErrTooManyRequests) {
				return nil, fmt.Errorf("%w: %s", model.ErrTemporarilyUnavailable, execErr.Error())
			}
			return nil, execErr
		}
		if raw == nil {
			return nil, nil
		}
		accrualResp, ok := raw.(*model.Accrual)
		if !ok {
			return nil, fmt.Errorf("unexpected breaker response type %T", raw)
		}
		return accrualResp, nil
	}, client.isRetryable)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (client *Client) doRequest(ctx context.Context, orderNumber string) (*model.Accrual, error) {
	url := fmt.Sprintf("%s/api/orders/%s", client.baseURL, orderNumber)

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	response, err := client.httpClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("http request: %w", err)
	}
	defer func() { _ = response.Body.Close() }()

	switch response.StatusCode {
	case http.StatusOK:
		var accrualResp model.Accrual
		if err := json.NewDecoder(response.Body).Decode(&accrualResp); err != nil {
			return nil, fmt.Errorf("decode response: %w", err)
		}
		return &accrualResp, nil

	case http.StatusNoContent:
		return (*model.Accrual)(nil), nil

	case http.StatusTooManyRequests:
		return (*model.Accrual)(nil), model.ErrTooManyRequests

	default:
		body, _ := io.ReadAll(response.Body)
		err := fmt.Errorf("unexpected status %d: %s", response.StatusCode, string(body))
		if response.StatusCode >= 500 {
			return nil, fmt.Errorf("%w: %v", retry.ErrRetryable, err)
		}
		return nil, err
	}
}

func (client *Client) isRetryable(err error) bool {
	// 429 не retry — воркер обрабатывает паузой
	if errors.Is(err, model.ErrTooManyRequests) {
		return false
	}
	if errors.Is(err, model.ErrTemporarilyUnavailable) {
		return false
	}
	return retry.IsRetryableNetwork(err)
}

func initBreaker() *gobreaker.CircuitBreaker {
	return gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name:        "accrual",
		MaxRequests: 3,
		Interval:    30 * time.Second,
		Timeout:     30 * time.Second,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			// Открываем breaker при серии подряд ошибок (типичный симптом деградации/падения сервиса).
			return counts.ConsecutiveFailures >= 5
		},
		IsSuccessful: func(err error) bool {
			return err == nil || errors.Is(err, model.ErrTooManyRequests)
		},
	})
}
