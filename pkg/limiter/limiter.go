package limiter

import (
	"context"
	"errors"
	"sync"
)

type Limiter struct {
	sem     chan struct{}
	counter int
	max     int
	mu      sync.Mutex
}

func NewLimiter(max int) *Limiter {
	l := &Limiter{
		sem: make(chan struct{}, max),
		max: max,
	}
	// Заполняем семафор
	for i := 0; i < max; i++ {
		l.sem <- struct{}{}
	}
	return l
}

func (l *Limiter) Acquire(ctx context.Context) error {
	select {
	case <-l.sem:
		l.mu.Lock()
		l.counter++
		l.mu.Unlock()
		return nil
	case <-ctx.Done():
		return ctx.Err()
	default:
		return ErrLimitExceeded
	}
}

func (l *Limiter) Release() {
	l.mu.Lock()
	if l.counter > 0 {
		l.sem <- struct{}{}
		l.counter--
	}
	l.mu.Unlock()
}

var ErrLimitExceeded = errors.New("too many requests")
