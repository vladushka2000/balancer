package detect

import "testing"

func TestDetectFIO(t *testing.T) {
	n := NewNERDetector(4000, 200)
	n.Preload()
	spans := n.Detect("Иванов Иван Иванович")
	if len(spans) != 1 || spans[0].Type != "fio" || spans[0].Source != "ner" {
		t.Fatalf("expected fio ner span, got %+v", spans)
	}
}

func TestDetectChunksOffset(t *testing.T) {
	n := NewNERDetector(50, 10)
	n.Preload()
	long := "текст текст текст текст текст текст текст текст текст текст Иванов Иван Иванович"
	spans := n.Detect(long)
	if len(spans) != 1 {
		t.Fatalf("expected fio span, got %+v", spans)
	}
	if spans[0].Start >= spans[0].End {
		t.Fatalf("invalid span offsets %+v", spans[0])
	}
}

func TestDetectDedup(t *testing.T) {
	n := NewNERDetector(30, 20)
	n.Preload()
	spans := n.Detect("Иванов Иван Иванович")
	seen := map[[3]int]struct{}{}
	for _, s := range spans {
		key := [3]int{s.Start, s.End, len(s.Type)}
		if _, ok := seen[key]; ok {
			t.Fatalf("duplicate span %+v", s)
		}
		seen[key] = struct{}{}
	}
}

func TestDetectBirthPlace(t *testing.T) {
	n := NewNERDetector(4000, 200)
	n.Preload()
	spans := n.Detect("Место рождения: Москва")
	if len(spans) != 1 || spans[0].Type != "birth_place" {
		t.Fatalf("expected birth_place span, got %+v", spans)
	}
}

func TestDetectCitizenship(t *testing.T) {
	n := NewNERDetector(4000, 200)
	n.Preload()
	spans := n.Detect("Гражданство: Российская Федерация")
	if len(spans) != 1 || spans[0].Type != "citizenship" {
		t.Fatalf("expected citizenship span, got %+v", spans)
	}
}

func TestDetectTwoWordFIOWithMarker(t *testing.T) {
	n := NewNERDetector(4000, 200)
	n.Preload()
	spans := n.Detect("Клиент Иванов Иван")
	found := false
	for _, s := range spans {
		if s.Type == "fio" && s.Source == "ner" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected 2-word fio span, got %+v", spans)
	}
}

func TestDetectTwoWordFIONoMarker(t *testing.T) {
	n := NewNERDetector(4000, 200)
	n.Preload()
	spans := n.Detect("Иванов Иван")
	for _, s := range spans {
		if s.Type == "fio" {
			t.Fatalf("expected no fio span without marker, got %+v", spans)
		}
	}
}

func TestDetectAddressComponents(t *testing.T) {
	n := NewNERDetector(4000, 200)
	n.Preload()
	spans := n.Detect("г. Москва, ул. Тверская, д. 1, кв. 5")
	found := false
	for _, s := range spans {
		if s.Type == "address" && s.Source == "ner" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected address span, got %+v", spans)
	}
}

func TestDetectDedupTypeCollision(t *testing.T) {
	n := NewNERDetector(4000, 200)
	n.Preload()
	spans := n.Detect("Клиент Иванов Иван")
	seen := map[[2]int]struct{}{}
	for _, s := range spans {
		key := [2]int{s.Start, s.End}
		if _, ok := seen[key]; ok {
			t.Fatalf("duplicate span at %+v in %+v", key, spans)
		}
		seen[key] = struct{}{}
	}
}
