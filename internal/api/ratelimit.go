package api

import (
	"sync"
	"time"
)

// TokenBucket is an in-memory rate limiter.
type TokenBucket struct {
	mu           sync.Mutex
	capacity     float64
	tokens       float64
	refillPerSec float64
	lastRefill   time.Time
}

// NewTokenBucket creates a token bucket.
func NewTokenBucket(capacity int, refillPerSec int) *TokenBucket {
	return &TokenBucket{
		capacity:     float64(capacity),
		tokens:       float64(capacity),
		refillPerSec: float64(refillPerSec),
		lastRefill:   time.Now(),
	}
}

// Acquire tries to take one token.
func (tb *TokenBucket) Acquire() bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(tb.lastRefill).Seconds()
	if elapsed < 0 {
		elapsed = 0
	}
	tb.tokens += elapsed * tb.refillPerSec
	if tb.tokens > tb.capacity {
		tb.tokens = tb.capacity
	}
	tb.lastRefill = now

	if tb.tokens >= 1 {
		tb.tokens--
		return true
	}
	return false
}
