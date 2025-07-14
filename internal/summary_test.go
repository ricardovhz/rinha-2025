package internal_test

import (
	"log"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/ricardovhz/rinha-2025/internal"
	"github.com/stretchr/testify/require"
)

func Test(t *testing.T) {
	service := internal.NewSummaryService("http://localhost:3000")
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
