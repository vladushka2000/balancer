package compute

import (
	"sync"
	"time"
)

// Direction is the processing direction.
type Direction int

const (
	DirectionMask Direction = iota
	DirectionDemask
)

// Histogram buckets latency values into fixed bounds (ms).
type Histogram struct {
	Bounds []float64 `json:"bounds"`
	Counts []uint64  `json:"counts"`
}

// NewHistogram creates a histogram with fixed bucket bounds.
func NewHistogram() Histogram {
	return Histogram{
		Bounds: []float64{1, 5, 10, 25, 50, 100, 250, 500, 1000},
		Counts: make([]uint64, 10),
	}
}

// Add records a latency value into the appropriate bucket.
func (h *Histogram) Add(v float64) {
	idx := len(h.Bounds)
	for i, b := range h.Bounds {
		if v <= b {
			idx = i
			break
		}
	}
	h.Counts[idx]++
}

// Percentile returns the p-th percentile via linear interpolation.
func (h *Histogram) Percentile(p float64) float64 {
	var total uint64
	for _, c := range h.Counts {
		total += c
	}
	if total == 0 {
		return 0
	}
	target := float64(total) * p / 100
	var cum uint64
	for i, c := range h.Counts {
		cum += c
		if float64(cum) >= target {
			if i == len(h.Bounds) {
				return h.Bounds[len(h.Bounds)-1]
			}
			lo := 0.0
			if i > 0 {
				lo = h.Bounds[i-1]
			}
			hi := h.Bounds[i]
			prev := cum - c
			if c == 0 {
				return hi
			}
			frac := (target - float64(prev)) / float64(c)
			return lo + frac*(hi-lo)
		}
	}
	return h.Bounds[len(h.Bounds)-1]
}

// Stats aggregates latency and counters.
type Stats struct {
	mu               sync.Mutex
	requestsTotal    uint64
	maskOK           uint64
	demaskOK         uint64
	count429         uint64
	detectionsByType map[string]uint64
	histogram        Histogram
	latencySum       float64
	tokensTotal      uint64
	window           []time.Time
}

// NewStats creates a stats aggregator.
func NewStats() *Stats {
	return &Stats{
		detectionsByType: map[string]uint64{},
		histogram:        NewHistogram(),
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
	cutoff := now.Add(-time.Second)
	keep := 0
	for keep < len(s.window) && !s.window[keep].After(cutoff) {
		keep++
	}
	if keep > 0 {
		s.window = s.window[keep:]
	}
	if dir == DirectionMask {
		s.maskOK++
	} else {
		s.demaskOK++
	}
	for _, t := range types {
		s.detectionsByType[t]++
	}
	s.histogram.Add(latencyMs)
	s.latencySum += latencyMs
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
	Histogram        Histogram         `json:"histogram"`
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
	if s.requestsTotal > 0 {
		snap.LatencyMeanMs = s.latencySum / float64(s.requestsTotal)
		snap.LatencyP50Ms = s.histogram.Percentile(50)
		snap.LatencyP95Ms = s.histogram.Percentile(95)
		snap.LatencyP99Ms = s.histogram.Percentile(99)
	}
	snap.Histogram = s.histogram
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

// windowStart returns the timestamp of the first recorded request.
func (s *Stats) windowStart() time.Time {
	if len(s.window) == 0 {
		return time.Now()
	}
	return s.window[0]
}

// Aggregate merges multiple snapshots into one.
func Aggregate(snaps []Snapshot) Snapshot {
	out := Snapshot{
		DetectionsByType: map[string]uint64{},
		Histogram:        NewHistogram(),
	}
	for _, s := range snaps {
		out.RequestsTotal += s.RequestsTotal
		out.MaskOK += s.MaskOK
		out.DemaskOK += s.DemaskOK
		out.Count429 += s.Count429
		out.TokensPerSec += s.TokensPerSec
		out.RPS += s.RPS
		for k, v := range s.DetectionsByType {
			out.DetectionsByType[k] += v
		}
		if len(s.Histogram.Counts) == len(out.Histogram.Counts) {
			for i := range out.Histogram.Counts {
				out.Histogram.Counts[i] += s.Histogram.Counts[i]
			}
		}
	}
	if out.RequestsTotal > 0 {
		out.LatencyP50Ms = out.Histogram.Percentile(50)
		out.LatencyP95Ms = out.Histogram.Percentile(95)
		out.LatencyP99Ms = out.Histogram.Percentile(99)
	}
	return out
}
