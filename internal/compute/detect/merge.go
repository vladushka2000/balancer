package detect

import (
	"sort"

	"pii/internal/models"
)

// MergeSpans sorts spans, resolves overlaps and applies the PIN gate.
func MergeSpans(spans []models.Span) []models.Span {
	sort.Slice(spans, func(i, j int) bool {
		if spans[i].Start != spans[j].Start {
			return spans[i].Start < spans[j].Start
		}
		return spans[i].End > spans[j].End
	})

	hasCard := false
	for _, s := range spans {
		if s.Type == "card" {
			hasCard = true
			break
		}
	}

	var out []models.Span
	for _, s := range spans {
		if s.Type == "pin" && !hasCard {
			continue
		}
		if len(out) > 0 && s.Start < out[len(out)-1].End {
			last := &out[len(out)-1]
			if prefer(s, *last) {
				out[len(out)-1] = s
			}
			continue
		}
		out = append(out, s)
	}
	return out
}

func prefer(a, b models.Span) bool {
	if a.Type == "card_holder" && b.Type == "fio" {
		return true
	}
	if b.Type == "card_holder" && a.Type == "fio" {
		return false
	}
	if a.Source == "regex" && b.Source != "regex" {
		return true
	}
	if b.Source == "regex" && a.Source != "regex" {
		return false
	}
	return a.End-a.Start > b.End-b.Start
}
