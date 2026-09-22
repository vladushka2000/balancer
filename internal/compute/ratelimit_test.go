package compute

import (
	"context"
	"testing"
	"time"
)

func TestSemaphoreCapacity(t *testing.T) {
	sem := NewSemaphore(2, 50*time.Millisecond)
	ctx := context.Background()
	if !sem.Acquire(ctx) {
		t.Fatalf("expected first acquire to succeed")
	}
	if !sem.Acquire(ctx) {
		t.Fatalf("expected second acquire to succeed")
	}
	if sem.Acquire(ctx) {
		t.Fatalf("expected third acquire to fail")
	}
}

func TestSemaphoreWaitTimeout(t *testing.T) {
	sem := NewSemaphore(1, 100*time.Millisecond)
	ctx := context.Background()
	if !sem.Acquire(ctx) {
		t.Fatalf("expected first acquire to succeed")
	}
	start := time.Now()
	if sem.Acquire(ctx) {
		t.Fatalf("expected second acquire to fail")
	}
	if elapsed := time.Since(start); elapsed < 90*time.Millisecond {
		t.Fatalf("expected wait ~100ms, got %v", elapsed)
	}
}

func TestSemaphoreRelease(t *testing.T) {
	sem := NewSemaphore(1, 50*time.Millisecond)
	ctx := context.Background()
	if !sem.Acquire(ctx) {
		t.Fatalf("expected first acquire to succeed")
	}
	sem.Release()
	if !sem.Acquire(ctx) {
		t.Fatalf("expected acquire after release to succeed")
	}
}
