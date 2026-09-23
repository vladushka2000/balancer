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
	if got != "77******93" {
		t.Fatalf("expected 77******93, got %q", got)
	}
}

func TestMaskINN12(t *testing.T) {
	text := "ИНН 500100732259"
	s := spanFor(text, "ИНН 500100732259")
	got := MaskINN(text, s)
	if got != "ИНН 50********59" {
		t.Fatalf("expected ИНН 50********59, got %q", got)
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

func TestMaskCVV4(t *testing.T) {
	text := "CVV 1234"
	s := spanFor(text, "1234")
	got := MaskCVV(text, s)
	if got != "****" {
		t.Fatalf("expected ****, got %q", got)
	}
}

func TestMaskPIN(t *testing.T) {
	text := "пин-код 1234"
	s := spanFor(text, "1234")
	got := MaskPIN(text, s)
	if got != "****" {
		t.Fatalf("expected ****, got %q", got)
	}
}

func TestMaskSNILS(t *testing.T) {
	text := "СНИЛС 123-456-789 01"
	s := spanFor(text, "123-456-789 01")
	got := MaskSNILS(text, s)
	if got != "123-***-*** **" {
		t.Fatalf("expected 123-***-*** **, got %q", got)
	}
}

func TestMaskSNILSOtherPrefix(t *testing.T) {
	text := "СНИЛС 987-654-321 00"
	s := spanFor(text, "987-654-321 00")
	got := MaskSNILS(text, s)
	if got != "987-***-*** **" {
		t.Fatalf("expected 987-***-*** **, got %q", got)
	}
}

func TestMaskAddress(t *testing.T) {
	text := "Москва, ул. Тверская, д. 1"
	s := spanFor(text, "Москва, ул. Тверская, д. 1")
	got := MaskAddress(text, s)
	if got != "М*****, ул. Т*******, д. *" {
		t.Fatalf("expected М*****, ул. Т*******, д. *, got %q", got)
	}
}

func TestMaskAddressOtherCity(t *testing.T) {
	text := "Казань, пр. Победы, д. 25"
	s := spanFor(text, "Казань, пр. Победы, д. 25")
	got := MaskAddress(text, s)
	if got != "К*****, пр. П*****, д. **" {
		t.Fatalf("expected К*****, пр. П*****, д. **, got %q", got)
	}
}

func TestMaskAddressFlatIndex(t *testing.T) {
	text := "г. Москва, ул. Тверская, д. 1, кв. 10, 123456"
	s := spanFor(text, text)
	got := MaskAddress(text, s)
	want := "г. М*****, ул. Т*******, д. *, кв. **, ******"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
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

func TestMaskPassportKeepsMarkers(t *testing.T) {
	text := "серия 4509 номер 123456"
	s := spanFor(text, text)
	got := MaskPassport(text, s)
	if got != "серия 45** номер ****56" {
		t.Fatalf("expected серия 45** номер ****56, got %q", got)
	}
}

func TestMaskPassportFallbackStarsDigits(t *testing.T) {
	text := "паспорт 4509"
	s := spanFor(text, text)
	got := MaskPassport(text, s)
	if got != "паспорт ****" {
		t.Fatalf("expected паспорт ****, got %q", got)
	}
}

func TestMaskCVVKeepsLabel(t *testing.T) {
	text := "CVV 123"
	s := spanFor(text, text)
	got := MaskCVV(text, s)
	if got != "CVV ***" {
		t.Fatalf("expected CVV ***, got %q", got)
	}
}

func TestMaskPINKeepsLabel(t *testing.T) {
	text := "пин-код 1234"
	s := spanFor(text, text)
	got := MaskPIN(text, s)
	if got != "пин-код ****" {
		t.Fatalf("expected пин-код ****, got %q", got)
	}
}

func TestMaskDepartmentKeepsLabel(t *testing.T) {
	text := "код подразделения 770-001"
	s := spanFor(text, text)
	got := MaskDepartmentCode(text, s)
	if got != "код подразделения ***-***" {
		t.Fatalf("expected код подразделения ***-***, got %q", got)
	}
}

func TestMaskDriverLicense(t *testing.T) {
	text := "в/у 77 АА 123456"
	s := spanFor(text, text)
	got := MaskDriverLicense(text, s)
	if got != "в/у 77** ****56" {
		t.Fatalf("expected в/у 77** ****56, got %q", got)
	}
}

func TestMaskDateText(t *testing.T) {
	text := "12 января 1990"
	s := spanFor(text, text)
	got := MaskDate(text, s)
	if got != "** ****** 1990" {
		t.Fatalf("expected ** ****** 1990, got %q", got)
	}
}

func TestMaskDateISO(t *testing.T) {
	text := "1990-01-12"
	s := spanFor(text, text)
	got := MaskDate(text, s)
	if got != "1990-**-**" {
		t.Fatalf("expected 1990-**-**, got %q", got)
	}
}

func TestMaskEmailRune(t *testing.T) {
	text := "ёлка@mail.ru"
	s := spanFor(text, text)
	got := MaskEmail(text, s)
	if got != "ё***@mail.ru" {
		t.Fatalf("expected ё***@mail.ru, got %q", got)
	}
}

func TestMaskCard18(t *testing.T) {
	text := "620000000000000001"
	s := spanFor(text, text)
	got := MaskCard(text, s)
	if got != "6200**********0001" {
		t.Fatalf("expected 6200**********0001, got %q", got)
	}
}

func TestMaskPhoneCompact(t *testing.T) {
	text := "+79123456789"
	s := spanFor(text, text)
	got := MaskPhone(text, s)
	if got != "+7 9** ***-**-89" {
		t.Fatalf("expected +7 9** ***-**-89, got %q", got)
	}
}

func TestMaskPhoneDigitFallback(t *testing.T) {
	text := "79123456789"
	s := spanFor(text, text)
	got := MaskPhone(text, s)
	if got != "79*******89" {
		t.Fatalf("expected 79*******89, got %q", got)
	}
}

func TestMaskFIOLower(t *testing.T) {
	text := "иванов иван иванович"
	s := spanFor(text, text)
	got := MaskFIO(text, s)
	if got != "И. И. И." {
		t.Fatalf("expected И. И. И., got %q", got)
	}
}
