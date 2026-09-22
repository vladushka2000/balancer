package detect

import (
	"testing"

	"pii/internal/models"
)

func TestFamousWithoutContext(t *testing.T) {
	c := NewContextRule(200)
	spans := []models.Span{{Start: 0, End: 6, Type: "fio", Source: "ner"}}
	out := c.Filter(spans, "Пушкин")
	if len(out) != 0 {
		t.Fatalf("expected famous filtered, got %+v", out)
	}
}

func TestFamousWithPassport(t *testing.T) {
	c := NewContextRule(200)
	spans := []models.Span{
		{Start: 0, End: 6, Type: "fio", Source: "ner"},
		{Start: 9, End: 23, Type: "passport", Source: "regex"},
	}
	out := c.Filter(spans, "Пушкин, паспорт 4509 123456")
	if len(out) != 2 {
		t.Fatalf("expected both spans kept, got %+v", out)
	}
}

func TestOrgAddressWithoutContext(t *testing.T) {
	c := NewContextRule(200)
	spans := []models.Span{{Start: 0, End: 30, Type: "address", Source: "ner"}}
	out := c.Filter(spans, "Отделение банка по адресу: Москва, ул. Тверская")
	if len(out) != 0 {
		t.Fatalf("expected org address filtered, got %+v", out)
	}
}

func TestAddressWithMarker(t *testing.T) {
	c := NewContextRule(200)
	spans := []models.Span{{Start: 6, End: 30, Type: "address", Source: "ner"}}
	out := c.Filter(spans, "адрес: Москва, ул. Тверская, д. 1")
	if len(out) != 1 {
		t.Fatalf("expected address kept, got %+v", out)
	}
}

func TestNERWithoutContext(t *testing.T) {
	c := NewContextRule(200)
	spans := []models.Span{{Start: 0, End: 20, Type: "fio", Source: "ner"}}
	out := c.Filter(spans, "Иванов Иван Иванович")
	if len(out) != 0 {
		t.Fatalf("expected fio filtered without context, got %+v", out)
	}
}
