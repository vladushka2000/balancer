package mask

import (
	"sort"
	"strings"

	"pii/internal/models"
)

// Masker applies masking rules to spans right-to-left.
type Masker struct {
	mode string
}

// NewMasker creates a masker with the given mode.
func NewMasker(mode string) *Masker {
	return &Masker{mode: mode}
}

// Apply masks text according to spans.
func (m *Masker) Apply(text string, spans []models.Span) string {
	if len(spans) == 0 {
		return text
	}
	sorted := make([]models.Span, len(spans))
	copy(sorted, spans)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Start > sorted[j].Start
	})

	rules := MaskRules()
	var b strings.Builder
	b.WriteString(text)
	for _, s := range sorted {
		var replacement string
		switch m.mode {
		case "redact":
			replacement = "[" + strings.ToUpper(s.Type) + "]"
		default:
			if fn, ok := rules[s.Type]; ok {
				replacement = fn(text, s)
			} else {
				replacement = MaskDefault(text, s)
			}
		}
		b.Reset()
		b.WriteString(text[:s.Start])
		b.WriteString(replacement)
		b.WriteString(text[s.End:])
		text = b.String()
	}
	return text
}
