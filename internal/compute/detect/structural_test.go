package detect

import "testing"

func TestPassportBasic(t *testing.T) {
	spans := PassportDetector{}.Detect("паспорт 4509 123456")
	if len(spans) != 1 || spans[0].Type != "passport" {
		t.Fatalf("expected passport span, got %+v", spans)
	}
}

func TestPassportWithWords(t *testing.T) {
	spans := PassportDetector{}.Detect("серия 4509 номер 123456")
	if len(spans) != 1 {
		t.Fatalf("expected passport span, got %+v", spans)
	}
}

func TestPassportCaseInsensitive(t *testing.T) {
	spans := PassportDetector{}.Detect("ПАСПОРТ 4509 123456")
	if len(spans) != 1 {
		t.Fatalf("expected passport span, got %+v", spans)
	}
}

func TestPassportNoMatch(t *testing.T) {
	spans := PassportDetector{}.Detect("123456")
	if len(spans) != 0 {
		t.Fatalf("expected no span, got %+v", spans)
	}
}

func TestINN10Valid(t *testing.T) {
	spans := INNDetector{}.Detect("ИНН 7707083893")
	if len(spans) != 1 || spans[0].Type != "inn" {
		t.Fatalf("expected inn span, got %+v", spans)
	}
}

func TestINN10Invalid(t *testing.T) {
	spans := INNDetector{}.Detect("ИНН 7707083894")
	if len(spans) != 0 {
		t.Fatalf("expected no span, got %+v", spans)
	}
}

func TestINN12Valid(t *testing.T) {
	spans := INNDetector{}.Detect("ИНН 500100732259")
	if len(spans) != 1 {
		t.Fatalf("expected inn span, got %+v", spans)
	}
}

func TestSNILSValid(t *testing.T) {
	spans := SNILSDetector{}.Detect("СНИЛС 112-233-445 95")
	if len(spans) != 1 {
		t.Fatalf("expected snils span, got %+v", spans)
	}
}

func TestSNILSInvalid(t *testing.T) {
	spans := SNILSDetector{}.Detect("СНИЛС 112-233-445 01")
	if len(spans) != 0 {
		t.Fatalf("expected no span, got %+v", spans)
	}
}

func TestPhoneBasic(t *testing.T) {
	spans := PhoneDetector{}.Detect("+7 912 345-67-89")
	if len(spans) != 1 || spans[0].Type != "phone" {
		t.Fatalf("expected phone span, got %+v", spans)
	}
}

func TestEmailBasic(t *testing.T) {
	spans := EmailDetector{}.Detect("ivan.ivanov@bank.ru")
	if len(spans) != 1 || spans[0].Type != "email" {
		t.Fatalf("expected email span, got %+v", spans)
	}
}

func TestCardLuhnValid(t *testing.T) {
	spans := CardDetector{}.Detect("4111 1111 1111 1111")
	if len(spans) != 1 || spans[0].Type != "card" {
		t.Fatalf("expected card span, got %+v", spans)
	}
}

func TestCardLuhnInvalid(t *testing.T) {
	spans := CardDetector{}.Detect("4111 1111 1111 1112")
	if len(spans) != 0 {
		t.Fatalf("expected no span, got %+v", spans)
	}
}

func TestCVVBasic(t *testing.T) {
	spans := CVVDetector{}.Detect("CVV 123")
	if len(spans) != 1 || spans[0].Type != "cvv" {
		t.Fatalf("expected cvv span, got %+v", spans)
	}
}

func TestPINBasic(t *testing.T) {
	spans := PINDetector{}.Detect("пин-код 1234")
	if len(spans) != 1 || spans[0].Type != "pin" {
		t.Fatalf("expected pin span, got %+v", spans)
	}
}

func TestDateDDMMYYYY(t *testing.T) {
	spans := DateDetector{}.Detect("дата рождения 12.01.1990")
	if len(spans) != 1 {
		t.Fatalf("expected date span, got %+v", spans)
	}
}

func TestDateBareNotPII(t *testing.T) {
	spans := DateDetector{}.Detect("встреча 12.01.2024")
	if len(spans) != 0 {
		t.Fatalf("expected no span, got %+v", spans)
	}
}

func TestDateInvalid(t *testing.T) {
	spans := DateDetector{}.Detect("дата рождения 32.13.2020")
	if len(spans) != 0 {
		t.Fatalf("expected no span, got %+v", spans)
	}
}

func TestPostalCode(t *testing.T) {
	spans := PostalCodeDetector{}.Detect("индекс 123456")
	if len(spans) != 1 || spans[0].Type != "postal_code" {
		t.Fatalf("expected postal_code span, got %+v", spans)
	}
}

func TestDepartmentCode(t *testing.T) {
	spans := DepartmentCodeDetector{}.Detect("код подразделения 123-456")
	if len(spans) != 1 || spans[0].Type != "department_code" {
		t.Fatalf("expected department_code span, got %+v", spans)
	}
}

func TestRegistryAllDetectors(t *testing.T) {
	r := CreateStructuralRegistry()
	if len(r.detectors) != 12 {
		t.Fatalf("expected 12 detectors, got %d", len(r.detectors))
	}
}

func TestDetectAllCombined(t *testing.T) {
	r := CreateStructuralRegistry()
	spans := r.DetectAll("паспорт 4509 123456, +7 912 345-67-89, ivan@bank.ru")
	if len(spans) != 3 {
		t.Fatalf("expected 3 spans, got %+v", spans)
	}
}
