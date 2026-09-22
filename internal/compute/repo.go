package compute

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"

	"pii/internal/models"
)

const repoCacheTTL = 2 * time.Second

type repoCacheEntry struct {
	config models.SystemConfig
	exp    time.Time
}

// Repo manages models.SystemConfig in Redis with a short TTL cache.
type Repo struct {
	rdb   *redis.Client
	ns    string
	mu    sync.Mutex
	cache map[string]repoCacheEntry
}

// NewRepo creates a system config repo.
func NewRepo(rdb *redis.Client, ns string) *Repo {
	return &Repo{
		rdb:   rdb,
		ns:    ns,
		cache: map[string]repoCacheEntry{},
	}
}

// SaveSystem stores a system config and bumps the config epoch.
func (r *Repo) SaveSystem(ctx context.Context, config models.SystemConfig) error {
	key := systemKey(r.ns, config.SystemID)
	raw, err := json.Marshal(config)
	if err != nil {
		return err
	}
	pipe := r.rdb.TxPipeline()
	pipe.Set(ctx, key, raw, 0)
	pipe.SAdd(ctx, systemsSet(r.ns), config.SystemID)
	pipe.Incr(ctx, configEpochKey(r.ns))
	if _, err := pipe.Exec(ctx); err != nil {
		return err
	}
	r.mu.Lock()
	r.cache[config.SystemID] = repoCacheEntry{config: config, exp: time.Now().Add(repoCacheTTL)}
	r.mu.Unlock()
	return nil
}

// GetSystem returns a system config using the cache.
func (r *Repo) GetSystem(ctx context.Context, systemID string) (*models.SystemConfig, error) {
	r.mu.Lock()
	if e, ok := r.cache[systemID]; ok && time.Now().Before(e.exp) {
		cfg := e.config
		r.mu.Unlock()
		return &cfg, nil
	}
	r.mu.Unlock()

	key := systemKey(r.ns, systemID)
	raw, err := r.rdb.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var cfg models.SystemConfig
	if err := json.Unmarshal([]byte(raw), &cfg); err != nil {
		return nil, err
	}
	r.mu.Lock()
	r.cache[systemID] = repoCacheEntry{config: cfg, exp: time.Now().Add(repoCacheTTL)}
	r.mu.Unlock()
	return &cfg, nil
}

// ListSystems returns all registered system configs.
func (r *Repo) ListSystems(ctx context.Context) ([]models.SystemConfig, error) {
	ids, err := r.rdb.SMembers(ctx, systemsSet(r.ns)).Result()
	if err != nil {
		return nil, err
	}
	var out []models.SystemConfig
	for _, id := range ids {
		cfg, err := r.GetSystem(ctx, id)
		if err != nil {
			return nil, err
		}
		if cfg != nil {
			out = append(out, *cfg)
		}
	}
	return out, nil
}

// BumpConfigEpoch increments the config epoch.
func (r *Repo) BumpConfigEpoch(ctx context.Context) error {
	return r.rdb.Incr(ctx, configEpochKey(r.ns)).Err()
}

// GetControlEpoch returns the current config epoch.
func (r *Repo) GetControlEpoch(ctx context.Context) (int64, error) {
	return r.rdb.Get(ctx, configEpochKey(r.ns)).Int64()
}
