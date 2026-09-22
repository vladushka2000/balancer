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
	return NewProcessor(store, newTestPipeline(), sem, stats, nil)
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
