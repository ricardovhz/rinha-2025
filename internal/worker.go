package internal

import (
	"context"
	"log"
	"sync"
	"time"
)

type WorkersPool struct {
	backend PaymentBackend
	q       *Queue
	wg      sync.WaitGroup
	c       chan PaymentsRequest
}

func (p *WorkersPool) start() {
	go func() {
		for {
			i, ok := p.q.Dequeue()
			if !ok {
				continue
			}
			p.c <- i
		}
	}()
}

func (p *WorkersPool) worker(id int) {
	defer p.wg.Done()
	for req := range p.c {
		var err error
		_, err = p.backend.ProcessPayment(context.Background(), req.CorrelationId, req.Amount, time.Now().UTC())
		if err != nil {
			log.Printf("Error processing payment: %v\n", err)
			p.q.Enqueue(req) // Re-enqueue the request for retry
		}
	}
}

func (p *WorkersPool) Stop() {
	close(p.c)
	p.wg.Wait()
}

func NewWorkersPool(backend PaymentBackend, q *Queue, numberWorkers int) *WorkersPool {
	pool := &WorkersPool{
		backend: backend,
		q:       q,
		c:       make(chan PaymentsRequest, numberWorkers),
	}
	for i := 0; i < numberWorkers; i++ {
		pool.wg.Add(1)
		go pool.worker(i)
	}
	pool.start()
	return pool
}
