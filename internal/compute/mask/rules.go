package mask

import (
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"

	"pii/internal/models"
)

var (
	digitRunRe    = regexp.MustCompile(`\d+`)
	passportValRe = regexp.MustCompile(`(\d{4})(\D*)(\d{6})`)
	phoneRe       = regexp.MustCompile(`(\+7|8)\s?[\(]?(\d{3})[\)]?\s?(\d{3})[-\s]?(\d{2})[-\s]?(\d{2})`)
	innValRe      = regexp.MustCompile(`\d{12}|\d{10}`)
	dateYMDRe     = regexp.MustCompile(`(\d{4})([./-])(\d{1,2})([./-])(\d{1,2})`)
	dateDMYRe     = regexp.MustCompile(`(\d{1,2})([./-])(\d{1,2})([./-])(\d{2,4})`)
	dateTextRe    = regexp.MustCompile(`(?i)(\d{1,2})(\s+)([а-яё]+)(\s+)(\d{4})`)
	shortDigitRe  = regexp.MustCompile(`\d{3,4}`)
	deptValRe     = regexp.MustCompile(`\d{3}-\d{3}`)
	dlValRe       = regexp.MustCompile(`(\d{2})(\s*)([А-ЯЁA-Zа-яёa-z]{2})(\s*)(\d{6})`)
	indexRe       = regexp.MustCompile(`\b\d{6}\b`)
	streetRe      = regexp.MustCompile(`(?i)(ул\.|улица|пр-т|пр\.|проспект|пер\.|переулок|бульвар|б-р|наб\.|шоссе)(\s+)([А-Яа-яЁёA-Za-z\-]+)`)
	houseRe       = regexp.MustCompile(`(?i)(д\.|дом)(\s*)(\d+)`)
	flatRe        = regexp.MustCompile(`(?i)(кв\.|квартира)(\s*)(\d+)`)
	cityPrefixRe  = regexp.MustCompile(`(?i)(г\.|город)(\s+)([А-Яа-яЁёA-Za-z\-]+)`)
	countryRe     = regexp.MustCompile(`(?i)россия|российская федерация`)
	snilsRe       = regexp.MustCompile(`(\d{3})-(\d{3})-(\d{3})\s?(\d{2})`)
)

func MaskPassport(text string, s models.Span) string {
	content := slice(text, s)
	if !passportValRe.MatchString(content) {
		return starDigits(content)
	}
	return passportValRe.ReplaceAllStringFunc(content, func(m string) string {
		g := passportValRe.FindStringSubmatch(m)
		return g[1][:2] + "**" + g[2] + "****" + g[3][4:]
	})
}

func MaskFIO(text string, s models.Span) string {
	words := strings.Fields(slice(text, s))
	initials := make([]string, 0, len(words))
	for _, w := range words {
		r, _ := utf8.DecodeRuneInString(w)
		if r == utf8.RuneError {
			continue
		}
		initials = append(initials, string(unicode.ToUpper(r))+".")
	}
	return strings.Join(initials, " ")
}

func MaskPhone(text string, s models.Span) string {
	content := slice(text, s)
	if phoneRe.MatchString(content) {
		return phoneRe.ReplaceAllStringFunc(content, func(m string) string {
			g := phoneRe.FindStringSubmatch(m)
			return g[1] + " " + g[2][:1] + "** ***-**-" + g[5]
		})
	}
	return maskDigitRuns(content, 2, 2)
}

func MaskEmail(text string, s models.Span) string {
	email := slice(text, s)
	at := strings.Index(email, "@")
	if at <= 0 {
		return starRunes(email)
	}
	local := []rune(email[:at])
	if len(local) == 0 {
		return starRunes(email)
	}
	return string(local[0]) + stars(len(local)-1) + email[at:]
}

func MaskCard(text string, s models.Span) string {
	content := slice(text, s)
	digits := extractDigits(content)
	if len(digits) < 13 || len(digits) > 19 {
		return maskDigitRuns(content, 2, 2)
	}
	masked := digits[:4] + stars(len(digits)-8) + digits[len(digits)-4:]
	return writeDigits(content, masked)
}

func MaskINN(text string, s models.Span) string {
	content := slice(text, s)
	if !innValRe.MatchString(content) {
		return maskDigitRuns(content, 2, 2)
	}
	return innValRe.ReplaceAllStringFunc(content, func(m string) string {
		return m[:2] + stars(len(m)-4) + m[len(m)-2:]
	})
}

func MaskSNILS(text string, s models.Span) string {
	content := slice(text, s)
	if !snilsRe.MatchString(content) {
		return starDigits(content)
	}
	return snilsRe.ReplaceAllStringFunc(content, func(m string) string {
		g := snilsRe.FindStringSubmatch(m)
		return g[1] + "-***-*** **"
	})
}

func MaskDate(text string, s models.Span) string {
	content := slice(text, s)
	if dateYMDRe.MatchString(content) {
		return dateYMDRe.ReplaceAllStringFunc(content, func(m string) string {
			g := dateYMDRe.FindStringSubmatch(m)
			return g[1] + g[2] + stars(len(g[3])) + g[4] + stars(len(g[5]))
		})
	}
	if dateTextRe.MatchString(content) {
		return dateTextRe.ReplaceAllStringFunc(content, func(m string) string {
			g := dateTextRe.FindStringSubmatch(m)
			return stars(len(g[1])) + g[2] + stars(utf8.RuneCountInString(g[3])) + g[4] + g[5]
		})
	}
	if dateDMYRe.MatchString(content) {
		return dateDMYRe.ReplaceAllStringFunc(content, func(m string) string {
			g := dateDMYRe.FindStringSubmatch(m)
			return stars(len(g[1])) + g[2] + stars(len(g[3])) + g[4] + g[5]
		})
	}
	return content
}

func MaskCVV(text string, s models.Span) string {
	return maskShortDigits(slice(text, s))
}

func MaskPIN(text string, s models.Span) string {
	return maskShortDigits(slice(text, s))
}

func MaskPostalCode(text string, s models.Span) string {
	return indexRe.ReplaceAllString(slice(text, s), "******")
}

func MaskDepartmentCode(text string, s models.Span) string {
	content := slice(text, s)
	if !deptValRe.MatchString(content) {
		return starDigits(content)
	}
	return deptValRe.ReplaceAllString(content, "***-***")
}

func MaskAddress(text string, s models.Span) string {
	content := slice(text, s)
	content = indexRe.ReplaceAllString(content, "******")
	content = countryRe.ReplaceAllStringFunc(content, func(m string) string {
		return maskWords(m)
	})
	content = streetRe.ReplaceAllStringFunc(content, func(m string) string {
		g := streetRe.FindStringSubmatch(m)
		return g[1] + g[2] + maskWordToken(g[3])
	})
	content = houseRe.ReplaceAllStringFunc(content, func(m string) string {
		g := houseRe.FindStringSubmatch(m)
		return g[1] + g[2] + stars(len(g[3]))
	})
	content = flatRe.ReplaceAllStringFunc(content, func(m string) string {
		g := flatRe.FindStringSubmatch(m)
		return g[1] + g[2] + stars(len(g[3]))
	})
	content = cityPrefixRe.ReplaceAllStringFunc(content, func(m string) string {
		g := cityPrefixRe.FindStringSubmatch(m)
		return g[1] + g[2] + maskWordToken(g[3])
	})
	return maskLeadingCity(content)
}

func MaskDriverLicense(text string, s models.Span) string {
	content := slice(text, s)
	if !dlValRe.MatchString(content) {
		return maskDigitRuns(content, 2, 2)
	}
	return dlValRe.ReplaceAllStringFunc(content, func(m string) string {
		g := dlValRe.FindStringSubmatch(m)
		num := g[5]
		return g[1] + "** ****" + num[4:]
	})
}

func MaskDefault(text string, s models.Span) string {
	return maskDigitRuns(slice(text, s), 1, 1)
}

func MaskWord(text string, s models.Span) string {
	return maskWords(slice(text, s))
}

var maskRules = map[string]func(string, models.Span) string{
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
	"org":             MaskWord,
	"birth_place":     MaskWord,
	"citizenship":     MaskWord,
	"card_holder":     MaskFIO,
}

func slice(text string, s models.Span) string {
	if s.Start < 0 || s.End > len(text) || s.Start > s.End {
		return ""
	}
	return text[s.Start:s.End]
}

func stars(n int) string {
	if n <= 0 {
		return ""
	}
	return strings.Repeat("*", n)
}

func starRunes(s string) string {
	return stars(utf8.RuneCountInString(s))
}

func starDigits(s string) string {
	return digitRunRe.ReplaceAllStringFunc(s, func(m string) string {
		return stars(len(m))
	})
}

func maskDigitRuns(s string, head, tail int) string {
	return digitRunRe.ReplaceAllStringFunc(s, func(m string) string {
		if len(m) <= head+tail {
			return stars(len(m))
		}
		return m[:head] + stars(len(m)-head-tail) + m[len(m)-tail:]
	})
}

func maskShortDigits(s string) string {
	if !shortDigitRe.MatchString(s) {
		return starDigits(s)
	}
	return shortDigitRe.ReplaceAllStringFunc(s, func(m string) string {
		return stars(len(m))
	})
}

func extractDigits(s string) string {
	var b strings.Builder
	for _, r := range s {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func writeDigits(s, digits string) string {
	var b strings.Builder
	b.Grow(len(s))
	i := 0
	for _, r := range s {
		if r >= '0' && r <= '9' && i < len(digits) {
			b.WriteByte(digits[i])
			i++
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

func maskWordToken(w string) string {
	r := []rune(w)
	if len(r) == 0 {
		return w
	}
	return string(unicode.ToUpper(r[0])) + stars(len(r)-1)
}

func keepToken(w string) bool {
	return strings.Contains(w, ".") && utf8.RuneCountInString(w) <= 4
}

func maskWords(s string) string {
	parts := strings.Fields(s)
	out := make([]string, 0, len(parts))
	for _, w := range parts {
		if keepToken(w) {
			out = append(out, w)
			continue
		}
		out = append(out, maskWordToken(w))
	}
	return strings.Join(out, " ")
}

func maskLeadingCity(s string) string {
	head, tail, found := strings.Cut(s, ",")
	trim := strings.TrimSpace(head)
	if trim == "" || strings.Contains(trim, "*") || streetRe.MatchString(trim) || houseRe.MatchString(trim) {
		return s
	}
	masked := maskWords(trim)
	lead := head[:len(head)-len(strings.TrimLeft(head, " \t"))]
	rest := head[len(lead)+len(trim):]
	if found {
		return lead + masked + rest + "," + tail
	}
	return lead + masked + rest
}
