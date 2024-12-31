package parser

import (
	"strings"
	"unicode"
)

// ParsePilotCallsign attempts to parse a callsign in one of the following formats:
//   - A single word, followed by a number consisting of any digits
//   - A number consisting of up to 3 digits
//
// Garbage in between the digits is ignored. The result is normalized so that each digit is lowercase and space-delimited.
func ParsePilotCallsign(tx string) (callsign string, isValid bool) {
	tx = normalize(tx)
	tx = spaceDigits(tx)
	for token, replacement := range map[string]string{
		"request": "",
		"this is": "",
		"on one":  "11",
		"want to": "12",
		"tutu":    "22",
		"to 8":    "28",
		"free 1":  "31",
	} {
		tx = strings.ReplaceAll(tx, token, replacement)
	}

	var builder strings.Builder
	numDigits := 0
	for _, char := range tx {
		if numDigits >= 3 {
			break
		}
		if unicode.IsDigit(char) {
			numDigits++
		}
		if numDigits == 0 || unicode.IsDigit(char) || unicode.IsSpace(char) {
			_, _ = builder.WriteRune(char)
		}
	}

	callsign = spaceDigits(normalize(builder.String()))
	if callsign == "" {
		return "", false
	}

	return callsign, true
}
