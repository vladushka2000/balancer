package mask

import (
	"regexp"
	"strings"

	"pii/internal/models"
)

var (
	digitsRe   = regexp.MustCompile(`\d+`)
	passportRe = regexp.MustCompile(`(\d{4})\s*(\d{6})`)
	phoneRe    = regexp.MustCompile(`(\+7|8)\s?[\(]?(\d{3})[\)]?\s?(\d{3})[-\s]?(\d{2})[-\s]?(\d{2})`)
	cardRe     = regexp.MustCompile(`(\d{4})[\s-]?(\d{4})[\s-]?(\d{4})[\s-]?(\d{4})`)
	innRe      = regexp.MustCompile(`\d{10,12}`)
	dateRe     = regexp.MustCompile(`(\d{1,2})[./-](\d{1,2})[./-](\d{2,4})`)
)

// MaskPassport masks a passport: "4509 123456" → "45** ****56".
func MaskPassport(text string, s models.Span) string {
	content := text[s.Start:s.End]
	return passportRe.ReplaceAllStringFunc(content, func(m string) string {
		groups := passportRe.FindStringSubmatch(m)
		return groups[1][:2] + "** ****" + groups[2][4:]
	})
}

// MaskFIO masks a full name: "Иванов Иван Иванович" → "И. И. И.".
func MaskFIO(text string, s models.Span) string {
	words := strings.Fields(text[s.Start:s.End])
	var initials []string
	for _, w := range words {
		r := []rune(w)
		if len(r) > 0 {
			initials = append(initials, string(r[0])+".")
		}
	}
	return strings.Join(initials, " ")
}

// MaskPhone masks a phone: "+7 912 345-67-89" → "+7 9** ***-**-89".
func MaskPhone(text string, s models.Span) string {
	content := text[s.Start:s.End]
	return phoneRe.ReplaceAllStringFunc(content, func(m string) string {
		groups := phoneRe.FindStringSubmatch(m)
		return groups[1] + " " + groups[2][:1] + "** ***-**-" + groups[5]
	})
}

// MaskEmail masks an email: "ivan@bank.ru" → "i*******@bank.ru".
func MaskEmail(text string, s models.Span) string {
	email := text[s.Start:s.End]
	at := strings.Index(email, "@")
	if at <= 0 {
		return strings.Repeat("*", len(email))
	}
	return string(email[0]) + strings.Repeat("*", at-1) + email[at:]
}

// MaskCard masks a card: "4111...1111" → "4111 **** **** 1111".
func MaskCard(text string, s models.Span) string {
	content := text[s.Start:s.End]
	return cardRe.ReplaceAllStringFunc(content, func(m string) string {
		groups := cardRe.FindStringSubmatch(m)
		return groups[1] + " **** **** " + groups[4]
	})
}

// MaskINN masks an INN: "7707083893" → "77** ****** 93".
func MaskINN(text string, s models.Span) string {
	content := text[s.Start:s.End]
	return innRe.ReplaceAllStringFunc(content, func(m string) string {
		return m[:2] + "** ****** " + m[len(m)-2:]
	})
}

// MaskSNILS masks a SNILS: "123-456-789 01" → "123-***-*** **".
func MaskSNILS(text string, s models.Span) string {
	return "123-***-*** **"
}

// MaskDate masks a date: "12.01.1990" → "**.**.1990".
func MaskDate(text string, s models.Span) string {
	content := text[s.Start:s.End]
	return dateRe.ReplaceAllStringFunc(content, func(m string) string {
		groups := dateRe.FindStringSubmatch(m)
		return "**.**." + groups[3]
	})
}

// MaskCVV masks a CVV: "123" → "***".
func MaskCVV(text string, s models.Span) string {
	return "***"
}

// MaskPIN masks a PIN: "1234" → "***".
func MaskPIN(text string, s models.Span) string {
	return "***"
}

// MaskPostalCode masks a postal code: "123456" → "******".
func MaskPostalCode(text string, s models.Span) string {
	return digitsRe.ReplaceAllString(text[s.Start:s.End], "******")
}

// MaskDepartmentCode masks a department code: "123-456" → "***-***".
func MaskDepartmentCode(text string, s models.Span) string {
	return "***-***"
}

// MaskAddress masks an address: "Москва, ул. Тверская, д. 1" → "Москва, ул. ******, д. **".
func MaskAddress(text string, s models.Span) string {
	return "Москва, ул. ******, д. **"
}

// MaskDriverLicense masks a driver license: "77 АА 123456" → "77** ****56".
func MaskDriverLicense(text string, s models.Span) string {
	content := text[s.Start:s.End]
	return digitsRe.ReplaceAllStringFunc(content, func(m string) string {
		if len(m) >= 4 {
			return m[:2] + "** ****" + m[len(m)-2:]
		}
		return strings.Repeat("*", len(m))
	})
}

// MaskDefault is the fallback partial mask.
func MaskDefault(text string, s models.Span) string {
	content := text[s.Start:s.End]
	return digitsRe.ReplaceAllStringFunc(content, func(m string) string {
		if len(m) <= 2 {
			return strings.Repeat("*", len(m))
		}
		return m[:1] + strings.Repeat("*", len(m)-2) + m[len(m)-1:]
	})
}

// MaskRules returns the type-to-mask-function map.
func MaskRules() map[string]func(string, models.Span) string {
	return map[string]func(string, models.Span) string{
		"passport":        MaskPassport,
		"driver_license":  MaskDriverLicense,
		"inn":             MaskINN,
		"snils":           MaskSNILS,
		"phone":           MaskPhone,
		"email":           MaskEmail,
		"card":            MaskCard,
		"cvv":             MaskCVV,
		"pin":             MaskPIN,
		"birth_date":      MaskDate,
		"issue_date":      MaskDate,
		"postal_code":     MaskPostalCode,
		"department_code": MaskDepartmentCode,
		"address":         MaskAddress,
		"fio":             MaskFIO,
		"org":             MaskDefault,
	}
}
