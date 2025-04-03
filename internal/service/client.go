package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type AccrualResponse struct {
	Order   string  `json:"order"`
	Status  string  `json:"status"`
	Accrual float64 `json:"accrual,omitempty"`
}

type AccrualClient interface {
	GetAccrual(ctx context.Context, orderNumber string) (*AccrualResponse, error)
}

type accrualClient struct {
	baseURL string
	client  *http.Client
}

func (a *accrualClient) GetAccrual(ctx context.Context, orderNumber string) (*AccrualResponse, error) {
	url := fmt.Sprintf("%s/api/orders/%s", a.baseURL, orderNumber)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("accrual system returned status: %d", resp.StatusCode)
	}

	var accrualResp AccrualResponse
	if err := json.NewDecoder(resp.Body).Decode(&accrualResp); err != nil {
		return nil, err
	}

	return &accrualResp, nil
}

func NewAccrualClient(baseURL string) AccrualClient {
	return &accrualClient{
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}
