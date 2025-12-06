package parser

import (
	"strings"
	"unicode"

	fuzz "github.com/hbollon/go-edlib"
	"github.com/rs/zerolog/log"
)

const (
	// similarityThreshold is the minimum Levenshtein similarity score
	// required for fuzzy string matching (0.0-1.0). A value of 0.6
	// allows for minor speech recognition errors while avoiding false positives.
	similarityThreshold = 0.6

	// halfFieldMinLength is the minimum length for half-field matching.
	// This handles cases where two words run together (e.g., "bogeydope").
	// Only fields longer than this value will be checked by splitting in half.
	halfFieldMinLength = 8
)

// isSimilar returns true if the two strings have a similarity score greater
// than similarityThreshold using the Levenshtein distance algorithm.
func isSimilar(a, b string) bool {
	v, err := fuzz.StringsSimilarity(strings.ToLower(a), strings.ToLower(b), fuzz.Levenshtein)
	if err != nil {
		log.Error().Err(err).Str("a", a).Str("b", b).Msg("failed to calculate similarity")
		return false
	}
	return v > similarityThreshold
}

func hasDigits(tx string) bool {
	for _, r := range tx {
		if unicode.IsDigit(r) {
			return true
		}
	}
	return false
}

// normalize the given string by applying the following transformations:
//
//   - Split on any "|" character and discard the tail.
//   - Convert to lowercase.
//   - Replace hyphens and underscores with spaces.
//   - Replace a period followed by a non-space character with a space.
//   - Remove any other characters
//     that are not letters, digits, or spaces.
//   - Insert a space between any letter immediately followed by a digit.
//   - Trim leading and trailing whitespace.
//   - Substitute alternate forms of request words with canonical forms.
//   - Remove extra spaces.
func normalize(tx string) string {
	// Cut on pipe delimiter
	tx, _, _ = strings.Cut(tx, "|")
	if tx == "" {
		return ""
	}

	// Single pass: lowercase + remove symbols + space numbers
	var builder strings.Builder
	builder.Grow(len(tx))

	for i, r := range tx {
		rLower := unicode.ToLower(r)

		periodBeforeNonSpace := r == '.' && i+1 < len(tx) && !unicode.IsSpace(rune(tx[i+1]))

		if r == '-' || r == '_' || periodBeforeNonSpace {
			builder.WriteRune(' ')
		} else if unicode.IsLetter(r) || unicode.IsDigit(r) || unicode.IsSpace(r) {
			builder.WriteRune(rLower)
			// Insert space between letter and following digit
			if i+1 < len(tx) && unicode.IsLetter(r) && unicode.IsDigit(rune(tx[i+1])) {
				builder.WriteRune(' ')
			}
		}
		// Other symbols are simply omitted
	}

	tx = builder.String()

	// Convert digit words to numerals
	tx = digitWords(tx)

	// Trim and collapse multiple spaces
	tx = strings.TrimSpace(tx)

	// Apply word replacements
	for _, repl := range replacements {
		tx = strings.ReplaceAll(tx, repl.Original, repl.Normal)
	}

	return tx
}

// digitWords converts digits from their word form to their numeral form.
// It is stricter than the numwords package.
func digitWords(tx string) string {
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

// spaceDigits normalizes the given string and inserts a space before each
// digit, e.g. "Eagle11" -> "Eagle 1 1".
func spaceDigits(tx string) string {
	builder := strings.Builder{}
	for _, char := range tx {
		if unicode.IsDigit(char) {
			_, _ = builder.WriteRune(' ')
		}
		_, _ = builder.WriteRune(char)
	}
	tx = builder.String()
	return normalize(tx)
}
