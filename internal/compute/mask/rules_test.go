package mask

import (
	"strings"
	"testing"

	"pii/internal/models"
)

func spanFor(text, substr string) models.Span {
	start := strings.Index(text, substr)
	return models.Span{Start: start, End: start + len(substr)}
}

func spanForType(text, substr, typ string) models.Span {
	s := spanFor(text, substr)
	s.Type = typ
	return s
}

func TestMaskPassport(t *testing.T) {
	text := "паспорт 4509 123456"
	s := spanFor(text, "4509 123456")
	got := MaskPassport(text, s)
	if got != "45** ****56" {
		t.Fatalf("expected 45** ****56, got %q", got)
	}
}

func TestMaskFIO(t *testing.T) {
	text := "Иванов Иван Иванович"
	s := spanFor(text, "Иванов Иван Иванович")
	got := MaskFIO(text, s)
	if got != "И. И. И." {
		t.Fatalf("expected И. И. И., got %q", got)
	}
}

func TestMaskPhone(t *testing.T) {
	text := "+7 912 345-67-89"
	s := spanFor(text, "+7 912 345-67-89")
	got := MaskPhone(text, s)
	if got != "+7 9** ***-**-89" {
		t.Fatalf("expected +7 9** ***-**-89, got %q", got)
	}
}

func TestMaskEmail(t *testing.T) {
	text := "ivan.ivanov@bank.ru"
	s := spanFor(text, "ivan.ivanov@bank.ru")
	got := MaskEmail(text, s)
	if got != "i**********@bank.ru" {
		t.Fatalf("expected i**********@bank.ru, got %q", got)
	}
}

func TestMaskCard(t *testing.T) {
	text := "4111 1111 1111 1111"
	s := spanFor(text, "4111 1111 1111 1111")
	got := MaskCard(text, s)
	if got != "4111 **** **** 1111" {
		t.Fatalf("expected 4111 **** **** 1111, got %q", got)
	}
}

func TestMaskINN(t *testing.T) {
	text := "7707083893"
	s := spanFor(text, "7707083893")
	got := MaskINN(text, s)
	if got != "77** ****** 93" {
		t.Fatalf("expected 77** ****** 93, got %q", got)
	}
}

func TestMaskDate(t *testing.T) {
	text := "12.01.1990"
	s := spanFor(text, "12.01.1990")
	got := MaskDate(text, s)
	if got != "**.**.1990" {
		t.Fatalf("expected **.**.1990, got %q", got)
	}
}

func TestMaskCVV(t *testing.T) {
	text := "CVV 123"
	s := spanFor(text, "123")
	got := MaskCVV(text, s)
	if got != "***" {
		t.Fatalf("expected ***, got %q", got)
	}
}

func TestMaskPIN(t *testing.T) {
	text := "пин-код 1234"
	s := spanFor(text, "1234")
	got := MaskPIN(text, s)
	if got != "***" {
		t.Fatalf("expected ***, got %q", got)
	}
}

func TestMaskAddress(t *testing.T) {
	text := "Москва, ул. Тверская, д. 1"
	s := spanFor(text, "Москва, ул. Тверская, д. 1")
	got := MaskAddress(text, s)
	if got != "Москва, ул. ******, д. **" {
		t.Fatalf("expected Москва, ул. ******, д. **, got %q", got)
	}
}

func TestMaskDefault(t *testing.T) {
	text := "1234567890"
	s := spanFor(text, "1234567890")
	got := MaskDefault(text, s)
	if got == "1234567890" {
		t.Fatalf("expected masked, got %q", got)
	}
}

func TestMaskWord(t *testing.T) {
	text := "Российская Федерация"
	s := spanFor(text, "Российская Федерация")
	got := MaskWord(text, s)
	if got != "Р********* Ф********" {
		t.Fatalf("expected Р********* Ф********, got %q", got)
	}
}
