package compute

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// Registry manages alive-heartbeats of instances in Redis and computes the
// number of live instances for dividing the global rate limit.
type Registry struct {
	rdb        *redis.Client
	ns         string
	instanceID string
	ttl        time.Duration
	interval   time.Duration
	mu         sync.Mutex
	aliveCache int
	aliveExp   time.Time
	cancel     context.CancelFunc
}

// NewRegistry creates an alive registry with a random instance id.
func NewRegistry(rdb *redis.Client, ns string, ttl, interval time.Duration) *Registry {
	return &Registry{
		rdb:        rdb,
		ns:         ns,
		instanceID: newInstanceID(),
		ttl:        ttl,
		interval:   interval,
		aliveCache: -1,
	}
}

// newInstanceID generates a random 16-byte hex instance id.
func newInstanceID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "unknown"
	}
	return hex.EncodeToString(b)
}

// InstanceID returns the registry's instance id.
func (r *Registry) InstanceID() string {
	return r.instanceID
}

// Start launches the background heartbeat goroutine.
func (r *Registry) Start(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	r.cancel = cancel
	go func() {
		ticker := time.NewTicker(r.interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := r.heartbeat(ctx); err != nil {
					slog.Error("heartbeat failed", "error", err)
				}
			}
		}
	}()
}

// Stop deletes the alive key and stops the heartbeat goroutine.
func (r *Registry) Stop(ctx context.Context) {
	if r.cancel != nil {
		r.cancel()
	}
	if r.rdb != nil {
		if err := r.rdb.Del(ctx, aliveKey(r.ns, r.instanceID)).Err(); err != nil {
			slog.Error("alive key delete failed", "error", err)
		}
	}
}

// AliveCount returns the number of live instances, cached for ttl/2.
func (r *Registry) AliveCount(ctx context.Context) (int, error) {
	r.mu.Lock()
	if r.aliveCache >= 0 && time.Now().Before(r.aliveExp) {
		n := r.aliveCache
		r.mu.Unlock()
		return n, nil
	}
	r.mu.Unlock()

	if r.rdb == nil {
		return 1, nil
	}
	keys, err := r.rdb.Keys(ctx, aliveSet(r.ns)+":*").Result()
	if err != nil {
		return 0, err
	}
	n := len(keys)
	r.mu.Lock()
	r.aliveCache = n
	r.aliveExp = time.Now().Add(r.ttl / 2)
	r.mu.Unlock()
	return n, nil
}

// heartbeat writes the alive key with a TTL.
func (r *Registry) heartbeat(ctx context.Context) error {
	if r.rdb == nil {
		return nil
	}
	return r.rdb.Set(ctx, aliveKey(r.ns, r.instanceID), time.Now().Unix(), r.ttl).Err()
}
