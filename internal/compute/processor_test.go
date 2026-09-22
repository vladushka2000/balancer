package compute

import (
	"context"
	"testing"
	"time"
)

func newTestProcessor() *Processor {
	store := NewStore(nil, "pii", 24*time.Hour, 1000, nil)
	stats := NewStats()
	sem := NewSemaphore(4, 100*time.Millisecond)
	bucket := NewTokenBucket(1000, 1000)
	return NewProcessor(store, newTestPipeline(), sem, bucket, stats, nil)
}

func TestProcessorMaskDemaskPair(t *testing.T) {
	p := newTestProcessor()
	ctx := context.Background()
	original := "Клиент Иванов Иван Иванович, паспорт 4509 123456"
	id := "pair-1"

	mask, err := p.Process(ctx, original, id)
	if err != nil {
		t.Fatalf("mask failed: %v", err)
	}
	demask, err := p.Process(ctx, mask, id)
	if err != nil {
		t.Fatalf("demask failed: %v", err)
	}
	if demask != original {
		t.Fatalf("expected demask == original, got %q", demask)
	}
}

func TestProcessorIdempotentRetry(t *testing.T) {
	p := newTestProcessor()
	ctx := context.Background()
	original := "паспорт 4509 123456"
	id := "retry-1"

	r1, err := p.Process(ctx, original, id)
	if err != nil {
		t.Fatalf("first failed: %v", err)
	}
	r2, err := p.Process(ctx, original, id)
	if err != nil {
		t.Fatalf("retry failed: %v", err)
	}
	if r1 != r2 {
		t.Fatalf("expected idempotent result, got %q vs %q", r1, r2)
	}
}

func TestProcessorStats(t *testing.T) {
	p := newTestProcessor()
	ctx := context.Background()
	_, err := p.Process(ctx, "паспорт 4509 123456", "stats-1")
	if err != nil {
		t.Fatalf("process failed: %v", err)
	}
	snap := p.stats.Snapshot()
	if snap.RequestsTotal == 0 {
		t.Fatalf("expected requests_total > 0")
	}
}

func TestProcessorDirectionCounts(t *testing.T) {
	p := newTestProcessor()
	ctx := context.Background()
	original := "паспорт 4509 123456"
	id := "dir-1"

	mask, err := p.Process(ctx, original, id)
	if err != nil {
		t.Fatalf("mask failed: %v", err)
	}
	if _, err := p.Process(ctx, original, id); err != nil {
		t.Fatalf("mask retry failed: %v", err)
	}
	if _, err := p.Process(ctx, mask, id); err != nil {
		t.Fatalf("demask failed: %v", err)
	}

	snap := p.stats.Snapshot()
	if snap.MaskOK != 2 {
		t.Fatalf("expected mask_ok=2, got %d", snap.MaskOK)
	}
	if snap.DemaskOK != 1 {
		t.Fatalf("expected demask_ok=1, got %d", snap.DemaskOK)
	}
}

func TestProcessorDemaskNotRateLimited(t *testing.T) {
	store := NewStore(nil, "pii", 24*time.Hour, 1000, nil)
	stats := NewStats()
	sem := NewSemaphore(4, 100*time.Millisecond)
	bucket := NewTokenBucket(1, 1)
	p := NewProcessor(store, newTestPipeline(), sem, bucket, stats, nil)
	ctx := context.Background()
	original := "паспорт 4509 123456"
	id := "demask-norl-1"

	mask, err := p.Process(ctx, original, id)
	if err != nil {
		t.Fatalf("mask failed: %v", err)
	}
	if _, err := p.Process(ctx, "паспорт 4509 123457", "demask-norl-2"); err != ErrRateLimited {
		t.Fatalf("expected new mask to be rate limited, got %v", err)
	}
	if _, err := p.Process(ctx, mask, id); err != nil {
		t.Fatalf("demask must not be rate limited, got %v", err)
	}
	if _, err := p.Process(ctx, original, id); err != nil {
		t.Fatalf("mask retry must not be rate limited, got %v", err)
	}
}

func TestProcessorWithRegistryNilRegistry(t *testing.T) {
	store := NewStore(nil, "pii", 24*time.Hour, 1000, nil)
	stats := NewStats()
	sem := NewSemaphore(4, 100*time.Millisecond)
	bucket := NewTokenBucket(1000, 1000)
	p := NewProcessorWithRegistry(store, newTestPipeline(), sem, bucket, stats, nil, nil, 1500)
	ctx := context.Background()
	if _, err := p.Process(ctx, "паспорт 4509 123456", "reg-nil-1"); err != nil {
		t.Fatalf("process failed: %v", err)
	}
	if int(p.bucket.capacity) != 1000 {
		t.Fatalf("expected bucket unchanged with nil registry, got %v", p.bucket.capacity)
	}
}

func TestProcessorWithRegistryDividesLimit(t *testing.T) {
	store := NewStore(nil, "pii", 24*time.Hour, 1000, nil)
	stats := NewStats()
	sem := NewSemaphore(4, 100*time.Millisecond)
	bucket := NewTokenBucket(1500, 1500)
	reg := NewRegistry(nil, "pii", 10*time.Second, 3*time.Second)
	p := NewProcessorWithRegistry(store, newTestPipeline(), sem, bucket, stats, nil, reg, 1500)
	ctx := context.Background()
	if _, err := p.Process(ctx, "паспорт 4509 123456", "reg-div-1"); err != nil {
		t.Fatalf("process failed: %v", err)
	}
	if int(p.bucket.capacity) != 1500 {
		t.Fatalf("expected bucket 1500 with 1 alive instance, got %v", p.bucket.capacity)
	}
}

func TestAdjustBucketNoRegistry(t *testing.T) {
	store := NewStore(nil, "pii", 24*time.Hour, 1000, nil)
	stats := NewStats()
	sem := NewSemaphore(4, 100*time.Millisecond)
	bucket := NewTokenBucket(100, 100)
	p := NewProcessor(store, newTestPipeline(), sem, bucket, stats, nil)
	p.adjustBucket(context.Background())
	if int(p.bucket.capacity) != 100 {
		t.Fatalf("expected bucket unchanged, got %v", p.bucket.capacity)
	}
}

func TestAdjustBucketZeroGlobalRPS(t *testing.T) {
	store := NewStore(nil, "pii", 24*time.Hour, 1000, nil)
	stats := NewStats()
	sem := NewSemaphore(4, 100*time.Millisecond)
	bucket := NewTokenBucket(100, 100)
	reg := NewRegistry(nil, "pii", 10*time.Second, 3*time.Second)
	p := NewProcessorWithRegistry(store, newTestPipeline(), sem, bucket, stats, nil, reg, 0)
	p.adjustBucket(context.Background())
	if int(p.bucket.capacity) != 100 {
		t.Fatalf("expected bucket unchanged with zero global rps, got %v", p.bucket.capacity)
	}
}

func TestAdjustBucketWithRegistry(t *testing.T) {
	store := NewStore(nil, "pii", 24*time.Hour, 1000, nil)
	stats := NewStats()
	sem := NewSemaphore(4, 100*time.Millisecond)
	bucket := NewTokenBucket(1500, 1500)
	reg := NewRegistry(nil, "pii", 10*time.Second, 3*time.Second)
	p := NewProcessorWithRegistry(store, newTestPipeline(), sem, bucket, stats, nil, reg, 1500)
	p.adjustBucket(context.Background())
	if int(p.bucket.capacity) != 1500 {
		t.Fatalf("expected bucket 1500 with 1 alive, got %v", p.bucket.capacity)
	}
}
