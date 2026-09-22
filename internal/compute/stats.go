package compute

import (
	"sort"
	"sync"
	"time"
)

// Direction is the processing direction.
type Direction int

const (
	DirectionMask Direction = iota
	DirectionDemask
)

// Stats aggregates latency and counters.
type Stats struct {
	mu               sync.Mutex
	requestsTotal    uint64
	maskOK           uint64
	demaskOK         uint64
	count429         uint64
	detectionsByType map[string]uint64
	latencies        []float64
	tokensTotal      uint64
	window           []time.Time
}

// NewStats creates a stats aggregator.
func NewStats() *Stats {
	return &Stats{
		detectionsByType: map[string]uint64{},
		latencies:        []float64{},
		window:           []time.Time{},
	}
}

// Record records a processed request.
func (s *Stats) Record(types []string, latencyMs float64, dir Direction) {
	s.RecordTokens(types, latencyMs, dir, 0)
}

// RecordTokens records a processed request with a token count.
func (s *Stats) RecordTokens(types []string, latencyMs float64, dir Direction, tokens int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.requestsTotal++
	s.tokensTotal += uint64(tokens)
	now := time.Now()
	s.window = append(s.window, now)
	if dir == DirectionMask {
		s.maskOK++
	} else {
		s.demaskOK++
	}
	for _, t := range types {
		s.detectionsByType[t]++
	}
	s.latencies = append(s.latencies, latencyMs)
}

// Record429 records a rate-limited request.
func (s *Stats) Record429() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.count429++
}

// Snapshot is a point-in-time view of stats.
type Snapshot struct {
	RequestsTotal    uint64            `json:"requests_total"`
	MaskOK           uint64            `json:"mask_ok"`
	DemaskOK         uint64            `json:"demask_ok"`
	Count429         uint64            `json:"count_429"`
	DetectionsByType map[string]uint64 `json:"detections_by_type"`
	LatencyMeanMs    float64           `json:"latency_mean_ms"`
	LatencyP50Ms     float64           `json:"latency_p50_ms"`
	LatencyP95Ms     float64           `json:"latency_p95_ms"`
	LatencyP99Ms     float64           `json:"latency_p99_ms"`
	RPS              float64           `json:"rps"`
	TokensPerSec     float64           `json:"tokens_per_sec"`
	CacheHitRate     float64           `json:"cache_hit_rate"`
	CorrStoreSize    int               `json:"corr_store_size"`
}

// Snapshot returns the current stats.
func (s *Stats) Snapshot() Snapshot {
	s.mu.Lock()
	defer s.mu.Unlock()
	snap := Snapshot{
		RequestsTotal:    s.requestsTotal,
		MaskOK:           s.maskOK,
		DemaskOK:         s.demaskOK,
		Count429:         s.count429,
		DetectionsByType: map[string]uint64{},
	}
	for k, v := range s.detectionsByType {
		snap.DetectionsByType[k] = v
	}
	if len(s.latencies) > 0 {
		sorted := make([]float64, len(s.latencies))
		copy(sorted, s.latencies)
		sort.Float64s(sorted)
		snap.LatencyMeanMs = mean(sorted)
		snap.LatencyP50Ms = percentile(sorted, 50)
		snap.LatencyP95Ms = percentile(sorted, 95)
		snap.LatencyP99Ms = percentile(sorted, 99)
	}
	now := time.Now()
	cutoff := now.Add(-time.Second)
	recent := 0
	for _, t := range s.window {
		if t.After(cutoff) {
			recent++
		}
	}
	snap.RPS = float64(recent)
	snap.TokensPerSec = float64(s.tokensTotal) / max(1, now.Sub(s.windowStart()).Seconds())
	return snap
}

func (s *Stats) windowStart() time.Time {
	if len(s.window) == 0 {
		return time.Now()
	}
	return s.window[0]
}

func mean(v []float64) float64 {
	var sum float64
	for _, x := range v {
		sum += x
	}
	return sum / float64(len(v))
}

func percentile(sorted []float64, p float64) float64 {
	if len(sorted) == 0 {
		return 0
	}
	idx := int(float64(len(sorted)-1) * p / 100)
	return sorted[idx]
}
