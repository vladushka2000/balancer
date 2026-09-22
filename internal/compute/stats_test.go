package compute

import (
	"testing"
)

func TestHistogramEmpty(t *testing.T) {
	h := NewHistogram()
	if p := h.Percentile(50); p != 0 {
		t.Fatalf("expected 0 for empty histogram, got %v", p)
	}
}

func TestHistogramSingleBucket(t *testing.T) {
	h := NewHistogram()
	h.Add(3)
	if p := h.Percentile(50); p != 3 {
		t.Fatalf("expected 3 for single value, got %v", p)
	}
}

func TestHistogramOverflowBucket(t *testing.T) {
	h := NewHistogram()
	h.Add(5000)
	if p := h.Percentile(100); p != 1000 {
		t.Fatalf("expected overflow to cap at 1000, got %v", p)
	}
}

func TestHistogramManyBuckets(t *testing.T) {
	h := NewHistogram()
	for i := 0; i < 100; i++ {
		h.Add(2)
	}
	for i := 0; i < 100; i++ {
		h.Add(20)
	}
	p50 := h.Percentile(50)
	if p50 < 2 || p50 > 20 {
		t.Fatalf("expected p50 between 2 and 20, got %v", p50)
	}
}

func TestAggregateCounters(t *testing.T) {
	s1 := Snapshot{RequestsTotal: 100, MaskOK: 60, DemaskOK: 40, Count429: 5}
	s2 := Snapshot{RequestsTotal: 100, MaskOK: 55, DemaskOK: 45, Count429: 3}
	out := Aggregate([]Snapshot{s1, s2})
	if out.RequestsTotal != 200 {
		t.Fatalf("expected requests_total 200, got %d", out.RequestsTotal)
	}
	if out.MaskOK != 115 {
		t.Fatalf("expected mask_ok 115, got %d", out.MaskOK)
	}
	if out.DemaskOK != 85 {
		t.Fatalf("expected demask_ok 85, got %d", out.DemaskOK)
	}
	if out.Count429 != 8 {
		t.Fatalf("expected count_429 8, got %d", out.Count429)
	}
}

func TestAggregateHistogram(t *testing.T) {
	h1 := NewHistogram()
	h1.Add(3)
	h1.Add(20)
	s1 := Snapshot{RequestsTotal: 2, Histogram: h1}
	h2 := NewHistogram()
	h2.Add(3)
	s2 := Snapshot{RequestsTotal: 1, Histogram: h2}
	out := Aggregate([]Snapshot{s1, s2})
	if out.RequestsTotal != 3 {
		t.Fatalf("expected requests_total 3, got %d", out.RequestsTotal)
	}
	if out.Histogram.Counts[1] != 2 {
		t.Fatalf("expected 2 values in bucket 1, got %d", out.Histogram.Counts[1])
	}
	if out.Histogram.Counts[3] != 1 {
		t.Fatalf("expected 1 value in bucket 3, got %d", out.Histogram.Counts[3])
	}
}

func TestAggregateDetections(t *testing.T) {
	s1 := Snapshot{DetectionsByType: map[string]uint64{"fio": 2, "passport": 1}}
	s2 := Snapshot{DetectionsByType: map[string]uint64{"fio": 3}}
	out := Aggregate([]Snapshot{s1, s2})
	if out.DetectionsByType["fio"] != 5 {
		t.Fatalf("expected fio 5, got %d", out.DetectionsByType["fio"])
	}
	if out.DetectionsByType["passport"] != 1 {
		t.Fatalf("expected passport 1, got %d", out.DetectionsByType["passport"])
	}
}

func TestStatsRecordUsesHistogram(t *testing.T) {
	s := NewStats()
	s.RecordTokens(nil, 3, DirectionMask, 10)
	s.RecordTokens(nil, 20, DirectionMask, 20)
	snap := s.Snapshot()
	if snap.RequestsTotal != 2 {
		t.Fatalf("expected requests_total 2, got %d", snap.RequestsTotal)
	}
	if snap.LatencyMeanMs != 11.5 {
		t.Fatalf("expected mean 11.5, got %v", snap.LatencyMeanMs)
	}
	if snap.Histogram.Counts[1] != 1 {
		t.Fatalf("expected 1 value in bucket 1, got %d", snap.Histogram.Counts[1])
	}
	if snap.Histogram.Counts[3] != 1 {
		t.Fatalf("expected 1 value in bucket 3, got %d", snap.Histogram.Counts[3])
	}
}

func TestStatsRecord(t *testing.T) {
	s := NewStats()
	s.Record([]string{"fio"}, 5, DirectionMask)
	s.Record(nil, 5, DirectionDemask)
	snap := s.Snapshot()
	if snap.RequestsTotal != 2 {
		t.Fatalf("expected requests_total 2, got %d", snap.RequestsTotal)
	}
	if snap.MaskOK != 1 || snap.DemaskOK != 1 {
		t.Fatalf("expected mask_ok=1 demask_ok=1, got %d/%d", snap.MaskOK, snap.DemaskOK)
	}
	if snap.DetectionsByType["fio"] != 1 {
		t.Fatalf("expected fio 1, got %d", snap.DetectionsByType["fio"])
	}
}

func TestStatsRecord429(t *testing.T) {
	s := NewStats()
	s.Record429()
	s.Record429()
	snap := s.Snapshot()
	if snap.Count429 != 2 {
		t.Fatalf("expected count_429 2, got %d", snap.Count429)
	}
}

func TestStatsWindowStartEmpty(t *testing.T) {
	s := NewStats()
	start := s.windowStart()
	if start.IsZero() {
		t.Fatalf("expected non-zero window start")
	}
}

func TestStatsWindowStartAfterRecord(t *testing.T) {
	s := NewStats()
	s.RecordTokens(nil, 1, DirectionMask, 0)
	start := s.windowStart()
	if start.IsZero() {
		t.Fatalf("expected non-zero window start")
	}
}

func TestStatsSnapshotEmpty(t *testing.T) {
	s := NewStats()
	snap := s.Snapshot()
	if snap.RequestsTotal != 0 {
		t.Fatalf("expected requests_total 0, got %d", snap.RequestsTotal)
	}
	if snap.LatencyMeanMs != 0 {
		t.Fatalf("expected mean 0, got %v", snap.LatencyMeanMs)
	}
	if snap.Histogram.Counts == nil {
		t.Fatalf("expected histogram counts initialized")
	}
}

func TestHistogramPercentileExactBucket(t *testing.T) {
	h := NewHistogram()
	for i := 0; i < 4; i++ {
		h.Add(1)
	}
	for i := 0; i < 4; i++ {
		h.Add(5)
	}
	if p := h.Percentile(50); p != 1 {
		t.Fatalf("expected p50 1, got %v", p)
	}
}

func TestHistogramPercentileZeroCountBucket(t *testing.T) {
	h := NewHistogram()
	h.Add(1)
	h.Add(100)
	if p := h.Percentile(50); p < 1 || p > 100 {
		t.Fatalf("expected p50 between 1 and 100, got %v", p)
	}
}

func TestHistogramPercentileAllOverflow(t *testing.T) {
	h := NewHistogram()
	h.Add(5000)
	h.Add(6000)
	if p := h.Percentile(50); p != 1000 {
		t.Fatalf("expected p50 1000 (overflow cap), got %v", p)
	}
}

func TestAggregateEmpty(t *testing.T) {
	out := Aggregate(nil)
	if out.RequestsTotal != 0 {
		t.Fatalf("expected requests_total 0, got %d", out.RequestsTotal)
	}
	if out.DetectionsByType == nil {
		t.Fatalf("expected detections map initialized")
	}
}

func TestAggregateMismatchedHistogram(t *testing.T) {
	s1 := Snapshot{RequestsTotal: 1, Histogram: NewHistogram()}
	s1.Histogram.Counts = make([]uint64, 5)
	s2 := Snapshot{RequestsTotal: 1, Histogram: NewHistogram()}
	out := Aggregate([]Snapshot{s1, s2})
	if out.RequestsTotal != 2 {
		t.Fatalf("expected requests_total 2, got %d", out.RequestsTotal)
	}
}
