package compute

import (
	"context"
	"errors"
	"strings"
	"sync"
	"time"

	"pii/internal/models"
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
	registry  *Registry
	globalRPS int
	bucketMu  sync.Mutex
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

// NewProcessorWithRegistry builds an engine processor with a registry for
// dividing the global rate limit across live instances.
func NewProcessorWithRegistry(store *Store, pipeline *Pipeline, semaphore *Semaphore, bucket *TokenBucket, stats *Stats, repo *Repo, registry *Registry, globalRPS int) *Processor {
	return &Processor{
		store:     store,
		pipeline:  pipeline,
		semaphore: semaphore,
		bucket:    bucket,
		stats:     stats,
		repo:      repo,
		registry:  registry,
		globalRPS: globalRPS,
	}
}

// Process handles a payload: idempotency lookup → detect → mask → store.
func (p *Processor) Process(ctx context.Context, payload, payloadID string) (string, error) {
	start := time.Now()
	systemID := systemIDFromPayloadID(payloadID)
	cfg := p.systemConfig(ctx, systemID)

	if !cfg.Enabled {
		return payload, nil
	}

	if result, dir, found, err := p.store.Lookup(ctx, payloadID, payload); err != nil {
		return "", err
	} else if found {
		if dir == DirectionDemask && !cfg.DemaskEnabled {
			return payload, nil
		}
		p.stats.RecordTokens(nil, elapsedMs(start), dir, tokenCount(payload))
		return result, nil
	}

	p.adjustBucket(ctx)
	p.bucketMu.Lock()
	ok := p.bucket.Acquire()
	p.bucketMu.Unlock()
	if !ok {
		p.stats.Record429()
		return "", ErrRateLimited
	}
	if !p.semaphore.Acquire(ctx) {
		p.stats.Record429()
		return "", ErrRateLimited
	}
	defer p.semaphore.Release()

	masked, types := p.pipeline.ProcessFiltered(payload, allowedTypes(cfg))
	if err := p.store.Put(ctx, payloadID, payload, masked, types); err != nil {
		return "", ErrStore
	}
	p.stats.RecordTokens(types, elapsedMs(start), DirectionMask, tokenCount(payload))
	return masked, nil
}

// systemConfig returns the per-system policy, defaulting to allow-all.
func (p *Processor) systemConfig(ctx context.Context, systemID string) models.SystemConfig {
	if p.repo == nil {
		return models.SystemConfig{SystemID: systemID, Enabled: true, DemaskEnabled: true}
	}
	cfg, err := p.repo.GetSystem(ctx, systemID)
	if err != nil || cfg == nil {
		return models.SystemConfig{SystemID: systemID, Enabled: true, DemaskEnabled: true}
	}
	return *cfg
}

// allowedTypes converts the config's comma-separated type list into a set.
func allowedTypes(c models.SystemConfig) map[string]bool {
	if c.Types == "" {
		return nil
	}
	set := map[string]bool{}
	for _, t := range strings.Split(c.Types, ",") {
		t = strings.TrimSpace(t)
		if t != "" {
			set[t] = true
		}
	}
	return set
}

// systemIDFromPayloadID derives the system id from the payload id prefix.
func systemIDFromPayloadID(payloadID string) string {
	for _, sep := range []string{":", "-", "_"} {
		if i := strings.Index(payloadID, sep); i > 0 {
			return payloadID[:i]
		}
	}
	return "default"
}

// adjustBucket resizes the token bucket to globalRPS / AliveCount.
func (p *Processor) adjustBucket(ctx context.Context) {
	if p.registry == nil || p.globalRPS <= 0 {
		return
	}
	n, err := p.registry.AliveCount(ctx)
	if err != nil || n < 1 {
		n = 1
	}
	perInstance := p.globalRPS / n
	if perInstance < 1 {
		perInstance = 1
	}
	p.bucketMu.Lock()
	if int(p.bucket.capacity) != perInstance {
		p.bucket = NewTokenBucket(perInstance, perInstance)
	}
	p.bucketMu.Unlock()
}

func tokenCount(s string) int {
	if s == "" {
		return 0
	}
	count := 0
	inWord := false
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case ' ', '\t', '\n', '\r':
			inWord = false
		default:
			if !inWord {
				count++
				inWord = true
			}
		}
	}
	return count
}

func elapsedMs(start time.Time) float64 {
	return float64(time.Since(start).Microseconds()) / 1000.0
}
