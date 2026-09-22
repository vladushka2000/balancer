package api

import (
	"sync"
	"testing"
	"time"
)

func TestTokenBucketCapacity(t *testing.T) {
	tb := NewTokenBucket(3, 3)
	for i := 0; i < 3; i++ {
		if !tb.Acquire() {
			t.Fatalf("expected acquire %d to succeed", i)
		}
	}
	if tb.Acquire() {
		t.Fatalf("expected 4th acquire to fail")
	}
}

func TestTokenBucketRefill(t *testing.T) {
	tb := NewTokenBucket(1, 10)
	if !tb.Acquire() {
		t.Fatalf("expected first acquire to succeed")
	}
	if tb.Acquire() {
		t.Fatalf("expected second acquire to fail")
	}
	time.Sleep(150 * time.Millisecond)
	if !tb.Acquire() {
		t.Fatalf("expected acquire after refill to succeed")
	}
}

func TestTokenBucketConcurrent(t *testing.T) {
	tb := NewTokenBucket(5, 5)
	var wg sync.WaitGroup
	var mu sync.Mutex
	success := 0
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if tb.Acquire() {
				mu.Lock()
				success++
				mu.Unlock()
			}
		}()
	}
	wg.Wait()
	if success != 5 {
		t.Fatalf("expected exactly 5 successes, got %d", success)
	}
}
