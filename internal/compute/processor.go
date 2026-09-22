package compute

import (
	"context"
	"errors"
	"strings"
	"time"
)

// ErrRateLimited is returned when the concurrency semaphore is exhausted.
var ErrRateLimited = errors.New("rate limited")

// ErrStore is returned when the correspondence store write fails.
var ErrStore = errors.New("store error")

// Processor is the engine entry point called by the api package.
type Processor struct {
	store     *Store
	pipeline  *Pipeline
	semaphore *Semaphore
	bucket    *TokenBucket
	stats     *Stats
	repo      *Repo
}

// NewProcessor builds an engine processor.
func NewProcessor(store *Store, pipeline *Pipeline, semaphore *Semaphore, bucket *TokenBucket, stats *Stats, repo *Repo) *Processor {
	return &Processor{
		store:     store,
		pipeline:  pipeline,
		semaphore: semaphore,
		bucket:    bucket,
		stats:     stats,
		repo:      repo,
	}
}

// Process handles a payload: idempotency lookup → detect → mask → store.
func (p *Processor) Process(ctx context.Context, payload, payloadID string) (string, error) {
	start := time.Now()

	if result, dir, found, err := p.store.Lookup(ctx, payloadID, payload); err != nil {
		return "", err
	} else if found {
		p.stats.RecordTokens(nil, elapsedMs(start), dir, tokenCount(payload))
		return result, nil
	}

	if !p.bucket.Acquire() {
		p.stats.Record429()
		return "", ErrRateLimited
	}
	if !p.semaphore.Acquire(ctx) {
		p.stats.Record429()
		return "", ErrRateLimited
	}
	defer p.semaphore.Release()

	masked, types := p.pipeline.Process(payload)
	if err := p.store.Put(ctx, payloadID, payload, masked, types); err != nil {
		return "", ErrStore
	}
	p.stats.RecordTokens(types, elapsedMs(start), DirectionMask, tokenCount(payload))
	return masked, nil
}

func tokenCount(s string) int {
	if s == "" {
		return 0
	}
	return len(strings.Fields(s))
}

func elapsedMs(start time.Time) float64 {
	return float64(time.Since(start).Microseconds()) / 1000.0
}
