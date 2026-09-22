package compute

import (
	"pii/internal/compute/detect"
	"pii/internal/compute/mask"
)

// Pipeline orchestrates detect → context → merge → mask.
type Pipeline struct {
	structural *detect.Registry
	ner        *detect.NERDetector
	context    *detect.ContextRule
	masker     *mask.Masker
}

// NewPipeline builds a detection pipeline.
func NewPipeline(structural *detect.Registry, ner *detect.NERDetector, context *detect.ContextRule, masker *mask.Masker) *Pipeline {
	return &Pipeline{
		structural: structural,
		ner:        ner,
		context:    context,
		masker:     masker,
	}
}

// Process masks text and returns the masked text and detected types.
func (p *Pipeline) Process(text string) (string, []string) {
	return p.ProcessFiltered(text, nil)
}

// ProcessFiltered masks text, keeping only spans whose type is allowed.
// A nil or empty allowed set keeps every detected type.
func (p *Pipeline) ProcessFiltered(text string, allowed map[string]bool) (string, []string) {
	spans := p.structural.DetectAll(text)
	nerSpans := p.ner.Detect(text)
	spans = append(spans, p.context.Filter(nerSpans, text)...)
	merged := detect.MergeSpans(spans)

	if len(allowed) > 0 {
		filtered := merged[:0]
		for _, s := range merged {
			if allowed[s.Type] {
				filtered = append(filtered, s)
			}
		}
		merged = filtered
	}

	masked := p.masker.Apply(text, merged)

	seen := map[string]struct{}{}
	var types []string
	for _, s := range merged {
		if _, ok := seen[s.Type]; ok {
			continue
		}
		seen[s.Type] = struct{}{}
		types = append(types, s.Type)
	}
	return masked, types
}
