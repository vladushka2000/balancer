package compute

import (
	"sort"
	"sync"
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
}

// NewStats creates a stats aggregator.
func NewStats() *Stats {
	return &Stats{
		detectionsByType: map[string]uint64{},
		latencies:        []float64{},
	}
}

// Record records a processed request.
func (s *Stats) Record(types []string, latencyMs float64, dir Direction) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.requestsTotal++
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
	return snap
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
