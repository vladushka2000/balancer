package mask

import (
	"strings"
	"testing"

	"pii/internal/models"
)

func TestMaskerPartialMode(t *testing.T) {
	m := NewMasker("partial")
	text := "паспорт 4509 123456"
	spans := []models.Span{spanForType(text, "4509 123456", "passport")}
	got := m.Apply(text, spans)
	if got != "паспорт 45** ****56" {
		t.Fatalf("expected паспорт 45** ****56, got %q", got)
	}
}

func TestMaskerRedactMode(t *testing.T) {
	m := NewMasker("redact")
	text := "паспорт 4509 123456"
	spans := []models.Span{spanForType(text, "4509 123456", "passport")}
	got := m.Apply(text, spans)
	if got != "паспорт [PASSPORT]" {
		t.Fatalf("expected паспорт [PASSPORT], got %q", got)
	}
}

func TestMaskerRightToLeft(t *testing.T) {
	m := NewMasker("partial")
	text := "Иванов Иван Иванович, паспорт 4509 123456"
	spans := []models.Span{
		spanForType(text, "Иванов Иван Иванович", "fio"),
		spanForType(text, "4509 123456", "passport"),
	}
	got := m.Apply(text, spans)
	if got != "И. И. И., паспорт 45** ****56" {
		t.Fatalf("unexpected result %q", got)
	}
}

func TestMaskerEmptySpans(t *testing.T) {
	m := NewMasker("partial")
	text := "Привет, мир!"
	got := m.Apply(text, nil)
	if got != text {
		t.Fatalf("expected unchanged, got %q", got)
	}
}

func TestMaskerContainsMasked(t *testing.T) {
	m := NewMasker("partial")
	text := "паспорт 4509 123456"
	spans := []models.Span{spanForType(text, "4509 123456", "passport")}
	got := m.Apply(text, spans)
	if !strings.Contains(got, "45** ****56") {
		t.Fatalf("expected masked digits, got %q", got)
	}
}
