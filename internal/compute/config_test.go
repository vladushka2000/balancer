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
	if cfg.GlobalRPS != 1500 {
		t.Fatalf("expected global rps 1500, got %d", cfg.GlobalRPS)
	}
	if cfg.AliveTTL != 10*time.Second {
		t.Fatalf("expected alive ttl 10s, got %v", cfg.AliveTTL)
	}
	if cfg.HeartbeatInterval != 3*time.Second {
		t.Fatalf("expected heartbeat interval 3s, got %v", cfg.HeartbeatInterval)
	}
	if cfg.StatsPublishEvery != 2*time.Second {
		t.Fatalf("expected stats publish 2s, got %v", cfg.StatsPublishEvery)
	}
	if cfg.StatsTTL != 10*time.Second {
		t.Fatalf("expected stats ttl 10s, got %v", cfg.StatsTTL)
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
	os.Setenv("PII_GLOBAL_RPS", "2000")
	os.Setenv("PII_ALIVE_TTL_SEC", "20")
	os.Setenv("PII_HEARTBEAT_SEC", "5")
	os.Setenv("PII_STATS_PUBLISH_SEC", "4")
	os.Setenv("PII_STATS_TTL_SEC", "30")
	defer func() {
		os.Unsetenv("PII_NS")
		os.Unsetenv("PII_CORR_TTL_SEC")
		os.Unsetenv("PII_CACHE_MAX")
		os.Unsetenv("PII_MAX_CONCURRENT")
		os.Unsetenv("PII_NER_CHUNK_CHARS")
		os.Unsetenv("PII_NER_OVERLAP")
		os.Unsetenv("PII_CONTEXT_WINDOW")
		os.Unsetenv("PII_SEM_WAIT_SEC")
		os.Unsetenv("PII_GLOBAL_RPS")
		os.Unsetenv("PII_ALIVE_TTL_SEC")
		os.Unsetenv("PII_HEARTBEAT_SEC")
		os.Unsetenv("PII_STATS_PUBLISH_SEC")
		os.Unsetenv("PII_STATS_TTL_SEC")
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
	if cfg.GlobalRPS != 2000 {
		t.Fatalf("expected global rps 2000, got %d", cfg.GlobalRPS)
	}
	if cfg.AliveTTL != 20*time.Second {
		t.Fatalf("expected alive ttl 20s, got %v", cfg.AliveTTL)
	}
	if cfg.HeartbeatInterval != 5*time.Second {
		t.Fatalf("expected heartbeat interval 5s, got %v", cfg.HeartbeatInterval)
	}
	if cfg.StatsPublishEvery != 4*time.Second {
		t.Fatalf("expected stats publish 4s, got %v", cfg.StatsPublishEvery)
	}
	if cfg.StatsTTL != 30*time.Second {
		t.Fatalf("expected stats ttl 30s, got %v", cfg.StatsTTL)
	}
}
