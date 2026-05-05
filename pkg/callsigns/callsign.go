// Package callsigns provides parsing and normalization of pilot callsigns.
package callsigns

import (
	"regexp"
	"strings"
	"unicode"

	"github.com/dharmab/skyeye/internal/normalize"
)

const (
	// maxCallsignDigits is the maximum number of digits allowed in a pilot callsign.
	maxCallsignDigits = 3
)

// digitHomophones maps common speech recognition misheard words to digits.
var digitHomophones = map[string]string{
	"won":   "1",
	"to":    "2",
	"too":   "2",
	"tu":    "2",
	"tutu":  "22",
	"free":  "3",
	"tree":  "3",
	"for":   "4",
	"fore":  "4",
	"ate":   "8",
	"niner": "9",
}

// replaceDigitHomophones replaces words that are homophones of digits,
// but only when they appear in digit positions of a callsign (i.e., after
// the callsign name or mixed with actual digits).
func replaceDigitHomophones(tx string) string {
	fields := strings.Fields(tx)
	firstDigitIdx := -1
	for i, f := range fields {
		if normalize.HasDigits(f) || digitHomophones[f] != "" {
			firstDigitIdx = i
			break
		}
	}
	if firstDigitIdx < 0 {
		return tx
	}
	for i := firstDigitIdx; i < len(fields); i++ {
		if d, ok := digitHomophones[fields[i]]; ok {
			fields[i] = d
		}
		fields[i] = stripOrdinalSuffix(fields[i])
	}
	return strings.Join(fields, " ")
}

// stripOrdinalSuffix removes ordinal suffixes (st, nd, rd, th) from a
// string that starts with digits, e.g. "5th" → "5".
func stripOrdinalSuffix(s string) string {
	for _, suffix := range []string{"st", "nd", "rd", "th"} {
		if strings.HasSuffix(s, suffix) {
			prefix := s[:len(s)-len(suffix)]
			if prefix != "" && normalize.HasDigits(prefix) {
				return prefix
			}
		}
	}
	return s
}

// ParsePilotCallsign attempts to parse a callsign in one of the following formats:
//   - A single word, followed by a number consisting of any digits
//   - A number consisting of up to 3 digits
//
// Garbage in between the digits is ignored. Clan tags in the format "[CLAN]"
// are also ignored. The result is normalized so that each digit is lowercase
// and space-delimited.
func ParsePilotCallsign(tx string) (callsign string, isValid bool) {
	tx = removeClanTags(tx)
	tx = normalize.Normalize(tx)
	tx = normalize.SpaceDigits(tx)

	tx = strings.ReplaceAll(tx, "this is", "")

	if idx := strings.Index(tx, "request"); idx >= 0 {
		tx = tx[:idx]
	}

	tx = replaceDigitHomophones(tx)

	var builder strings.Builder
	n := 0
	for _, char := range tx {
		if n >= maxCallsignDigits {
			break
		}
		if unicode.IsDigit(char) {
			n++
		}
		if n == 0 || unicode.IsDigit(char) || unicode.IsSpace(char) {
			_, _ = builder.WriteRune(char)
		}
	}

	callsign = normalize.SpaceDigits(normalize.Normalize(builder.String()))
	if callsign == "" {
		return "", false
	}

	return callsign, true
}

var clanTagRe = regexp.MustCompile(`\[.*?\]`)

func removeClanTags(tx string) string {
	return clanTagRe.ReplaceAllString(tx, "")
}
