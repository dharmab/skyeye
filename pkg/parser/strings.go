package parser

import (
	"strings"

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
