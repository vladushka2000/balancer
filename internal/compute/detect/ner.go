package detect

import (
	"embed"
	"regexp"
	"strings"
	"sync"

	"pii/internal/models"
)

//go:embed dicts/*.txt
var dictFS embed.FS

var (
	fioRe         = regexp.MustCompile(`[А-ЯЁ][а-яё]+(?:\s+[А-ЯЁ][а-яё]+){2,}`)
	addressRe     = regexp.MustCompile(`(?i)(?:г\.|ул\.|пр\.|пер\.|бульвар|проспект)\s+[А-Яа-яЁё0-9.,\s-]+`)
	orgRe         = regexp.MustCompile(`(?i)(?:выдан|отделение|уфмс|мвд|гу)\s+[А-Яа-яЁё0-9.,\s-]+`)
	orgMarkerRe   = regexp.MustCompile(`(?i)выдан|отделение|уфмс|мвд|гу`)
	birthPlaceRe  = regexp.MustCompile(`(?i)(?:место рождения|родился|родилась)\s*[:]?\s*[А-Яа-яЁё][А-Яа-яЁё\s-]+`)
	citizenshipRe = regexp.MustCompile(`(?i)(?:гражданство|гражданин)\s*[:]?\s*[А-Яа-яЁё][А-Яа-яЁё\s-]+`)
)

// NERDetector finds FIO, addresses and issuing organs via regex+dict.
type NERDetector struct {
	famous     map[string]struct{}
	chunkChars int
	overlap    int
}

// NewNERDetector creates an NER detector.
func NewNERDetector(chunkChars, overlap int) *NERDetector {
	return &NERDetector{
		famous:     map[string]struct{}{},
		chunkChars: chunkChars,
		overlap:    overlap,
	}
}

// Preload loads dictionaries.
func (n *NERDetector) Preload() {
	data, err := dictFS.ReadFile("dicts/famous.txt")
	if err != nil {
		return
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			n.famous[strings.ToLower(line)] = struct{}{}
		}
	}
}

// Detect finds NER spans, processing chunks in parallel.
func (n *NERDetector) Detect(text string) []models.Span {
	chunks := chunkText(text, n.chunkChars, n.overlap)
	results := make([][]models.Span, len(chunks))
	var wg sync.WaitGroup
	for i, c := range chunks {
		wg.Add(1)
		go func(i int, c chunk) {
			defer wg.Done()
			results[i] = n.detectChunk(c)
		}(i, c)
	}
	wg.Wait()

	seen := map[[3]int]struct{}{}
	var spans []models.Span
	for _, rs := range results {
		for _, s := range rs {
			key := [3]int{s.Start, s.End, len(s.Type)}
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			spans = append(spans, s)
		}
	}
	return spans
}

type chunk struct {
	text   string
	offset int
}

func chunkText(text string, size, overlap int) []chunk {
	if size <= 0 {
		size = 4000
	}
	runes := []rune(text)
	if len(runes) <= size {
		return []chunk{{text: text, offset: 0}}
	}
	var chunks []chunk
	for start := 0; start < len(runes); start += size - overlap {
		end := start + size
		if end > len(runes) {
			end = len(runes)
		}
		chunks = append(chunks, chunk{text: string(runes[start:end]), offset: start})
		if end == len(runes) {
			break
		}
	}
	return chunks
}

func (n *NERDetector) detectChunk(c chunk) []models.Span {
	var spans []models.Span
	for _, m := range fioRe.FindAllStringIndex(c.text, -1) {
		run := c.text[m[0]:m[1]]
		words := strings.Fields(run)
		if len(words) < 3 {
			continue
		}
		last := words[len(words)-3:]
		start := strings.Index(run, last[0])
		spans = append(spans, models.Span{
			Start: c.offset + m[0] + start, End: c.offset + m[0] + start + len(strings.Join(last, " ")),
			Type: "fio", Confidence: 0.8, Source: "ner",
		})
	}
	for _, m := range addressRe.FindAllStringIndex(c.text, -1) {
		spans = append(spans, models.Span{
			Start: c.offset + m[0], End: c.offset + m[1], Type: "address", Confidence: 0.7, Source: "ner",
		})
	}
	for _, m := range orgRe.FindAllStringIndex(c.text, -1) {
		if !orgMarkerRe.MatchString(c.text[m[0]:m[1]]) {
			continue
		}
		spans = append(spans, models.Span{
			Start: c.offset + m[0], End: c.offset + m[1], Type: "org", Confidence: 0.7, Source: "ner",
		})
	}
	for _, m := range birthPlaceRe.FindAllStringIndex(c.text, -1) {
		spans = append(spans, models.Span{
			Start: c.offset + m[0], End: c.offset + m[1], Type: "birth_place", Confidence: 0.7, Source: "ner",
		})
	}
	for _, m := range citizenshipRe.FindAllStringIndex(c.text, -1) {
		spans = append(spans, models.Span{
			Start: c.offset + m[0], End: c.offset + m[1], Type: "citizenship", Confidence: 0.7, Source: "ner",
		})
	}
	return spans
}
