package internal_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/ricardovhz/rinha-2025/internal"
	"github.com/stretchr/testify/require"
)

func TestBackendHealth(t *testing.T) {

	client := &http.Client{
		Timeout: time.Second * 10,
	}

	b := internal.NewBackend("http://localhost:8002", client)
	failing, respTime, err := b.GetHealth(context.Background())
	require.NoError(t, err)
	require.False(t, failing)
	require.GreaterOrEqual(t, respTime, int64(0), "Response time should be greater than 0")
}

func TestBackendPayments(t *testing.T) {
	client := &http.Client{
		Timeout: time.Second * 10,
	}

	b := internal.NewBackend("http://localhost:8002", client)
	msg, err := b.ProcessPayment(context.Background(), "4a7901b8-7d26-4e9d-aa19-4dc1c7cf60b3", 100.0, time.Now())
	require.NoError(t, err)
	require.Equal(t, "payment processed successfully", msg)
}
