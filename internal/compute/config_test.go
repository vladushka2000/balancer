package compute

import (
	"os"
	"testing"
	"time"
)

func TestConfigDefaults(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.NS != "pii" {
		t.Fatalf("expected ns pii, got %q", cfg.NS)
	}
	if cfg.CorrTTL != 24*time.Hour {
		t.Fatalf("expected corr ttl 24h, got %v", cfg.CorrTTL)
	}
	if cfg.CacheMax != 100000 {
		t.Fatalf("expected cache max 100000, got %d", cfg.CacheMax)
	}
	if cfg.NERChunkChars != 4000 {
		t.Fatalf("expected ner chunk 4000, got %d", cfg.NERChunkChars)
	}
	if cfg.NEROverlap != 200 {
		t.Fatalf("expected ner overlap 200, got %d", cfg.NEROverlap)
	}
	if cfg.ContextWindow != 200 {
		t.Fatalf("expected context window 200, got %d", cfg.ContextWindow)
	}
	if cfg.SemWait != 300*time.Millisecond {
		t.Fatalf("expected sem wait 300ms, got %v", cfg.SemWait)
	}
}

func TestConfigFromEnv(t *testing.T) {
	os.Setenv("PII_NS", "testns")
	os.Setenv("PII_CORR_TTL_SEC", "3600")
	os.Setenv("PII_CACHE_MAX", "42")
	os.Setenv("PII_MAX_CONCURRENT", "7")
	os.Setenv("PII_NER_CHUNK_CHARS", "1000")
	os.Setenv("PII_NER_OVERLAP", "50")
	os.Setenv("PII_CONTEXT_WINDOW", "100")
	os.Setenv("PII_SEM_WAIT_SEC", "0.5")
	defer func() {
		os.Unsetenv("PII_NS")
		os.Unsetenv("PII_CORR_TTL_SEC")
		os.Unsetenv("PII_CACHE_MAX")
		os.Unsetenv("PII_MAX_CONCURRENT")
		os.Unsetenv("PII_NER_CHUNK_CHARS")
		os.Unsetenv("PII_NER_OVERLAP")
		os.Unsetenv("PII_CONTEXT_WINDOW")
		os.Unsetenv("PII_SEM_WAIT_SEC")
	}()

	cfg := LoadConfig()
	if cfg.NS != "testns" {
		t.Fatalf("expected ns testns, got %q", cfg.NS)
	}
	if cfg.CorrTTL != time.Hour {
		t.Fatalf("expected corr ttl 1h, got %v", cfg.CorrTTL)
	}
	if cfg.CacheMax != 42 {
		t.Fatalf("expected cache max 42, got %d", cfg.CacheMax)
	}
	if cfg.MaxConcurrent != 7 {
		t.Fatalf("expected max concurrent 7, got %d", cfg.MaxConcurrent)
	}
	if cfg.NERChunkChars != 1000 {
		t.Fatalf("expected ner chunk 1000, got %d", cfg.NERChunkChars)
	}
	if cfg.NEROverlap != 50 {
		t.Fatalf("expected ner overlap 50, got %d", cfg.NEROverlap)
	}
	if cfg.ContextWindow != 100 {
		t.Fatalf("expected context window 100, got %d", cfg.ContextWindow)
	}
	if cfg.SemWait != 500*time.Millisecond {
		t.Fatalf("expected sem wait 500ms, got %v", cfg.SemWait)
	}
}
