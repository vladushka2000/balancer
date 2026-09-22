package detect

import (
	"testing"

	"pii/internal/models"
)

func TestMergeNoOverlap(t *testing.T) {
	spans := []models.Span{
		{Start: 0, End: 5, Type: "phone", Source: "regex"},
		{Start: 10, End: 15, Type: "email", Source: "regex"},
	}
	out := MergeSpans(spans)
	if len(out) != 2 {
		t.Fatalf("expected 2 spans, got %+v", out)
	}
}

func TestMergeOverlapPreferRegex(t *testing.T) {
	spans := []models.Span{
		{Start: 0, End: 20, Type: "fio", Source: "ner"},
		{Start: 0, End: 10, Type: "passport", Source: "regex"},
	}
	out := MergeSpans(spans)
	if len(out) != 1 || out[0].Source != "regex" {
		t.Fatalf("expected regex preferred, got %+v", out)
	}
}

func TestMergeOverlapPreferLonger(t *testing.T) {
	spans := []models.Span{
		{Start: 0, End: 10, Type: "phone", Source: "regex"},
		{Start: 0, End: 20, Type: "email", Source: "regex"},
	}
	out := MergeSpans(spans)
	if len(out) != 1 || out[0].End != 20 {
		t.Fatalf("expected longer span, got %+v", out)
	}
}

func TestPINGateNoCard(t *testing.T) {
	spans := []models.Span{{Start: 0, End: 10, Type: "pin", Source: "regex"}}
	out := MergeSpans(spans)
	if len(out) != 0 {
		t.Fatalf("expected pin dropped, got %+v", out)
	}
}

func TestPINGateWithCard(t *testing.T) {
	spans := []models.Span{
		{Start: 0, End: 10, Type: "pin", Source: "regex"},
		{Start: 20, End: 40, Type: "card", Source: "regex"},
	}
	out := MergeSpans(spans)
	if len(out) != 2 {
		t.Fatalf("expected both spans, got %+v", out)
	}
}

func TestMergeSorted(t *testing.T) {
	spans := []models.Span{
		{Start: 20, End: 30, Type: "email", Source: "regex"},
		{Start: 0, End: 10, Type: "phone", Source: "regex"},
	}
	out := MergeSpans(spans)
	if out[0].Start != 0 || out[1].Start != 20 {
		t.Fatalf("expected sorted, got %+v", out)
	}
}
