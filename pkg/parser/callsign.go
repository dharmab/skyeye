package parser

import (
	"regexp"
	"strings"
	"unicode"
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
	// Find the first field that is or looks like a digit.
	// Everything before that is the callsign name.
	firstDigitIdx := -1
	for i, f := range fields {
		if hasDigits(f) || digitHomophones[f] != "" {
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
		// Strip ordinal suffixes: "1st" → "1", "2nd" → "2", etc.
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
			if prefix != "" && hasDigits(prefix) {
				return prefix
			}
		}
	}
	return s
}

// isDigitLike reports whether s contains digits or is a digit homophone.
func isDigitLike(s string) bool {
	return hasDigits(s) || digitHomophones[s] != ""
}

// deduplicateConsecutiveWords removes consecutive duplicate words,
// e.g. "eagle eagle 2 7" → "eagle 2 7". This handles STT stutter
// where words are repeated.
func deduplicateConsecutiveWords(tx string) string {
	fields := strings.Fields(tx)
	if len(fields) <= 1 {
		return tx
	}
	result := []string{fields[0]}
	for i := 1; i < len(fields); i++ {
		// Only deduplicate words that are not digits or digit homophones.
		// This collapses "eagle eagle" but preserves "won won" (→ "1 1")
		// and "1 1".
		if fields[i] == fields[i-1] && !isDigitLike(fields[i]) {
			continue
		}
		result = append(result, fields[i])
	}
	return strings.Join(result, " ")
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
	tx = normalize(tx)
	tx = spaceDigits(tx)

	// Discard "this is" prefix.
	tx = strings.ReplaceAll(tx, "this is", "")

	// Truncate at "request" — not proper brevity, but some players say it.
	// Anything after it is part of the request, not the callsign.
	if idx := strings.Index(tx, "request"); idx >= 0 {
		tx = tx[:idx]
	}

	tx = deduplicateConsecutiveWords(tx)
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

	callsign = spaceDigits(normalize(builder.String()))
	if callsign == "" {
		return "", false
	}

	return callsign, true
}

var clanTagRe = regexp.MustCompile(`\[.*?\]`)

func removeClanTags(tx string) string {
	return clanTagRe.ReplaceAllString(tx, "")
}
