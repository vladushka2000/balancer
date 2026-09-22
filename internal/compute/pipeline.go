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
	spans := p.structural.DetectAll(text)
	nerSpans := p.ner.Detect(text)
	spans = append(spans, p.context.Filter(nerSpans, text)...)
	merged := detect.MergeSpans(spans)
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
