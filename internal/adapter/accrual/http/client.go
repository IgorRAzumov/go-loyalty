package http

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"loyalty/internal/domain/accrual/model"
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
	res, err := client.breaker.Execute(func() (any, error) {
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
			return nil, fmt.Errorf("unexpected status %d: %s", response.StatusCode, string(body))
		}
	})
	if err != nil {
		if errors.Is(err, gobreaker.ErrOpenState) || errors.Is(err, gobreaker.ErrTooManyRequests) {
			return nil, fmt.Errorf("%w: %s", model.ErrTemporarilyUnavailable, err.Error())
		}
		return nil, err
	}

	if res == nil {
		return nil, nil
	}
	accrualResp, ok := res.(*model.Accrual)
	if !ok {
		return nil, fmt.Errorf("unexpected breaker response type %T", res)
	}
	return accrualResp, nil
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
