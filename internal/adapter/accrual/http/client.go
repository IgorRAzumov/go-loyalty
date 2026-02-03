package http

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"loyalty/internal/domain/accrual/model"
	"net/http"
	"strings"
	"time"
)

// Client реализует accrual.AccrualClient через HTTP.
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// NewClient создаёт HTTP-клиент для системы accrual.
func NewClient(baseURL string, timeout time.Duration) *Client {
	return &Client{
		baseURL:    strings.TrimSuffix(baseURL, "/"),
		httpClient: &http.Client{Timeout: timeout},
	}
}

// GetOrderAccrual получает информацию о начислении для заказа.
func (client *Client) GetOrderAccrual(ctx context.Context, orderNumber string) (*model.Accrual, error) {
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
		return nil, nil

	case http.StatusTooManyRequests:
		return nil, model.ErrTooManyRequests

	default:
		body, _ := io.ReadAll(response.Body)
		return nil, fmt.Errorf("unexpected status %d: %s", response.StatusCode, string(body))
	}
}
