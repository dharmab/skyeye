// Package normalize provides text normalization for callsign and brevity strings.
package normalize

import (
	"strings"
	"unicode"
)

// String normalizes the given string by applying the following transformations:
//
//   - Split on any "|" character and discard the tail.
//   - Convert to lowercase.
//   - Replace hyphens and underscores with spaces.
//   - Replace a period followed by a non-space character with a space.
//   - Remove any other characters
//     that are not letters, digits, or spaces.
//   - Insert a space between any letter immediately followed by a digit.
//   - Trim leading and trailing whitespace.
//   - Convert digit words to numerals.
//   - Remove extra spaces.
func String(tx string) string {
	tx, _, _ = strings.Cut(tx, "|")
	if tx == "" {
		return ""
	}

	var builder strings.Builder
	builder.Grow(len(tx))

	for i, r := range tx {
		rLower := unicode.ToLower(r)

		periodBeforeNonSpace := r == '.' && i+1 < len(tx) && !unicode.IsSpace(rune(tx[i+1]))

		if r == '-' || r == '_' || periodBeforeNonSpace {
			builder.WriteRune(' ')
		} else if unicode.IsLetter(r) || unicode.IsDigit(r) || unicode.IsSpace(r) {
			builder.WriteRune(rLower)
			if i+1 < len(tx) && unicode.IsLetter(r) && unicode.IsDigit(rune(tx[i+1])) {
				builder.WriteRune(' ')
			}
		}
	}

	tx = builder.String()

	tx = DigitWords(tx)

	tx = strings.TrimSpace(tx)

	return tx
}

// DigitWords converts digits from their word form to their numeral form.
func DigitWords(tx string) string {
	fields := strings.Fields(tx)
	for i, field := range fields {
		var n string
		switch field {
		case "zero":
			n = "0"
		case "one":
			n = "1"
		case "two":
			n = "2"
		case "three":
			n = "3"
		case "four":
			n = "4"
		case "five":
			n = "5"
		case "six":
			n = "6"
		case "seven":
			n = "7"
		case "eight":
			n = "8"
		case "nine":
			n = "9"
		}
		if n != "" {
			fields[i] = n
		}
	}
	tx = strings.Join(fields, " ")
	return tx
}

// SpaceDigits inserts a space before each digit, then normalizes the result.
func SpaceDigits(tx string) string {
	builder := strings.Builder{}
	for _, char := range tx {
		if unicode.IsDigit(char) {
			_, _ = builder.WriteRune(' ')
		}
		_, _ = builder.WriteRune(char)
	}
	tx = builder.String()
	return String(tx)
}

// HasDigits returns true if the string contains any digit character.
func HasDigits(tx string) bool {
	for _, r := range tx {
		if unicode.IsDigit(r) {
			return true
		}
	}
	return false
}
