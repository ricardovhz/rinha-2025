package internal_test

import (
	"log"
	"testing"
	"time"

	"github.com/gocql/gocql"
	_ "github.com/mattn/go-sqlite3"
	"github.com/ricardovhz/rinha-2025/internal"
	"github.com/stretchr/testify/require"
)

func setupDatabase(t *testing.T) {
	// s, err := sql.Open("sqlite3", ":memory:")
	// if err != nil {
	// 	t.Fail()
	// }
	cluster := gocql.NewCluster("localhost:45991")
	exec, err := gocql.NewSingleHostQueryExecutor(cluster)
	// session, err := cluster.CreateSession()
	if err != nil {
		t.Fail()
	}
	defer exec.Close()
	err = exec.Exec(`CREATE KEYSPACE IF NOT EXISTS paymentsks WITH REPLICATION = {'class': 'NetworkTopologyStrategy', 'replication_factor': 1}`)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}
	err = exec.Exec(`DELETE from paymentsks.payments WHERE backend in (0,1)`)

	err = exec.Exec(`CREATE TABLE IF NOT EXISTS paymentsks.payments (requested_at timestamp,id uuid,backend int,amount decimal,PRIMARY KEY (backend, requested_at, id));`)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// _, err = s.Exec(`CREATE TABLE payments_summary(
	// backend INTEGET NOT NULL,
	// amount REAL NOT NULL,
	// requested_at DATETIME NOT NULL
	// )`)
	// if err != nil {
	// 	t.Fatalf("Failed to create table: %v", err)
	// }
	// return s
}
func Test(t *testing.T) {

	setupDatabase(t)
	service := internal.NewSummaryService("localhost:45991")
	err := service.AddPayment(t.Context(), 0, 100.0, time.Now().UTC())
	require.NoError(t, err)

	sum, err := service.GetSummary(t.Context(), time.Now().Add(-24*time.Hour).UTC(), time.Now().UTC())
	require.NoError(t, err)
	require.NotNil(t, sum)
	require.Equal(t, int64(1), sum.DefaultBackend.TotalRequests)
	require.Equal(t, int64(100), sum.DefaultBackend.TotalAmount)
	require.Equal(t, int64(0), sum.FallbackBackend.TotalRequests)
	require.Equal(t, int64(0), sum.FallbackBackend.TotalAmount)

	service.AddPayment(t.Context(), 1, 199, time.Now().UTC())
	sum, err = service.GetSummary(t.Context(), time.Now().Add(-24*time.Hour), time.Now())
	require.NoError(t, err)
	require.NotNil(t, sum)
	require.Equal(t, int64(1), sum.DefaultBackend.TotalRequests)
	require.Equal(t, int64(100), sum.DefaultBackend.TotalAmount)
	require.Equal(t, int64(1), sum.FallbackBackend.TotalRequests)
	require.Equal(t, int64(199), sum.FallbackBackend.TotalAmount)
}

func TestTruncate(t *testing.T) {
	ts := time.Now().UTC()
	log.Printf("amount: %d", int64((19.9 * 100.0)))
	log.Printf("Truncated time: %s", ts.Truncate(time.Second).Format(time.RFC3339Nano))
	log.Printf("Truncated time: %s", ts.Truncate(time.Millisecond).Format(time.RFC3339Nano))
}
