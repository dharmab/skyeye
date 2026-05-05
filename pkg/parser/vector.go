package parser

import (
	"strings"

	"github.com/dharmab/skyeye/internal/parser/token"
	"github.com/dharmab/skyeye/pkg/brevity"
)

// parseVector parses a VECTOR request. locations is the full set of candidate
// location names (including the tanker alias) and must not be mutated.
func parseVector(callsign string, locations []string, stream *token.Stream) (*brevity.VectorRequest, bool) {
	var words []string
	for !stream.AtEnd() {
		word := strings.ToLower(stream.Text())
		words = append(words, word)
		stream.Advance()
	}

	// Collect all fuzzy matches across every (i, j) span, then pick the
	// longest span (break ties by earliest start) so "home plate" beats
	// "home" and iteration order cannot change the result.
	bestLocation := ""
	bestSpan := 0
	bestStart := 0
	found := false
	for i := range words {
		for j := i; j < len(words); j++ {
			sequence := strings.Join(words[i:j+1], " ")
			for _, location := range locations {
				if !isSimilar(sequence, location) {
					continue
				}
				span := j - i + 1
				if !found || span > bestSpan || (span == bestSpan && i < bestStart) {
					bestLocation = location
					bestSpan = span
					bestStart = i
					found = true
				}
			}
		}
	}

	if !found {
		return nil, false
	}
	return &brevity.VectorRequest{Callsign: callsign, Location: bestLocation}, true
}
