package detect

import (
	"embed"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"sync"

	"pii/internal/models"
)

//go:embed dicts/*.txt
var dictFS embed.FS

var (
	fioRe         = regexp.MustCompile(`[А-ЯЁ][а-яё]+(?:\s+[А-ЯЁ][а-яё]+){2,}`)
	fioTwoRe      = regexp.MustCompile(`[А-ЯЁ][а-яё]+\s+[А-ЯЁ][а-яё]+`)
	addressRe     = regexp.MustCompile(`(?i)(?:г\.|ул\.|пр\.|пер\.|бульвар|проспект)\s+[А-Яа-яЁё0-9.,\s-]+`)
	orgRe         = regexp.MustCompile(`(?i)(?:выдан|отделение|уфмс|мвд|гу)\s+[А-Яа-яЁё0-9.,\s-]+`)
	orgMarkerRe   = regexp.MustCompile(`(?i)выдан|отделение|уфмс|мвд|гу`)
	birthPlaceRe  = regexp.MustCompile(`(?i)(?:место рождения|родился|родилась)\s*[:]?\s*[А-Яа-яЁё][А-Яа-яЁё\s-]+`)
	citizenshipRe = regexp.MustCompile(`(?i)(?:гражданство|гражданин)\s*[:]?\s*[А-Яа-яЁё][А-Яа-яЁё\s-]+`)
	fioMarkerRe   = regexp.MustCompile(`(?i)клиент|заемщик|паспорт|родился|родилась|держатель|получатель|заявитель`)
	countryRe     = regexp.MustCompile(`(?i)(?:страна|государство)\s*[:]?\s*[А-Яа-яЁё][А-Яа-яЁё\s-]+`)
	cityRe        = regexp.MustCompile(`(?i)(?:г\.|город)\s+[А-ЯЁ][а-яё]+`)
	streetRe      = regexp.MustCompile(`(?i)(?:ул\.|улица|пр\.|проспект|пер\.|переулок|бульвар)\s+[А-ЯЁ][а-яё]+`)
	houseRe       = regexp.MustCompile(`(?i)(?:д\.|дом)\s*\d+`)
	flatRe        = regexp.MustCompile(`(?i)(?:кв\.|квартира)\s*\d+`)
	cardHolderRe  = regexp.MustCompile(`(?i:держатель|holder|имя на карте|cardholder)\s*(?:карты\s+)?[:]?\s*([А-ЯЁ][а-яё]+(?:\s+[А-ЯЁ][а-яё]+){1,2})`)
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
	if len(chunks) == 1 {
		results[0] = n.detectChunk(chunks[0])
	} else {
		workers := len(chunks)
		if ncpu := runtime.NumCPU(); workers > ncpu {
			workers = ncpu
		}
		var wg sync.WaitGroup
		ch := make(chan int)
		for w := 0; w < workers; w++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for i := range ch {
					results[i] = n.detectChunk(chunks[i])
				}
			}()
		}
		for i := range chunks {
			ch <- i
		}
		close(ch)
		wg.Wait()
	}

	seen := map[spanKey]struct{}{}
	var spans []models.Span
	for _, rs := range results {
		for _, s := range rs {
			key := spanKey{Start: s.Start, End: s.End, Type: s.Type}
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

type spanKey struct {
	Start int
	End   int
	Type  string
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
	for _, m := range fioTwoRe.FindAllStringIndex(c.text, -1) {
		if !hasMarkerNear(c.text, m[0], m[1]) {
			continue
		}
		spans = append(spans, models.Span{
			Start: c.offset + m[0], End: c.offset + m[1], Type: "fio", Confidence: 0.7, Source: "ner",
		})
	}
	for _, m := range addressRe.FindAllStringIndex(c.text, -1) {
		spans = append(spans, models.Span{
			Start: c.offset + m[0], End: c.offset + m[1], Type: "address", Confidence: 0.7, Source: "ner",
		})
	}
	spans = append(spans, detectAddressComponents(c)...)
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
	for _, m := range cardHolderRe.FindAllStringSubmatchIndex(c.text, -1) {
		spans = append(spans, models.Span{
			Start: c.offset + m[2], End: c.offset + m[3], Type: "card_holder", Confidence: 0.8, Source: "ner",
		})
	}
	return spans
}

func hasMarkerNear(text string, start, end int) bool {
	lo := start - 200
	if lo < 0 {
		lo = 0
	}
	hi := end + 200
	if hi > len(text) {
		hi = len(text)
	}
	return fioMarkerRe.MatchString(text[lo:hi])
}

func detectAddressComponents(c chunk) []models.Span {
	var comps []models.Span
	for _, re := range []*regexp.Regexp{countryRe, cityRe, streetRe, houseRe, flatRe} {
		for _, m := range re.FindAllStringIndex(c.text, -1) {
			comps = append(comps, models.Span{
				Start: c.offset + m[0], End: c.offset + m[1], Type: "address", Confidence: 0.6, Source: "ner",
			})
		}
	}
	if len(comps) < 2 {
		return nil
	}
	sort.Slice(comps, func(i, j int) bool { return comps[i].Start < comps[j].Start })
	merged := []models.Span{comps[0]}
	for _, s := range comps[1:] {
		last := &merged[len(merged)-1]
		if s.Start <= last.End+2 {
			if s.End > last.End {
				last.End = s.End
			}
			continue
		}
		merged = append(merged, s)
	}
	return merged
}
