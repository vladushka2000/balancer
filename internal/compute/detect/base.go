package detect

import "pii/internal/models"

// Detector finds PII spans in text.
type Detector interface {
	Detect(text string) []models.Span
}

// Registry holds a set of detectors and runs text through all of them.
type Registry struct {
	detectors []Detector
}

// NewRegistry creates an empty detector registry.
func NewRegistry() *Registry {
	return &Registry{}
}

// Register adds a detector to the registry.
func (r *Registry) Register(d Detector) {
	r.detectors = append(r.detectors, d)
}

// DetectAll runs text through every detector and collects spans.
func (r *Registry) DetectAll(text string) []models.Span {
	var spans []models.Span
	for _, d := range r.detectors {
		spans = append(spans, d.Detect(text)...)
	}
	return spans
}
