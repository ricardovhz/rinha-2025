package internal

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"
)

type HealthBackend interface {
	GetHealth(context.Context) (bool, int64, error)
}

type PaymentBackend interface {
	ProcessPayment(context.Context, string, float64, time.Time) (string, error)
}

type Backend interface {
	HealthBackend
	PaymentBackend
}

type HealthResponse struct {
	Failing         bool  `json:"failing"`
	MinResponseTime int64 `json:"minResponseTime"`
}

type PaymentsRequest struct {
	CorrelationId string  `json:"correlationId"`
	Amount        float64 `json:"amount"`
	RequestedAt   string  `json:"requestedAt"`
}

type PaymentsResponse struct {
	Message string `json:"message"`
}

type backend struct {
	host         string
	client       *http.Client
	healthPath   string
	paymentsPath string
}

func (b *backend) GetHealth(ctx context.Context) (bool, int64, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", b.healthPath, nil)
	if err != nil {
		return false, -1, err
	}
	resp, err := b.client.Do(req)
	if err != nil {
		return false, -1, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return false, -1, nil
	}
	var healthResp HealthResponse
	if err := json.NewDecoder(resp.Body).Decode(&healthResp); err != nil {
		return false, -1, err
	}
	return healthResp.Failing, healthResp.MinResponseTime, nil
}

func (b *backend) ProcessPayment(ctx context.Context, correlationId string, amount float64, timestamp time.Time) (string, error) {
	jsonReq := &PaymentsRequest{
		CorrelationId: correlationId,
		Amount:        amount,
		RequestedAt:   timestamp.Format(time.RFC3339Nano),
	}
	jsonReqBytes, err := json.Marshal(jsonReq)
	if err != nil {
		return "", err
	}
	request, err := http.NewRequestWithContext(ctx, "POST", b.paymentsPath, bytes.NewReader(jsonReqBytes))
	if err != nil {
		return "", err
	}
	request.Header.Set("Content-Type", "application/json")

	var resp PaymentsResponse
	response, err := b.client.Do(request)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		if response.StatusCode == http.StatusUnprocessableEntity { // 422
			log.Printf("Received 422 Unprocessable Entity for request: %s", correlationId)
			return "", nil // drop the request if it's 422
		}
		return "", fmt.Errorf("error %d", response.StatusCode) // Handle non-200 responses as needed
	}
	decoder := json.NewDecoder(response.Body)
	if err := decoder.Decode(&resp); err != nil {
		return "", err
	}

	return resp.Message, nil
}

func NewBackend(host string, client *http.Client) Backend {
	healthPath, _ := url.JoinPath(host, "/payments/service-health")
	paymentsPath, _ := url.JoinPath(host, "/payments")
	return &backend{
		host:         host,
		client:       client,
		healthPath:   healthPath,
		paymentsPath: paymentsPath,
	}
}
