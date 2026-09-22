package compute

import (
	"container/list"
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"

	"pii/internal/models"
)

// Store is the correspondence store with in-memory LRU+TTL cache and Redis write-through.
type Store struct {
	rdb    *redis.Client
	ns     string
	ttl    time.Duration
	cache  *lruCache
	encKey []byte
	mu     sync.Mutex
	hits   uint64
	misses uint64
}

// NewStore creates a correspondence store.
func NewStore(rdb *redis.Client, ns string, ttl time.Duration, cacheMax int, encKey []byte) *Store {
	return &Store{
		rdb:    rdb,
		ns:     ns,
		ttl:    ttl,
		cache:  newLRUCache(cacheMax),
		encKey: encKey,
	}
}

// Get returns a correspondence record by payload id.
func (s *Store) Get(ctx context.Context, payloadID string) (*models.CorrRecord, error) {
	if rec, ok := s.cache.get(payloadID); ok {
		s.recordHit()
		return rec, nil
	}
	s.recordMiss()
	if s.rdb == nil {
		return nil, nil
	}
	key := corrKey(s.ns, payloadID)
	raw, err := s.rdb.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	plain, err := decrypt([]byte(raw), s.encKey)
	if err != nil {
		return nil, err
	}
	var rec models.CorrRecord
	if err := json.Unmarshal(plain, &rec); err != nil {
		return nil, err
	}
	s.cache.put(payloadID, &rec)
	return &rec, nil
}

// Put stores a correspondence record (write-through: cache + Redis with TTL).
func (s *Store) Put(ctx context.Context, payloadID, original, mask string, types []string) error {
	rec := &models.CorrRecord{
		Original:  original,
		Mask:      mask,
		Types:     types,
		CreatedTS: float64(time.Now().Unix()),
	}
	plain, err := json.Marshal(rec)
	if err != nil {
		return err
	}
	enc, err := encrypt(plain, s.encKey)
	if err != nil {
		return err
	}
	key := corrKey(s.ns, payloadID)
	if s.rdb != nil {
		if err := s.rdb.Set(ctx, key, enc, s.ttl).Err(); err != nil {
			return err
		}
	}
	s.cache.put(payloadID, rec)
	return nil
}

// Lookup resolves the direction: original→mask or mask→original.
func (s *Store) Lookup(ctx context.Context, payloadID, payload string) (string, Direction, bool, error) {
	rec, err := s.Get(ctx, payloadID)
	if err != nil || rec == nil {
		return "", DirectionMask, false, err
	}
	if payload == rec.Original {
		return rec.Mask, DirectionMask, true, nil
	}
	if payload == rec.Mask {
		return rec.Original, DirectionDemask, true, nil
	}
	return "", DirectionMask, false, nil
}

// Size returns the number of cached records.
func (s *Store) Size() int {
	return s.cache.len()
}

func (s *Store) recordHit() {
	s.mu.Lock()
	s.hits++
	s.mu.Unlock()
}

func (s *Store) recordMiss() {
	s.mu.Lock()
	s.misses++
	s.mu.Unlock()
}

// CacheHitRate returns the in-memory cache hit ratio.
func (s *Store) CacheHitRate() float64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	total := s.hits + s.misses
	if total == 0 {
		return 0
	}
	return float64(s.hits) / float64(total)
}

type lruEntry struct {
	key   string
	value *models.CorrRecord
	exp   time.Time
}

type lruCache struct {
	mu    sync.Mutex
	max   int
	ll    *list.List
	items map[string]*list.Element
}

func newLRUCache(max int) *lruCache {
	if max <= 0 {
		max = 100000
	}
	return &lruCache{
		max:   max,
		ll:    list.New(),
		items: map[string]*list.Element{},
	}
}

func (c *lruCache) get(key string) (*models.CorrRecord, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	el, ok := c.items[key]
	if !ok {
		return nil, false
	}
	entry := el.Value.(*lruEntry)
	if time.Now().After(entry.exp) {
		c.remove(el)
		return nil, false
	}
	c.ll.MoveToFront(el)
	return entry.value, true
}

func (c *lruCache) put(key string, rec *models.CorrRecord) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if el, ok := c.items[key]; ok {
		el.Value.(*lruEntry).value = rec
		el.Value.(*lruEntry).exp = time.Now().Add(24 * time.Hour)
		c.ll.MoveToFront(el)
		return
	}
	entry := &lruEntry{key: key, value: rec, exp: time.Now().Add(24 * time.Hour)}
	el := c.ll.PushFront(entry)
	c.items[key] = el
	if c.ll.Len() > c.max {
		c.remove(c.ll.Back())
	}
}

func (c *lruCache) remove(el *list.Element) {
	entry := el.Value.(*lruEntry)
	delete(c.items, entry.key)
	c.ll.Remove(el)
}

func (c *lruCache) len() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.ll.Len()
}
