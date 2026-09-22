package detect

import (
	"strings"

	"pii/internal/models"
)

// ContextRule filters NER spans by dictionaries and context window.
type ContextRule struct {
	window       int
	famous       map[string]struct{}
	orgAddresses map[string]struct{}
	markers      map[string]struct{}
}

// NewContextRule creates a context rule and loads dictionaries.
func NewContextRule(window int) *ContextRule {
	c := &ContextRule{
		window:       window,
		famous:       loadDict("dicts/famous.txt"),
		orgAddresses: loadDict("dicts/org_addresses.txt"),
		markers:      loadDict("dicts/markers.txt"),
	}
	return c
}

func loadDict(name string) map[string]struct{} {
	m := map[string]struct{}{}
	data, err := dictFS.ReadFile(name)
	if err != nil {
		return m
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			m[strings.ToLower(line)] = struct{}{}
		}
	}
	return m
}

// Filter keeps NER spans that belong to PII.
func (c *ContextRule) Filter(spans []models.Span, text string) []models.Span {
	var out []models.Span
	for _, s := range spans {
		if s.Source != "ner" {
			out = append(out, s)
			continue
		}
		value := strings.ToLower(text[s.Start:s.End])
		if s.Type == "fio" {
			if _, ok := c.famous[value]; ok {
				continue
			}
		}
		if s.Type == "address" {
			if _, ok := c.orgAddresses[value]; ok {
				continue
			}
			if c.hasOrgMarker(text, s) {
				continue
			}
		}
		if !c.hasContext(text, s) {
			continue
		}
		out = append(out, s)
	}
	return out
}

func (c *ContextRule) hasContext(text string, s models.Span) bool {
	start := s.Start - c.window
	if start < 0 {
		start = 0
	}
	end := s.End + c.window
	if end > len(text) {
		end = len(text)
	}
	window := strings.ToLower(text[start:end])
	for marker := range c.markers {
		if strings.Contains(window, marker) {
			return true
		}
	}
	return false
}

func (c *ContextRule) hasOrgMarker(text string, s models.Span) bool {
	start := s.Start - c.window
	if start < 0 {
		start = 0
	}
	end := s.End + c.window
	if end > len(text) {
		end = len(text)
	}
	window := strings.ToLower(text[start:end])
	for _, m := range []string{"отделение", "офис", "филиал", "банк по адресу"} {
		if strings.Contains(window, m) {
			return true
		}
	}
	return false
}
