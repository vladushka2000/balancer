package compute

import (
	"os"
	"runtime"
	"strconv"
	"time"
)

// Config holds engine settings loaded from environment variables.
type Config struct {
	RedisURL      string
	NS            string
	CorrTTL       time.Duration
	CacheMax      int
	MaxConcurrent int
	SemWait       time.Duration
	NERChunkChars int
	NEROverlap    int
	ContextWindow int
	StoreKey      string
}

// DefaultConfig returns engine defaults.
func DefaultConfig() Config {
	return Config{
		RedisURL:      "redis://localhost:6379/0",
		NS:            "pii",
		CorrTTL:       24 * time.Hour,
		CacheMax:      100000,
		MaxConcurrent: runtime.NumCPU(),
		SemWait:       300 * time.Millisecond,
		NERChunkChars: 4000,
		NEROverlap:    200,
		ContextWindow: 200,
	}
}

// LoadConfig reads engine configuration from environment variables.
func LoadConfig() Config {
	cfg := DefaultConfig()
	if v := os.Getenv("REDIS_URL"); v != "" {
		cfg.RedisURL = v
	}
	if v := os.Getenv("PII_NS"); v != "" {
		cfg.NS = v
	}
	if v := os.Getenv("PII_CORR_TTL_SEC"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			cfg.CorrTTL = time.Duration(n) * time.Second
		}
	}
	if v := os.Getenv("PII_CACHE_MAX"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			cfg.CacheMax = n
		}
	}
	if v := os.Getenv("PII_MAX_CONCURRENT"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			cfg.MaxConcurrent = n
		}
	}
	if v := os.Getenv("PII_SEM_WAIT_SEC"); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil && f > 0 {
			cfg.SemWait = time.Duration(f * float64(time.Second))
		}
	}
	if v := os.Getenv("PII_NER_CHUNK_CHARS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			cfg.NERChunkChars = n
		}
	}
	if v := os.Getenv("PII_NER_OVERLAP"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			cfg.NEROverlap = n
		}
	}
	if v := os.Getenv("PII_CONTEXT_WINDOW"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			cfg.ContextWindow = n
		}
	}
	if v := os.Getenv("PII_STORE_KEY"); v != "" {
		cfg.StoreKey = v
	}
	return cfg
}
