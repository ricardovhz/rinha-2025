package internal

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

type BackendSummary struct {
	TotalRequests int64 `json:"totalRequests"`
	TotalAmount   int64 `json:"totalAmount"`
}

type SummaryModel struct {
	DefaultBackend  *BackendSummary `json:"default"`
	FallbackBackend *BackendSummary `json:"fallback"`
}

type SummaryService interface {
	GetSummary(ctx context.Context, from, to time.Time) (*SummaryModel, error)
	AddPayment(ctx context.Context, backend int, amount int64, requestedAt time.Time) error
}

type mysqlSummaryService struct {
	// db *sql.DB
	client         *http.Client
	addPaymentsUrl string
	summaryUrl     string
}

func (s *mysqlSummaryService) GetSummary(ctx context.Context, from, to time.Time) (*SummaryModel, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf(s.summaryUrl+"?from=%s&to=%s",
		from.Format(time.RFC3339Nano), to.Format(time.RFC3339Nano)), nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("could not request summary: %v", err)
	}
	var respBody SummaryModel
	if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
		return nil, fmt.Errorf("could not decode summary response: %v", err)
	}

	return &respBody, nil
}

func (s *mysqlSummaryService) AddPayment(ctx context.Context, backend int, amount int64, requestedAt time.Time) error {
	saveRequestBody := &SavePaymentRequest{
		Backend:     backend,
		Amount:      amount,
		RequestedAt: requestedAt.Format(time.RFC3339Nano),
	}
	body, _ := json.Marshal(saveRequestBody)
	req, err := http.NewRequestWithContext(ctx, "POST", s.addPaymentsUrl, bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	b, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer b.Body.Close()
	if b.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to add payment: %s", b.Status)
	}
	return nil
}

func NewSummaryService(host string) SummaryService {
	addPaymentsUrl, _ := url.JoinPath(host, "/payments")
	summaryUrl, _ := url.JoinPath(host, "/payments-summary")

	return &mysqlSummaryService{
		client:         http.DefaultClient,
		addPaymentsUrl: addPaymentsUrl,
		summaryUrl:     summaryUrl}
}
