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
