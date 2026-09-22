package compute

import (
	"context"
	"time"
)

// Semaphore limits concurrent detect pipelines with a bounded wait.
type Semaphore struct {
	sem  chan struct{}
	wait time.Duration
}

// NewSemaphore creates a concurrency semaphore.
func NewSemaphore(capacity int, wait time.Duration) *Semaphore {
	return &Semaphore{
		sem:  make(chan struct{}, capacity),
		wait: wait,
	}
}

// Acquire tries to take a slot within the wait window.
func (s *Semaphore) Acquire(ctx context.Context) bool {
	select {
	case s.sem <- struct{}{}:
		return true
	case <-time.After(s.wait):
		return false
	case <-ctx.Done():
		return false
	}
}

// Release returns a slot.
func (s *Semaphore) Release() {
	<-s.sem
}
