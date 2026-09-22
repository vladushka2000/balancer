package api

import (
	"context"
	"testing"
	"time"

	"pii/internal/compute"
	"pii/internal/compute/detect"
	"pii/internal/compute/mask"
)

func newTestDoor(capacity int) *Door {
	ner := detect.NewNERDetector(4000, 200)
	ner.Preload()
	store := compute.NewStore(nil, "pii", 24*time.Hour, 1000, nil)
	stats := compute.NewStats()
	sem := compute.NewSemaphore(4, 100*time.Millisecond)
	pipeline := compute.NewPipeline(
		detect.CreateStructuralRegistry(),
		ner,
		detect.NewContextRule(200),
		mask.NewMasker("partial"),
	)
	proc := compute.NewProcessor(store, pipeline, sem, stats, nil)
	return NewDoor(NewTokenBucket(capacity, capacity), proc)
}

func TestDoorForwardSuccess(t *testing.T) {
	d := newTestDoor(10)
	result, err := d.Process(context.Background(), "паспорт 4509 123456", "door-1")
	if err != nil {
		t.Fatalf("process failed: %v", err)
	}
	if result == "" {
		t.Fatalf("expected non-empty result")
	}
}

func TestDoor429RateLimit(t *testing.T) {
	d := newTestDoor(1)
	if _, err := d.Process(context.Background(), "паспорт 4509 123456", "door-2"); err != nil {
		t.Fatalf("first should succeed: %v", err)
	}
	if _, err := d.Process(context.Background(), "паспорт 4509 123456", "door-3"); err != ErrRateLimited {
		t.Fatalf("expected ErrRateLimited, got %v", err)
	}
}

func TestDoor422InvalidBody(t *testing.T) {
	d := newTestDoor(10)
	if _, err := d.Process(context.Background(), "", "door-4"); err != ErrValidation {
		t.Fatalf("expected ErrValidation, got %v", err)
	}
	if _, err := d.Process(context.Background(), "text", ""); err != ErrValidation {
		t.Fatalf("expected ErrValidation, got %v", err)
	}
}
