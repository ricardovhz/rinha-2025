package internal

import (
	"context"
	"fmt"
	"time"

	"github.com/gocql/gocql"
	"github.com/google/uuid"
	"gopkg.in/inf.v0"
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
	session *gocql.Session
}

// func (s *mysqlSummaryService) Get_Summary(ctx context.Context, from, to time.Time) (*SummaryModel, error) {
// 	query := `
// 		SELECT
// 			backend,
// 			COUNT(*) AS total_requests,
// 			ROUND(SUM(amount),2) AS total_amount
// 		FROM payments_summary
// 		WHERE requested_at BETWEEN ? AND ?
// 		GROUP BY backend
// 	`

// 	summaries := map[int]*BackendSummary{
// 		0: {TotalRequests: 0, TotalAmount: 0.0},
// 		1: {TotalRequests: 0, TotalAmount: 0.0},
// 	}

// 	var backendNumber int
// 	row, err := s.db.QueryContext(ctx, query, from, to)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer row.Close()

// 	for row.Next() {
// 		summary := &BackendSummary{}
// 		err = row.Scan(&backendNumber, &summary.TotalRequests, &summary.TotalAmount)
// 		if err != nil {
// 			return nil, err
// 		}
// 		summaries[backendNumber] = summary
// 	}

// 	return &SummaryModel{
// 		DefaultBackend:  summaries[0],
// 		FallbackBackend: summaries[1],
// 	}, nil
// }

// func (s *mysqlSummaryService) Add_Payment(ctx context.Context, backend int, amount float64, requestedAt time.Time) error {
// 	query := `
// 		INSERT INTO payments_summary (backend, amount, requested_at)
// 		VALUES (?, ?, ?)
// 	`
// 	_, err := s.db.ExecContext(ctx, query, backend, amount, requestedAt.Truncate(time.Second))
// 	return err
// }

func (s *mysqlSummaryService) fillSummary(ctx context.Context, id int, from, to time.Time) *BackendSummary {
	q := s.session.Query(`SELECT backend,amount FROM paymentsks.payments WHERE backend=? and requested_at >= ? AND requested_at <= ?`, id, from, to).
		WithContext(ctx).
		PageSize(1000)
	// p, err := s.session.Prepare(ctx, `SELECT backend,amount FROM paymentsks.payments WHERE requested_at BETWEEN ? AND ?`)
	it := q.Iter()
	defer it.Close()
	resultSummary := &BackendSummary{TotalRequests: 0, TotalAmount: 0.0}
	var backend int
	var amount *inf.Dec
	totalAmount := inf.NewDec(0, 2)
	for it.Scan(&backend, &amount) {
		// fmt.Printf("Backend: %d, Amount: %.2f\n", backend, amount)
		resultSummary.TotalRequests++
		totalAmount = totalAmount.Add(totalAmount, amount)
	}
	err := it.Scanner().Err()
	if err != nil {
		fmt.Printf("Error scanning results: %v\n", err)
	}
	unscaled, _ := totalAmount.Unscaled()
	resultSummary.TotalAmount = unscaled // Convert to float64 and scale to 2 decimal places

	return resultSummary
}

func (s *mysqlSummaryService) GetSummary(ctx context.Context, from, to time.Time) (*SummaryModel, error) {
	result := make(map[int]*BackendSummary)
	result[0] = s.fillSummary(ctx, 0, from, to)
	result[1] = s.fillSummary(ctx, 1, from, to)

	return &SummaryModel{
		DefaultBackend:  result[0],
		FallbackBackend: result[1],
	}, nil
}

func (s *mysqlSummaryService) AddPayment(ctx context.Context, backend int, amount int64, requestedAt time.Time) error {
	q := s.session.Bind(`INSERT INTO paymentsks.payments (backend, id, requested_at, amount) VALUES (?,?,?,?)`, func(q *gocql.QueryInfo) ([]interface{}, error) {
		return []any{backend, uuid.New().String(), requestedAt.Truncate(time.Second), inf.NewDec(amount, 2)}, nil
	})
	err := q.Exec()
	// p, err := s.session.Prepare(ctx, `INSERT INTO paymentsks.payments (backend, requested_at, amount) VALUES (?,?,?)`)
	if err != nil {
		return err
	}
	return nil
}

func NewSummaryService(host string) SummaryService {
	cluster := gocql.NewCluster(host)
	session, err := cluster.CreateSession()
	// cfg := scylla.DefaultSessionConfig("paymentsks", host)
	// session, err := scylla.NewSession(context.Background(), cfg)
	if err != nil {
		panic(err)
	}

	return &mysqlSummaryService{session: session}
}
