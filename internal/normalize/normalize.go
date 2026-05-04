// Package normalize provides text normalization for callsign and brevity strings.
package normalize

import (
	"strings"
	"unicode"
)

var digitWordMap = map[string]string{
	"zero": "0", "one": "1", "two": "2", "three": "3", "four": "4",
	"five": "5", "six": "6", "seven": "7", "eight": "8", "nine": "9",
}

// Normalize applies text normalization: split on "|" and discard the tail,
// lowercase, replace hyphens/underscores/mid-word periods with spaces, strip
// other non-alphanumeric characters, insert a space between a letter and a
// following digit, convert digit words to numerals, and collapse whitespace.
func Normalize(s string) string {
	s, _, _ = strings.Cut(s, "|")
	if s == "" {
		return ""
	}

	var b strings.Builder
	b.Grow(len(s))
	for i, r := range s {
		isSeparator := r == '-' || r == '_'
		isPeriodBeforeNonSpace := r == '.' && i+1 < len(s) && !unicode.IsSpace(rune(s[i+1]))
		if isSeparator || isPeriodBeforeNonSpace {
			b.WriteRune(' ')
			continue
		}

		if unicode.IsLetter(r) || unicode.IsDigit(r) || unicode.IsSpace(r) {
			b.WriteRune(unicode.ToLower(r))
			isLetterBeforeDigit := unicode.IsLetter(r) && i+1 < len(s) && unicode.IsDigit(rune(s[i+1]))
			if isLetterBeforeDigit {
				b.WriteRune(' ')
			}
		}
	}

	return strings.TrimSpace(DigitWords(b.String()))
}

// DigitWords converts digit words (zero through nine) to their numeral form.
func DigitWords(s string) string {
	fields := strings.Fields(s)
	for i, f := range fields {
		if n, ok := digitWordMap[f]; ok {
			fields[i] = n
		}
	}
	return strings.Join(fields, " ")
}

// SpaceDigits inserts a space before each digit, then normalizes the result.
func SpaceDigits(s string) string {
	var b strings.Builder
	for _, r := range s {
		if unicode.IsDigit(r) {
			b.WriteRune(' ')
		}
		b.WriteRune(r)
	}
	return Normalize(b.String())
}

// HasDigits returns true if the string contains any digit character.
func HasDigits(s string) bool {
	return strings.ContainsFunc(s, unicode.IsDigit)
}
