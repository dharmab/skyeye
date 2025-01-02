package parser

import (
	"strings"
	"unicode"

	fuzz "github.com/hbollon/go-edlib"
	"github.com/rs/zerolog/log"
)

// isSimilar returns true if the two strings have a similarity score greater
// than 0.6 using the Levenshtein distance algorithm.
func isSimilar(a, b string) bool {
	v, err := fuzz.StringsSimilarity(strings.ToLower(a), strings.ToLower(b), fuzz.Levenshtein)
	if err != nil {
		log.Error().Err(err).Str("a", a).Str("b", b).Msg("failed to calculate similarity")
		return false
	}
	return v > 0.6
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
	tx, _, _ = strings.Cut(tx, "|")
	tx = strings.ToLower(tx)
	tx = removeSymbols(tx)
	tx = spaceNumbers(tx)
	tx = strings.TrimSpace(tx)
	for _, repl := range replacements {
		tx = strings.ReplaceAll(tx, repl.Original, repl.Normal)
	}
	tx = strings.Join(strings.Fields(tx), " ")
	return tx
}

// removeSymbols removes any characters that are not letters, digits, or
// spaces. Hyphens and underscores are replaced with spaces. A period followed
// by a non-space character is replaced with a space. Other symbols are
// removed.
func removeSymbols(tx string) string {
	var builder strings.Builder
	for i, r := range tx {
		isPeriodBeforeNonSpace := r == '.' && i+1 < len(tx) && !unicode.IsSpace(rune(tx[i+1]))
		if r == '-' || r == '_' || isPeriodBeforeNonSpace {
			_, _ = builder.WriteRune(' ')
		} else if unicode.IsLetter(r) || unicode.IsDigit(r) || unicode.IsSpace(r) {
			_, _ = builder.WriteRune(r)
		}
	}
	return builder.String()
}

// spaceNumbers inserts spaces between letters and numbers, e.g. "BRAA090" -> "BRAA 090".
func spaceNumbers(tx string) string {
	builder := strings.Builder{}
	for i, char := range tx {
		_, _ = builder.WriteRune(char)
		if i+1 < len(tx) && unicode.IsLetter(char) && unicode.IsDigit(rune(tx[i+1])) {
			_, _ = builder.WriteRune(' ')
		}
	}
	return builder.String()
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
