package detect

import (
	"regexp"
	"strconv"
	"strings"

	"pii/internal/models"
)

var (
	passportRe       = regexp.MustCompile(`(?i)(?:серия\s*(\d{4})\s*(?:номер\s*)?(\d{6}))|(?:паспорт.{0,20}?(\d{4})\s*(\d{6}))`)
	passportBareRe   = regexp.MustCompile(`\b(\d{4})\s(\d{6})\b`)
	driverLicenseRe  = regexp.MustCompile(`(?i)(?:в/у|водительск).{0,20}?(\d{2})\s?([А-ЯA-Z]{2})\s?(\d{6})`)
	innRe            = regexp.MustCompile(`(?i)\b(\d{10}|\d{12})\b`)
	snilsRe          = regexp.MustCompile(`\b(\d{3})-(\d{3})-(\d{3})\s?(\d{2})\b`)
	phoneRe          = regexp.MustCompile(`(?:\+7|8)\s?[\(]?\d{3}[\)]?\s?\d{3}[-\s]?\d{2}[-\s]?\d{2}`)
	emailRe          = regexp.MustCompile(`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`)
	cardRe           = regexp.MustCompile(`\b\d{4}[\s-]?\d{4}[\s-]?\d{4}[\s-]?\d{4}\b`)
	cvvRe            = regexp.MustCompile(`(?i)(?:cvv|cvc|код безопасности).{0,10}?(\d{3}|\d{4})`)
	pinRe            = regexp.MustCompile(`(?i)(?:пин|pin).{0,15}?(\d{4})`)
	dateRe           = regexp.MustCompile(`\b(\d{1,2})[./-](\d{1,2})[./-](\d{2}|\d{4})\b`)
	dateTextRe       = regexp.MustCompile(`(?i)\b(\d{1,2})\s+([а-яё]+)\s+(\d{4})\b`)
	postalCodeRe     = regexp.MustCompile(`(?i)(?:индекс|почтовый код).{0,10}?(\d{6})`)
	departmentCodeRe = regexp.MustCompile(`\b(\d{3})-(\d{3})\b`)
)

var (
	passportMarkerRe = regexp.MustCompile(`(?i)паспорт|серия|номер`)
	innMarkerRe      = regexp.MustCompile(`(?i)инн`)
	dateMarkerRe     = regexp.MustCompile(`(?i)родил|дата рождения|выдан|дата выдачи`)
	issueMarkerRe    = regexp.MustCompile(`(?i)выдан|дата выдачи`)
	postalMarkerRe   = regexp.MustCompile(`(?i)индекс|почтовый код|адрес`)
	deptMarkerRe     = regexp.MustCompile(`(?i)код подразделения|выдан`)
	snilsMarkerRe    = regexp.MustCompile(`(?i)снилс`)
	driverMarkerRe   = regexp.MustCompile(`(?i)в/у|водительск`)
	fioNearRe        = regexp.MustCompile(`(?i)[А-ЯЁ][а-яё]+\s+[А-ЯЁ][а-яё]+\s+[А-ЯЁ][а-яё]+`)
)

var monthNames = loadMonths()

// PassportDetector finds Russian passport series and number.
type PassportDetector struct{}

func (PassportDetector) Detect(text string) []models.Span {
	var spans []models.Span
	for _, m := range passportRe.FindAllStringSubmatchIndex(text, -1) {
		start, end := m[0], m[1]
		if !passportMarkerRe.MatchString(text[start:end]) && !passportMarkerRe.MatchString(text[max(0, start-60):start]) {
			continue
		}
		spans = append(spans, models.Span{Start: start, End: end, Type: "passport", Confidence: 1, Source: "regex"})
	}
	for _, m := range passportBareRe.FindAllStringSubmatchIndex(text, -1) {
		start, end := m[0], m[1]
		if !passportMarkerRe.MatchString(text[max(0, start-60):start]) {
			continue
		}
		if overlapsAny(spans, start, end) {
			continue
		}
		spans = append(spans, models.Span{Start: start, End: end, Type: "passport", Confidence: 1, Source: "regex"})
	}
	return spans
}

func overlapsAny(spans []models.Span, start, end int) bool {
	for _, s := range spans {
		if start < s.End && end > s.Start {
			return true
		}
	}
	return false
}

// DriverLicenseDetector finds driver license numbers.
type DriverLicenseDetector struct{}

func (DriverLicenseDetector) Detect(text string) []models.Span {
	var spans []models.Span
	for _, m := range driverLicenseRe.FindAllStringSubmatchIndex(text, -1) {
		start, end := m[0], m[1]
		region := text[m[2]:m[3]]
		if !validRegion(region) {
			continue
		}
		spans = append(spans, models.Span{Start: start, End: end, Type: "driver_license", Confidence: 1, Source: "regex"})
	}
	return spans
}

// INNDetector finds taxpayer identification numbers.
type INNDetector struct{}

func (INNDetector) Detect(text string) []models.Span {
	var spans []models.Span
	for _, m := range innRe.FindAllStringSubmatchIndex(text, -1) {
		start, end := m[0], m[1]
		digits := text[m[2]:m[3]]
		if !validINN(digits) {
			continue
		}
		before := text[max(0, start-60):start]
		if !innMarkerRe.MatchString(before) && !fioNearRe.MatchString(before) {
			continue
		}
		spans = append(spans, models.Span{Start: start, End: end, Type: "inn", Confidence: 1, Source: "regex"})
	}
	return spans
}

// SNILSDetector finds SNILS numbers.
type SNILSDetector struct{}

func (SNILSDetector) Detect(text string) []models.Span {
	var spans []models.Span
	for _, m := range snilsRe.FindAllStringSubmatchIndex(text, -1) {
		start, end := m[0], m[1]
		digits := text[m[2]:m[3]] + text[m[4]:m[5]] + text[m[6]:m[7]] + text[m[8]:m[9]]
		if !validSNILS(digits) {
			continue
		}
		if !snilsMarkerRe.MatchString(text[max(0, start-40):end]) {
			continue
		}
		spans = append(spans, models.Span{Start: start, End: end, Type: "snils", Confidence: 1, Source: "regex"})
	}
	return spans
}

// PhoneDetector finds phone numbers.
type PhoneDetector struct{}

func (PhoneDetector) Detect(text string) []models.Span {
	var spans []models.Span
	for _, m := range phoneRe.FindAllStringIndex(text, -1) {
		spans = append(spans, models.Span{Start: m[0], End: m[1], Type: "phone", Confidence: 1, Source: "regex"})
	}
	return spans
}

// EmailDetector finds email addresses.
type EmailDetector struct{}

func (EmailDetector) Detect(text string) []models.Span {
	var spans []models.Span
	for _, m := range emailRe.FindAllStringIndex(text, -1) {
		spans = append(spans, models.Span{Start: m[0], End: m[1], Type: "email", Confidence: 1, Source: "regex"})
	}
	return spans
}

// CardDetector finds payment card numbers with Luhn validation.
type CardDetector struct{}

func (CardDetector) Detect(text string) []models.Span {
	var spans []models.Span
	for _, m := range cardRe.FindAllStringIndex(text, -1) {
		digits := extractDigits(text[m[0]:m[1]])
		if !validLuhn(digits) {
			continue
		}
		spans = append(spans, models.Span{Start: m[0], End: m[1], Type: "card", Confidence: 1, Source: "regex"})
	}
	return spans
}

// CVVDetector finds CVV codes.
type CVVDetector struct{}

func (CVVDetector) Detect(text string) []models.Span {
	var spans []models.Span
	for _, m := range cvvRe.FindAllStringSubmatchIndex(text, -1) {
		spans = append(spans, models.Span{Start: m[0], End: m[1], Type: "cvv", Confidence: 1, Source: "regex"})
	}
	return spans
}

// PINDetector finds PIN codes.
type PINDetector struct{}

func (PINDetector) Detect(text string) []models.Span {
	var spans []models.Span
	for _, m := range pinRe.FindAllStringSubmatchIndex(text, -1) {
		spans = append(spans, models.Span{Start: m[0], End: m[1], Type: "pin", Confidence: 1, Source: "regex"})
	}
	return spans
}

// DateDetector finds birth/issue dates only near markers.
type DateDetector struct{}

func (DateDetector) Detect(text string) []models.Span {
	var spans []models.Span
	for _, m := range dateRe.FindAllStringSubmatchIndex(text, -1) {
		start, end := m[0], m[1]
		if !dateMarkerRe.MatchString(text[max(0, start-60):start]) {
			continue
		}
		day, _ := strconv.Atoi(text[m[2]:m[3]])
		month, _ := strconv.Atoi(text[m[4]:m[5]])
		year, _ := strconv.Atoi(text[m[6]:m[7]])
		if !validDate(day, month, year) {
			continue
		}
		spans = append(spans, models.Span{Start: start, End: end, Type: dateType(text, start), Confidence: 1, Source: "regex"})
	}
	for _, m := range dateTextRe.FindAllStringSubmatchIndex(text, -1) {
		start, end := m[0], m[1]
		if !dateMarkerRe.MatchString(text[max(0, start-60):start]) {
			continue
		}
		day, _ := strconv.Atoi(text[m[2]:m[3]])
		month, ok := monthNames[strings.ToLower(text[m[4]:m[5]])]
		if !ok {
			continue
		}
		year, _ := strconv.Atoi(text[m[6]:m[7]])
		if !validDate(day, month, year) {
			continue
		}
		spans = append(spans, models.Span{Start: start, End: end, Type: dateType(text, start), Confidence: 1, Source: "regex"})
	}
	return spans
}

func dateType(text string, start int) string {
	if issueMarkerRe.MatchString(text[max(0, start-60):start]) {
		return "issue_date"
	}
	return "birth_date"
}

// PostalCodeDetector finds postal codes.
type PostalCodeDetector struct{}

func (PostalCodeDetector) Detect(text string) []models.Span {
	var spans []models.Span
	for _, m := range postalCodeRe.FindAllStringSubmatchIndex(text, -1) {
		start, end := m[0], m[1]
		code, _ := strconv.Atoi(text[m[2]:m[3]])
		if code < 100000 || code > 699999 {
			continue
		}
		spans = append(spans, models.Span{Start: start, End: end, Type: "postal_code", Confidence: 1, Source: "regex"})
	}
	return spans
}

// DepartmentCodeDetector finds passport department codes.
type DepartmentCodeDetector struct{}

func (DepartmentCodeDetector) Detect(text string) []models.Span {
	var spans []models.Span
	for _, m := range departmentCodeRe.FindAllStringSubmatchIndex(text, -1) {
		start, end := m[0], m[1]
		if !deptMarkerRe.MatchString(text[max(0, start-60):start]) {
			continue
		}
		spans = append(spans, models.Span{Start: start, End: end, Type: "department_code", Confidence: 1, Source: "regex"})
	}
	return spans
}

// CreateStructuralRegistry builds a registry with all structural detectors.
func CreateStructuralRegistry() *Registry {
	r := NewRegistry()
	r.Register(PassportDetector{})
	r.Register(DriverLicenseDetector{})
	r.Register(INNDetector{})
	r.Register(SNILSDetector{})
	r.Register(PhoneDetector{})
	r.Register(EmailDetector{})
	r.Register(CardDetector{})
	r.Register(CVVDetector{})
	r.Register(PINDetector{})
	r.Register(DateDetector{})
	r.Register(PostalCodeDetector{})
	r.Register(DepartmentCodeDetector{})
	return r
}

func validRegion(s string) bool {
	n, err := strconv.Atoi(s)
	return err == nil && n >= 1 && n <= 99
}

func extractDigits(s string) string {
	buf := make([]byte, 0, len(s))
	for i := 0; i < len(s); i++ {
		if s[i] >= '0' && s[i] <= '9' {
			buf = append(buf, s[i])
		}
	}
	return string(buf)
}

func validLuhn(digits string) bool {
	if len(digits) < 16 || len(digits) > 19 {
		return false
	}
	sum := 0
	double := false
	for i := len(digits) - 1; i >= 0; i-- {
		d := int(digits[i] - '0')
		if double {
			d *= 2
			if d > 9 {
				d -= 9
			}
		}
		sum += d
		double = !double
	}
	return sum%10 == 0
}

func validINN(digits string) bool {
	if len(digits) == 10 {
		weights := []int{2, 4, 10, 3, 5, 9, 4, 6, 8}
		return innCheck(digits, weights) == int(digits[9]-'0')
	}
	if len(digits) == 12 {
		w1 := []int{7, 2, 4, 10, 3, 5, 9, 4, 6, 8}
		w2 := []int{3, 7, 2, 4, 10, 3, 5, 9, 4, 6, 8}
		return innCheck(digits, w1) == int(digits[10]-'0') && innCheck(digits, w2) == int(digits[11]-'0')
	}
	return false
}

func innCheck(digits string, weights []int) int {
	sum := 0
	for i, w := range weights {
		sum += int(digits[i]-'0') * w
	}
	return sum % 11 % 10
}

func validSNILS(digits string) bool {
	if len(digits) != 11 {
		return false
	}
	sum := 0
	for i := 0; i < 9; i++ {
		sum += int(digits[i]-'0') * (9 - i)
	}
	control := sum % 101
	if control == 100 {
		control = 0
	}
	return control == int(digits[9]-'0')*10+int(digits[10]-'0')
}

func validDate(day, month, year int) bool {
	if month < 1 || month > 12 || year < 1900 || year > 2100 {
		return false
	}
	daysInMonth := []int{31, 28, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31}
	if isLeap(year) && month == 2 {
		daysInMonth[1] = 29
	}
	return day >= 1 && day <= daysInMonth[month-1]
}

func isLeap(year int) bool {
	return year%4 == 0 && (year%100 != 0 || year%400 == 0)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func loadMonths() map[string]int {
	months := map[string]int{}
	data, err := dictFS.ReadFile("dicts/months.txt")
	if err != nil {
		return months
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		months[strings.ToLower(line)] = len(months) + 1
	}
	return months
}
