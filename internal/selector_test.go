package internal_test

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/ricardovhz/rinha-2025/internal"
	"github.com/stretchr/testify/require"
)

type MockBackend struct {
	id           int
	failing      bool
	responseTime int64
}

func (m *MockBackend) GetHealth(ctx context.Context) (bool, int64, error) {
	return m.failing, m.responseTime, nil
}
func (m *MockBackend) ProcessPayment(ctx context.Context, correlationId string, amount float64, requestedAt time.Time) (string, error) {
	return "", nil
}

func TestSelector(t *testing.T) {
	m1 := &MockBackend{id: 1, failing: false, responseTime: 500}
	m2 := &MockBackend{id: 2, failing: false, responseTime: 700}
	// var expectedBackend int

	selector := internal.NewBackendSelector(context.Background(), m1, m2, func(backend int, amount float64, requestedAt time.Time) {
		// expectedBackend = backend
		fmt.Printf("Payment processed on backend %d with amount %.2f\n", backend, amount)
	})
	selector.Start()

	q := make(chan int)
	wg := sync.WaitGroup{}
	wg.Add(2)

	for i := range 3 {
		go func() {
			defer wg.Done()
			for {
				select {
				case <-q:
					return
				default:
				}
				be, _ := selector.GetBackend()
				fmt.Printf("[%d] selected backend: %d\n", i, be.(*MockBackend).id)
				time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond) // Simulate periodic backend checks
			}
		}()
	}

	be, _ := selector.GetBackend()
	require.Equal(t, 1, be.(*MockBackend).id)

	m1.failing = true
	time.Sleep(6 * time.Second) // Wait for the selector to switch backends
	be, _ = selector.GetBackend()
	require.Equal(t, 2, be.(*MockBackend).id)

	m1.failing = false
	m2.responseTime = 1200      // Simulate high response time
	time.Sleep(6 * time.Second) // Wait for the selector to check health again
	be, _ = selector.GetBackend()
	require.Equal(t, 1, be.(*MockBackend).id)

	close(q)
	wg.Wait()
}
