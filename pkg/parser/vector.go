package parser

import (
	"strings"

	"github.com/dharmab/skyeye/internal/parser/token"
	"github.com/dharmab/skyeye/pkg/brevity"
)

const (
	LocationTanker = "tanker"
)

func parseVector(callsign string, locations []string, stream *token.Stream) (*brevity.VectorRequest, bool) {
	request := &brevity.VectorRequest{Callsign: callsign}
	locations = append(locations, LocationTanker)

	var words []string
	for !stream.AtEnd() {
		word := strings.ToLower(stream.Text())
		words = append(words, word)
		stream.Advance()
	}

	for i := range words {
		for j := i; j < len(words); j++ {
			sequence := strings.Join(words[i:j+1], " ")
			for _, location := range locations {
				if isSimilar(sequence, location) {
					request.Location = location
					return request, true
				}
			}
		}
	}

	return nil, false
}
