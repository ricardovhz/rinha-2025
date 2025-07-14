package internal

import "sync"

type Queue struct {
	items []PaymentsRequest
	l     sync.Mutex
}

func (q *Queue) Enqueue(item PaymentsRequest) {
	q.l.Lock()
	defer q.l.Unlock()
	q.items = append(q.items, item)
}

func (q *Queue) Dequeue() (PaymentsRequest, bool) {
	q.l.Lock()
	defer q.l.Unlock()
	if len(q.items) == 0 {
		return PaymentsRequest{}, false
	}
	item := q.items[0]
	q.items = q.items[1:]
	return item, true
}

func NewQueue() *Queue {
	return &Queue{
		items: make([]PaymentsRequest, 0),
	}
}
