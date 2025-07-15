package internal

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type BackendSelector struct {
	ctx                 context.Context
	selected            int
	defaultBackend      Backend
	fallbackBackend     Backend
	mtx                 sync.RWMutex
	onProcessedCallback func(int, float64, time.Time)
}

func (s *BackendSelector) ProcessPayment(ctx context.Context, correlationId string, amount float64, requestedAt time.Time) (string, error) {
	// call the selected backend's ProcessPayment method
	be, index := s.GetBackend()

	res, err := be.ProcessPayment(ctx, correlationId, amount, requestedAt)
	if err != nil {
		if index%2 == 0 {
			be = s.fallbackBackend
			res, err = be.ProcessPayment(ctx, correlationId, amount, requestedAt)
			if err != nil {
				return "", fmt.Errorf("failed to process payment: %w", err)
			} else {
				index = 1 // switch to fallback backend
			}
		}
	}
	s.onProcessedCallback(index, amount, requestedAt)
	return res, nil
}

func (s *BackendSelector) switchBackend() {
	s.mtx.Lock()
	fmt.Printf("Switching backend from %d to %d due to health check failure or high response time\n", s.selected%2, (s.selected+1)%2)
	s.selected++
	s.mtx.Unlock()
}

func (s *BackendSelector) GetBackend() (PaymentBackend, int) {
	s.mtx.RLock()
	defer s.mtx.RUnlock()
	idx := s.selected % 2
	if idx == 0 {
		return s.defaultBackend.(PaymentBackend), idx
	}
	return s.fallbackBackend.(PaymentBackend), idx
}

func (s *BackendSelector) isFallback() bool {
	return s.selected%2 == 1
}

func (s *BackendSelector) verifyBackends(ctx context.Context) {
	s.mtx.RLock()
	var backend HealthBackend = s.defaultBackend
	isFallback := s.isFallback()
	failing, minResponseTime, err := backend.GetHealth(ctx)
	s.mtx.RUnlock()
	if err != nil || failing || minResponseTime > 1000 {
		fmt.Printf("Health check failed or response time too high: %v, %v, %dms\n", err, failing, minResponseTime)
		if isFallback {
			// is it's in fallback mode, we don't switch backends
			return
		}
		s.switchBackend()
	} else if isFallback {
		fmt.Printf("default backend ok: %v, %v, %dms\n", err, failing, minResponseTime)
		s.switchBackend()
	}
}

func (s *BackendSelector) Start() {
	go func() {
		t := time.NewTicker(5 * time.Second)
		defer t.Stop()

		timer := time.NewTimer(10 * time.Second)
		for {
			select {
			case <-t.C:
				s.verifyBackends(s.ctx)
			case <-timer.C:
			}
			timer.Reset(10 * time.Second)
		}
	}()
}

func NewBackendSelector(ctx context.Context, defaultBackend, fallbackBackend Backend, onProcessedCallback func(int, float64, time.Time)) *BackendSelector {
	return &BackendSelector{
		ctx:                 ctx,
		selected:            0,
		defaultBackend:      defaultBackend,
		fallbackBackend:     fallbackBackend,
		mtx:                 sync.RWMutex{},
		onProcessedCallback: onProcessedCallback,
	}
}
