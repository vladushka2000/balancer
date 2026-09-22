package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"

	"pii/internal/api"
	"pii/internal/compute"
	"pii/internal/compute/detect"
	"pii/internal/compute/mask"
	"pii/internal/models"
)

type serverConfig struct {
	appHost    string
	appPort    int
	computeCfg compute.Config
	storeKey   string
}

func loadServerConfig() serverConfig {
	cfg := serverConfig{
		appHost:    "0.0.0.0",
		appPort:    8080,
		computeCfg: compute.LoadConfig(),
	}
	if v := os.Getenv("PII_APP_HOST"); v != "" {
		cfg.appHost = v
	}
	if v := os.Getenv("PII_APP_PORT"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			cfg.appPort = n
		}
	}
	cfg.storeKey = os.Getenv("PII_STORE_KEY")
	return cfg
}

func main() {
	cfg := loadServerConfig()
	logger := compute.InitLogging()
	slog.SetDefault(logger)

	rdb := redis.NewClient(&redis.Options{
		Addr:         redisAddr(cfg.computeCfg.RedisURL),
		PoolSize:     256,
		MinIdleConns: 16,
	})
	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		logger.Error("redis unavailable", "error", err)
	}

	ner := detect.NewNERDetector(cfg.computeCfg.NERChunkChars, cfg.computeCfg.NEROverlap)
	ner.Preload()

	store := compute.NewStore(rdb, cfg.computeCfg.NS, cfg.computeCfg.CorrTTL, cfg.computeCfg.CacheMax, []byte(cfg.storeKey))
	repo := compute.NewRepo(rdb, cfg.computeCfg.NS)
	stats := compute.NewStats()
	sem := compute.NewSemaphore(cfg.computeCfg.MaxConcurrent, cfg.computeCfg.SemWait)
	pipeline := compute.NewPipeline(
		detect.CreateStructuralRegistry(),
		ner,
		detect.NewContextRule(cfg.computeCfg.ContextWindow),
		mask.NewMasker("partial"),
	)
	registry := compute.NewRegistry(rdb, cfg.computeCfg.NS, cfg.computeCfg.AliveTTL, cfg.computeCfg.HeartbeatInterval)
	registry.Start(ctx)
	defer registry.Stop(ctx)
	proc := compute.NewProcessorWithRegistry(store, pipeline, sem, compute.NewTokenBucket(cfg.computeCfg.GlobalRPS, cfg.computeCfg.GlobalRPS), stats, repo, registry, cfg.computeCfg.GlobalRPS)
	door := api.NewDoor(proc)

	go publishStatsLoop(ctx, repo, registry, stats, cfg.computeCfg.StatsPublishEvery, cfg.computeCfg.StatsTTL)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /process", processHandler(door))
	mux.HandleFunc("GET /app/health", healthHandler(rdb))
	mux.HandleFunc("GET /health", healthHandler(rdb))
	mux.HandleFunc("GET /stats", statsHandler(stats, store, repo))
	mux.HandleFunc("POST /systems", saveSystemHandler(repo))
	mux.HandleFunc("GET /systems", listSystemsHandler(repo))
	mux.HandleFunc("POST /clear", clearHandler(repo))

	addr := cfg.appHost + ":" + strconv.Itoa(cfg.appPort)
	srv := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	logger.Info("server started", "addr", addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Error("server error", "error", err)
		os.Exit(1)
	}
}

func redisAddr(redisURL string) string {
	if redisURL == "" {
		return "localhost:6379"
	}
	u, err := url.Parse(redisURL)
	if err != nil {
		return redisURL
	}
	return u.Host
}

func processHandler(door *api.Door) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req models.ProcessRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"error": "invalid body"})
			return
		}
		result, err := door.Process(r.Context(), req.Payload, req.PayloadID)
		switch err {
		case nil:
			writeJSON(w, http.StatusOK, models.ProcessResponse{Result: result})
		case api.ErrValidation:
			writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"error": "payload and payload_id required"})
		case api.ErrRateLimited, compute.ErrRateLimited:
			w.Header().Set("Retry-After", "1")
			writeJSON(w, http.StatusTooManyRequests, map[string]string{"error": "too many requests"})
		default:
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		}
	}
}

func healthHandler(rdb *redis.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		redisOK := rdb.Ping(r.Context()).Err() == nil
		writeJSON(w, http.StatusOK, map[string]any{"status": "ok", "redis": redisOK, "role": "worker"})
	}
}

func statsHandler(stats *compute.Stats, store *compute.Store, repo *compute.Repo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		snaps, err := repo.ReadAllStats(r.Context())
		if err != nil || len(snaps) == 0 {
			snap := stats.Snapshot()
			snap.CacheHitRate = store.CacheHitRate()
			snap.CorrStoreSize = store.Size()
			writeJSON(w, http.StatusOK, snap)
			return
		}
		snap := compute.Aggregate(snaps)
		writeJSON(w, http.StatusOK, snap)
	}
}

func publishStatsLoop(ctx context.Context, repo *compute.Repo, registry *compute.Registry, stats *compute.Stats, every, ttl time.Duration) {
	ticker := time.NewTicker(every)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := repo.PublishStats(ctx, registry.InstanceID(), stats.Snapshot(), ttl); err != nil {
				slog.Error("stats publish failed", "error", err)
			}
		}
	}
}

func saveSystemHandler(repo *compute.Repo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var cfg models.SystemConfig
		if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
			writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"error": "invalid body"})
			return
		}
		if err := repo.SaveSystem(r.Context(), cfg); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "save failed"})
			return
		}
		writeJSON(w, http.StatusOK, cfg)
	}
}

func listSystemsHandler(repo *compute.Repo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		systems, err := repo.ListSystems(r.Context())
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "list failed"})
			return
		}
		writeJSON(w, http.StatusOK, systems)
	}
}

func clearHandler(repo *compute.Repo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := repo.BumpConfigEpoch(r.Context()); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "clear failed"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
